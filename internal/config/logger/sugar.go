// Package sugar содержит реализацию логгера с использованием библиотеки Zap.
package sugar

import (
	"github.com/oegegr/shortener/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger возвращает новый экземпляр логгера с использованием конфигурации приложения.
// Эта функция принимает конфигурацию приложения и возвращает логгер и ошибку.
func NewLogger(c config.Config) (*zap.SugaredLogger, error) {
	level, err := zapcore.ParseLevel(c.LogLevel)

	if err != nil {
		panic(err)
	}

	logCfg := zap.NewDevelopmentConfig()
	logCfg.Level = zap.NewAtomicLevelAt(level)
	logger, err := logCfg.Build()

	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
