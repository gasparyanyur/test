package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tel-io/tel/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const storageGracefulTimeout = 5 * time.Second

type Config struct {
	URI      string
	User     string
	Password string
	DB       string
}

type Storage struct {
	client   *mongo.Client
	DB       *mongo.Database
	observer *tel.Telemetry
}

func NewStorage(ctx context.Context, cfg Config) (*Storage, error) {

	opts := options.Client().ApplyURI(cfg.URI)
	// TODO add trace filter by span name (cabinet-api ping)

	if cfg.User != "" {
		opts = opts.SetAuth(
			options.Credential{
				Username: cfg.User,
				Password: cfg.Password,
			},
		)
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("initialize connection: %w", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Storage{
		client:   client,
		DB:       client.Database(cfg.DB),
		observer: tel.FromCtx(ctx),
	}, nil
}

func (s *Storage) Close() {

	ctx, cancel := context.WithTimeout(context.Background(), storageGracefulTimeout)
	defer cancel()

	done := make(chan error)
	go func() {
		if err := s.client.Disconnect(ctx); err != nil {
			done <- err
		}
		close(done)
	}()

	select {
	case err := <-done:
		if err != nil {
			s.observer.Error("cannot close mongo db connection", tel.Error(err))
			return
		}
		s.observer.Info("mongo db connection gracefully closed")
	case <-ctx.Done():
		s.observer.Error("close mongo db connection, error by graceful timeout", tel.Error(ctx.Err()))
	}
}

func IsNotFoundError(err error) bool {
	return errors.Is(err, mongo.ErrNoDocuments)
}
