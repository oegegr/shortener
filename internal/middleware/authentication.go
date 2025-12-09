// Package middleware содержит middleware-функции для обработки HTTP-запросов.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/oegegr/shortener/internal/service"
)

// contextKey представляет ключ для контекста запроса.
type contextKey string

// userIDKey представляет ключ для идентификатора пользователя в контексте запроса.
const (
	userIDKey           contextKey = "userID"
	cookieName          string     = "auth"
	authorizationHeader string     = "Authorization"
)

// AuthContextUserIDPovider предоставляет провайдер для получения идентификатора пользователя из контекста запроса.
type AuthContextUserIDPovider struct{}

// Get возвращает идентификатор пользователя из контекста запроса.
func (a *AuthContextUserIDPovider) Get(ctx context.Context) (string, error) {
	value := ctx.Value(userIDKey)
	if value == nil {
		return "", fmt.Errorf("failed to get userID from context")
	}

	userID, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("failed to convert userID to string")
	}
	return userID, nil
}

// AuthMiddleware возвращает middleware-функцию для аутентификации пользователей.
func AuthMiddleware(logger zap.SugaredLogger, jwt service.JWTParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {

		authHandler := func(w http.ResponseWriter, r *http.Request) {

			var userID string
			var err error

			userID, err = tryAuthorization(r, jwt)

			if err != nil {
				http.Error(w, "bad authorization header", http.StatusUnauthorized)
			}

			if userID == "" {
				userID, err = tryCookies(r, jwt)

				if err != nil {
					http.Error(w, "bad cookie", http.StatusUnauthorized)
				}
			}

			if userID == "" {
				userID = generateUserID()
				logger.Info("new userID has been created")
			}

			token, err := jwt.CreateNewJWTToken(userID)

			if err != nil {
				logger.Errorf("Failed to create jwt token with userID %s", userID, err)
				http.Error(w, "", http.StatusInternalServerError)
			}

			setAuthCookie(w, token)

			setAuthorizationHeader(w, token)

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(authHandler)
	}
}

// GRPCAuthInterceptor предоставляет интерцептор для аутентификации gRPC запросов
func GRPCAuthInterceptor(jwtParser service.JWTParser, logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var userID string
		var token string

		// Извлекаем токен из метаданных
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
				token = authHeaders[0]
				// Убираем префикс "Bearer " если есть
				token = strings.TrimPrefix(token, "Bearer ")
				var err error
				userID, err = jwtParser.UserFromJWTToken(token)
				if err != nil {
					logger.Debugf("Invalid JWT token: %v", err)
				}
			}
		}

		// Если токен не валиден или отсутствует, создаем нового пользователя
		if userID == "" {
			userID = uuid.New().String()
			var err error
			token, err = jwtParser.CreateNewJWTToken(userID)
			if err != nil {
				logger.Errorf("Failed to create JWT token: %v", err)
				return handler(ctx, req)
			}

			// Устанавливаем новый токен в исходящие метаданные
			header := metadata.Pairs("authorization", token)
			ctx = metadata.NewOutgoingContext(ctx, header)
		}

		// Добавляем userID в контекст
		ctx = context.WithValue(ctx, userIDKey, userID)
		
		return handler(ctx, req)
	}
}

// generateUserID генерирует новый идентификатор пользователя.
func generateUserID() string {
	return uuid.New().String()
}

// setAuthorizationHeader устанавливает заголовок Authorization в HTTP-ответе.
func setAuthorizationHeader(w http.ResponseWriter, token string) {
	w.Header().Set(authorizationHeader, token)
}

// setAuthCookie устанавливает cookie с именем auth в HTTP-ответе.
func setAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, cookie)
}

// tryAuthorization пытается получить идентификатор пользователя из заголовка Authorization.
func tryAuthorization(r *http.Request, v service.JWTParser) (string, error) {
	header := r.Header.Get(authorizationHeader)

	if header != "" {
		return v.UserFromJWTToken(header)
	}

	return "", nil
}

// tryCookies пытается получить идентификатор пользователя из cookie с именем auth.
func tryCookies(r *http.Request, v service.JWTParser) (string, error) {
	c, err := r.Cookie(cookieName)

	if err == nil {
		return v.UserFromJWTToken(c.Value)
	}

	if errors.Is(err, http.ErrNoCookie) {
		return "", nil
	}

	return "", err
}
