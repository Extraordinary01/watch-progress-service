package watchProgress

import (
	"context"
	"errors"
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
	if req.StartTime > req.EndTime {
		return errors.New("start time must be less than end time")
	}

	if req.EndTime > req.Duration {
		return errors.New("end time must be less than duration")
	}

	if req.EndTime-req.StartTime < 1 {
		return nil
	}
	return l.svcCtx.Scylla.SetLastWatchTime(req)
}
