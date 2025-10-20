package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/golang-jwt/jwt/v5"
)

// Test JWT Secret - 与 main.go 中的默认值保持一致
var testJWTSecret = []byte("dev-local-secret-change-me")

// --------------------------- Mock 辅助结构 ---------------------------

// MockStore 用于测试 API 路由和鉴权，隔离数据库依赖。
type MockStore struct {
	ListPayoutsFn         func(limit, offset int) ([]PayoutRecord, error)
	ListMerchantPayoutsFn func(merchant common.Address, limit, offset int) ([]PayoutRecord, error)
}

// 模拟 ListPayouts 方法
func (m *MockStore) ListPayouts(limit, offset int) ([]PayoutRecord, error) {
	if m.ListPayoutsFn != nil {
		return m.ListPayoutsFn(limit, offset)
	}
	return []PayoutRecord{}, nil // 默认返回空列表
}

// 模拟 ListMerchantPayouts 方法
func (m *MockStore) ListMerchantPayouts(merchant common.Address, limit, offset int) ([]PayoutRecord, error) {
	if m.ListMerchantPayoutsFn != nil {
		return m.ListMerchantPayoutsFn(merchant, limit, offset)
	}
	return []PayoutRecord{}, nil
}

// 占位符方法：Store 结构体中必须存在的方法 (尽管在 API 测试中可能用不到)
func (m *MockStore) Close() error                                             { return nil }
func (m *MockStore) GetLastProcessedBlock(chainID int) (uint64, error)        { return 0, nil }
func (m *MockStore) SetLastProcessedBlock(chainID int, blockNum uint64) error { return nil }
func (m *MockStore) InsertEventIfNotExists(txHash string, logIndex uint, blockNumber uint64, rawLog string) (bool, error) {
	return true, nil
}
func (m *MockStore) MarkEventAsParsed(txHash string, logIndex uint) error { return nil }
func (m *MockStore) UpsertPayout(rec PayoutRecord) error                  { return nil }
func (m *MockStore) ListPendingPayouts(limit int) ([]PayoutRecord, error) { return nil, nil }
func (m *MockStore) UpdatePayoutStatus(txHash, status string) error       { return nil }

// 注意：由于 Store 是具体结构体，为通过编译，必须实现所有方法。

// generateTestJWT 生成测试用的 JWT
func generateTestJWT(merchant string, role string) (string, error) {
	// 设置 24 小时过期时间
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"merchant": merchant,
		"role":     role,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用测试密钥签名
	return token.SignedString(testJWTSecret)
}

// createTestServer 创建一个用于测试的 Server 实例
func createTestServer(store *MockStore) *Server {
	// 由于 Server.routes() 依赖于全局变量 jwtSecret，我们需要在测试前设置它
	jwtSecret = testJWTSecret

	// 需要 ethclient 占位符，但不必是有效的
	mockClient, _ := ethclient.Dial("http://127.0.0.1:8545")

	// 创建一个 Processor 占位符
	mockProcessor := NewProcessor(mockClient, nil)

	return NewServer(
		store, // 直接以接口注入，无需强转
		mockClient,
		common.Address{},
		common.Hash{},
		mockProcessor,
	)
}

// --------------------------- 测试用例 ---------------------------

func TestAuthMiddleware(t *testing.T) {
	testMerchantAddr := "0x1111111111111111111111111111111111111111"
	testAdminAddr := "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	// 1. 设置 Mock Store
	var merchantCallAddr common.Address
	mockStore := &MockStore{
		ListMerchantPayoutsFn: func(merchant common.Address, limit, offset int) ([]PayoutRecord, error) {
			merchantCallAddr = merchant // 记录被调用的商家地址
			return []PayoutRecord{{TxHash: "test_tx_merch"}}, nil
		},
		ListPayoutsFn: func(limit, offset int) ([]PayoutRecord, error) {
			return []PayoutRecord{{TxHash: "test_tx_all"}}, nil
		},
	}

	// 2. 创建 Server 和 Router
	server := createTestServer(mockStore)
	router := server.routes()

	// 3. 生成测试 Token
	merchantToken, err := generateTestJWT(testMerchantAddr, "merchant")
	if err != nil {
		t.Fatalf("Failed to generate merchant token: %v", err)
	}
	adminToken, err := generateTestJWT(testAdminAddr, "admin")
	if err != nil {
		t.Fatalf("Failed to generate admin token: %v", err)
	}

	tests := []struct {
		name         string
		method       string
		path         string
		token        string
		expectedCode int
		expectedTx   string // 预期返回的交易哈希，用于验证 Store 调用
		isMerchantFn bool   // 预期调用 ListMerchantPayoutsFn
	}{
		// --- 商家路由测试 ---
		{
			name:         "Merchant_Success",
			method:       "GET",
			path:         "/merchant/payouts",
			token:        merchantToken,
			expectedCode: http.StatusOK,
			expectedTx:   "test_tx_merch",
			isMerchantFn: true,
		},
		{
			name:         "Merchant_NoToken_401",
			method:       "GET",
			path:         "/merchant/payouts",
			token:        "",
			expectedCode: http.StatusUnauthorized,
			isMerchantFn: false,
		},
		{
			name:         "Admin_CanAccessMerchant",
			method:       "GET",
			path:         "/merchant/payouts",
			token:        adminToken,
			expectedCode: http.StatusOK,
			expectedTx:   "test_tx_merch",
			isMerchantFn: true,
		},
		// --- 管理员路由测试 ---
		{
			name:         "Admin_Success",
			method:       "GET",
			path:         "/admin/payouts",
			token:        adminToken,
			expectedCode: http.StatusOK,
			expectedTx:   "test_tx_all",
			isMerchantFn: false,
		},
		{
			name:         "Admin_NoToken_401",
			method:       "GET",
			path:         "/admin/payouts",
			token:        "",
			expectedCode: http.StatusUnauthorized,
			isMerchantFn: false,
		},
		{
			name:         "Merchant_AccessAdmin_403",
			method:       "GET",
			path:         "/admin/payouts",
			token:        merchantToken,
			expectedCode: http.StatusForbidden,
			isMerchantFn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置记录
			merchantCallAddr = common.Address{}

			req, _ := http.NewRequest(tt.method, tt.path, nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedCode {
				t.Errorf("Handler returned wrong status code: got %v, want %v. Body: %s",
					status, tt.expectedCode, rr.Body.String())
			}

			// 进一步验证成功状态下的内容和调用
			if tt.expectedCode == http.StatusOK {
				bodyBytes, _ := io.ReadAll(rr.Body)
				var records []PayoutRecord
				if err := json.Unmarshal(bodyBytes, &records); err != nil {
					t.Fatalf("Could not unmarshal response body: %v", err)
				}

				if len(records) == 0 {
					t.Fatalf("Expected records, got 0")
				}

				// 验证返回的 TxHash 是否正确（用于区分 ListMerchant/ListAll）
				if records[0].TxHash != tt.expectedTx {
					t.Errorf("Handler returned wrong record: got TxHash %s, want %s",
						records[0].TxHash, tt.expectedTx)
				}

				// 验证 ListMerchantPayouts 是否被调用了正确的地址
				if tt.isMerchantFn && merchantCallAddr.Hex() != testMerchantAddr {
					// 如果是商家路由，且成功，必须验证商家地址是否被正确传递
					// 注意：AdminToken 也可以访问 /merchant，此时应该传递 AdminToken 载荷中的地址
					expectedAddr := common.HexToAddress(testMerchantAddr).Hex()
					if strings.Contains(tt.name, "Admin") {
						expectedAddr = common.HexToAddress(testAdminAddr).Hex()
					}

					if merchantCallAddr.Hex() != expectedAddr {
						t.Errorf("ListMerchantPayouts not called with correct merchant address. Got %s, want %s",
							merchantCallAddr.Hex(), expectedAddr)
					}
				}
			}
		})
	}
}
