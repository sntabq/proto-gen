package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	auth "github.com/sntabq/proto-gen/gen/go/auth"
	catalogue "github.com/sntabq/proto-gen/gen/go/catalogue"
	order "github.com/sntabq/proto-gen/gen/go/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
)

func main() {

	mux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			// Extract headers and return as metadata
			md := metadata.Pairs("authorization", req.Header.Get("Authorization"))
			log.Printf("Extracted metadata: %v", md)
			return md
		}),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := auth.RegisterAuthHandlerFromEndpoint(context.Background(), mux, "localhost:44044", opts)
	if err != nil {
		panic(err)
	}

	err = catalogue.RegisterCatalogueServiceHandlerFromEndpoint(context.Background(), mux, "localhost:44045", opts)
	if err != nil {
		panic(err)
	}

	err = order.RegisterOrderServiceHandlerFromEndpoint(context.Background(), mux, "localhost:44046", opts)
	if err != nil {
		panic(err)
	}

	handler := cors.Default().Handler(mux)

	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		panic(err)
	}
}
