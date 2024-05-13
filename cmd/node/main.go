package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/tel-io/tel/v2"
	"go.uber.org/zap"

	"node-test/internal/node/config"
	"node-test/internal/node/handler/rest"
	"node-test/internal/node/repository"
	"node-test/internal/node/service"
	"node-test/pkg/http"
	"node-test/pkg/mongodb"
)

const (
	appGracefulTimeout = 30 * time.Second
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	sugar := logger.Sugar()

	cfg, err := config.GetConfig(ctx)
	if err != nil {
		sugar.Error("get config", tel.Error(err))
		return
	}

	mongoStorage, err := mongodb.NewStorage(ctx, cfg.Mongo.External())
	if err != nil {
		sugar.Error("initialize mongo connection", tel.Error(err))
		return
	}
	defer mongoStorage.Close()

	fsRepository, err := repository.NewNodeRepository(mongoStorage.DB)
	if err != nil {
		sugar.Error("initialize fs repository", tel.Error(err))
		return
	}

	nodeService := service.NewNodeService(cfg, sugar, fsRepository)

	routes := rest.MakeRoutes(&rest.RouterDependencies{
		NodeService: nodeService,
	})

	srv, err := http.NewEchoServer(ctx, cfg.Server.Port, routes, cancel)
	if err != nil {
		sugar.Error("initialize server", tel.Error(err))
		return
	}

	srv.Start()

	// WAIT INTERRUPT SIGNAL

	<-ctx.Done()
	sugar.Info("signal received, stop application...")

	stopCtx, stopCancel := context.WithTimeout(context.Background(), appGracefulTimeout)
	defer stopCancel()

	// stop http server
	srv.Stop(stopCtx)
}
