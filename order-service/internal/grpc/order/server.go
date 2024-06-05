package orderGrpc

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	cataloguep "github.com/sntabq/proto-gen/gen/go/catalogue"
	orderp "github.com/sntabq/proto-gen/gen/go/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"order-service/internal/data"
	"order-service/internal/data/dto"
	"order-service/internal/data/models"
	"strconv"
	"strings"
)

type OrderService interface {
	CreateOrder(context.Context, *dto.OrderDTO) (*dto.OrderDTO, error)
	ListOrders(context.Context) ([]*dto.OrderDTO, error)
	GetOrder(context.Context, int) (*models.Order, error)
	GetOrdersByUserId(context.Context, int) ([]*dto.OrderDTO, error)
}

type orderService struct {
	orderp.UnimplementedOrderServiceServer
	order OrderService
}

func Register(gRPCServer *grpc.Server, order OrderService) {
	orderp.RegisterOrderServiceServer(gRPCServer, &orderService{order: order})
}

func (os *orderService) CreateOrder(ctx context.Context, req *orderp.CreateOrderRequest) (*orderp.CreateOrderResponse, error) {
	var ord dto.OrderDTO
	err := copier.Copy(&ord, req.Order)
	if err != nil {
		log.Fatalf("failed to copy %v", err)
		return nil, err
	}

	orderDTO, err := os.order.CreateOrder(ctx, &ord)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), data.ErrItemDoesNotExist.Error()):
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("item with id %d does not exist", req.Order.ItemId))
		case strings.Contains(err.Error(), data.ErrUserDoesNotExist.Error()):
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("user with id %d does not exist", req.Order.ItemId))
		default:
			return nil, status.Error(codes.Internal, "error with create order")
		}
	}

	var response orderp.Order
	err = copier.Copy(&response, orderDTO)
	if err != nil {
		return nil, status.Error(codes.Internal, "error with copying to dto")
	}

	return &orderp.CreateOrderResponse{Order: &response}, nil
}

func (os *orderService) ListOrders(ctx context.Context, req *orderp.ListOrdersRequest) (*orderp.ListOrdersResponse, error) {
	var responseOrders []*orderp.Order

	orders, err := os.order.ListOrders(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range orders {
		var ord orderp.Order
		err := copier.Copy(&ord, &item)
		if err != nil {
			log.Fatalf("failed to copy %v", err)
			return nil, err
		}

		responseOrders = append(responseOrders, &ord)
	}

	return &orderp.ListOrdersResponse{Orders: responseOrders}, nil
}

func (os *orderService) GetOrder(ctx context.Context, req *orderp.GetOrderRequest) (*orderp.GetOrderResponse, error) {
	id, err := strconv.Atoi(req.Id)
	if err != nil {
		return nil, err
	}

	order, err := os.order.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	orderResponse := &orderp.Order{
		Id:     order.ID,
		ItemId: order.ItemId,
		UserId: order.UserId,
	}
	return &orderp.GetOrderResponse{Order: orderResponse}, nil
}

func (os *orderService) GetOrderByUserId(ctx context.Context, req *orderp.GetOrdersByUserId) (*orderp.ListOrdersResponse, error) {
	userId := req.GetUserId()
	if userId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	ordersByUserId, err := os.order.GetOrdersByUserId(ctx, int(userId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get orders of user")
	}

	var ordersResponse []*orderp.Order
	for _, order := range ordersByUserId {
		var o orderp.Order
		err := copier.Copy(&o, order)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to copy to response orderp")
		}

		ordersResponse = append(ordersResponse, &o)
	}

	var items []*cataloguep.Item
	items = append(items, &cataloguep.Item{Id: 1})
	items = append(items, &cataloguep.Item{Id: 2})
	items = append(items, &cataloguep.Item{Id: 3})
	items = append(items, &cataloguep.Item{Id: 4})

	return &orderp.ListOrdersResponse{Orders: ordersResponse}, nil
}
