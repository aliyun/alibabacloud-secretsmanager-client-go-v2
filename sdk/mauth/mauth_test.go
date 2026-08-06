package mauth

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/config"
	"github.com/stretchr/testify/assert"
)

func TestNewMAuth_ClientKey_Success(t *testing.T) {
	tmpConfigFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpConfigFile.Name())

	cfg := config.AuthConfig{
		AuthMethod:          config.ClientKey,
		ClientKeyConfigPath: tmpConfigFile.Name(),
		ClientKeyPassword:   "test_password",
		KmsEndpoint:         "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, config.ClientKey, m.config.AuthMethod)
}

func TestNewMAuth_ACKOidcJwt_Success(t *testing.T) {
	cfg := config.AuthConfig{
		AuthMethod:  config.ACKOidcJwt,
		AapArn:      "arn:test:aap",
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func TestNewMAuth_ECSInstanceIdentity_Success(t *testing.T) {
	cfg := config.AuthConfig{
		AuthMethod:  config.ECSInstanceIdentity,
		AapArn:      "arn:test:aap",
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func TestNewMAuth_InvalidAuthMethod(t *testing.T) {
	cfg := config.AuthConfig{
		AuthMethod: "InvalidMethod",
	}

	m, err := NewMAuth(cfg, nil)
	assert.NotNil(t, err)
	assert.Nil(t, m)
}

func TestNewMAuth_EmptyAuthMethod(t *testing.T) {
	cfg := config.AuthConfig{
		AuthMethod: "",
	}

	m, err := NewMAuth(cfg, nil)
	assert.NotNil(t, err)
	assert.Nil(t, m)
}

func TestMAuth_GetToken_CacheHit(t *testing.T) {
	tmpConfigFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpConfigFile.Name())

	cfg := config.AuthConfig{
		AuthMethod:          config.ClientKey,
		ClientKeyConfigPath: tmpConfigFile.Name(),
		ClientKeyPassword:   "test_password",
		KmsEndpoint:         "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.Nil(t, err)

	// 手动设置缓存 token，过期时间大于 48 小时
	m.cacheToken = map[string]*AuthToken{
		"ClientKey": {
			TokenType:  AuthTokenTypeClientKey,
			TokenKeyId: "test-key-id",
			TokenValue: "test-token-value",
			ExpiresAt:  time.Now().Add(72 * time.Hour),
		},
	}

	ctx := context.Background()
	token, err := m.GetToken(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "test-key-id", token.TokenKeyId)
	assert.Equal(t, "test-token-value", token.TokenValue)
}

func TestMAuth_ClearToken(t *testing.T) {
	tmpConfigFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpConfigFile.Name())

	cfg := config.AuthConfig{
		AuthMethod:          config.ClientKey,
		ClientKeyConfigPath: tmpConfigFile.Name(),
		ClientKeyPassword:   "test_password",
		KmsEndpoint:         "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.Nil(t, err)

	m.cacheToken = map[string]*AuthToken{
		"ClientKey": {
			TokenType:  AuthTokenTypeClientKey,
			TokenKeyId: "test-key-id",
			TokenValue: "test-token-value",
			ExpiresAt:  time.Now().Add(72 * time.Hour),
		},
	}

	ctx := context.Background()
	err = m.ClearToken(ctx)
	assert.Nil(t, err)
	assert.Empty(t, m.cacheToken)
}

func TestMAuth_getClientKeyByLocalFile_ConfigFileNotExist(t *testing.T) {
	cfg := config.AuthConfig{
		AuthMethod:          config.ClientKey,
		ClientKeyConfigPath: "/not/exist/config.json",
		ClientKeyPassword:   "test_password",
		KmsEndpoint:         "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.NotNil(t, err)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "client key config file does not exist")
}

func TestMAuth_getClientKeyByLocalFile_InvalidConfigFormat(t *testing.T) {
	tmpConfigFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpConfigFile.Name())

	_, err = tmpConfigFile.WriteString("invalid json")
	assert.Nil(t, err)
	tmpConfigFile.Close()

	cfg := config.AuthConfig{
		AuthMethod:          config.ClientKey,
		ClientKeyConfigPath: tmpConfigFile.Name(),
		ClientKeyPassword:   "test_password",
		KmsEndpoint:         "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.Nil(t, err)

	ctx := context.Background()
	_, err = m.getClientKeyByLocalFile(ctx)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal client key config")
}

func TestMAuth_getClientKeyByLocalFile_PasswordFileNotExist(t *testing.T) {
	tmpConfigFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpConfigFile.Name())

	_, err = tmpConfigFile.WriteString(`{"KeyId":"test","PrivateKeyData":"test"}`)
	assert.Nil(t, err)
	tmpConfigFile.Close()

	cfg := config.AuthConfig{
		AuthMethod:            config.ClientKey,
		ClientKeyConfigPath:   tmpConfigFile.Name(),
		ClientKeyPasswordPath: "/not/exist/password.txt",
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.NotNil(t, err)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "client key password file does not exist")
}

func TestMAuth_getClientKeyByLocalFile_MissingPassword(t *testing.T) {
	tmpConfigFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpConfigFile.Name())

	_, err = tmpConfigFile.WriteString(`{"KeyId":"test","PrivateKeyData":"test"}`)
	assert.Nil(t, err)
	tmpConfigFile.Close()

	cfg := config.AuthConfig{
		AuthMethod:          config.ClientKey,
		ClientKeyConfigPath: tmpConfigFile.Name(),
		KmsEndpoint:         "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.NotNil(t, err)
	assert.Nil(t, m)
	assert.Contains(t, err.Error(), "ClientKeyPassword or ClientKeyPasswordPath is required")
}

func TestNewMAuth_MultiCloudIDaaS_Success(t *testing.T) {
	cfg := config.AuthConfig{
		AuthMethod:  config.AwsEc2PKCS7,
		AapArn:      "arn:test:aap",
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
	}

	m, err := NewMAuth(cfg, nil)
	assert.Nil(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, config.AwsEc2PKCS7, m.config.AuthMethod)
}
