package grpcapp

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/jinzhu/copier"
	authp "github.com/sntabq/proto-gen/gen/go/auth"
	orderv1 "github.com/sntabq/proto-gen/gen/go/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	"strings"
)

func InterceptorCreateOrder(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod != "/order.OrderService/CreateOrder" {
		return handler(ctx, req)
	}
	createOrderRequest := req.(*orderv1.CreateOrderRequest)

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("failed to get metadata from context")
	}
	tkn, found := md["authorization"]
	if !found && len(tkn) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authentication is required")
	}

	userInfo, err := UserInfoServiceClient.GetUserInfo(ctx, &authp.GetUserInfoRequest{Token: tkn[0]})
	if err != nil {
		log.Printf("failed to get user info from sso service")
		return nil, status.Errorf(codes.Internal, "failed to get user info from sso service")
	}
	if createOrderRequest.Order.UserId != userInfo.User.Id {
		return nil, status.Errorf(codes.InvalidArgument, "you can not create an order for others")
	}

	return handler(ctx, req)
}

func InterceptorGetOrdersOfUser(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod != "/order.OrderService/GetOrderByUserId" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("failed to get metadata from context")
	}
	tkn, found := md["authorization"]
	if !found && len(tkn) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authentication is required")
	}

	userInfo, err := UserInfoServiceClient.GetUserInfo(ctx, &authp.GetUserInfoRequest{Token: tkn[0]})
	if err != nil {
		log.Printf("failed to get user info from sso service")
		return nil, status.Errorf(codes.Internal, "failed to get user info from sso service")
	}

	var request orderv1.GetOrdersByUserId
	err = copier.Copy(&request, req)
	if err != nil {
		return nil, err
	}
	if userInfo.User.Id == request.UserId {
		return handler(ctx, req)
	}

	isAdminRequest := &authp.IsAdminRequest{UserId: int64(userInfo.User.Id)}
	isAdminResponse, err := AuthServiceClient.IsAdmin(ctx, isAdminRequest)
	if err != nil {
		log.Printf("permissions fail %v", err)
		return nil, status.Errorf(codes.PermissionDenied, "permission failed")
	}

	if !isAdminResponse.IsAdmin {
		return nil, status.Errorf(codes.PermissionDenied, "permission failed")
	}

	return handler(ctx, req)
}

func AdminInterceptorGetAllOrders(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod != "/order.OrderService/ListOrders" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("failed to get metadata from context")
	}
	tkn, found := md["authorization"]
	if !found || len(tkn) == 0 || tkn[0] == "" {
		return nil, status.Errorf(codes.PermissionDenied, "lack of permission")
	}

	userInfo, err := UserInfoServiceClient.GetUserInfo(ctx, &authp.GetUserInfoRequest{Token: tkn[0]})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "unknown user"):
			return nil, status.Errorf(codes.InvalidArgument, "invalid user")
		default:
			return nil, status.Errorf(codes.Internal, "failed to get user info from sso service")
		}
	}

	isAdminRequest := &authp.IsAdminRequest{UserId: int64(userInfo.User.Id)}
	isAdminResponse, err := AuthServiceClient.IsAdmin(ctx, isAdminRequest)
	if err != nil {
		log.Printf("permissions fail %v", err)
		return nil, status.Errorf(codes.Internal, "permission failed")
	}

	if !isAdminResponse.IsAdmin {
		return nil, status.Errorf(codes.Internal, "permission failed")
	}

	return handler(ctx, req)
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
