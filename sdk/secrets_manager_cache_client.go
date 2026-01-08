package sdk

import (
	"errors"
	"fmt"
	"github.com/alibabacloud-go/tea/tea"
	"sync"
	"time"

	kms "github.com/alibabacloud-go/kms-20160120/v3/client"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/cache"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/logger"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/models"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/service"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/utils"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	// defaultTtl 默认TTL时间
	defaultTtl                 int64 = 60 * 60 * 1000
	defaultJsonTtlPropertyName       = "ttl"
)

type SecretManagerCacheClient struct {
	jsonTTLPropertyName      string
	stage                    string
	secretManagerClient      service.SecretManagerClient
	cacheSecretStoreStrategy cache.SecretCacheStoreStrategy
	refreshSecretStrategy    service.RefreshSecretStrategy
	cacheHook                cache.SecretCacheHook
	secretTTLMap             map[string]int64

	scheduledMap     cmap.ConcurrentMap
	secretNameMtx    sync.Mutex
	secretNameMtxMap map[string]*sync.Mutex
}

type runnable interface {
	getRunnable() func()
}

type refreshSecretTask struct {
	client     *SecretManagerCacheClient
	secretName string
}

func NewSecretCacheClient() *SecretManagerCacheClient {
	return &SecretManagerCacheClient{
		jsonTTLPropertyName: defaultJsonTtlPropertyName,
		stage:               utils.StageAcsCurrent,
		secretTTLMap:        make(map[string]int64),
		scheduledMap:        cmap.New(),
		secretNameMtxMap:    make(map[string]*sync.Mutex),
	}
}

func (scc *SecretManagerCacheClient) Init() error {
	if scc.secretManagerClient == nil {
		scc.secretManagerClient = service.NewDefaultSecretManagerClientBuilder().Build()
	}
	err := scc.secretManagerClient.Init()
	if err != nil {
		return err
	}
	if scc.cacheSecretStoreStrategy == nil {
		scc.cacheSecretStoreStrategy = cache.NewMemoryCacheSecretStoreStrategy()
	}
	err = scc.cacheSecretStoreStrategy.Init()
	if err != nil {
		return err
	}
	if scc.refreshSecretStrategy == nil {
		scc.refreshSecretStrategy = service.NewDefaultRefreshSecretStrategy(scc.jsonTTLPropertyName)
	}
	err = scc.refreshSecretStrategy.Init()
	if err != nil {
		return err
	}
	if scc.cacheHook == nil {
		scc.cacheHook = cache.NewDefaultSecretCacheHook(scc.stage)
	}
	err = scc.cacheHook.Init()
	if err != nil {
		return err
	}
	for secretName := range scc.secretTTLMap {
		secretInfo, err := scc.getSecretValue(secretName)
		if err != nil {
			logger.GetCommonLogger(utils.ModeName).Errorf("action:initSecretCacheClient", err)
			if scc.judgeSkipRefreshException(err) {
				return err
			}
		}
		err = scc.storeAndRefresh(secretName, secretInfo)
		if err != nil {
			return err
		}
	}
	logger.GetCommonLogger(utils.ModeName).Infof("secretCacheClient init success")
	return nil
}

// 根据凭据名称获取secretInfo信息
func (scc *SecretManagerCacheClient) GetSecretInfo(secretName string) (*models.SecretInfo, error) {
	if secretName == "" {
		return nil, errors.New(fmt.Sprintf("the argument secretName must not be empty"))
	}
	cacheSecretInfo, err := scc.cacheSecretStoreStrategy.GetCacheSecretInfo(secretName)
	if err == nil && !scc.judgeCacheExpire(cacheSecretInfo) {
		return scc.cacheHook.Get(cacheSecretInfo)
	} else {
		lck := scc.getLock(secretName)
		lck.Lock()
		defer lck.Unlock()
		cacheSecretInfo, err = scc.cacheSecretStoreStrategy.GetCacheSecretInfo(secretName)
		if err == nil && !scc.judgeCacheExpire(cacheSecretInfo) {
			return scc.cacheHook.Get(cacheSecretInfo)
		} else {
			secretInfo, err := scc.getSecretValue(secretName)
			if err != nil {
				return nil, err
			}
			err = scc.storeAndRefreshLocked(secretName, secretInfo)
			if err != nil {
				return nil, err
			}
			cacheSecretInfo, err = scc.cacheHook.Put(secretInfo)
			if err != nil {
				return nil, err
			}
			if cacheSecretInfo == nil {
				return nil, errors.New(fmt.Sprintf("cacheSecretInfo is nil"))
			}
			return cacheSecretInfo.SecretInfo, nil
		}
	}
}

// 根据凭据名称获取凭据存储值文本信息
func (scc *SecretManagerCacheClient) GetStringValue(secretName string) (string, error) {
	secretInfo, err := scc.GetSecretInfo(secretName)
	if err != nil {
		return "", err
	}
	if utils.TextDataType != secretInfo.SecretDataType {
		return "", errors.New(fmt.Sprintf("the secret named[%s] do not support text value", secretName))
	}
	return secretInfo.SecretValue, nil
}

// 根据凭据名称获取凭据存储的二进制信息
func (scc *SecretManagerCacheClient) GetBinaryValue(secretName string) ([]byte, error) {
	secretInfo, err := scc.GetSecretInfo(secretName)
	if err != nil {
		return nil, err
	}
	if utils.BinaryDataType != secretInfo.SecretDataType {
		return nil, errors.New(fmt.Sprintf("the secret named[%s] do not support binary value", secretName))
	}
	return []byte(secretInfo.SecretValue), nil
}

// 强制刷新指定的凭据名称
func (scc *SecretManagerCacheClient) RefreshNow(secretName string) (bool, error) {
	if secretName == "" {
		return false, errors.New(fmt.Sprintf("the argument[%s] must not be null", secretName))
	}
	return scc.refreshNow(secretName, nil)
}

func (scc *SecretManagerCacheClient) Close() error {
	if scc.cacheSecretStoreStrategy != nil {
		if err := scc.cacheSecretStoreStrategy.Close(); err != nil {
			logger.GetCommonLogger(utils.ModeName).Errorf("action:closeCacheSecretStoreStrategy", err)
		}
	}
	if scc.refreshSecretStrategy != nil {
		if err := scc.refreshSecretStrategy.Close(); err != nil {
			logger.GetCommonLogger(utils.ModeName).Errorf("action:closeRefreshSecretStrategy", err)
		}
	}
	if scc.secretManagerClient != nil {
		if err := scc.secretManagerClient.Close(); err != nil {
			logger.GetCommonLogger(utils.ModeName).Errorf("action:closeSecretManagerClient", err)
		}
	}
	if scc.cacheHook != nil {
		if err := scc.cacheHook.Close(); err != nil {
			logger.GetCommonLogger(utils.ModeName).Errorf("action:closeCacheHook", err)
		}
	}
	return nil
}

func (scc *SecretManagerCacheClient) judgeCacheExpire(cacheSecretInfo *models.CacheSecretInfo) bool {
	ttl := scc.refreshSecretStrategy.ParseTTL(cacheSecretInfo.SecretInfo)
	if ttl <= 0 {
		if ttl0, ok := scc.secretTTLMap[cacheSecretInfo.SecretInfo.SecretName]; !ok {
			ttl = defaultTtl
		} else {
			ttl = ttl0
		}
	}
	return (time.Now().UnixNano()/1e6)-cacheSecretInfo.RefreshTimestamp > ttl
}

func (scc *SecretManagerCacheClient) getSecretValue(secretName string) (*models.SecretInfo, error) {
	request := &kms.GetSecretValueRequest{}
	request.SetSecretName(secretName)
	request.SetVersionStage(scc.stage)
	request.SetFetchExtendedConfig(true)
	resp, err := scc.secretManagerClient.GetSecretValue(request)
	if err == nil {
		return &models.SecretInfo{
			SecretName:        tea.StringValue(resp.Body.SecretName),
			VersionId:         tea.StringValue(resp.Body.VersionId),
			SecretValue:       tea.StringValue(resp.Body.SecretData),
			SecretDataType:    tea.StringValue(resp.Body.SecretDataType),
			CreateTime:        tea.StringValue(resp.Body.CreateTime),
			SecretType:        tea.StringValue(resp.Body.SecretType),
			AutomaticRotation: tea.StringValue(resp.Body.AutomaticRotation),
			ExtendedConfig:    tea.StringValue(resp.Body.ExtendedConfig),
			RotationInterval:  tea.StringValue(resp.Body.RotationInterval),
			NextRotationDate:  tea.StringValue(resp.Body.NextRotationDate),
		}, nil
	} else {
		logger.GetCommonLogger(utils.ModeName).Errorf("action:getSecretValue", err)
		if utils.JudgeNeedRecoveryException(err) {
			secretInfo, inErr := scc.cacheHook.RecoveryGetSecret(secretName)
			if inErr != nil {
				logger.GetCommonLogger(utils.ModeName).Errorf("action:recoveryGetSecret", inErr)
				return nil, inErr
			}
			if secretInfo == nil {
				return nil, err
			}
			return secretInfo, nil
		}
	}
	return nil, err
}

func (scc *SecretManagerCacheClient) storeAndRefresh(secretName string, secretInfo *models.SecretInfo) error {
	_, err := scc.refreshNow(secretName, secretInfo)
	if err != nil {
		return err
	}
	return nil
}

func (scc *SecretManagerCacheClient) storeAndRefreshLocked(secretName string, secretInfo *models.SecretInfo) error {
	_, err := scc.refreshNowLocked(secretName, secretInfo)
	if err != nil {
		return err
	}
	return nil
}

func (scc *SecretManagerCacheClient) refresh(secretName string, secretInfo *models.SecretInfo) (err error) {
	if secretInfo == nil {
		secretInfo, err = scc.getSecretValue(secretName)
		if err != nil {
			return err
		}
	}
	cacheSecretInfo, err := scc.cacheHook.Put(secretInfo)
	if err != nil {
		return err
	}
	if cacheSecretInfo != nil {
		err = scc.cacheSecretStoreStrategy.StoreSecret(cacheSecretInfo)
		if err != nil {
			return err
		}
	}
	logger.GetCommonLogger(utils.ModeName).Infof("secretName:%s refresh success", secretName)
	return nil
}

func (scc *SecretManagerCacheClient) removeRefreshTask(secretName string) {
	if v, ok := scc.scheduledMap.Get(secretName); ok {
		if task, okk := v.(*time.Timer); okk {
			task.Stop()
			scc.scheduledMap.Remove(secretName)
		}
	}
}

func (scc *SecretManagerCacheClient) addRefreshTask(secretName string, runnable runnable) error {
	cacheSecretInfo, err := scc.cacheSecretStoreStrategy.GetCacheSecretInfo(secretName)
	if err != nil {
		return err
	}
	executeTime := scc.refreshSecretStrategy.ParseNextExecuteTime(cacheSecretInfo)
	if executeTime <= 0 {
		refreshTimestamp := cacheSecretInfo.RefreshTimestamp
		ttl := defaultTtl
		if t, ok := scc.secretTTLMap[secretName]; ok {
			ttl = t
		}
		executeTime = scc.refreshSecretStrategy.GetNextExecuteTime(secretName, ttl, refreshTimestamp)
		if executeTime < (time.Now().UnixNano() / 1e6) {
			executeTime = time.Now().UnixNano() / 1e6
		}
	}
	delay := executeTime - time.Now().UnixNano()/1e6
	if delay < 0 {
		delay = 0
	}
	schedule := time.AfterFunc(time.Duration(delay)*time.Millisecond, runnable.getRunnable())
	scc.scheduledMap.Set(secretName, schedule)
	logger.GetCommonLogger(utils.ModeName).Infof("secretName:%s addRefreshTask success", secretName)
	return nil
}

func (scc *SecretManagerCacheClient) refreshNow(secretName string, secretInfo *models.SecretInfo) (bool, error) {
	lck := scc.getLock(secretName)
	lck.Lock()
	defer lck.Unlock()
	return scc.refreshNowLocked(secretName, secretInfo)
}

func (scc *SecretManagerCacheClient) refreshNowLocked(secretName string, secretInfo *models.SecretInfo) (bool, error) {
	err := scc.refresh(secretName, secretInfo)
	if err != nil {
		return false, err
	}
	scc.removeRefreshTask(secretName)
	err = scc.addRefreshTask(secretName, &refreshSecretTask{
		secretName: secretName,
		client:     scc,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (scc *SecretManagerCacheClient) judgeServerException(err error) bool {
	return utils.JudgeNeedBackoff(err)
}

func (scc *SecretManagerCacheClient) judgeSkipRefreshException(err error) bool {
	return !scc.judgeServerException(err) && !func(err error) bool {
		var teaErr *tea.SDKError
		switch {
		case errors.As(err, &teaErr):
			errorCode := tea.StringValue(teaErr.Code)
			if utils.ErrorCodeForbiddenInDebtOverDue == errorCode || utils.ErrorCodeForbiddenInDebt == errorCode {
				return true
			}
		}
		return false
	}(err)
}

func (scc *SecretManagerCacheClient) getLock(key string) *sync.Mutex {
	scc.secretNameMtx.Lock()
	defer scc.secretNameMtx.Unlock()
	mtx, ok := scc.secretNameMtxMap[key]
	if ok {
		return mtx
	}
	scc.secretNameMtxMap[key] = &sync.Mutex{}
	return scc.secretNameMtxMap[key]
}

func (rst *refreshSecretTask) getRunnable() func() {
	return func() {
		err := rst.client.refresh(rst.secretName, nil)
		if err != nil {
			logger.GetCommonLogger(utils.ModeName).Errorf("action:refreshSecretTask", err)
		}
		rst.client.removeRefreshTask(rst.secretName)
		err = rst.client.addRefreshTask(rst.secretName, rst)
		if err != nil {
			logger.GetCommonLogger(utils.ModeName).Errorf("action:addRefreshTask", err)
		}
	}
}
