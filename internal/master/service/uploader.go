package service

import (
	"context"

	"go.uber.org/zap"

	"node-test/internal/domain"
	"node-test/internal/gateway"
)

type (
	uploadService struct {
		logger         *zap.SugaredLogger
		storageGateway gateway.StorageNodeGateway
	}

	// UploadService represents an interface for uploader service
	UploadService interface {
		UploadChunkedAsync(ctx context.Context) chan *domain.Chunk
		DownloadChunked(ctx context.Context, id string) ([]*domain.Chunk, error)
	}
)

func NewStorageService(
	logger *zap.SugaredLogger,
	storageGateway gateway.StorageNodeGateway,
) UploadService {
	return &uploadService{
		logger:         logger,
		storageGateway: storageGateway,
	}
}

// UploadChunkedAsync register jobs for worker pool to upload file chunk async
func (s *uploadService) UploadChunkedAsync(ctx context.Context) chan *domain.Chunk {

	var uploadChan = make(chan *domain.Chunk, 1)

	go func() {
	upload:
		for {
			select {
			case <-ctx.Done():
				break upload
			case chunk := <-uploadChan:
				if chunk == nil {
					break upload
				}
				s.storageGateway.SendAsync(chunk)
			}
		}
	}()

	return uploadChan
}

func (s *uploadService) DownloadChunked(ctx context.Context, id string) ([]*domain.Chunk, error) {

	var list []*domain.Chunk

	downloadChann := s.storageGateway.DownloadAsync(id)

download:
	for {
		select {
		case <-ctx.Done():
			break download
		case msg := <-downloadChann:
			if msg == nil {
				break download
			}

			if list == nil {
				list = make([]*domain.Chunk, 0, msg.TotalChunks)
			}
			list[msg.ChunkNumber-1] = msg
		}
	}

	return list, nil

}
