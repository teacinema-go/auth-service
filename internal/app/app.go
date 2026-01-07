package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/teacinema-go/auth-service/internal/config"
	"github.com/teacinema-go/auth-service/internal/database"
	teacinema "github.com/teacinema-go/auth-service/internal/database/sqlc"
	"github.com/teacinema-go/auth-service/internal/handler"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
	"google.golang.org/grpc"
)

type App struct {
	cfg        *config.Config
	logger     *slog.Logger
	grpcServer *grpc.Server
	db         *pgxpool.Pool
	rdb        *redis.Client
}

func New(cfg *config.Config, logger *slog.Logger) *App {
	return &App{
		cfg:    cfg,
		logger: logger,
	}
}

func (a *App) Run() error {
	ctx := context.Background()

	db, err := database.NewPostgresClient(ctx, &a.cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	a.db = db
	queries := teacinema.New(db)

	a.logger.Info("database connection established")

	rdb, err := database.NewRedisClient(ctx, &a.cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	a.rdb = rdb
	a.logger.Info("redis connection established")

	a.grpcServer = grpc.NewServer()
	h := handler.NewHandler(a.logger, queries, db, rdb)

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
	a.logger.Info("gRPC server stopped")

	if a.db != nil {
		a.db.Close()
		a.logger.Info("database connection closed")
	}

	if a.rdb != nil {
		err := a.rdb.Close()
		if err != nil {
			return fmt.Errorf("failed to close redis client: %w", err)
		}
		a.logger.Info("redis connection closed")
	}

	a.logger.Info("server stopped gracefully")
	return nil
}
