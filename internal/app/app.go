package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teacinema-go/auth-service/internal/auth/repositories/account"
	"github.com/teacinema-go/auth-service/internal/auth/repositories/refreshToken"
	"github.com/teacinema-go/auth-service/internal/auth/services"
	"github.com/teacinema-go/auth-service/internal/config"
	"github.com/teacinema-go/auth-service/internal/infra/storage/postgres"
	"github.com/teacinema-go/auth-service/internal/infra/storage/postgres/sqlc"
	"github.com/teacinema-go/auth-service/internal/infra/storage/redis"
	"github.com/teacinema-go/auth-service/internal/services/txmanager"
	"github.com/teacinema-go/auth-service/internal/transport/grpc/handlers"
	accountv1 "github.com/teacinema-go/contracts/gen/go/account/v1"
	authv1 "github.com/teacinema-go/contracts/gen/go/auth/v1"
	"github.com/teacinema-go/core/logger"
	"google.golang.org/grpc"
)

type App struct {
	cfg         *config.Config
	grpcServer  *grpc.Server
	db          *pgxpool.Pool
	redisClient *redis.Client
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run() error {
	ctx := context.Background()

	db, err := postgres.NewPostgresClient(ctx, &a.cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	a.db = db
	sqlcQuerier := sqlc.New(db)

	logger.Info("database connection established")

	redisClient, err := redis.NewClient(ctx, &a.cfg.Redis)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	a.redisClient = redisClient
	logger.Info("redis connection established")

	a.grpcServer = grpc.NewServer()

	txManager := txmanager.NewPostgresTxManager(db)
	postgresAccountRepo := account.NewPostgresAccountRepository(sqlcQuerier)
	postgresRefreshTokenRepo := refreshToken.NewPostgresRefreshTokenRepository(sqlcQuerier)

	authService := services.NewAuthService(postgresAccountRepo, postgresRefreshTokenRepo, redisClient, txManager, a.cfg.App.SecretKey)

	authHandler := handlers.NewAuthHandler(authService)
	accountHandler := handlers.NewAccountHandler(authService)

	authv1.RegisterAuthServiceServer(a.grpcServer, authHandler)
	accountv1.RegisterAccountServiceServer(a.grpcServer, accountHandler)

	grpcAddr := fmt.Sprintf(":%d", a.cfg.App.Port)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("starting gRPC server", "port", a.cfg.App.Port)
		if err = a.grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("gRPC server error", "error", err)
			quit <- syscall.SIGTERM
		}
	}()

	sig := <-quit
	logger.Info("received shutdown signal", "signal", sig.String())
	logger.Info("shutting down server...")

	a.grpcServer.GracefulStop()
	logger.Info("gRPC server stopped")

	if a.db != nil {
		a.db.Close()
		logger.Info("database connection closed")
	}

	if a.redisClient != nil {
		err := a.redisClient.Close()
		if err != nil {
			return fmt.Errorf("failed to close redis client: %w", err)
		}
		logger.Info("redis connection closed")
	}

	logger.Info("server stopped gracefully")
	return nil
}
