package exchanger

import (
	"context"
	"fmt"
)

type Exchanger interface {
	ExchangeCredential(ctx context.Context, identity string) (string, string, error)
}

func NewExchanger(config Config) (Exchanger, error) {
	// 首先验证配置
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var exchanger Exchanger

	switch config.ExchangerType {
	case ExchangerTypeKms:
		exchanger = &KMSExchanger{
			kmsEndpoint: config.KmsEndpoint,
			aapArn:      config.AapArn,
			ca:          config.Ca,
		}
	default:
		return nil, fmt.Errorf("unsupported exchanger type: %s", config.ExchangerType)
	}
	return exchanger, nil
}
