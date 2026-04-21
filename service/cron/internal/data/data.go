package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewFinanceClient)

// Data .
type Data struct {
	financeClient *FinanceClient
	log           *log.Helper
}

// NewData .
func NewData(fc *FinanceClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.Info("closing the data resources")
	}
	return &Data{
		financeClient: fc,
		log:           log.NewHelper(logger),
	}, cleanup, nil
}

// FinanceClient 获取 finance 客户端
func (d *Data) FinanceClient() *FinanceClient {
	return d.financeClient
}
