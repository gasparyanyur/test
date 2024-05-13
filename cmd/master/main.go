package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/tel-io/tel/v2"
	"go.uber.org/zap"

	"node-test/internal/common/pool"
	"node-test/internal/gateway"
	"node-test/internal/master/config"
	masterRoutes "node-test/internal/master/handler/rest"
	"node-test/internal/master/service"
	"node-test/pkg/http"
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

	pool := pool.NewWorkerPool(cfg.FileStorage.WorkerCount)
	pool.Start(ctx)

	storageGateway, err := gateway.NewStorageNodeGateway(cfg.FileStorage, pool)
	if err != nil {
		sugar.Error("storage gateway", tel.Error(err))
		return
	}

	storageService := service.NewStorageService(sugar, storageGateway)

	routes := masterRoutes.MakeRoutes(&masterRoutes.RouterDependencies{
		StorageService: storageService,
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
