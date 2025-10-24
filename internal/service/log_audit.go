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

type LogAuditManager interface {
	NotifyAllAuditors(ctx context.Context, logItem model.LogAuditItem)
}

type DefaultLogAuditManager struct {
	auditors []LogAuditor
}

func NewDefaultLogAuditManager(auditors []LogAuditor) *DefaultLogAuditManager {
	return &DefaultLogAuditManager{
		auditors: auditors,
	}
}

func (m *DefaultLogAuditManager) NotifyAllAuditors(ctx context.Context, logItem model.LogAuditItem) {
	for _, auditor := range m.auditors {
		auditor.SaveLogItem(ctx, logItem)
	}
}

type MockLogAuditManager struct {
	mock.Mock
}

func (m *MockLogAuditManager) NotifyAllAuditors(ctx context.Context, logItem model.LogAuditItem) {}

type LogAuditor interface {
	SaveLogItem(ctx context.Context, item model.LogAuditItem) error
}

type FileLogAuditor struct {
	fileLog string
	mu      sync.Mutex
}

func NewFileLogAuditor(fileLog string) *FileLogAuditor {
	return &FileLogAuditor{fileLog: fileLog}
}

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

type HTTPLogAuditor struct {
	httpAddress string
	client      *http.Client
}

func NewHTTPLogAuditor(httpAddress string) *HTTPLogAuditor{
	return &HTTPLogAuditor{
		httpAddress: httpAddress,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

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
