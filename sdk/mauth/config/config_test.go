package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthConfig_Validate_ClientKey_WithPassword(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpFile.Name())

	cfg := &AuthConfig{
		AuthMethod:          ClientKey,
		ClientKeyConfigPath: tmpFile.Name(),
		ClientKeyPassword:   "test_password",
		KmsEndpoint:         "https://kms.cn-hangzhou.aliyuncs.com",
	}

	err = cfg.Validate()
	assert.Nil(t, err)
}

func TestAuthConfig_Validate_ClientKey_WithPasswordPath(t *testing.T) {
	tmpConfigFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpConfigFile.Name())

	tmpPwdFile, err := ioutil.TempFile("", "client_key_password_*.txt")
	assert.Nil(t, err)
	defer os.Remove(tmpPwdFile.Name())

	cfg := &AuthConfig{
		AuthMethod:            ClientKey,
		ClientKeyConfigPath:   tmpConfigFile.Name(),
		ClientKeyPasswordPath: tmpPwdFile.Name(),
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
	}

	err = cfg.Validate()
	assert.Nil(t, err)
}

func TestAuthConfig_Validate_ClientKey_MissingPassword(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpFile.Name())

	cfg := &AuthConfig{
		AuthMethod:          ClientKey,
		ClientKeyConfigPath: tmpFile.Name(),
		KmsEndpoint:         "https://kms.cn-hangzhou.aliyuncs.com",
	}

	err = cfg.Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "ClientKeyPassword or ClientKeyPasswordPath is required")
}

func TestAuthConfig_Validate_ClientKey_ConfigFileNotExist(t *testing.T) {
	cfg := &AuthConfig{
		AuthMethod:          ClientKey,
		ClientKeyConfigPath: "/not/exist/config.json",
		ClientKeyPassword:   "test_password",
		KmsEndpoint:         "https://kms.cn-hangzhou.aliyuncs.com",
	}

	err := cfg.Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "client key config file does not exist")
}

func TestAuthConfig_Validate_ClientKey_PasswordFileNotExist(t *testing.T) {
	tmpConfigFile, err := ioutil.TempFile("", "client_key_config_*.json")
	assert.Nil(t, err)
	defer os.Remove(tmpConfigFile.Name())

	cfg := &AuthConfig{
		AuthMethod:            ClientKey,
		ClientKeyConfigPath:   tmpConfigFile.Name(),
		ClientKeyPasswordPath: "/not/exist/password.txt",
		KmsEndpoint:           "https://kms.cn-hangzhou.aliyuncs.com",
	}

	err = cfg.Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "client key password file does not exist")
}

func TestAuthConfig_Validate_ACKOidcJwt_Success(t *testing.T) {
	tmpTokenFile, err := ioutil.TempFile("", "token_*.txt")
	assert.Nil(t, err)
	defer os.Remove(tmpTokenFile.Name())

	cfg := &AuthConfig{
		AuthMethod:  ACKOidcJwt,
		AapArn:      "arn:test:aap",
		TokenPath:   tmpTokenFile.Name(),
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
	}

	err = cfg.Validate()
	assert.Nil(t, err)
}

func TestAuthConfig_Validate_ACKOidcJwt_TokenFileNotExist(t *testing.T) {
	cfg := &AuthConfig{
		AuthMethod:  ACKOidcJwt,
		AapArn:      "arn:test:aap",
		TokenPath:   "/not/exist/token.txt",
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
	}

	err := cfg.Validate()
	assert.Nil(t, err)
}

func TestAuthConfig_Validate_ECSInstanceIdentity_Success(t *testing.T) {
	cfg := &AuthConfig{
		AuthMethod:  ECSInstanceIdentity,
		AapArn:      "arn:test:aap",
		KmsEndpoint: "https://kms.cn-hangzhou.aliyuncs.com",
	}

	err := cfg.Validate()
	assert.Nil(t, err)
}

func TestAuthConfig_Validate_EmptyAuthMethod(t *testing.T) {
	cfg := &AuthConfig{
		AuthMethod: "",
	}

	err := cfg.Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "AuthMethod is required")
}
