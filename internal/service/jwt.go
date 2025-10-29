// Package service содержит реализацию парсера JWT-токенов.
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// Claims представляет структуру данных для хранения в JWT-токене.
type Claims struct {
	// UserID представляет идентификатор пользователя.
	UserID string `json:"user_id"`
	// RegisteredClaims представляет зарегистрированные данные в JWT-токене.
	jwt.RegisteredClaims
}

// JWTParser представляет парсер JWT-токенов.
type JWTParser struct {
	// jwtSecret представляет секретный ключ для подписи JWT-токенов.
	jwtSecret string
	// logger представляет логгер для записи сообщений.
	logger zap.SugaredLogger
}

// ErrInvalidJWTToken представляет ошибку, которая возникает при невалидном JWT-токене.
var ErrInvalidJWTToken = errors.New("invalid token")

// NewJWTParser возвращает новый экземпляр JWTParser.
// Эта функция принимает секретный ключ для подписи JWT-токенов и логгер.
func NewJWTParser(jwtSecret string, logger zap.SugaredLogger) JWTParser {
	return JWTParser{jwtSecret, logger}
}

// CreateNewJWTToken создает новый JWT-токен для пользователя.
// Эта функция принимает идентификатор пользователя и возвращает сгенерированный JWT-токен.
func (v *JWTParser) CreateNewJWTToken(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(v.jwtSecret))
}

// UserFromJWTToken извлекает идентификатор пользователя из JWT-токена.
// Эта функция принимает JWT-токен и возвращает идентификатор пользователя, если токен валиден.
func (v *JWTParser) UserFromJWTToken(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", ErrInvalidJWTToken
	}

	return claims.UserID, nil
}
