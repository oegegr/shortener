// Package middleware содержит middleware-функции для обработки HTTP-запросов.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

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
