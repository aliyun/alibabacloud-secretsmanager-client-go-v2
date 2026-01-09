package utils

import (
	"encoding/json"
	"fmt"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/models"
	"github.com/aliyun/credentials-go/credentials"
	"strconv"
)

var (
	CheckParamIllegalMessage = "%s credentials param[%s] is illegal"
)

func CredentialsWithAccessKey(accessKeyId, accessKeySecret string) (credentials.Credential, error) {
	config := new(credentials.Config).
		SetType("access_key").
		SetAccessKeyId(accessKeyId).
		SetAccessKeySecret(accessKeySecret)
	return credentials.NewCredential(config)
}

func CredentialsWithOIDCRoleArn(roleArn, oidcProviderArn, oidcTokenFilePath, roleSessionName, policy, stsEndpoint string, durationSeconds int) (credentials.Credential, error) {
	config := new(credentials.Config).
		SetType("oidc_role_arn").
		SetRoleArn(roleArn).
		SetOIDCTokenFilePath(oidcTokenFilePath).
		SetOIDCProviderArn(oidcProviderArn).
		SetRoleSessionName(roleSessionName).
		SetPolicy(policy).
		SetSTSEndpoint(stsEndpoint).
		SetRoleSessionExpiration(durationSeconds)
	return credentials.NewCredential(config)
}
func CredentialsWithSimpleOIDCRoleArn(roleArn, oidcProviderArn, oidcTokenFilePath string) (credentials.Credential, error) {
	config := new(credentials.Config).
		SetType("oidc_role_arn").
		SetRoleArn(roleArn).
		SetOIDCTokenFilePath(oidcTokenFilePath).
		SetOIDCProviderArn(oidcProviderArn)
	return credentials.NewCredential(config)
}
func CredentialsWithEcsRamRole(roleName string) (credentials.Credential, error) {
	config := new(credentials.Config).
		// 凭证类型。
		SetType("ecs_ram_role").
		SetRoleName(roleName)
	return credentials.NewCredential(config)
}

// InitKmsRegions 初始化KMS区域信息
//
// @param properties 属性配置
// @param sourceType 来源类型
// @return 区域信息列表
func InitKmsRegions(properties map[string]string, sourceType string) ([]*models.RegionInfo, error) {
	var regionInfoList []*models.RegionInfo
	regionIds, exists := properties[VariableCacheClientRegionIdKey]
	if !exists || regionIds == "" {
		return regionInfoList, nil
	}

	// 尝试解析JSON格式的区域信息
	var list []map[string]interface{}
	err := json.Unmarshal([]byte(regionIds), &list)
	if err != nil {
		return nil, fmt.Errorf(CheckParamIllegalMessage, sourceType, VariableCacheClientRegionIdKey)
	}

	// 遍历并构建区域信息对象
	for _, regionInfoMap := range list {
		regionInfo := &models.RegionInfo{}

		regionId, err := ParseString(regionInfoMap[VariableRegionRegionIdNameKey])
		if err != nil {
			return nil, err
		}
		regionInfo.RegionId = regionId
		endpoint, err := ParseString(regionInfoMap[VariableRegionEndpointNameKey])
		if err != nil {
			return nil, err
		}
		regionInfo.Endpoint = endpoint
		var vpc bool
		if regionInfoMap[VariableRegionVpcNameKey] == "" {
			vpc = false
		} else {
			vpc, err = ParseBool(regionInfoMap[VariableRegionVpcNameKey])
			if err != nil {
				return nil, err
			}
		}
		regionInfo.Vpc = vpc
		caFilePath, err := ParseString(regionInfoMap[VariableRegionCaFilePathNameKey])
		if err != nil {
			return nil, err
		}
		regionInfo.CaFilePath = caFilePath
		regionInfoList = append(regionInfoList, regionInfo)
	}

	return regionInfoList, nil
}

// InitCredential 初始化凭据提供程序
//
// @param map 凭据配置映射
// @param sourceTypeName 来源类型名称
// @return 凭据提供程序
func InitCredential(configMap map[string]string, sourceTypeName string) (credentials.Credential, error) {
	credentialsType, exists := configMap[VariableCredentialsTypeKey]
	if !exists || credentialsType == "" {
		return nil, nil
	}

	switch credentialsType {
	case "ak":
		accessKeyId, exists := configMap[VariableCredentialsAccessKeyIdKey]
		if !exists || accessKeyId == "" {
			return nil, fmt.Errorf(CheckParamErrorMessage, sourceTypeName, VariableCredentialsAccessKeyIdKey)
		}

		accessSecret, exists := configMap[VariableCredentialsAccessSecretKey]
		if !exists || accessSecret == "" {
			return nil, fmt.Errorf(CheckParamErrorMessage, sourceTypeName, VariableCredentialsAccessSecretKey)
		}
		return CredentialsWithAccessKey(accessKeyId, accessSecret)

	case "ecs_ram_role":
		roleName, exists := configMap[VariableCredentialsRoleNameKey]
		if !exists || roleName == "" {
			return nil, fmt.Errorf(CheckParamErrorMessage, sourceTypeName, VariableCredentialsRoleNameKey)
		}
		return CredentialsWithEcsRamRole(roleName)

	case "oidc_role_arn":
		var roleSessionExpiration int
		if durationStr, exists := configMap[VariableCredentialsOidcDurationSecondsKey]; exists && durationStr != "" {
			if duration, err := strconv.Atoi(durationStr); err == nil {
				roleSessionExpiration = duration
			}
		}
		roleSessionName := configMap[VariableCredentialsOidcRoleSessionNameKey]
		roleArn := configMap[VariableCredentialsOidcRoleArnKey]
		oidcProviderArn := configMap[VariableCredentialsOidcProviderArnKey]
		oidcTokenFilePath := configMap[VariableCredentialsOidcTokenFilePathKey]
		policy := configMap[VariableCredentialsOidcPolicyKey]
		stsEndpoint := configMap[VariableCredentialsOidcStsEndpointKey]
		return CredentialsWithOIDCRoleArn(roleArn, oidcProviderArn, oidcTokenFilePath, roleSessionName, policy, stsEndpoint, roleSessionExpiration)
	default:
		return nil, fmt.Errorf("%s credentials type[%s] is illegal", sourceTypeName, credentialsType)
	}
}
