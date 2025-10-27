// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package watchProgress

import (
	"net/http"
	"watch-progress-service/services/gateway/api/internal/handler/response"
	watchProgress "watch-progress-service/services/gateway/api/internal/logic/watchProgress"
	"watch-progress-service/services/gateway/api/internal/svc"
	"watch-progress-service/services/gateway/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SetLastWatchTimeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetWatchTimeReq
		if err := httpx.Parse(r, &req); err != nil {
			response.NewErrorResponse(http.StatusBadRequest, err.Error(), w)
			return
		}

		// TODO: add validation for st <= et, and eps > 0.5.
		// TODO: add validation for et <= duration.
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
