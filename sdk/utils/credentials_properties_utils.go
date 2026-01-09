package utils

import (
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/models"
)

func LoadCredentialsProperties(fileName string) (*models.CredentialsProperties, error) {
	if fileName == "" {
		fileName = CredentialsPropertiesConfigName
	}
	configMap, err := LoadProperties(fileName)
	if err != nil {
		return nil, err
	}
	if configMap != nil && len(configMap) > 0 {
		regionInfos, err := InitKmsRegions(configMap, SourceTypeConfig)
		if err != nil {
			return nil, err
		}
		credential, err := InitCredential(configMap, SourceTypeConfig)
		if err != nil {
			return nil, err
		}
		return models.NewCredentialsProperties(credential, regionInfos, configMap), nil
	}
	return nil, nil
}
