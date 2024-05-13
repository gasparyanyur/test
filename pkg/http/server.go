package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/tel-io/tel/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	echoServerReadTimeout       = 5 * time.Second
	echoServerReadHeaderTimeout = 1 * time.Second
	echoServerWriteTimeout      = 30 * time.Second
	echoServerIdleTimeout       = 120 * time.Second
	echoServerGracefulTimeout   = 5 * time.Second
)

type EchoServer struct {
	server   *echo.Echo
	cancel   context.CancelFunc
	observer *tel.Telemetry
}

func NewEchoServer(
	ctx context.Context,
	port int,
	e *echo.Echo,
	cancel context.CancelFunc,
) (*EchoServer, error) {

	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadTimeout:       echoServerReadTimeout,
		ReadHeaderTimeout: echoServerReadHeaderTimeout,
		WriteTimeout:      echoServerWriteTimeout,
		IdleTimeout:       echoServerIdleTimeout,
		ErrorLog: func() *log.Logger {
			// error checking is muted because inside a function
			// an error can be returned only if an incorrect logging level is entered
			errorLog, _ := zap.NewStdLogAt(
				tel.FromCtx(ctx).Logger.With(zap.String("system", "http")),
				zapcore.WarnLevel,
			)

			return errorLog
		}(),
	}

	e.Server = &srv
	e.HideBanner = true
	e.HidePort = true

	return &EchoServer{
		server:   e,
		cancel:   cancel,
		observer: tel.FromCtx(ctx),
	}, nil
}

func (s *EchoServer) Start() {

	s.observer.Info("starting echo server", tel.String("addr", s.server.Server.Addr))

	go func() {
		if err := s.server.Start(s.server.Server.Addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.observer.Error("echo server listener closed due to the error", tel.Error(err))
			s.cancel()
		}
	}()

	time.Sleep(time.Millisecond * 500)
}

func (s *EchoServer) Use(f echo.MiddlewareFunc) {
	s.server.Use(f)
}

func (s *EchoServer) Add(method, path string, h echo.HandlerFunc) {
	s.server.Add(method, path, h)
}

func (s *EchoServer) Stop(ctx context.Context) {

	ctx, cancel := context.WithTimeout(ctx, echoServerGracefulTimeout)
	defer cancel()

	done := make(chan error)
	go func() {
		if err := s.server.Shutdown(ctx); err != nil {
			done <- err
		}
		close(done)
	}()

	select {
	case err := <-done:
		if err != nil {
			s.observer.Error("cannot stop echo server", tel.Error(err))
			return
		}
		s.observer.Info("echo server gracefully stopped")
	case <-ctx.Done():
		s.observer.Error("cannot stop echo server, error by graceful timeout", tel.Error(ctx.Err()))
	}
}
