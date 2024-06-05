package app

import (
	grpcapp "auth-service/internal/app/grpc"
	"auth-service/internal/data/storage"
	"auth-service/internal/services/auth"
	"auth-service/internal/services/user_info"
	"log/slog"
	"time"
)

// App wrapper for grpcapp.App
type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	dsn string,
	tokenTTL time.Duration,
) *App {
	// TODO: database setup
	authStorage, err := storage.NewAuthStorage(dsn)
	if err != nil {
		panic(err)
	}

	userInfoStorage, err := storage.NewUserInfoStorage(dsn)
	if err != nil {
		panic(err)
	}

	// TODO: auth service setup
	authService := auth.New(log, tokenTTL, authStorage)

	userInfoService := user_info.New(log, userInfoStorage, tokenTTL)

	// TODO: grpc app setup
	grpcApp := grpcapp.New(log, authService, userInfoService, grpcPort)

	go authStorage.CheckTokens()
	return &App{
		GRPCServer: grpcApp,
	}
}
