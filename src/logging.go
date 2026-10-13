package main

import (
	"context"
	"depozit-server/src/config"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		addr := ""

		if p, ok := peer.FromContext(ctx); ok {
			addr = p.Addr.String()
		}

		config.Logger.Info("request",
			"method", info.FullMethod,
			"peer", addr,
			"code", status.Code(err).String(),
			"duration", time.Since(start),
		)

		return resp, err
	}
}
