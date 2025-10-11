package main

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/mux"
)

// Server 承载 API，并可以触发后端操作（如 backfill）
type Server struct {
	store    *Store
	httpsCli *ethclient.Client
	oappAddr common.Address
	topic    common.Hash
	proc     *Processor

	// backfill control channel，用于在同一进程内触发回填（可扩展）
	backfillCh chan backfillRequest
}

type backfillRequest struct {
	FromBlock uint64
	ToBlock   uint64
	// 我们也可以加入回调或 request id
}

// NewServer 构造 Server
func NewServer(s *Store, httpsCli *ethclient.Client, oappAddr common.Address, topic common.Hash, proc *Processor) *Server {
	srv := &Server{
		store:      s,
		httpsCli:   httpsCli,
		oappAddr:   oappAddr,
		topic:      topic,
		proc:       proc,
		backfillCh: make(chan backfillRequest, 4),
	}
	// 后台 goroutine 负责实际执行 backfill，以避免在 HTTP handler 中阻塞
	go srv.backfillWorker()
	return srv
}

func (s *Server) routes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/health", s.handleHealth).Methods("GET")
	r.HandleFunc("/payouts", s.handleListPayouts).Methods("GET")
	r.HandleFunc("/admin/backfill", s.handleBackfill).Methods("POST")
	// 可选：增加手动 reconcile、status endpoints
	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"ok":        true,
		"db":        s.store != nil,
		"wssStatus": getWssStatus(),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleListPayouts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset, _ := strconv.Atoi(q.Get("offset"))
	list, err := s.store.ListPayouts(limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(list)
}

// handleBackfill: 支持可选 JSON body { "from_block": <n>, "to_block": <m> }
// 若不提供，将使用默认后端逻辑 (e.g., last 200 blocks)
func (s *Server) handleBackfill(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		FromBlock uint64 `json:"from_block"`
		ToBlock   uint64 `json:"to_block"`
	}
	var req Req
	_ = json.NewDecoder(r.Body).Decode(&req)

	// 将请求异步放入队列，由后台 worker 处理
	s.backfillCh <- backfillRequest{FromBlock: req.FromBlock, ToBlock: req.ToBlock}

	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"status":"backfill queued"}`))
	log.Printf("API: backfill queued from=%d to=%d", req.FromBlock, req.ToBlock)
}

// backfillWorker: 处理队列中的 backfill 请求
func (s *Server) backfillWorker() {
	for req := range s.backfillCh {
		// 如果用户没有指定区间（0,0），使用默认 backfill 最近 N blocks（由 backfillHistoricalEvents 处理）
		log.Printf("backfillWorker: received request from=%d to=%d", req.FromBlock, req.ToBlock)
		if req.FromBlock == 0 && req.ToBlock == 0 {
			// 没有指定：使用默认策略（最近 200 区块）
			_ = s.doBackfillDefault()
		} else {
			// 指定区间：执行定制查询
			_ = s.doBackfillRange(req.FromBlock, req.ToBlock)
		}
	}
}

func (s *Server) doBackfillDefault() error {
	// 调用之前实现的 backfillHistoricalEvents（它会查询最近 200 blocks）
	// 我们在这里传入 s.httpsCli, s.oappAddr, s.topic, s.proc
	go func() {
		// 另起协程避免阻塞 worker loop（但保证并发队列不被阻塞）
		_ = backfillHistoricalEvents(s.httpsCli, s.oappAddr, s.topic, s.proc)
	}()
	return nil
}

func (s *Server) doBackfillRange(from, to uint64) error {
	// 我们需要在这里使用 FilterLogs 显式按区间查询并调用 proc.ParseAndPersist
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(from),
		ToBlock:   new(big.Int).SetUint64(to),
		Addresses: []common.Address{s.oappAddr},
		Topics:    [][]common.Hash{{s.topic}},
	}
	logs, err := s.httpsCli.FilterLogs(ctx, query)
	if err != nil {
		log.Printf("doBackfillRange: FilterLogs error: %v", err)
		return err
	}
	log.Printf("doBackfillRange: found %d logs in [%d - %d]", len(logs), from, to)
	for i := len(logs) - 1; i >= 0; i-- {
		if err := s.proc.ParseAndPersist(context.Background(), logs[i]); err != nil {
			log.Printf("doBackfillRange: parse error tx=%s idx=%d: %v", logs[i].TxHash.Hex(), logs[i].Index, err)
		}
	}
	return nil
}

// helper to read wss status safely
func getWssStatus() string {
	mu.Lock()
	defer mu.Unlock()
	return wssStatus
}
