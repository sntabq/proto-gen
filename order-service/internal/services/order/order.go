package order

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	ssov1 "github.com/sntabq/proto-gen/gen/go/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	grpcapp "order-service/internal/app/grpc"
	"order-service/internal/data/dto"
	"order-service/internal/data/models"
	"order-service/internal/sl"
	"time"
)

type Order struct {
	log           *slog.Logger
	orderProvider OrderRepo
	channel       *amqp.Channel
	tokenTTL      time.Duration
}

func New(
	log *slog.Logger,
	orderProvider OrderRepo,
	tokenTtl time.Duration,
) *Order {
	return &Order{
		log:           log,
		orderProvider: orderProvider,
		tokenTTL:      tokenTtl,
	}
}

type OrderRepo interface {
	SaveOrder(ctx context.Context, orderDTO *dto.OrderDTO) (*dto.OrderDTO, error)
	GetAllOrders(context.Context) ([]*dto.OrderDTO, error)
	GetOrderById(context.Context, int) (*models.Order, error)
	GetOrdersByUserId(context.Context, int) ([]*dto.OrderDTO, error)
}

func (o *Order) CreateOrder(ctx context.Context, orderDTO *dto.OrderDTO) (*dto.OrderDTO, error) {
	const op = "Order.CreateOrder"

	log := o.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to create orderDTO")

	orderDTO, err := o.orderProvider.SaveOrder(ctx, orderDTO)
	if err != nil {
		o.log.Warn("failed to save orderDTO", sl.Err(err))
		return nil, fmt.Errorf("%s", op)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("%s", op)
	}

	tkn, found := md["authorization"]
	if !found && len(tkn) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authentication is required")
	}
	userInfo, err := grpcapp.UserInfoServiceClient.GetUserInfo(ctx, &ssov1.GetUserInfoRequest{Token: tkn[0]})
	if err != nil {
		switch {
		}
		log.Info("failed to get user info from auth service")
		return nil, status.Errorf(codes.Internal, "failed to get user info from auth service")
	}
	err = sendNotification(ctx, orderDTO, userInfo)
	if err != nil {
		o.log.Warn("failed to publish message", sl.Err(err))
		return nil, fmt.Errorf("%s", op)
	}

	return orderDTO, nil
}

func (o *Order) ListOrders(ctx context.Context) ([]*dto.OrderDTO, error) {
	const op = "Order.ListOrders"
	log := o.log.With(
		slog.String("op", op),
	)

	log.Info("attempting to get all orders")

	items, err := o.orderProvider.GetAllOrders(ctx)
	if err != nil {
		o.log.Warn("failed to get all items", sl.Err(err))
		return nil, err
	}

	return items, nil
}

func (o *Order) GetOrder(ctx context.Context, id int) (*models.Order, error) {
	const op = "Order.GetOrder"
	log := o.log.With(
		slog.String("op", op),
		slog.Int("item id", id),
	)

	log.Info("attempting to get order")
	item, err := o.orderProvider.GetOrderById(ctx, id)
	if err != nil {
		o.log.Warn("failed to get order", sl.Err(err))
		return nil, err
	}

	return item, nil
}

func (o *Order) GetOrdersByUserId(ctx context.Context, userId int) ([]*dto.OrderDTO, error) {
	const op = "Order.GetOrder"
	log := o.log.With(
		slog.String("op", op),
		slog.Int("user id", userId),
	)

	log.Info("attempting to get orders of user with id ", userId)
	ordersByUserId, err := o.orderProvider.GetOrdersByUserId(ctx, userId)
	if err != nil {
		o.log.Warn("failed to get orders of user", sl.Err(err))
		return nil, err
	}

	return ordersByUserId, nil
}

func sendNotification(ctx context.Context, orderDTO *dto.OrderDTO, userInfo *ssov1.GetUserInfoResponse) error {
	const op = "Order.sendNotification"
	conn, err := amqp.Dial("amqp://user:password@localhost:5672/")
	if err != nil {
		log.Print("failed to connect in new connection", sl.Err(err))
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Print("failed to create channel in new MQ", sl.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	q, err := ch.QueueDeclare(
		"shop", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		log.Print("failed to declare queue", sl.Err(err))
		return fmt.Errorf("%s", op)
	}

	data := map[string]interface{}{
		"user_info":  userInfo.User,
		"order_info": orderDTO,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Print("Failed to marshal data", sl.Err(err))
		return fmt.Errorf("%s", op)
	}
	body := dataBytes
	err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to publish %v", err))
	}
	return nil
}
