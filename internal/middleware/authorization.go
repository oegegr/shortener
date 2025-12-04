package middleware

import (
	"net/http"

	"go.uber.org/zap"

	pkgnet "github.com/oegegr/shortener/pkg/net"
)

// TurstedSubnetAuthorizer возвращает middleware-функцию авторизации HTTP-запросов по принадлежности IP к доверенной сети.
func TrusteSubnetAuthorizer(trustedSubnet *pkgnet.Subnet, logger zap.SugaredLogger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == nil {
				logger.Warn("Trusted subnet is empty")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				logger.Warn("X-Real-IP header is missing, denying access")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Проверка, что IP-адрес клиента входит в доверенную подсеть
			if !trustedSubnet.Contains(clientIP) {
				logger.Warn("Client IP is not in trusted subnet, denying access")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			logger.Debug("Client IP is in trusted subnet, allowing access")
			next(w, r)
		}
	}
}
