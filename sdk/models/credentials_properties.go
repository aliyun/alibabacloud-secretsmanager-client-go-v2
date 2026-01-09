package models

import (
	"github.com/aliyun/credentials-go/credentials"
)

type CredentialsProperties struct {
	Credential       credentials.Credential
	RegionInfoSlice  []*RegionInfo
	SourceProperties map[string]string
}

func NewCredentialsProperties(credential credentials.Credential, regionInfoSlice []*RegionInfo, sourceProperties map[string]string) *CredentialsProperties {
	return &CredentialsProperties{Credential: credential, RegionInfoSlice: regionInfoSlice, SourceProperties: sourceProperties}
}
