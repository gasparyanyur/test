package service

import (
	"context"
	"fmt"

	validatorEngine "github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	commonDomain "node-test/internal/domain"
	"node-test/internal/node/config"
	"node-test/internal/node/repository"
	"node-test/internal/node/service/domain"
)

type (
	nodeService struct {
		cfg            *config.Config
		nodeRepository repository.NodeRepository
		logger         *zap.SugaredLogger
		validator      *validatorEngine.Validate
	}

	NodeService interface {
		State(ctx context.Context) (*domain.State, error)
		Upload(chunk *commonDomain.Chunk) error
	}
)

func NewNodeService(
	cfg *config.Config,
	logger *zap.SugaredLogger,
	nodeRepository repository.NodeRepository) NodeService {
	return &nodeService{
		cfg:            cfg,
		nodeRepository: nodeRepository,
		logger:         logger,
		validator:      validatorEngine.New(),
	}
}

func (s *nodeService) State(ctx context.Context) (*domain.State, error) {

	used, err := s.nodeRepository.State(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to determine used bytes of fs %w", err)
	}

	return &domain.State{
		Size: s.cfg.Server.Size,
		Free: s.cfg.Server.Size - used,
		Used: used,
	}, nil
}

func (s *nodeService) Upload(chunk *commonDomain.Chunk) error {

	if err := s.validator.Struct(chunk); err != nil {
		return fmt.Errorf("chunk validation %w", err)
	}

	if err := s.nodeRepository.Add(chunk); err != nil {
		return fmt.Errorf("add file to fs %w", err)
	}

	return nil
}
