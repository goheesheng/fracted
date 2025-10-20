package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// TestAdminSecurity 测试管理员安全机制
func TestAdminSecurity(t *testing.T) {
	// 重置配置
	adminConfig = LoadAdminConfig()
	merchantConfig = LoadMerchantConfig()

	// 创建测试服务器
	mockStore := &MockStore{}
	server := createTestServer(mockStore)

	tests := []struct {
		name           string
		address        string
		role           string
		expectedStatus int
		description    string
	}{
		{
			name:           "Valid Admin Login",
			address:        "0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6",
			role:           "admin",
			expectedStatus: http.StatusOK,
			description:    "白名单中的管理员地址应该能够登录",
		},
		{
			name:           "Invalid Admin Login",
			address:        "0x1234567890123456789012345678901234567890",
			role:           "admin",
			expectedStatus: http.StatusForbidden,
			description:    "不在白名单中的地址不应该能够获得管理员权限",
		},
		{
			name:           "Valid Merchant Login",
			address:        "0x77Ed7f6455FE291728A48785090292e3D10F53Bb",
			role:           "merchant",
			expectedStatus: http.StatusOK,
			description:    "白名单中的商家地址应该能够登录",
		},
		{
			name:           "Invalid Merchant Login",
			address:        "0x1234567890123456789012345678901234567890",
			role:           "merchant",
			expectedStatus: http.StatusForbidden,
			description:    "不在白名单中的地址不应该能够获得商家权限",
		},
		{
			name:           "Invalid Address Format",
			address:        "invalid-address",
			role:           "admin",
			expectedStatus: http.StatusBadRequest,
			description:    "无效地址格式应该被拒绝",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建登录请求
			loginReq := LoginRequest{
				Address: tt.address,
				Role:    tt.role,
			}
			reqBody, _ := json.Marshal(loginReq)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// 执行请求
			w := httptest.NewRecorder()
			server.handleLogin(w, req)

			// 验证响应状态码
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. %s", tt.expectedStatus, w.Code, tt.description)
			}

			// 如果是成功登录，验证响应内容
			if w.Code == http.StatusOK {
				var response LoginResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.Address != tt.address {
					t.Errorf("Expected address %s, got %s", tt.address, response.Address)
				}
				if response.Role != tt.role {
					t.Errorf("Expected role %s, got %s", tt.role, response.Role)
				}
				if response.Token == "" {
					t.Error("Token should not be empty")
				}
			}
		})
	}
}

// TestAdminManagement 测试管理员管理功能
func TestAdminManagement(t *testing.T) {
	// 重置配置
	adminConfig = LoadAdminConfig()
	merchantConfig = LoadMerchantConfig()

	mockStore := &MockStore{}
	server := createTestServer(mockStore)

	// 首先以管理员身份登录
	adminToken := getAdminToken(t, server)

	t.Run("List Admins", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/admins", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		server.handleListAdmins(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		admins, ok := response["admins"].([]interface{})
		if !ok {
			t.Error("Response should contain admins array")
		}

		if len(admins) == 0 {
			t.Error("Should have at least one admin")
		}
	})

	t.Run("Add Admin", func(t *testing.T) {
		newAdminReq := map[string]string{
			"address": "0x1234567890123456789012345678901234567890",
		}
		reqBody, _ := json.Marshal(newAdminReq)
		req := httptest.NewRequest("POST", "/admin/admins", bytes.NewBuffer(reqBody))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.handleAddAdmin(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// 验证新管理员是否被添加
		if !adminConfig.IsAdminAddress("0x1234567890123456789012345678901234567890") {
			t.Error("New admin should be added to config")
		}
	})

	t.Run("Remove Admin", func(t *testing.T) {
		// 确保地址是管理员
		if !adminConfig.IsAdminAddress("0x1234567890123456789012345678901234567890") {
			adminConfig.AddAdminAddress("0x1234567890123456789012345678901234567890")
		}

		req := httptest.NewRequest("DELETE", "/admin/admins/0x1234567890123456789012345678901234567890", nil)
		req = mux.SetURLVars(req, map[string]string{"address": "0x1234567890123456789012345678901234567890"})
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		server.handleRemoveAdmin(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		// 验证管理员是否被移除
		if adminConfig.IsAdminAddress("0x1234567890123456789012345678901234567890") {
			t.Error("Admin should be removed from config")
		}
	})

	t.Run("List Merchants", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/merchants", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		server.handleListMerchants(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}

		merchants, ok := response["merchants"].([]interface{})
		if !ok {
			t.Error("Response should contain merchants array")
		}

		if len(merchants) == 0 {
			t.Error("Should have at least one merchant")
		}
	})

	t.Run("Add Merchant", func(t *testing.T) {
		newMerchantReq := map[string]string{
			"address": "0x1234567890123456789012345678901234567890",
		}
		reqBody, _ := json.Marshal(newMerchantReq)
		req := httptest.NewRequest("POST", "/admin/merchants", bytes.NewBuffer(reqBody))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.handleAddMerchant(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// 验证新商家是否被添加
		if !merchantConfig.IsMerchantAddress("0x1234567890123456789012345678901234567890") {
			t.Error("New merchant should be added to config")
		}
	})

	t.Run("Remove Merchant", func(t *testing.T) {
		// 确保地址是商家
		if !merchantConfig.IsMerchantAddress("0x1234567890123456789012345678901234567890") {
			merchantConfig.AddMerchantAddress("0x1234567890123456789012345678901234567890")
		}

		req := httptest.NewRequest("DELETE", "/admin/merchants/0x1234567890123456789012345678901234567890", nil)
		req = mux.SetURLVars(req, map[string]string{"address": "0x1234567890123456789012345678901234567890"})
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		server.handleRemoveMerchant(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Response: %s", w.Code, w.Body.String())
		}

		// 验证商家是否被移除
		if merchantConfig.IsMerchantAddress("0x1234567890123456789012345678901234567890") {
			t.Error("Merchant should be removed from config")
		}
	})
}

// 辅助函数：获取管理员token
func getAdminToken(t *testing.T, server *Server) string {
	loginReq := LoginRequest{
		Address: "0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6",
		Role:    "admin",
	}
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.handleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get admin token: %d", w.Code)
	}

	var response LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}

	return response.Token
}
