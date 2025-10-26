// gateway/internal/middleware/readonly.go
package middleware

import (
	"net/http"
	"os"
)

// ReadOnlyMiddleware блокирует мутации, если сервис в режиме read-only
func ReadOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serviceMode := os.Getenv("SERVICE_MODE")

		// Если режим read-only и метод не безопасный
		if serviceMode == "ro" && (r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" || r.Method == "DELETE") {
			http.Error(w, "Write operations are not allowed in read-only mode", http.StatusForbidden)
			return
		}

		// Пропускаем дальше
		next.ServeHTTP(w, r)
	})
}
