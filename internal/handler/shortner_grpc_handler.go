package handler

import (
	"context"
	"errors"

	"github.com/oegegr/shortener/api"
	app_error "github.com/oegegr/shortener/internal/error"
	"github.com/oegegr/shortener/internal/model"
	"github.com/oegegr/shortener/internal/repository"
	"github.com/oegegr/shortener/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCServer реализует gRPC сервер для сервиса сокращения URL
type GRPCServer struct {
	api.UnimplementedShortenerServiceServer
	URLService     service.URLShortener
	userIDProvider UserIDProvider
	logAudit       service.LogAuditManager
	logger         *zap.SugaredLogger
}

// NewGRPCServer создает новый экземпляр gRPC сервера
func NewGRPCServer(
	service service.URLShortener,
	provider UserIDProvider,
	logAudit service.LogAuditManager,
	logger *zap.SugaredLogger,
) *GRPCServer {
	return &GRPCServer{
		URLService:     service,
		userIDProvider: provider,
		logAudit:       logAudit,
		logger:         logger,
	}
}

// ShortenURL обрабатывает gRPC запрос на сокращение URL
func (s *GRPCServer) ShortenURL(ctx context.Context, req *api.URLShortenRequest) (*api.URLShortenResponse, error) {
	userID, err := s.userIDProvider.Get(ctx)
	if err != nil {
		s.logger.Debugf("Failed to get userID from context: %v", err)
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	if req.Url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	// Валидируем URL
	if err := validateURL(req.Url); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid URL")
	}

	shortURL, err := s.URLService.GetShortURL(ctx, req.Url, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRepoURLAlreadyExists) {
			s.logAudit.NotifyAllAuditors(ctx, *model.NewLogAuditItem(req.Url, userID, model.LogActionShorten))
			return &api.URLShortenResponse{Result: shortURL}, nil
		}
		s.logger.Errorf("Failed to shorten URL: %v", err)
		return nil, status.Error(codes.Internal, "failed to shorten URL")
	}

	s.logAudit.NotifyAllAuditors(ctx, *model.NewLogAuditItem(req.Url, userID, model.LogActionShorten))
	return &api.URLShortenResponse{Result: shortURL}, nil
}

// ExpandURL обрабатывает gRPC запрос на получение оригинального URL
func (s *GRPCServer) ExpandURL(ctx context.Context, req *api.URLExpandRequest) (*api.URLExpandResponse, error) {
	userID, _ := s.userIDProvider.Get(ctx)

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	originalURL, err := s.URLService.GetOriginalURL(ctx, req.Id)
	if err != nil {
		if errors.Is(err, app_error.ErrServiceURLGone) {
			return nil, status.Error(codes.NotFound, "URL not found or deleted")
		}
		s.logger.Errorf("Failed to expand URL: %v", err)
		return nil, status.Error(codes.Internal, "failed to expand URL")
	}

	s.logAudit.NotifyAllAuditors(ctx, *model.NewLogAuditItem(originalURL, userID, model.LogActionFollow))
	return &api.URLExpandResponse{Result: originalURL}, nil
}

// ListUserURLs обрабатывает gRPC запрос на получение URL пользователя
func (s *GRPCServer) ListUserURLs(ctx context.Context, _ *api.Empty) (*api.UserURLsResponse, error) {
	userID, err := s.userIDProvider.Get(ctx)
	if err != nil {
		s.logger.Debugf("Failed to get userID from context: %v", err)
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	urlItems, err := s.URLService.GetUserURL(ctx, userID)
	if err != nil {
		s.logger.Errorf("Failed to get user URLs: %v", err)
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	if len(urlItems) == 0 {
		return &api.UserURLsResponse{}, nil
	}

	response := &api.UserURLsResponse{
		Url: make([]*api.URLData, 0, len(urlItems)),
	}

	for _, item := range urlItems {
		response.Url = append(response.Url, &api.URLData{
			ShortUrl:    item.ShortURL,
			OriginalUrl: item.URL,
		})
	}

	return response, nil
}