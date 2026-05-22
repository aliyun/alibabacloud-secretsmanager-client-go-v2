package utils

import (
	"os"
	"testing"

	mauthconfig "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/config"
	"github.com/stretchr/testify/assert"
)

func TestInitMAuthConfig_ClientKey_WithEnvPassword(t *testing.T) {
	envName := "TEST_CLIENT_KEY_PASSWORD_ENV"
	os.Setenv(envName, "test_password_from_env")
	defer os.Unsetenv(envName)

	configMap := map[string]string{
		VariableCredentialsTypeKey:                     VariableCredentialsTypeClientKey,
		VariableCredentialsClientKeyConfigPathKey:      "/path/to/client_key.json",
		VariableCredentialsClientKeyPasswordEnvNameKey: envName,
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, mauthconfig.ClientKey, cfg.AuthMethod)
	assert.Equal(t, "/path/to/client_key.json", cfg.ClientKeyConfigPath)
	assert.Equal(t, "test_password_from_env", cfg.ClientKeyPassword)
	assert.Equal(t, "", cfg.ClientKeyPasswordPath)
}

func TestInitMAuthConfig_ClientKey_WithEnvPassword_NotSet(t *testing.T) {
	os.Unsetenv("TEST_CLIENT_KEY_PASSWORD_NOT_SET")

	configMap := map[string]string{
		VariableCredentialsTypeKey:                     VariableCredentialsTypeClientKey,
		VariableCredentialsClientKeyConfigPathKey:      "/path/to/client_key.json",
		VariableCredentialsClientKeyPasswordEnvNameKey: "TEST_CLIENT_KEY_PASSWORD_NOT_SET",
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), VariableCredentialsClientKeyPasswordEnvNameKey)
	assert.Contains(t, err.Error(), VariableCredentialsClientKeyPasswordPathKey)
}

func TestInitMAuthConfig_ClientKey_WithPasswordPath(t *testing.T) {
	configMap := map[string]string{
		VariableCredentialsTypeKey:                  VariableCredentialsTypeClientKey,
		VariableCredentialsClientKeyConfigPathKey:   "/path/to/client_key.json",
		VariableCredentialsClientKeyPasswordPathKey: "/path/to/password",
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "/path/to/password", cfg.ClientKeyPasswordPath)
	assert.Equal(t, "", cfg.ClientKeyPassword)
}

func TestInitMAuthConfig_ClientKey_MissingPassword(t *testing.T) {
	configMap := map[string]string{
		VariableCredentialsTypeKey:                VariableCredentialsTypeClientKey,
		VariableCredentialsClientKeyConfigPathKey: "/path/to/client_key.json",
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), VariableCredentialsClientKeyPasswordEnvNameKey)
	assert.Contains(t, err.Error(), VariableCredentialsClientKeyPasswordPathKey)
}

func TestInitMAuthConfig_ClientKey_MissingConfigPath(t *testing.T) {
	configMap := map[string]string{
		VariableCredentialsTypeKey:                     VariableCredentialsTypeClientKey,
		VariableCredentialsClientKeyPasswordEnvNameKey: "TEST_PWD",
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), VariableCredentialsClientKeyConfigPathKey)
}

func TestInitMAuthConfig_ACKOidcJwt_Success(t *testing.T) {
	configMap := map[string]string{
		VariableCredentialsTypeKey:      VariableCredentialsTypeAckOidcJwt,
		VariableCredentialsAapArnKey:    "arn:test:aap",
		VariableCredentialsTokenPathKey: "/path/to/token",
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, mauthconfig.ACKOidcJwt, cfg.AuthMethod)
	assert.Equal(t, "arn:test:aap", cfg.AapArn)
	assert.Equal(t, "/path/to/token", cfg.TokenPath)
}

func TestInitMAuthConfig_ACKOidcJwt_MissingAapArn(t *testing.T) {
	configMap := map[string]string{
		VariableCredentialsTypeKey:      VariableCredentialsTypeAckOidcJwt,
		VariableCredentialsTokenPathKey: "/path/to/token",
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), VariableCredentialsAapArnKey)
}

func TestInitMAuthConfig_ACKOidcJwt_MissingTokenPath(t *testing.T) {
	configMap := map[string]string{
		VariableCredentialsTypeKey:   VariableCredentialsTypeAckOidcJwt,
		VariableCredentialsAapArnKey: "arn:test:aap",
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), VariableCredentialsTokenPathKey)
}

func TestInitMAuthConfig_ECSInstanceIdentity_Success(t *testing.T) {
	configMap := map[string]string{
		VariableCredentialsTypeKey:   VariableCredentialsTypeEcsInstanceIdentity,
		VariableCredentialsAapArnKey: "arn:test:aap",
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.Nil(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, mauthconfig.ECSInstanceIdentity, cfg.AuthMethod)
	assert.Equal(t, "arn:test:aap", cfg.AapArn)
}

func TestInitMAuthConfig_ECSInstanceIdentity_MissingAapArn(t *testing.T) {
	configMap := map[string]string{
		VariableCredentialsTypeKey: VariableCredentialsTypeEcsInstanceIdentity,
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), VariableCredentialsAapArnKey)
}

func TestInitMAuthConfig_EmptyCredentialsType(t *testing.T) {
	configMap := map[string]string{}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.Nil(t, err)
	assert.Nil(t, cfg)
}

func TestInitMAuthConfig_UnsupportedCredentialsType(t *testing.T) {
	configMap := map[string]string{
		VariableCredentialsTypeKey: "unsupported_type",
	}

	cfg, err := InitMAuthConfig(configMap, SourceTypeConfig)
	assert.Nil(t, err)
	assert.Nil(t, cfg)
}
