package config

import (
	"fmt"
	"os"
)

// AuthConfig 认证配置
type AuthConfig struct {
	// 认证方式
	AuthMethod string `json:"AuthMethod"`
	// AAP ID
	AapArn string `json:"AapArn"`
	// 令牌路径（用于ACK OIDC）
	TokenPath string `json:"TokenPath,omitempty"`
	// ClientKey配置路径
	ClientKeyConfigPath string `json:"ClientKeyConfigPath,omitempty"`
	// ClientKeyPassword ClientKey密码值（可从环境变量间接获取）
	ClientKeyPassword string `json:"ClientKeyPassword,omitempty"`
	// ClientKeyPasswordPath ClientKey密码文件路径
	ClientKeyPasswordPath string `json:"ClientKeyPasswordPath,omitempty"`
	// KmsEndpoint KMS端点
	KmsEndpoint string `json:"KmsEndpoint,omitempty"`
	Ca          string `json:"Ca,omitempty"`
}

const (
	ECSInstanceIdentity = "ECSInstanceIdentity"
	ACKOidcJwt          = "ACKOidcJwt"
	ClientKey           = "ClientKey"
)

/*
写一个config 校验文件
1. 文件路径，文件历经参数不为空的话，要校验是否存在这文件
2. 除 Client Key 以外，都要存在 AapArn
*/

// Validate 校验配置参数
func (c *AuthConfig) Validate() error {
	// 检查认证方式是否为空
	if c.AuthMethod == "" {
		return fmt.Errorf("AuthMethod is required")
	}
	if c.KmsEndpoint == "" {
		return fmt.Errorf("KmsEndpoint is required")
	}

	// 检查不同认证方式的特定要求
	switch c.AuthMethod {
	case ACKOidcJwt:
	case ECSInstanceIdentity:
	case ClientKey:
		// Client Key 认证不需要 AAP ID
		// ClientKeyConfigPath 不为空时，需要校验文件是否存在
		if c.ClientKeyConfigPath == "" {
			return fmt.Errorf("ClientKeyConfigPath is required for Client Key authentication")
		}
		if _, err := os.Stat(c.ClientKeyConfigPath); os.IsNotExist(err) {
			return fmt.Errorf("client key config file does not exist: %s", c.ClientKeyConfigPath)
		}
		// 如果 ClientKeyPassword 为空，则需要检查 ClientKeyPasswordPath
		if c.ClientKeyPassword == "" {
			if c.ClientKeyPasswordPath == "" {
				return fmt.Errorf("ClientKeyPassword or ClientKeyPasswordPath is required for Client Key authentication")
			}
			if _, err := os.Stat(c.ClientKeyPasswordPath); os.IsNotExist(err) {
				return fmt.Errorf("client key password file does not exist: %s", c.ClientKeyPasswordPath)
			}
		}
	default:
		return fmt.Errorf("unsupported authentication method: %s", c.AuthMethod)
	}

	return nil
}
