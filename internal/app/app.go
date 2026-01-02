package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/teacinema-go/auth-service/internal/config"
	"github.com/teacinema-go/auth-service/internal/handler"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
	"google.golang.org/grpc"
)

type App struct {
	cfg        *config.Config
	logger     *slog.Logger
	grpcServer *grpc.Server
}

func New(cfg *config.Config, logger *slog.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run() error {
	a.grpcServer = grpc.NewServer()

	h := handler.NewHandler(a.logger)

	authv1.RegisterAuthServiceServer(a.grpcServer, h)

	grpcAddr := fmt.Sprintf(":%d", a.cfg.App.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		a.logger.Info("starting gRPC server", "port", a.cfg.App.Port)
		if err = a.grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			a.logger.Error("gRPC server error", "error", err)
			quit <- syscall.SIGTERM
		}
	}()

	sig := <-quit
	a.logger.Info("received shutdown signal", "signal", sig.String())
	a.logger.Info("shutting down server...")
	a.grpcServer.GracefulStop()
	a.logger.Info("server stopped gracefully")
	return nil
}
