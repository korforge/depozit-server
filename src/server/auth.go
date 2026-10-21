package server

import (
	"context"
	"crypto/subtle"
	"depozit-server/src/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func checkAPIKey(ctx context.Context, apiKey string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	tokens := md.Get("api-key")
	if len(tokens) == 0 {
		return status.Error(codes.Unauthenticated, "missing api-key")
	}

	if subtle.ConstantTimeCompare([]byte(tokens[0]), []byte(apiKey)) != 1 {
		return status.Error(codes.Unauthenticated, "invalid api-key")
	}

	return nil
}

func AuthInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := checkAPIKey(ctx, cfg.APIKey); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func AuthStreamInterceptor(cfg *config.Config) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := checkAPIKey(ss.Context(), cfg.APIKey); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}
