package middleware

import (
	"net/http"
	"quickflow/shared/logger"
	"strings"

	"quickflow/config/cors"
	http2 "quickflow/gateway/utils/http"
)

// CORSMiddleware adds CORS headers to the response.
func CORSMiddleware(config *cors_config.CORSConfig) func(http.Handler) http.Handler {
	if config == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(http2.SetRequestId(r.Context()))

			origin := r.Header.Get("Origin")
			logger.Info(r.Context(), "CORS middleware got request from origin: "+origin+", with method: "+r.Method)

			// Проверяем, не установлен ли уже CORS-заголовок
			if _, exists := w.Header()["Access-Control-Allow-Origin"]; !exists {
				logger.Info(r.Context(), "Setting CORS headers for origin: %v", origin)
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				logger.Info(r.Context(), "allowed methods: %v", strings.Join(config.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			}

			// Если OPTIONS-запрос, отвечаем 204
			if r.Method == http.MethodOptions {
				logger.Info(r.Context(), "Setting CORS options for request: %v", r.URL)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
