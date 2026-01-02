package main

import (
	"log"

	"github.com/teacinema-go/auth-service/internal/app"
	"github.com/teacinema-go/auth-service/internal/config"
	"github.com/teacinema-go/core/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	l := logger.New(cfg.App.Env)
	l.Info("config loaded successfully", "env", cfg.App.Env)
	application := app.New(cfg, l)

	if err = application.Run(); err != nil {
		l.Error("application stopped with error", "error", err)
	}
}
