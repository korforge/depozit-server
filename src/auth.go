package main

import (
	"context"
	"depozit-server/src/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md.Get("api-key")
		if len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing api-key")
		}

		token := tokens[0]

		if token != cfg.APIKey {
			return nil, status.Error(codes.Unauthenticated, "invalid api-key")
		}

		return handler(ctx, req)
	}
}
