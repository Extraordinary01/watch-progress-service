// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"watch-progress-service/third_party/clickhouse"
	"watch-progress-service/third_party/scylla"
)

type Config struct {
	Rest       rest.RestConf
	Scylla     scylla.Params
	ApiKeys    []string
	Clickhouse clickhouse.Config
}
