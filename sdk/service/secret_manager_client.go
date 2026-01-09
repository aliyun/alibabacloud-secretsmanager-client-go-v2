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
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/models"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/utils"
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
}

// defaultSecretManagerClient 是默认的SecretManager客户端实现
// 实现了SecretManagerClient接口的所有方法
type defaultSecretManagerClient struct {
	*DefaultSecretManagerClientBuilder
	clientMap map[*models.RegionInfo]*kms20160120.Client // KMS客户端映射
	clientMtx sync.Mutex                                 // 客户端访问互斥锁
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
	// 注意>go1.8才有sort.Slice
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
	if dmc.regionInfos != nil && len(dmc.regionInfos) > 1 {
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
	return nil
}

func (dmc *defaultSecretManagerClient) getSecretValue(regionInfo *models.RegionInfo, req *kms20160120.GetSecretValueRequest) (*kms20160120.GetSecretValueResponse, error) {
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
		if regionInfo.Endpoint != "" {
			config.SetEndpoint(regionInfo.Endpoint)
			if strings.HasSuffix(regionInfo.Endpoint, utils.InstanceGatewayDomainSuffix) {
				if regionInfo.CaFilePath != "" {
					content, err := ioutil.ReadFile(regionInfo.CaFilePath)
					if err != nil {
						return nil, err
					}
					config.SetCa(string(content))
				} else {
					caContent, exists := utils.RegionIdAndCaMap[regionInfo.RegionId]
					if !exists {
						return nil, fmt.Errorf("cannot find the built-in ca certificate for region[%s], please provide the caFilePath parameter", regionInfo.RegionId)
					}
					config.SetCa(caContent)
				}
			}
		} else if regionInfo.Vpc {
			config.SetEndpoint(utils.GetVpcEndpoint(regionInfo.RegionId))
		} else {
			config.SetEndpoint(utils.GetEndpoint(regionInfo.RegionId))
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
