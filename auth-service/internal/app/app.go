package app

import (
	grpcapp "auth-service/internal/app/grpc"
	"auth-service/internal/data/storage"
	"auth-service/internal/services/auth"
	"auth-service/internal/services/user_info"
	"log/slog"
	"time"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(
	log *slog.Logger,
	grpcPort int,
	dsn string,
	tokenTTL time.Duration,
) *App {
	// database setup
	authStorage, err := storage.NewAuthStorage(dsn)
	if err != nil {
		panic(err)
	}

	userInfoStorage, err := storage.NewUserInfoStorage(dsn)
	if err != nil {
		panic(err)
	}

	// auth service setup
	authService := auth.New(log, tokenTTL, authStorage)

	userInfoService := user_info.New(log, userInfoStorage, tokenTTL)

	// grpc app setup
	grpcApp := grpcapp.New(log, authService, userInfoService, grpcPort)

	go authStorage.CheckTokens()
	return &App{
		GRPCServer: grpcApp,
	}
}
