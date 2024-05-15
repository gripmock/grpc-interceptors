package grpc_interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type TimeoutCallOption struct {
	grpc.EmptyCallOption

	forcedTimeout time.Duration
}

func WithForcedTimeout(forceTimeout time.Duration) TimeoutCallOption {
	return TimeoutCallOption{forcedTimeout: forceTimeout} //nolint:exhaustruct
}

func getForcedTimeout(callOptions []grpc.CallOption) (time.Duration, bool) {
	for _, opt := range callOptions {
		if co, ok := opt.(TimeoutCallOption); ok {
			return co.forcedTimeout, true
		}
	}

	return 0, false
}

func UnaryTimeoutInterceptor(t time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) error {
		timeout := t
		if v, ok := getForcedTimeout(opts); ok {
			timeout = v
		}

		if timeout <= 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamTimeoutInterceptor(t time.Duration) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string,
		streamer grpc.Streamer, opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		timeout := t
		if v, ok := getForcedTimeout(opts); ok {
			timeout = v
		}

		if timeout <= 0 {
			return streamer(ctx, desc, cc, method, opts...)
		}

		ctx, _ = context.WithTimeout(ctx, timeout) //nolint:govet

		return streamer(ctx, desc, cc, method, opts...)
	}
}
