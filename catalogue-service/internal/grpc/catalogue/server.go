package catalogueGrpc

import (
	"catalogue-service/internal/data"
	"catalogue-service/internal/data/models"
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	cataloguep "github.com/sntabq/proto-gen/gen/go/catalogue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
)

var ErrInvalidCredentials = errors.New("item not found")

type Catalogue interface {
	CreateItem(
		ctx context.Context,
		item *models.Item,
	) (int32, error)
	ListItems(
		context.Context,
	) ([]*models.Item, error)
	GetItem(
		context.Context,
		int,
	) (*models.Item, error)
}

type catalogueService struct {
	cataloguep.UnimplementedCatalogueServiceServer
	catalogue Catalogue
}

// Register - for registering gRPC server
func Register(gRPCServer *grpc.Server, catalogue Catalogue) {
	cataloguep.RegisterCatalogueServiceServer(gRPCServer, &catalogueService{catalogue: catalogue})
}

func (cs *catalogueService) CreateItem(ctx context.Context, req *cataloguep.CreateItemRequest) (*cataloguep.CreateItemResponse, error) {
	if req.Item.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Item.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}
	if req.Item.Quantity < 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity cannot be negative")
	}
	if req.Item.Price < 0 {
		return nil, status.Error(codes.InvalidArgument, "price cannot be negative")
	}

	// Create a new item
	var item models.Item
	err := copier.Copy(&item, req.Item)
	if err != nil {
		log.Fatalf("failed to copy %v", err)
		return nil, err
	}

	id, err := cs.catalogue.CreateItem(ctx, &item)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrItemAlreadyExist):
			return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("item '%s' already exists", req.Item.Name))
		}
		return nil, status.Error(codes.Internal, "error with create item")
	}

	req.Item.Id = id

	// Return the response
	return &cataloguep.CreateItemResponse{Item: req.Item}, nil
}

func (cs *catalogueService) ListItems(ctx context.Context, req *cataloguep.ListItemsRequest) (*cataloguep.ListItemsResponse, error) {
	var responseItems []*cataloguep.Item

	items, err := cs.catalogue.ListItems(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		var it cataloguep.Item
		err := copier.Copy(&it, &item)
		if err != nil {
			log.Fatalf("failed to copy %v", err)
			return nil, err
		}

		responseItems = append(responseItems, &it)
	}

	return &cataloguep.ListItemsResponse{Items: responseItems}, nil
}

func (cs *catalogueService) GetItem(ctx context.Context, req *cataloguep.GetItemRequest) (*cataloguep.GetItemResponse, error) {
	id, err := strconv.Atoi(req.Id)
	if err != nil {
		return nil, err
	}

	item, err := cs.catalogue.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}

	itemResponse := &cataloguep.Item{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		Quantity:    item.Quantity,
	}
	return &cataloguep.GetItemResponse{Item: itemResponse}, nil
}
