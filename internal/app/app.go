package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"service-account/internal/config"
	"service-account/internal/path"
	"service-account/internal/service"
	"service-account/internal/transport/http/handler"
	"service-account/internal/transport/http/server"
	"service-account/pkg/logger"
	"syscall"
	"time"
)

func Run() {
	// Init logger.
	logsDir := path.GetLogsDir()
	var tops = []logger.TeeOption{
		{
			Filename: logsDir + "/access.log",
			Ropt: logger.RotateOptions{
				MaxSize:    1,
				MaxAge:     1,
				MaxBackups: 3,
				Compress:   true,
			},
			Lef: func(lvl logger.Level) bool {
				return lvl <= logger.InfoLevel
			},
		},
		{
			Filename: logsDir + "/error.log",
			Ropt: logger.RotateOptions{
				MaxSize:    1,
				MaxAge:     1,
				MaxBackups: 3,
				Compress:   true,
			},
			Lef: func(lvl logger.Level) bool {
				return lvl > logger.InfoLevel
			},
		},
	}

	logger.ResetDefault(logger.NewTeeWithRotate(tops))
	defer logger.Sync()

	// Init config from file.
	serviceConfig := config.NewConfig()
	if err := serviceConfig.Init(path.ConfigFile); err != nil {
		logger.Error("Init config", logger.NamedError("error", err))
		return
	}

	// Init dependencies.
	services := service.NewService(serviceConfig)

	// Init HTTP handlers.
	handlerHttp := handler.NewHandler(services)

	// Init HTTP server.
	serverHttp := server.NewServer(serviceConfig, handlerHttp.Init())

	// For graceful shutdown.
	doneChan := make(chan os.Signal, 1)
	signal.Notify(doneChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Run HTTP server in new goroutine.
	go func() {
		if err := serverHttp.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serverHttp.Run()",
				logger.NamedError("error", err),
			)
		}
	}()

	logger.Info("Services started")

	// Graceful shutdown.
	<-doneChan
	logger.Info("Services stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// Extra handling here.
		// Close database, redis, truncate message queues, etc.
		cancel()
	}()

	if err := serverHttp.Stop(ctx); err != nil {
		logger.Fatal("Services shutdown failed",
			logger.NamedError("error", err),
		)
	}

	logger.Info("Services exited properly")
}
