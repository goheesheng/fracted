package main

import (
	"context"
	"encoding/json" // 新增
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv" // 新增
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient" // 新增
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type contextKey string

const (
	ctxKeyMerchant contextKey = "merchant"
	ctxKeyRole     contextKey = "role"
)

// 全局配置
var adminConfig *AdminConfig
var merchantConfig *MerchantConfig

// 初始化配置
func init() {
	adminConfig = LoadAdminConfig()
	merchantConfig = LoadMerchantConfig()
}

// 从 Authorization: Bearer <token> 中提取 token
func parseBearerToken(r *http.Request) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("invalid Authorization header format")
	}
	return parts[1], nil
}

// 校验并解析 JWT，返回 merchant 地址与 role
func generateJWT(merchant, role string) (string, error) {
	claims := jwt.MapClaims{
		"merchant": merchant,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // 24小时过期
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func verifyAndExtractClaims(tokenStr string, secret []byte) (string, string, error) {
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil || !tok.Valid {
		return "", "", fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}
	merchant, _ := claims["merchant"].(string)
	role, _ := claims["role"].(string)
	if merchant == "" {
		return "", "", fmt.Errorf("missing merchant claim")
	}
	return merchant, role, nil
}

// 强制登录（商家或管理员都可通过）
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := parseBearerToken(r)
		if err != nil {
			log.Printf("authMiddleware: failed to parse token: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// 使用 main.go 中定义的 jwtSecret
		merchantStr, role, err := verifyAndExtractClaims(tokenStr, jwtSecret)
		if err != nil {
			log.Printf("authMiddleware: failed to verify token: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// 验证地址格式（支持 EVM 和 Solana）
		if !isValidAddress(merchantStr) {
			log.Printf("authMiddleware: invalid address format: %s", merchantStr)
			http.Error(w, "Invalid address format", http.StatusUnauthorized)
			return
		}
		// 写入上下文（对于 Solana 地址，需要特殊处理）
		var merchantAddr common.Address
		if isValidEVMAddress(merchantStr) {
			merchantAddr = common.HexToAddress(merchantStr)
		} else {
			// Solana 地址：转换为 EVM 格式用于查询
			merchantAddr = solanaAddressToEVMAddressShared(merchantStr)
		}
		ctx := context.WithValue(r.Context(), ctxKeyMerchant, merchantAddr)
		ctx = context.WithValue(ctx, ctxKeyRole, role)
		// 保存原始地址字符串，用于查询
		ctx = context.WithValue(ctx, "merchant_original", merchantStr)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// 仅管理员可访问
func adminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(ctxKeyRole).(string)
		if role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// 仅供 API 使用的最小 Store 接口，便于在测试中注入 Mock
type PayoutStore interface {
	ListPayouts(limit, offset int) ([]PayoutRecord, error)
	ListMerchantPayouts(merchant common.Address, limit, offset int) ([]PayoutRecord, error)
	ListMerchantPayoutsByString(merchantAddr string, limit, offset int) ([]PayoutRecord, error)
	GetAllEvents(limit, offset int) ([]RawEvent, error)
	GetEventCount() (int, error)
}

// Server 承载 API，并可以触发后端操作（如 backfill）
type Server struct {
	store    PayoutStore
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

// NewServer 构造 Server（接受满足 PayoutStore 接口的实现，生产中传 *Store 即可）
func NewServer(s PayoutStore, httpsCli *ethclient.Client, oappAddr common.Address, topic common.Hash, proc *Processor) *Server {
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

	// 根路径重定向到 dashboard
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/dashboard/", http.StatusMovedPermanently)
			return
		}
	}).Methods("GET")

	// Dashboard API 路由（需要在静态文件之前）
	r.HandleFunc("/dashboard/api/payouts", s.handleDashboardPayoutsWithAuth).Methods("GET")
	r.HandleFunc("/dashboard/api/merchant/{address}/payouts", s.handleDashboardMerchantPayouts).Methods("GET")

	// 静态文件服务（Dashboard）
	r.PathPrefix("/dashboard/").Handler(http.StripPrefix("/dashboard/", http.FileServer(http.Dir("./dashboard/"))))
	r.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard/", http.StatusMovedPermanently)
	})

	// 公共
	r.HandleFunc("/health", s.handleHealth).Methods("GET")
	// 调试页面（仅开发环境）
	r.HandleFunc("/debug", s.handleDebug).Methods("GET")
	// 认证接口
	r.HandleFunc("/auth/login", s.handleLogin).Methods("POST")
	r.HandleFunc("/auth/me", s.handleGetUserInfo).Methods("GET")

	// 管理员接口（支持 URL 参数认证，方便浏览器直接访问）- 必须放在前面避免路由冲突
	r.HandleFunc("/admin/payouts", s.handleListPayoutsWithURLAuth).Methods("GET")
	r.HandleFunc("/admin/events", s.handleListEventsWithURLAuth).Methods("GET")

	// 移除公开路由，所有商家数据访问都需要认证

	// 管理员管理接口（需要管理员权限）
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(authMiddleware)
	admin.Use(adminOnlyMiddleware)
	admin.HandleFunc("/backfill", s.handleBackfill).Methods("POST")
	admin.HandleFunc("/admins", s.handleListAdmins).Methods("GET")
	admin.HandleFunc("/admins", s.handleAddAdmin).Methods("POST")
	admin.HandleFunc("/admins/{address}", s.handleRemoveAdmin).Methods("DELETE")
	admin.HandleFunc("/merchants", s.handleListMerchants).Methods("GET")
	admin.HandleFunc("/merchants", s.handleAddMerchant).Methods("POST")
	admin.HandleFunc("/merchants/{address}", s.handleRemoveMerchant).Methods("DELETE")

	// 商家需要登录
	merchant := r.PathPrefix("/merchant").Subrouter()
	merchant.Use(authMiddleware)
	merchant.HandleFunc("/payouts", s.handleListMerchantPayouts).Methods("GET")

	// 如果你仍希望提供未受保护的全量列表，请取消注释下面这行
	// r.HandleFunc("/payouts", s.handleListPayouts).Methods("GET")

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

// handleDebug 提供调试页面（仅开发环境）
func (s *Server) handleDebug(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "debug.html")
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Address string `json:"address"`
	Role    string `json:"role"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Token   string `json:"token"`
	Address string `json:"address"`
	Role    string `json:"role"`
}

// UserInfoResponse 用户信息响应结构
type UserInfoResponse struct {
	Address string `json:"address"`
	Role    string `json:"role"`
}

// PayoutResponse API响应的支付记录结构，包含格式化后的金额
type PayoutResponse struct {
	TxHash         string    `json:"TxHash"`
	BlockNumber    int64     `json:"BlockNumber"`
	DstEid         int64     `json:"DstEid"`
	DstChain       string    `json:"DstChain"` // 目标链名称
	Payer          string    `json:"Payer"`
	Merchant       string    `json:"Merchant"`
	SrcToken       string    `json:"SrcToken"`
	DstToken       string    `json:"DstToken"`
	GrossAmount    string    `json:"GrossAmount"`    // 原始值（字符串）
	NetAmount      string    `json:"NetAmount"`      // 原始值（字符串）
	GrossAmountUSD string    `json:"GrossAmountUSD"` // 格式化后的USD值
	NetAmountUSD   string    `json:"NetAmountUSD"`   // 格式化后的USD值
	Status         string    `json:"Status"`
	Timestamp      time.Time `json:"Timestamp"`
	SolanaMerchant string    `json:"SolanaMerchant,omitempty"` // Solana 原始地址
	SolanaPayer    string    `json:"SolanaPayer,omitempty"`    // Solana 原始地址
}

// formatAmountToUSD 将原始金额转换为USD显示格式
func formatAmountToUSD(amount *big.Int) string {
	if amount == nil {
		return "0.00"
	}

	// 转换为浮点数，除以10^6（USDT/USDC的6位小数）
	f := new(big.Float).SetInt(amount)
	divisor := big.NewFloat(1000000) // 10^6
	f.Quo(f, divisor)

	// 格式化为2位小数
	return fmt.Sprintf("%.2f", f)
}

// getChainName 根据 EID 返回链名称
func getChainName(eid int64) string {
	switch eid {
	case 40245:
		return "Base Sepolia"
	case 40231:
		return "Arbitrum Sepolia"
	case 40168:
		return "Solana Devnet"
	case 30168:
		return "Solana Mainnet"
	default:
		return fmt.Sprintf("Chain %d", eid)
	}
}

// isSolanaChain 判断是否为 Solana 链
func isSolanaChain(eid int64) bool {
	return eid == 40168 || eid == 30168
}

// convertPayoutToResponse 转换 PayoutRecord 到 PayoutResponse
func convertPayoutToResponse(payout PayoutRecord) PayoutResponse {
	resp := PayoutResponse{
		TxHash:         payout.TxHash,
		BlockNumber:    payout.BlockNumber,
		DstEid:         payout.DstEid,
		DstChain:       getChainName(payout.DstEid),
		SrcToken:       payout.SrcToken.Hex(),
		DstToken:       payout.DstToken.Hex(),
		GrossAmount:    payout.GrossAmount.String(),
		NetAmount:      payout.NetAmount.String(),
		GrossAmountUSD: formatAmountToUSD(payout.GrossAmount),
		NetAmountUSD:   formatAmountToUSD(payout.NetAmount),
		Status:         payout.Status,
		Timestamp:      payout.Timestamp,
	}

	// 根据链类型决定显示哪种地址格式
	if isSolanaChain(payout.DstEid) && payout.SolanaMerchant != "" {
		// Solana 目标链：Merchant 显示 Solana 地址
		resp.Merchant = payout.SolanaMerchant
		// Payer 仍然是源链（EVM）地址
		resp.Payer = payout.Payer.Hex()
		// 如果需要也保存 Solana Payer（未来功能）
		if payout.SolanaPayer != "" {
			resp.SolanaPayer = payout.SolanaPayer
		}
	} else {
		// EVM 链：全部显示 0x 地址
		resp.Merchant = payout.Merchant.Hex()
		resp.Payer = payout.Payer.Hex()
	}

	return resp
}

// handleLogin 处理登录请求
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证地址格式（支持 EVM 和 Solana）
	if !isValidAddress(req.Address) {
		http.Error(w, "Invalid address format. Must be EVM (0x...) or Solana (Base58)", http.StatusBadRequest)
		return
	}

	// 验证角色
	if req.Role != "merchant" && req.Role != "admin" {
		http.Error(w, "Invalid role. Must be 'merchant' or 'admin'", http.StatusBadRequest)
		return
	}

	// 标准化地址（小写）
	normalizedAddr := normalizeAddress(req.Address)

	// 如果请求管理员权限，检查地址是否在管理员白名单中
	if req.Role == "admin" && !adminConfig.IsAdminAddress(normalizedAddr) {
		http.Error(w, "Address not authorized for admin access", http.StatusForbidden)
		return
	}

	// 如果请求商家权限，检查地址是否在商家白名单中
	if req.Role == "merchant" && !merchantConfig.IsMerchantAddress(normalizedAddr) {
		http.Error(w, "Address not authorized for merchant access", http.StatusForbidden)
		return
	}

	// 生成 JWT token
	token, err := generateJWT(req.Address, req.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		Token:   token,
		Address: req.Address,
		Role:    req.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetUserInfo 获取当前用户信息
func (s *Server) handleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	// 从 Authorization header 获取 token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	// 验证 JWT token
	merchant, role, err := verifyAndExtractClaims(tokenStr, jwtSecret)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	response := UserInfoResponse{
		Address: merchant,
		Role:    role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListPayoutsWithURLAuth 支持 URL 参数认证的管理员接口
func (s *Server) handleListPayoutsWithURLAuth(w http.ResponseWriter, r *http.Request) {
	// 从 URL 参数获取 token
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token parameter. Use: /admin/payouts?token=YOUR_JWT_TOKEN", http.StatusUnauthorized)
		return
	}

	// 验证 JWT token
	merchant, role, err := verifyAndExtractClaims(token, jwtSecret)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// 检查是否为管理员
	if role != "admin" {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	// 将认证信息添加到请求上下文
	ctx := context.WithValue(r.Context(), ctxKeyMerchant, common.HexToAddress(merchant))
	ctx = context.WithValue(ctx, ctxKeyRole, role)
	r = r.WithContext(ctx)

	// 调用原有的处理逻辑
	s.handleListPayouts(w, r)
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

	// 转换为API响应格式
	responses := make([]PayoutResponse, len(list))
	for i, payout := range list {
		responses[i] = convertPayoutToResponse(payout)
	}

	_ = json.NewEncoder(w).Encode(responses)
}

// 已移除公开路由，所有商家数据访问都需要认证

// 占位函数：模拟从认证/会话中获取商家地址
// 【重要】在实际项目中，你需要替换为从用户的认证信息（如 JWT token）中安全提取地址的逻辑。
func getAuthenticatedMerchantAddress(r *http.Request) (common.Address, error) {
	v := r.Context().Value(ctxKeyMerchant)
	addr, ok := v.(common.Address)
	if !ok {
		return common.Address{}, fmt.Errorf("no merchant in context or context is not set")
	}
	return addr, nil
}

// handleListMerchantPayouts 处理 /merchant/payouts 请求，只返回与当前商家相关的交易数据
func (s *Server) handleListMerchantPayouts(w http.ResponseWriter, r *http.Request) {
	// 1. 【安全】获取当前请求的商家地址（原始地址）
	merchantOriginal := r.Context().Value("merchant_original")
	if merchantOriginal == nil {
		http.Error(w, "Authentication failed or merchant address missing", http.StatusUnauthorized)
		return
	}
	merchantStr, ok := merchantOriginal.(string)
	if !ok {
		http.Error(w, "Invalid merchant address format", http.StatusUnauthorized)
		return
	}

	// 2. 解析分页参数
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	// 3. 调用 Store 层的新方法进行筛选
	// 支持 Solana 地址查询
	list, err := s.store.ListMerchantPayoutsByString(merchantStr, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. 转换为API响应格式
	responses := make([]PayoutResponse, len(list))
	for i, payout := range list {
		responses[i] = convertPayoutToResponse(payout)
	}

	// 5. 返回结果
	_ = json.NewEncoder(w).Encode(responses)
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

// 管理员管理接口

// handleListAdmins 列出所有管理员地址
func (s *Server) handleListAdmins(w http.ResponseWriter, r *http.Request) {
	addresses := adminConfig.GetAdminAddresses()
	response := map[string]interface{}{
		"admins": addresses,
		"count":  len(addresses),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAddAdmin 添加管理员地址
func (s *Server) handleAddAdmin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证地址格式
	if req.Address == "" || len(req.Address) != 42 || !strings.HasPrefix(req.Address, "0x") {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

	// 检查是否已经是管理员
	if adminConfig.IsAdminAddress(req.Address) {
		http.Error(w, "Address is already an admin", http.StatusConflict)
		return
	}

	// 添加管理员
	adminConfig.AddAdminAddress(req.Address)

	response := map[string]interface{}{
		"message": "Admin added successfully",
		"address": req.Address,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRemoveAdmin 移除管理员地址
func (s *Server) handleRemoveAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// 验证地址格式
	if address == "" || len(address) != 42 || !strings.HasPrefix(address, "0x") {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

	// 检查是否是管理员
	if !adminConfig.IsAdminAddress(address) {
		http.Error(w, "Address is not an admin", http.StatusNotFound)
		return
	}

	// 移除管理员
	adminConfig.RemoveAdminAddress(address)

	response := map[string]interface{}{
		"message": "Admin removed successfully",
		"address": address,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 商家管理接口

// handleListMerchants 列出所有商家地址
func (s *Server) handleListMerchants(w http.ResponseWriter, r *http.Request) {
	addresses := merchantConfig.GetMerchantAddresses()
	response := map[string]interface{}{
		"merchants": addresses,
		"count":     len(addresses),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAddMerchant 添加商家地址
func (s *Server) handleAddMerchant(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证地址格式
	if req.Address == "" || len(req.Address) != 42 || !strings.HasPrefix(req.Address, "0x") {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

	// 检查是否已经是商家
	if merchantConfig.IsMerchantAddress(req.Address) {
		http.Error(w, "Address is already a merchant", http.StatusConflict)
		return
	}

	// 添加商家
	merchantConfig.AddMerchantAddress(req.Address)

	response := map[string]interface{}{
		"message": "Merchant added successfully",
		"address": req.Address,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRemoveMerchant 移除商家地址
func (s *Server) handleRemoveMerchant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// 验证地址格式
	if address == "" || len(address) != 42 || !strings.HasPrefix(address, "0x") {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

	// 检查是否是商家
	if !merchantConfig.IsMerchantAddress(address) {
		http.Error(w, "Address is not a merchant", http.StatusNotFound)
		return
	}

	// 移除商家
	merchantConfig.RemoveMerchantAddress(address)

	response := map[string]interface{}{
		"message": "Merchant removed successfully",
		"address": address,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// helper to read wss status safely
func getWssStatus() string {
	mu.Lock()
	defer mu.Unlock()
	return wssStatus
}

// handleListEventsWithURLAuth 列出所有原始事件（支持 URL 参数认证）
func (s *Server) handleListEventsWithURLAuth(w http.ResponseWriter, r *http.Request) {
	// 从 URL 参数获取 token
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token parameter", http.StatusUnauthorized)
		return
	}

	// 验证 token
	merchant, role, err := verifyAndExtractClaims(token, jwtSecret)
	if err != nil {
		log.Printf("handleListEventsWithURLAuth: invalid token: %v", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 验证是否为管理员
	if role != "admin" || !adminConfig.IsAdminAddress(merchant) {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	// 解析分页参数
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // 默认50条
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 500 {
				limit = 500 // 最多500条
			}
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// 获取事件
	events, err := s.store.GetAllEvents(limit, offset)
	if err != nil {
		log.Printf("handleListEventsWithURLAuth: GetAllEvents error: %v", err)
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	// 获取总数
	total, err := s.store.GetEventCount()
	if err != nil {
		log.Printf("handleListEventsWithURLAuth: GetEventCount error: %v", err)
		total = 0
	}

	response := map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDashboardPayoutsWithAuth Dashboard API - 获取所有交易（需要管理员认证）
func (s *Server) handleDashboardPayoutsWithAuth(w http.ResponseWriter, r *http.Request) {
	// 检查 Bearer Token
	tokenStr, err := parseBearerToken(r)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// 验证 token
	merchant, role, err := verifyAndExtractClaims(tokenStr, jwtSecret)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 验证是否为管理员
	if role != "admin" || !adminConfig.IsAdminAddress(merchant) {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	// 解析分页参数
	limit := 100
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	payouts, err := s.store.ListPayouts(limit, offset)
	if err != nil {
		log.Printf("handleDashboardPayoutsWithAuth: ListPayouts error: %v", err)
		http.Error(w, "Failed to fetch payouts", http.StatusInternalServerError)
		return
	}

	// 转换为带 USD 字段的响应
	responses := make([]PayoutResponse, len(payouts))
	for i, p := range payouts {
		responses[i] = convertPayoutToResponse(p)
	}

	resp := map[string]interface{}{
		"list": responses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleDashboardMerchantPayouts Dashboard API - 获取特定商户的交易
func (s *Server) handleDashboardMerchantPayouts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	addressStr := vars["address"]

	// 验证地址格式（支持 EVM 和 Solana）
	if !isValidAddress(addressStr) {
		http.Error(w, "Invalid address format", http.StatusBadRequest)
		return
	}

	// 验证是否为白名单商户（使用标准化地址）
	normalizedAddr := normalizeAddress(addressStr)
	if !merchantConfig.IsMerchantAddress(normalizedAddr) {
		http.Error(w, "Merchant not found or not authorized", http.StatusNotFound)
		return
	}

	// 转换为 EVM 格式用于数据库查询
	var merchantAddr common.Address
	if isValidEVMAddress(addressStr) {
		merchantAddr = common.HexToAddress(addressStr)
	} else {
		// Solana 地址：转换为 EVM 格式查询
		merchantAddr = solanaAddressToEVMAddressShared(addressStr)
	}

	limit := 100
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	payouts, err := s.store.ListMerchantPayouts(merchantAddr, limit, offset)
	if err != nil {
		log.Printf("handleDashboardMerchantPayouts: ListMerchantPayouts error: %v", err)
		http.Error(w, "Failed to fetch payouts", http.StatusInternalServerError)
		return
	}

	// 转换为带 USD 字段的响应
	responses := make([]PayoutResponse, len(payouts))
	for i, p := range payouts {
		responses[i] = convertPayoutToResponse(p)
	}

	resp := map[string]interface{}{
		"list": responses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
