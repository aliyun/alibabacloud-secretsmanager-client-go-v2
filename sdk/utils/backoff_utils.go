package utils

import (
	"errors"
	"net"
	"strings"

	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	// RejectedThrottling KMS限流返回错误码
	RejectedThrottling = "Rejected.Throttling"

	// ServiceUnavailableTemporary KMS服务不可用返回错误码
	ServiceUnavailableTemporary = "ServiceUnavailableTemporary"

	// InternalFailure KMS服务内部错误返回错误码
	InternalFailure = "InternalFailure"
)

// JudgeNeedBackoff 根据Client异常判断是否进行规避重试
func JudgeNeedBackoff(err error) bool {
	var sdkErr *dara.SDKError
	if errors.As(err, &sdkErr) {
		errorCode := tea.StringValue(sdkErr.GetCode())
		return RejectedThrottling == errorCode ||
			ServiceUnavailableTemporary == errorCode ||
			InternalFailure == errorCode
	}
	return false
}

// JudgeNeedRecoveryException 根据Client异常判断是否进行容灾重试
func JudgeNeedRecoveryException(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return isConnectionError(netErr)
	}

	var teaErr *tea.SDKError
	if errors.As(err, &teaErr) {
		errorCode := tea.StringValue(teaErr.Code)
		return SdkReadTimeout == errorCode || JudgeNeedBackoff(err)
	}
	return false
}

// isConnectionError 判断是否为连接错误
func isConnectionError(err net.Error) bool {
	errStr := err.Error()

	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"No connection could be made",
		"A connection attempt failed because the connected party did not properly respond",
		"broken pipe",
		"network is unreachable",
		"established connection failed because connected host has failed to respond",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(errStr, connErr) {
			return true
		}
	}
	return false
}
