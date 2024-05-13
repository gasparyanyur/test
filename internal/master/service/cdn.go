package service

import (
	"context"

	"go.uber.org/zap"

	"node-test/internal/gateway"
)

type (
	cdnService struct {
		logger         *zap.SugaredLogger
		storageGateway gateway.StorageNodeGateway
	}

	// CDNService represents an interface for uploader service
	CDNService interface {
		Download(ctx context.Context, id string)
	}
)

func NewCDNService(
	logger *zap.SugaredLogger,
	storageGateway gateway.StorageNodeGateway,
) CDNService {
	return &cdnService{
		logger:         logger,
		storageGateway: storageGateway,
	}
}

// Download downloads chunks of the specified file
func (s *cdnService) Download(ctx context.Context, id string) {}
