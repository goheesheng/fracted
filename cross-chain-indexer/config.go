package main

import (
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
)

// AdminConfig 管理员配置
type AdminConfig struct {
	// 管理员地址白名单
	AdminAddresses map[string]bool
}

// MerchantConfig 商家配置
type MerchantConfig struct {
	// 商家地址白名单
	MerchantAddresses map[string]bool
}

// LoadAdminConfig 加载管理员配置
func LoadAdminConfig() *AdminConfig {
	config := &AdminConfig{
		AdminAddresses: make(map[string]bool),
	}

	// 从环境变量读取管理员地址（用逗号分隔）
	adminEnv := os.Getenv("ADMIN_ADDRESSES")
	if adminEnv != "" {
		addresses := strings.Split(adminEnv, ",")
		for _, addr := range addresses {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				config.AdminAddresses[strings.ToLower(addr)] = true
			}
		}
	}

	// 如果没有设置环境变量，使用默认的管理员地址
	if len(config.AdminAddresses) == 0 {
		config.AdminAddresses["0x27f9b6a7c1fd66ac4d0e76a2d43b35e8590165f6"] = true
	}

	return config
}

// IsAdminAddress 检查地址是否为管理员
func (c *AdminConfig) IsAdminAddress(address string) bool {
	return c.AdminAddresses[strings.ToLower(address)]
}

// AddAdminAddress 添加管理员地址（运行时动态添加）
func (c *AdminConfig) AddAdminAddress(address string) {
	c.AdminAddresses[strings.ToLower(address)] = true
}

// RemoveAdminAddress 移除管理员地址
func (c *AdminConfig) RemoveAdminAddress(address string) {
	delete(c.AdminAddresses, strings.ToLower(address))
}

// GetAdminAddresses 获取所有管理员地址
func (c *AdminConfig) GetAdminAddresses() []string {
	var addresses []string
	for addr := range c.AdminAddresses {
		addresses = append(addresses, addr)
	}
	return addresses
}

// LoadMerchantConfig 加载商家配置
func LoadMerchantConfig() *MerchantConfig {
	config := &MerchantConfig{
		MerchantAddresses: make(map[string]bool),
	}

	// 从环境变量读取商家地址（用逗号分隔）
	merchantEnv := os.Getenv("MERCHANT_ADDRESSES")
	if merchantEnv != "" {
		addresses := strings.Split(merchantEnv, ",")
		for _, addr := range addresses {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				config.MerchantAddresses[strings.ToLower(addr)] = true
			}
		}
	}

	// 如果没有设置环境变量，使用默认的商家地址
	if len(config.MerchantAddresses) == 0 {
		// EVM 商家地址
		config.MerchantAddresses["0x77ed7f6455fe291728a48785090292e3d10f53bb"] = true
		config.MerchantAddresses["0x27f9b6a7c1fd66ac4d0e76a2d43b35e8590165f6"] = true
		config.MerchantAddresses["0xb7aa464b19037cf3db7f723504dfafe7b63aab84"] = true
		// 测试商家
		config.MerchantAddresses["0xfedcba0987654321fedcba0987654321fedcba09"] = true
		config.MerchantAddresses["0x9876543210987654321098765432109876543210"] = true
		config.MerchantAddresses["0xabcdef1234567890abcdef1234567890abcdef12"] = true

		// Solana 商家地址（从数据库中提取的真实商家）
		config.MerchantAddresses["6h7aykpuhnmuca92gc82oarxc48igkli14mczh9xnlpp"] = true // 最常见的商家
		config.MerchantAddresses["a9qyh2sten3xffk95wzr2hslFMC2781oPwKexPySNJrt"] = true // vault_authority
		config.MerchantAddresses["awun8gk6x3xkr73ybrw2h8wxc6qgbjrvehs5dgejx3zs"] = true // 另一个商家
		config.MerchantAddresses["7xkxtg2cw87d97txjsdpbd5jbkhetqa83tzrujosgasu"] = true // 测试商家（Arb->Solana跨链）
	}

	return config
}

// IsMerchantAddress 检查地址是否为商家
func (c *MerchantConfig) IsMerchantAddress(address string) bool {
	return c.MerchantAddresses[strings.ToLower(address)]
}

// AddMerchantAddress 添加商家地址（运行时动态添加）
func (c *MerchantConfig) AddMerchantAddress(address string) {
	c.MerchantAddresses[strings.ToLower(address)] = true
}

// RemoveMerchantAddress 移除商家地址
func (c *MerchantConfig) RemoveMerchantAddress(address string) {
	delete(c.MerchantAddresses, strings.ToLower(address))
}

// GetMerchantAddresses 获取所有商家地址
func (c *MerchantConfig) GetMerchantAddresses() []string {
	var addresses []string
	for addr := range c.MerchantAddresses {
		addresses = append(addresses, addr)
	}
	return addresses
}

// isValidEVMAddress 验证 EVM 地址格式
func isValidEVMAddress(address string) bool {
	if address == "" {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	if len(address) != 42 {
		return false
	}
	return true
}

// isValidSolanaAddress 验证 Solana 地址格式
func isValidSolanaAddress(address string) bool {
	if address == "" {
		return false
	}
	// Solana 地址通常是 32-44 字符的 Base58 编码
	if len(address) < 32 || len(address) > 44 {
		return false
	}
	// 尝试解析为 Solana 公钥
	_, err := solana.PublicKeyFromBase58(address)
	return err == nil
}

// isValidAddress 验证地址格式（支持 EVM 和 Solana）
func isValidAddress(address string) bool {
	return isValidEVMAddress(address) || isValidSolanaAddress(address)
}

// normalizeAddress 标准化地址（统一小写，便于比较）
func normalizeAddress(address string) string {
	return strings.ToLower(strings.TrimSpace(address))
}

// solanaAddressToEVMAddressShared 将 Solana 地址转换为 EVM 地址（用于数据库查询）
func solanaAddressToEVMAddressShared(solAddr string) common.Address {
	if solAddr == "" {
		return common.HexToAddress("0x0000000000000000000000000000000000000000")
	}

	pubkey, err := solana.PublicKeyFromBase58(solAddr)
	if err != nil {
		return common.HexToAddress("0x0000000000000000000000000000000000000000")
	}

	// 使用公钥的前20字节作为 EVM 地址
	var evmAddr common.Address
	copy(evmAddr[:], pubkey[:20])
	return evmAddr
}
