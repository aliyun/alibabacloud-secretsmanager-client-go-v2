package service

import (
	"errors"
	"fmt"
	"github.com/aliyun/credentials-go/credentials"
	"io/ioutil"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	kms20160120 "github.com/alibabacloud-go/kms-20160120/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/logger"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth"
	mauthconfig "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/config"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/models"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/utils"
	idaasconfig "github.com/cloud-idaas/idaas-go-core-sdk/config"
)

// SecretManagerClient 是阿里云凭据管理服务客户端接口
// 提供初始化、获取凭据值和关闭连接的功能
type SecretManagerClient interface {
	// Init 初始化Client
	Init() error

	// GetSecretValue 获取指定凭据信息
	GetSecretValue(req *kms20160120.GetSecretValueRequest) (*kms20160120.GetSecretValueResponse, error)

	// Close 关闭Client
	Close() error
}

// BaseSecretManagerClientBuilder 是基础的SecretManager客户端构建器结构体
// 用于创建标准的SecretManager客户端构建器实例
type BaseSecretManagerClientBuilder struct {
}

// DefaultSecretManagerClientBuilder 是默认的SecretManager客户端构建器
// 包含构建SecretManager客户端所需的各种配置参数
type DefaultSecretManagerClientBuilder struct {
	BaseSecretManagerClientBuilder
	regionInfos      []*models.RegionInfo                       // 地域信息列表
	credential       credentials.Credential                     // 认证凭证
	backoffStrategy  BackoffStrategy                            // 退避策略
	configMap        map[*models.RegionInfo]*openapiutil.Config // 地域配置映射
	customConfigFile string                                     // 自定义配置文件路径
	mauthConfig      *mauthconfig.AuthConfig                    // mauth认证配置
}

// defaultSecretManagerClient 是默认的SecretManager客户端实现
// 实现了SecretManagerClient接口的所有方法
type defaultSecretManagerClient struct {
	*DefaultSecretManagerClientBuilder
	clientMap     map[*models.RegionInfo]*kms20160120.Client // KMS客户端映射
	clientMtx     sync.Mutex                                 // 客户端访问互斥锁
	bearerClients map[string]*secretManagerClientWithBearer  // 以regionId为key的bearer认证客户端
}

func NewBaseSecretManagerClientBuilder() *BaseSecretManagerClientBuilder {
	return &BaseSecretManagerClientBuilder{}
}

func NewDefaultSecretManagerClientBuilder() *DefaultSecretManagerClientBuilder {
	return &DefaultSecretManagerClientBuilder{
		configMap: make(map[*models.RegionInfo]*openapiutil.Config),
	}
}

func (base *BaseSecretManagerClientBuilder) Standard() *DefaultSecretManagerClientBuilder {
	return NewDefaultSecretManagerClientBuilder()
}

// WithAccessKey 使用AccessKey配置认证信息
// 参数accessKeyId和accessKeySecret分别对应阿里云的AccessKey ID和Secret
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithAccessKey(accessKeyId, accessKeySecret string) *DefaultSecretManagerClientBuilder {
	dsb.credential, _ = utils.CredentialsWithAccessKey(accessKeyId, accessKeySecret)
	return dsb
}

// WithCredential 设置自定义认证凭证
// 参数credential是一个实现了阿里云凭证接口的对象
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithCredential(credential credentials.Credential) *DefaultSecretManagerClientBuilder {
	dsb.credential = credential
	return dsb
}

// WithRegion 指定多个调用地域Id
// WithRegion 指定多个调用地域Id
// 参数regionIds是可变长度的地域ID字符串数组
// 为每个地域ID创建RegionInfo对象并添加到构建器中
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithRegion(regionIds ...string) *DefaultSecretManagerClientBuilder {
	for _, regionId := range regionIds {
		dsb.AddRegionInfo(&models.RegionInfo{RegionId: regionId})
	}
	return dsb
}

// AddRegionInfo 指定调用地域信息
// AddRegionInfo 添加地域信息
// 参数regionInfo是要添加的地域信息对象
// 将指定的地域信息添加到构建器的地域信息列表中
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) AddRegionInfo(regionInfo *models.RegionInfo) *DefaultSecretManagerClientBuilder {
	dsb.regionInfos = append(dsb.regionInfos, regionInfo)
	return dsb
}

// WithBackoffStrategy 设置退避策略
// 参数backoffStrategy是实现了BackoffStrategy接口的对象
// 用于控制请求重试的时间间隔策略
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithBackoffStrategy(backoffStrategy BackoffStrategy) *DefaultSecretManagerClientBuilder {
	dsb.backoffStrategy = backoffStrategy
	return dsb
}

// AddConfig 添加地域配置
// 参数config是OpenAPI的配置对象，包含地域ID和终端节点等信息
// 根据配置创建对应的RegionInfo对象，并将其添加到地域信息列表中
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) AddConfig(config *openapiutil.Config) *DefaultSecretManagerClientBuilder {
	regionInfo := &models.RegionInfo{
		RegionId: tea.StringValue(config.RegionId),
		Endpoint: tea.StringValue(config.Endpoint),
	}
	dsb.configMap[regionInfo] = config
	dsb.AddRegionInfo(regionInfo)
	return dsb
}

// WithCustomConfigFile 设置自定义配置文件路径
// 参数customConfigFile是配置文件的路径
// 允许使用自定义的配置文件来初始化客户端
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithCustomConfigFile(customConfigFile string) *DefaultSecretManagerClientBuilder {
	dsb.customConfigFile = customConfigFile
	return dsb
}

// WithClientKey 使用ClientKey配置认证信息
// 参数clientKeyConfigPath为客户端密钥配置文件路径
// 参数clientKeyPasswordPath为客户端密钥密码文件路径
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithClientKey(clientKeyConfigPath, clientKeyPasswordPath string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:            mauthconfig.ClientKey,
		ClientKeyConfigPath:   clientKeyConfigPath,
		ClientKeyPasswordPath: clientKeyPasswordPath,
	}
	return dsb
}

// WithACKOidcJwt 使用ACK OIDC JWT配置认证信息
// 参数aapArn为应用访问点ARN
// 参数tokenPath为OIDC令牌文件路径
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithACKOidcJwt(aapArn, tokenPath string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod: mauthconfig.ACKOidcJwt,
		AapArn:     aapArn,
		TokenPath:  tokenPath,
	}
	return dsb
}

// WithECSInstanceIdentity 使用ECS实例身份配置认证信息
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithECSInstanceIdentity(aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod: mauthconfig.ECSInstanceIdentity,
		AapArn:     aapArn,
	}
	return dsb
}

// WithAwsEc2PKCS7 使用AWS EC2 PKCS7配置认证信息（对象版本）
// 参数clientConfig为IDaaS客户端配置对象
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithAwsEc2PKCS7(clientConfig *idaasconfig.IDaaSClientConfig, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:        mauthconfig.AwsEc2PKCS7,
		AapArn:            aapArn,
		IDaaSClientConfig: clientConfig,
	}
	return dsb
}

// WithAwsEc2PKCS7Path 使用AWS EC2 PKCS7配置认证信息（配置文件路径版本）
// 参数configPath为IDaaS配置文件路径；为空时走IDaaS原生默认读取逻辑
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithAwsEc2PKCS7Path(configPath, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:      mauthconfig.AwsEc2PKCS7,
		AapArn:          aapArn,
		IDaaSConfigPath: configPath,
	}
	return dsb
}

// WithAwsEksOIDC 使用AWS EKS OIDC配置认证信息（对象版本）
// 参数clientConfig为IDaaS客户端配置对象
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithAwsEksOIDC(clientConfig *idaasconfig.IDaaSClientConfig, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:        mauthconfig.AwsEksOIDC,
		AapArn:            aapArn,
		IDaaSClientConfig: clientConfig,
	}
	return dsb
}

// WithAwsEksOIDCPath 使用AWS EKS OIDC配置认证信息（配置文件路径版本）
// 参数configPath为IDaaS配置文件路径；为空时走IDaaS原生默认读取逻辑
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithAwsEksOIDCPath(configPath, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:      mauthconfig.AwsEksOIDC,
		AapArn:          aapArn,
		IDaaSConfigPath: configPath,
	}
	return dsb
}

// WithGcpVmOIDC 使用GCP VM OIDC配置认证信息（对象版本）
// 参数clientConfig为IDaaS客户端配置对象
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithGcpVmOIDC(clientConfig *idaasconfig.IDaaSClientConfig, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:        mauthconfig.GcpVmOIDC,
		AapArn:            aapArn,
		IDaaSClientConfig: clientConfig,
	}
	return dsb
}

// WithGcpVmOIDCPath 使用GCP VM OIDC配置认证信息（配置文件路径版本）
// 参数configPath为IDaaS配置文件路径；为空时走IDaaS原生默认读取逻辑
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithGcpVmOIDCPath(configPath, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:      mauthconfig.GcpVmOIDC,
		AapArn:          aapArn,
		IDaaSConfigPath: configPath,
	}
	return dsb
}

// WithGcpGkeOIDC 使用GCP GKE OIDC配置认证信息（对象版本）
// 参数clientConfig为IDaaS客户端配置对象
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithGcpGkeOIDC(clientConfig *idaasconfig.IDaaSClientConfig, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:        mauthconfig.GcpGkeOIDC,
		AapArn:            aapArn,
		IDaaSClientConfig: clientConfig,
	}
	return dsb
}

// WithGcpGkeOIDCPath 使用GCP GKE OIDC配置认证信息（配置文件路径版本）
// 参数configPath为IDaaS配置文件路径；为空时走IDaaS原生默认读取逻辑
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithGcpGkeOIDCPath(configPath, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:      mauthconfig.GcpGkeOIDC,
		AapArn:          aapArn,
		IDaaSConfigPath: configPath,
	}
	return dsb
}

// WithAzureVmOIDC 使用Azure VM OIDC配置认证信息（对象版本）
// 参数clientConfig为IDaaS客户端配置对象
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithAzureVmOIDC(clientConfig *idaasconfig.IDaaSClientConfig, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:        mauthconfig.AzureVmOIDC,
		AapArn:            aapArn,
		IDaaSClientConfig: clientConfig,
	}
	return dsb
}

// WithAzureVmOIDCPath 使用Azure VM OIDC配置认证信息（配置文件路径版本）
// 参数configPath为IDaaS配置文件路径；为空时走IDaaS原生默认读取逻辑
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithAzureVmOIDCPath(configPath, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:      mauthconfig.AzureVmOIDC,
		AapArn:          aapArn,
		IDaaSConfigPath: configPath,
	}
	return dsb
}

// WithAzureAksOIDC 使用Azure AKS OIDC配置认证信息（对象版本）
// 参数clientConfig为IDaaS客户端配置对象
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithAzureAksOIDC(clientConfig *idaasconfig.IDaaSClientConfig, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:        mauthconfig.AzureAksOIDC,
		AapArn:            aapArn,
		IDaaSClientConfig: clientConfig,
	}
	return dsb
}

// WithAzureAksOIDCPath 使用Azure AKS OIDC配置认证信息（配置文件路径版本）
// 参数configPath为IDaaS配置文件路径；为空时走IDaaS原生默认读取逻辑
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithAzureAksOIDCPath(configPath, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:      mauthconfig.AzureAksOIDC,
		AapArn:          aapArn,
		IDaaSConfigPath: configPath,
	}
	return dsb
}

// WithGenericKubernetesOIDC 使用Generic Kubernetes OIDC配置认证信息（对象版本）
// 参数clientConfig为IDaaS客户端配置对象
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithGenericKubernetesOIDC(clientConfig *idaasconfig.IDaaSClientConfig, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:        mauthconfig.GenericKubernetesOIDC,
		AapArn:            aapArn,
		IDaaSClientConfig: clientConfig,
	}
	return dsb
}

// WithGenericKubernetesOIDCPath 使用Generic Kubernetes OIDC配置认证信息（配置文件路径版本）
// 参数configPath为IDaaS配置文件路径；为空时走IDaaS原生默认读取逻辑
// 参数aapArn为应用访问点ARN
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) WithGenericKubernetesOIDCPath(configPath, aapArn string) *DefaultSecretManagerClientBuilder {
	dsb.mauthConfig = &mauthconfig.AuthConfig{
		AuthMethod:      mauthconfig.GenericKubernetesOIDC,
		AapArn:          aapArn,
		IDaaSConfigPath: configPath,
	}
	return dsb
}

// Build 构建SecretManager客户端
// 根据已设置的配置参数创建并返回SecretManagerClient实例
// 返回实现SecretManagerClient接口的对象
func (dsb *DefaultSecretManagerClientBuilder) Build() SecretManagerClient {
	return &defaultSecretManagerClient{
		DefaultSecretManagerClientBuilder: dsb,
		clientMap:                         make(map[*models.RegionInfo]*kms20160120.Client),
	}
}

// AddRegion 指定调用地域Id
// AddRegion 指定调用地域Id
// 参数regionId是要添加的地域ID
// 创建RegionInfo对象并添加到构建器中
// 返回构建器本身以支持链式调用
func (dsb *DefaultSecretManagerClientBuilder) AddRegion(regionId string) *DefaultSecretManagerClientBuilder {
	return dsb.AddRegionInfo(&models.RegionInfo{RegionId: regionId})
}

func (dsb *DefaultSecretManagerClientBuilder) sortRegionInfos(regionInfos []*models.RegionInfo) []*models.RegionInfo {
	var regionInfoResp []*models.RegionInfo
	var regionInfoExtends []*models.RegionInfoExtend
	var wg sync.WaitGroup
	for _, regionInfo := range regionInfos {
		wg.Add(1)
		regionInfo := regionInfo
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			var pingDelay float64
			regionInfoExtend := &models.RegionInfoExtend{
				RegionInfo: regionInfo,
			}
			if regionInfo.Endpoint != "" {
				pingDelay = utils.Ping(regionInfo.Endpoint)
			} else if regionInfo.Vpc {
				pingDelay = utils.Ping(utils.GetVpcEndpoint(regionInfo.RegionId))
			} else {
				pingDelay = utils.Ping(utils.GetEndpoint(regionInfo.RegionId))
			}
			if pingDelay >= 0 {
				regionInfoExtend.Elapsed = pingDelay
			} else {
				regionInfoExtend.Elapsed = math.MaxFloat64
			}
			regionInfoExtend.Reachable = pingDelay >= 0
			regionInfoExtends = append(regionInfoExtends, regionInfoExtend)
		}(&wg)
	}
	wg.Wait()
	sort.Slice(regionInfoExtends, func(i, j int) bool {
		return regionInfoExtends[i].Elapsed < regionInfoExtends[j].Elapsed
	})
	for _, regionInfoExtend := range regionInfoExtends {
		regionInfoResp = append(regionInfoResp, regionInfoExtend.RegionInfo)
	}
	return regionInfoResp
}

func (dmc *defaultSecretManagerClient) Init() error {
	err := dmc.initFromConfigFile()
	if err != nil {
		return err
	}
	err = dmc.initFromEnv()
	if err != nil {
		return err
	}
	if len(dmc.regionInfos) == 0 {
		return errors.New("the param[regionInfo] is needed")
	}
	UserAgentManager.RegisterUserAgent(utils.UserAgentOfSecretsManagerV2Go, 0, utils.ProjectVersion)
	if dmc.backoffStrategy == nil {
		dmc.backoffStrategy = &FullJitterBackoffStrategy{}
	}
	err = dmc.backoffStrategy.Init()
	if err != nil {
		return err
	}

	if dmc.mauthConfig != nil {
		dmc.bearerClients = make(map[string]*secretManagerClientWithBearer)
		for _, regionInfo := range dmc.regionInfos {
			mauthCfg := *dmc.mauthConfig
			if mauthCfg.KmsEndpoint == "" {
				endpoint, ca, err := resolveEndpointAndCa(regionInfo)
				if err != nil {
					return fmt.Errorf("failed to resolve endpoint and ca for region[%s]: %w", regionInfo.RegionId, err)
				}
				mauthCfg.KmsEndpoint = endpoint
				if mauthCfg.Ca == "" {
					mauthCfg.Ca = ca
				}
			}
			mauthObj, err := mauth.NewMAuth(mauthCfg, logger.GetCommonLogger(utils.ModeName))
			if err != nil {
				return fmt.Errorf("failed to init mauth for region[%s]: %w", regionInfo.RegionId, err)
			}
			dmc.bearerClients[regionInfo.RegionId] = newSecretManagerClientWithBearer(mauthObj)
		}
		return nil
	}
	if len(dmc.regionInfos) > 1 {
		dmc.regionInfos = dmc.sortRegionInfos(dmc.regionInfos)
	}
	for _, regionInfo := range dmc.regionInfos {
		_, err := dmc.getClient(regionInfo)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dmc *defaultSecretManagerClient) GetSecretValue(req *kms20160120.GetSecretValueRequest) (*kms20160120.GetSecretValueResponse, error) {
	var results []*kms20160120.GetSecretValueResponse
	var errs []error
	var wg sync.WaitGroup
	finished := int32(len(dmc.regionInfos))
	retryEnd := make(chan struct{})
	for i, regionInfo := range dmc.regionInfos {
		if i == 0 {
			resp, err := dmc.getSecretValue(regionInfo, req)
			if err == nil {
				return resp, nil
			}
			logger.GetCommonLogger(utils.ModeName).Errorf("action:getSecretValue, regionInfo:%+v, %+v", regionInfo, err)
			if !utils.JudgeNeedRecoveryException(err) {
				return nil, err
			}
			wg.Add(1)
		}
		regionInfo := regionInfo
		request := &kms20160120.GetSecretValueRequest{}
		request.SecretName = req.SecretName
		request.VersionStage = req.VersionStage
		request.FetchExtendedConfig = req.FetchExtendedConfig
		go func(wg *sync.WaitGroup, finished *int32, retryEnd <-chan struct{}) {
			if resp, err := dmc.retryGetSecretValue(request, regionInfo, retryEnd); err == nil {
				results = append(results, resp)
				wg.Done()
			} else {
				errs = append(errs, err)
				for {
					val := atomic.LoadInt32(finished)
					if atomic.CompareAndSwapInt32(finished, val, val-1) {
						break
					}
				}
				if atomic.LoadInt32(finished) == 0 {
					wg.Done()
				}
			}
		}(&wg, &finished, retryEnd)
	}
	dmc.waitTimeout(&wg, time.Duration(utils.RequestWaitingTime)*time.Millisecond)
	close(retryEnd)
	if len(results) == 0 {
		var errStr string
		for _, err := range errs {
			errStr += fmt.Sprintf("%+v;", err)
		}
		return nil, errors.New(fmt.Sprintf("action:retryGetSecretValueTask:%s", errStr))
	}
	return results[0], nil
}

func (dmc *defaultSecretManagerClient) Close() error {
	for _, bc := range dmc.bearerClients {
		if bc.mauth != nil {
			bc.mauth.Close()
		}
	}
	return nil
}

func resolveEndpoint(regionInfo *models.RegionInfo) string {
	if regionInfo.Endpoint != "" {
		return regionInfo.Endpoint
	} else if regionInfo.Vpc {
		return utils.GetVpcEndpoint(regionInfo.RegionId)
	}
	return utils.GetEndpoint(regionInfo.RegionId)
}

func resolveCa(regionInfo *models.RegionInfo) (string, error) {
	if regionInfo.CaFilePath != "" {
		content, err := ioutil.ReadFile(regionInfo.CaFilePath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	caContent, exists := utils.RegionIdAndCaMap[regionInfo.RegionId]
	if !exists {
		return "", fmt.Errorf("cannot find the built-in ca certificate for region[%s], please provide the caFilePath parameter", regionInfo.RegionId)
	}
	return caContent, nil
}

func resolveEndpointAndCa(regionInfo *models.RegionInfo) (endpoint string, ca string, err error) {
	endpoint = resolveEndpoint(regionInfo)
	if strings.HasSuffix(endpoint, utils.InstanceGatewayDomainSuffix) {
		ca, err = resolveCa(regionInfo)
		if err != nil {
			return "", "", err
		}
	}
	return endpoint, ca, nil
}

func (dmc *defaultSecretManagerClient) getSecretValue(regionInfo *models.RegionInfo, req *kms20160120.GetSecretValueRequest) (*kms20160120.GetSecretValueResponse, error) {
	if bc, ok := dmc.bearerClients[regionInfo.RegionId]; ok {
		return bc.getSecretValue(regionInfo, req)
	}
	client, err := dmc.getClient(regionInfo)
	if err != nil {
		return nil, err
	}
	return client.GetSecretValue(req)
}

func (dmc *defaultSecretManagerClient) getClient(regionInfo *models.RegionInfo) (*kms20160120.Client, error) {
	if client, ok := dmc.clientMap[regionInfo]; ok {
		return client, nil
	}
	dmc.clientMtx.Lock()
	defer dmc.clientMtx.Unlock()
	if client, ok := dmc.clientMap[regionInfo]; ok {
		return client, nil
	}
	kmsClient, err := dmc.buildKmsClient(regionInfo)
	if err != nil {
		return nil, err
	}
	dmc.clientMap[regionInfo] = kmsClient
	return dmc.clientMap[regionInfo], nil
}

func (dmc *defaultSecretManagerClient) buildKmsClient(regionInfo *models.RegionInfo) (*kms20160120.Client, error) {
	config := dmc.configMap[regionInfo]
	if config == nil {
		config = &openapiutil.Config{}
		endpoint, ca, err := resolveEndpointAndCa(regionInfo)
		if err != nil {
			return nil, err
		}
		config.SetEndpoint(endpoint)
		if ca != "" {
			config.SetCa(ca)
		}
		if dmc.credential == nil {
			credential, err := credentials.NewCredential(nil)
			if err != nil {
				return nil, err
			}
			dmc.credential = credential
		}
		config.SetCredential(dmc.credential)
		config.SetProtocol(utils.DefaultProtocol)
	}
	if config.Ca != nil && *config.Ca != "" {
		config.SetUserAgent(fmt.Sprintf("%s/%s %s_ca_expiration_utc_date/%s", UserAgentManager.GetUserAgent(), UserAgentManager.GetProjectVersion(), regionInfo.RegionId, utils.GetCaExpirationUtcDate(*config.Ca)))
	} else {
		config.SetUserAgent(fmt.Sprintf("%s/%s", UserAgentManager.GetUserAgent(), UserAgentManager.GetProjectVersion()))
	}
	return kms20160120.NewClient(config)
}

func (dmc *defaultSecretManagerClient) initFromConfigFile() error {
	credentialsProperties, err := utils.LoadCredentialsProperties(dmc.customConfigFile)
	if err != nil {
		return err
	}
	if credentialsProperties != nil {
		if credentialsProperties.Credential != nil {
			dmc.credential = credentialsProperties.Credential
		}
		dmc.regionInfos = append(dmc.regionInfos, credentialsProperties.RegionInfoSlice...)
		if err := dmc.initMAuthConfig(credentialsProperties.SourceProperties, utils.SourceTypeConfig); err != nil {
			return err
		}
	}
	return nil
}

func (dmc *defaultSecretManagerClient) initFromEnv() error {
	envMap := utils.GetAllEnvAsMap()
	credential, err := utils.InitCredential(envMap, utils.SourceTypeEnv)
	if err != nil {
		return err
	}
	if credential != nil {
		dmc.credential = credential
	}
	regionInfos, err := utils.InitKmsRegions(envMap, utils.SourceTypeEnv)
	if err != nil {
		return err
	}
	dmc.regionInfos = append(dmc.regionInfos, regionInfos...)
	if err := dmc.initMAuthConfig(envMap, utils.SourceTypeEnv); err != nil {
		return err
	}
	return nil
}

func (dmc *defaultSecretManagerClient) initMAuthConfig(properties map[string]string, sourceType string) error {
	if dmc.mauthConfig == nil {
		mauthCfg, err := utils.InitMAuthConfig(properties, sourceType)
		if err != nil {
			return err
		}
		dmc.mauthConfig = mauthCfg
	}
	return nil
}

func (dmc *defaultSecretManagerClient) retryGetSecretValue(req *kms20160120.GetSecretValueRequest, regionInfo *models.RegionInfo, retryEnd <-chan struct{}) (*kms20160120.GetSecretValueResponse, error) {
	retryTimes := 0
	for {
		select {
		case <-retryEnd:
			return nil, errors.New(fmt.Sprintf("action:retryGetSecretValue, retry end"))
		default:
			waitTimeExponential := dmc.backoffStrategy.GetWaitTimeExponential(retryTimes)
			if waitTimeExponential < 0 {
				return nil, errors.New(fmt.Sprintf("action:retryGetSecretValue, Times limit exceeded"))
			}

			time.Sleep(time.Duration(waitTimeExponential) * time.Millisecond)

			resp, err := dmc.getSecretValue(regionInfo, req)
			if err == nil {
				return resp, nil
			}
			logger.GetCommonLogger(utils.ModeName).Errorf("action:retryGetSecretValue, regionInfo:%+v, %+v", regionInfo, err)
			if !utils.JudgeNeedRecoveryException(err) {
				return nil, err
			}
			retryTimes += 1
		}
	}
}

func (dmc *defaultSecretManagerClient) waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()
	select {
	case <-done:
		return false
	case <-time.After(timeout):
		return true
	}
}
