package watchProgress

import (
	"context"
	"watch-progress-service/services/gateway/api/internal/svc"
	"watch-progress-service/services/gateway/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetLastWatchTimeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetLastWatchTimeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetLastWatchTimeLogic {
	return &SetLastWatchTimeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetLastWatchTimeLogic) SetLastWatchTime(req *types.SetWatchTimeReq) error {
	return l.svcCtx.Scylla.SetLastWatchTime(req)
}
