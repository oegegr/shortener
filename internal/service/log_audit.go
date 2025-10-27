// Package service содержит реализацию менеджера аудита логов.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/oegegr/shortener/internal/model"
	"github.com/stretchr/testify/mock"
)

// LogAuditManager представляет интерфейс для менеджера аудита логов.
type LogAuditManager interface {
	// NotifyAllAuditors уведомляет всех аудиторов о новом лог-элементе.
	NotifyAllAuditors(ctx context.Context, logItem model.LogAuditItem)
}

// DefaultLogAuditManager представляет реализацию менеджера аудита логов по умолчанию.
type DefaultLogAuditManager struct {
	// auditors представляет список аудиторов.
	auditors []LogAuditor
}

// NewDefaultLogAuditManager возвращает новый экземпляр DefaultLogAuditManager.
// Эта функция принимает список аудиторов.
func NewDefaultLogAuditManager(auditors []LogAuditor) *DefaultLogAuditManager {
	return &DefaultLogAuditManager{
		auditors: auditors,
	}
}

// NotifyAllAuditors уведомляет всех аудиторов о новом лог-элементе.
func (m *DefaultLogAuditManager) NotifyAllAuditors(ctx context.Context, logItem model.LogAuditItem) {
	for _, auditor := range m.auditors {
		auditor.SaveLogItem(ctx, logItem)
	}
}

// MockLogAuditManager представляет мок-реализацию менеджера аудита логов для тестирования.
type MockLogAuditManager struct {
	mock.Mock
}

// NotifyAllAuditors уведомляет всех аудиторов о новом лог-элементе (мок-реализация).
func (m *MockLogAuditManager) NotifyAllAuditors(ctx context.Context, logItem model.LogAuditItem) {}

// LogAuditor представляет интерфейс для аудитора логов.
type LogAuditor interface {
	// SaveLogItem сохраняет лог-элемент в аудиторе.
	SaveLogItem(ctx context.Context, item model.LogAuditItem) error
}

// FileLogAuditor представляет реализацию аудитора логов, который записывает логи в файл.
type FileLogAuditor struct {
	// fileLog представляет путь к файлу логов.
	fileLog string
	// mu представляет mutex для синхронизации доступа к файлу логов.
	mu sync.Mutex
}

// NewFileLogAuditor возвращает новый экземпляр FileLogAuditor.
// Эта функция принимает путь к файлу логов.
func NewFileLogAuditor(fileLog string) *FileLogAuditor {
	return &FileLogAuditor{fileLog: fileLog}
}

// SaveLogItem сохраняет лог-элемент в файле логов.
func (a *FileLogAuditor) SaveLogItem(ctx context.Context, item model.LogAuditItem) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal log item: %w", err)
	}

	data = append(data, '\n')

	file, err := os.OpenFile(a.fileLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}

// HTTPLogAuditor представляет реализацию аудитора логов, который отправляет логи по HTTP.
type HTTPLogAuditor struct {
	// httpAddress представляет адрес HTTP-эндпоинта для отправки логов.
	httpAddress string
	// client представляет HTTP-клиент для отправки логов.
	client *http.Client
}

// NewHTTPLogAuditor возвращает новый экземпляр HTTPLogAuditor.
// Эта функция принимает адрес HTTP-эндпоинта для отправки логов.
func NewHTTPLogAuditor(httpAddress string) *HTTPLogAuditor {
	return &HTTPLogAuditor{
		httpAddress: httpAddress,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SaveLogItem сохраняет лог-элемент, отправляя его по HTTP.
func (a *HTTPLogAuditor) SaveLogItem(ctx context.Context, item model.LogAuditItem) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal log item: %w", err)
	}

	req, err := http.NewRequest("POST", a.httpAddress, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP request failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}
