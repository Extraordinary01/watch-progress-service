// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"github.com/zeromicro/go-zero/rest"
	"watch-progress-service/services/gateway/api/internal/config"
	"watch-progress-service/services/gateway/api/internal/middleware"
	"watch-progress-service/third_party/scylla"
)

type ServiceContext struct {
	Config           config.Config
	Scylla           *Scylla
	ApiKeyMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	scyllaClient, err := scylla.NewClient(c.Scylla)
	if err != nil {
		return nil, err
	}

	scyllaLogic := NewScylla(scyllaClient)

	return &ServiceContext{
		Config:           c,
		Scylla:           scyllaLogic,
		ApiKeyMiddleware: middleware.NewApiKeyMiddleware(c).Handle,
	}, nil
}
