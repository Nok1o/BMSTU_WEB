package middleware

import (
	"context"
	"net/http"

	"quickflow/shared/logger"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(context.Background(), "Panic: %v, URL: %s", err, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500 - Internal Server Error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
