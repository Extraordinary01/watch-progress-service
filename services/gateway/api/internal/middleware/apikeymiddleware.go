package middleware

import (
	"net/http"
	"watch-progress-service/services/gateway/api/internal/config"
	"watch-progress-service/services/gateway/api/internal/handler/response"
)

type ApiKeyMiddleware struct {
	validKeys map[string]bool
}

func NewApiKeyMiddleware(cfg config.Config) *ApiKeyMiddleware {
	validKeys := make(map[string]bool)
	for _, key := range cfg.ApiKeys {
		validKeys[key] = true
	}
	return &ApiKeyMiddleware{
		validKeys: validKeys,
	}
}

func (m *ApiKeyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")

		if apiKey == "" {
			response.NewErrorResponse(http.StatusUnauthorized, "API Key is missing", w)
			return
		}
		if !m.isValid(apiKey) {
			response.NewErrorResponse(http.StatusUnauthorized, "Invalid API key", w)
			return
		}

		next(w, r)
	}
}

func (m *ApiKeyMiddleware) isValid(key string) bool {
	return m.validKeys[key]
}
