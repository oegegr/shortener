// Package service содержит реализацию стратегии удаления URL-адресов.
package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/oegegr/shortener/internal/repository"
	"go.uber.org/zap"
)

// ErrDeleteQueueIsFull представляет ошибку, которая возникает при переполнении очереди удаления.
var ErrDeleteQueueIsFull = errors.New("delete queue is full")

// QueueDeletionStrategy представляет реализацию стратегии удаления URL-адресов с использованием очереди.
type QueueDeletionStrategy struct {
	// urlRepository представляет репозиторий URL-адресов.
	urlRepository repository.URLRepository
	// logger представляет логгер для записи сообщений.
	logger zap.SugaredLogger
	// workerNum представляет количество рабочих потоков для удаления URL-адресов.
	workerNum int
	// deleteQueue представляет очередь удаления URL-адресов.
	deleteQueue chan deleteTask
	// workerWG представляет группу ожидания для рабочих потоков.
	workerWG sync.WaitGroup
	// waitTimeout представляет время ожидания для удаления URL-адресов.
	waitTimeout time.Duration
}

// deleteTask представляет задачу удаления URL-адресов.
type deleteTask struct {
	// ctx представляет контекст задачи.
	ctx context.Context
	// shortIDs представляет список идентификаторов URL-адресов для удаления.
	shortIDs []string
}

// NewQueueURLDeletionStrategy возвращает новый экземпляр QueueDeletionStrategy.
// Эта функция принимает репозиторий URL-адресов, логгер, количество рабочих потоков, размер очереди и время ожидания.
func NewQueueURLDeletionStrategy(
	repo repository.URLRepository,
	logger zap.SugaredLogger,
	workerNum int,
	taskNum int,
	waitTimeout time.Duration,
) *QueueDeletionStrategy {
	delStrategy := &QueueDeletionStrategy{
		urlRepository: repo,
		logger:        logger,
		workerNum:     workerNum,
		deleteQueue:   make(chan deleteTask, taskNum),
		waitTimeout:   waitTimeout,
	}
	delStrategy.Start()
	return delStrategy
}

// Start запускает рабочие потоки для удаления URL-адресов.
func (s *QueueDeletionStrategy) Start() {
	s.workerWG.Add(s.workerNum)
	for workerID := 0; workerID < s.workerNum; workerID++ {
		go s.worker()
	}
}

// Stop останавливает рабочие потоки для удаления URL-адресов.
func (s *QueueDeletionStrategy) Stop() {
	close(s.deleteQueue)
	s.workerWG.Wait()
}

// DeleteURL добавляет задачу удаления URL-адресов в очередь.
// Эта функция принимает контекст и список идентификаторов URL-адресов для удаления.
func (s *QueueDeletionStrategy) DeleteURL(ctx context.Context, shortIDs []string) error {
	select {
	case s.deleteQueue <- deleteTask{ctx, shortIDs}:
		return nil
	default:
		s.logger.Error("delete queue is full")
		return ErrDeleteQueueIsFull
	}
}

// worker представляет рабочий поток для удаления URL-адресов.
func (s *QueueDeletionStrategy) worker() {
	defer s.workerWG.Done()

	for task := range s.deleteQueue {
		err := s.urlRepository.DeleteURL(task.ctx, task.shortIDs)
		if err != nil {
			s.logger.Error(err)
			continue
		}
	}
}
