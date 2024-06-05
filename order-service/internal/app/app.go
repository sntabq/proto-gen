package app

import (
	"log/slog"
	grpcapp "order-service/internal/app/grpc"
	"order-service/internal/data"
	"order-service/internal/services/order"
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
	storage, err := data.New(dsn)
	if err != nil {
		panic(err)
	}

	orderService := order.New(log, storage, tokenTTL)

	grpcApp := grpcapp.New(log, orderService, grpcPort)

	return &App{
		GRPCServer: grpcApp,
	}
}
