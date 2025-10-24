package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/oegegr/shortener/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrDeleteQueueIsFull = errors.New("delete queue is full")
)

type QueueDeletionStrategy struct {
	urlRepository repository.URLRepository
	logger        zap.SugaredLogger
	workerNum     int
	deleteQueue   chan deleteTask
	workerWG      sync.WaitGroup
	waitTimeout   time.Duration
}

type deleteTask struct {
	ctx      context.Context
	shortIDs []string
}

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

func (s *QueueDeletionStrategy) Start() {
	s.workerWG.Add(s.workerNum)
	for workerID := 0; workerID < s.workerNum; workerID++ {
		go s.worker()
	}
}

func (s *QueueDeletionStrategy) Stop() {
	close(s.deleteQueue)
	s.workerWG.Wait()
}

func (s *QueueDeletionStrategy) DeleteURL(ctx context.Context, shortIDs []string) error {

	select {
	case s.deleteQueue <- deleteTask{ctx, shortIDs}:
		return nil
	default:
		s.logger.Error("delete queue is full")
		return ErrDeleteQueueIsFull
	}
}

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
