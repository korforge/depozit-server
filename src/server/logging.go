package server

import (
	"context"
	"depozit-server/src/config"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func logRequest(ctx context.Context, method string, start time.Time, err error) {
	addr := ""

	if p, ok := peer.FromContext(ctx); ok {
		addr = p.Addr.String()
	}

	config.Logger.Info("request",
		"method", method,
		"peer", addr,
		"code", status.Code(err).String(),
		"duration", time.Since(start),
	)
}

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		logRequest(ctx, info.FullMethod, start, err)

		return resp, err
	}
}

func LoggingStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		err := handler(srv, ss)

		logRequest(ss.Context(), info.FullMethod, start, err)

		return err
	}
}
