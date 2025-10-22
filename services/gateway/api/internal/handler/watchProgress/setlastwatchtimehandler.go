// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package watchProgress

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"watch-progress-service/services/gateway/api/internal/handler/response"
	watchProgress "watch-progress-service/services/gateway/api/internal/logic/watchProgress"
	"watch-progress-service/services/gateway/api/internal/svc"
	"watch-progress-service/services/gateway/api/internal/types"
)

func SetLastWatchTimeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetWatchTimeReq
		if err := httpx.Parse(r, &req); err != nil {
			response.NewErrorResponse(http.StatusBadRequest, err.Error(), w)
			return
		}
		l := watchProgress.NewSetLastWatchTimeLogic(r.Context(), svcCtx)
		err := l.SetLastWatchTime(&req)
		if err != nil {
			response.NewErrorResponse(http.StatusInternalServerError, err.Error(), w)
		} else {
			w.WriteHeader(http.StatusCreated)
			httpx.Ok(w)
		}
	}
}
