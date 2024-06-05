package user_info

import (
	"auth-service/internal/data/models"
	"auth-service/internal/services/auth"
	"context"
	"errors"
	"github.com/jinzhu/copier"
	authp "github.com/sntabq/proto-gen/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserInfo interface {
	GetUserInfo(ctx context.Context, token string) (*models.User, error)
}

type serverAPI struct {
	authp.UnimplementedUserInfoServer
	userinfo UserInfo
}

func Register(grpcServer *grpc.Server, ui UserInfo) {
	authp.RegisterUserInfoServer(grpcServer, &serverAPI{userinfo: ui})
}

func (s *serverAPI) GetUserInfo(ctx context.Context, in *authp.GetUserInfoRequest) (*authp.GetUserInfoResponse, error) {
	userInfo, err := s.userinfo.GetUserInfo(ctx, in.GetToken())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrNotValidJwt):
			return nil, status.Error(codes.PermissionDenied, "unknown user")
		}
		return nil, err
	}

	var uiResponse authp.User
	err = copier.Copy(&uiResponse, userInfo)
	if err != nil {
		return nil, err
	}
	return &authp.GetUserInfoResponse{User: &uiResponse}, nil
}
