package service

import (
	"context"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	kms20160120 "github.com/alibabacloud-go/kms-20160120/v3/client"
	teautil "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/models"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/utils"
	"strings"
)

// secretManagerClientWithBearer 负责 bearer 签名方式的单次凭据获取
type secretManagerClientWithBearer struct {
	mauth *mauth.MAuth
}

func newSecretManagerClientWithBearer(mauth *mauth.MAuth) *secretManagerClientWithBearer {
	return &secretManagerClientWithBearer{mauth: mauth}
}

func (sm *secretManagerClientWithBearer) getSecretValue(regionInfo *models.RegionInfo, req *kms20160120.GetSecretValueRequest) (*kms20160120.GetSecretValueResponse, error) {
	resp, err := sm.doGetSecretValueWithBearer(regionInfo, req)
	if err != nil {
		if sdkErr, ok := err.(*tea.SDKError); ok && sdkErr.Code != nil && *sdkErr.Code == utils.ErrorCodeInvalidAccessKeyIdNotFound {
			clearTokenErr := sm.mauth.ClearToken(context.Background())
			if clearTokenErr != nil {
				return nil, fmt.Errorf("encountered InvalidAccessKeyId.NotFound[%v] and clear token error[%v]", sdkErr, clearTokenErr)
			}
			resp, err = sm.doGetSecretValueWithBearer(regionInfo, req)
		}
	}
	return resp, err
}

func (sm *secretManagerClientWithBearer) doGetSecretValueWithBearer(regionInfo *models.RegionInfo, req *kms20160120.GetSecretValueRequest) (*kms20160120.GetSecretValueResponse, error) {
	// 1. 获取认证令牌
	token, err := sm.mauth.GetToken(context.Background())
	if err != nil || token == nil {
		return nil, fmt.Errorf("get token error: %w", err)
	}

	// 2. 构建签名参数，进行签名
	queryParams, header := utils.ExtractSignParams(req, token.TokenKeyId, utils.ActionGetSecretValue)
	signature, err := utils.Sign(queryParams, header, token.TokenValue)
	if err != nil {
		return nil, fmt.Errorf("sign error: %w", err)
	}

	// 3. 复制配置并添加签名信息
	config := &openapi.Config{}
	endpoint := resolveEndpoint(regionInfo)
	config.SetEndpoint(endpoint)
	if strings.HasSuffix(endpoint, utils.InstanceGatewayDomainSuffix) {
		ca, err := resolveCa(regionInfo)
		if err != nil {
			return nil, err
		}
		config.SetCa(ca)
	}
	config.SetProtocol(utils.DefaultProtocol)
	config.SetBearerToken(signature)

	kmsClient, err := kms20160120.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create sm client: %v", err)
	}

	// 把那些header 加到client 里去
	if kmsClient.Headers == nil {
		kmsClient.Headers = map[string]*string{}
	}
	for key, value := range header {
		kmsClient.Headers[key] = tea.String(value)
	}

	runtime := &teautil.RuntimeOptions{}
	return kmsClient.GetSecretValueWithOptions(req, runtime)
}
