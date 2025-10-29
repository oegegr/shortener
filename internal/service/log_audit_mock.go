// Package service содержит реализацию менеджера аудита логов.
package service

import (
	"context"

	"github.com/oegegr/shortener/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockLogAuditManager представляет мок-реализацию менеджера аудита логов для тестирования.
type MockLogAuditManager struct {
	mock.Mock
}

// NotifyAllAuditors уведомляет всех аудиторов о новом лог-элементе (мок-реализация).
func (m *MockLogAuditManager) NotifyAllAuditors(ctx context.Context, logItem model.LogAuditItem) {}
