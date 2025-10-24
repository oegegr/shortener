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

type contextKey string

const (
	userIDKey           contextKey = "userID"
	cookieName          string     = "auth"
	authorizationHeader string     = "Authrorization"
)

type AuthContextUserIDPovider struct{}

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

func generateUserID() string {
	return uuid.New().String()
}

func setAuthorizationHeader(w http.ResponseWriter, token string) {
	w.Header().Set(authorizationHeader, token)
}

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

func tryAuthorization(r *http.Request, v service.JWTParser) (string, error) {
	header := r.Header.Get(authorizationHeader)

	if header != "" {
		return v.UserFromJWTToken(header)
	}

	return "", nil
}

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
