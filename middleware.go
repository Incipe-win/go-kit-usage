package main

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	"golang.org/x/time/rate"
)

func loggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request any) (any, error) {
			// logger.Log("msg", "calling endpoint")
			// defer logger.Log("msg", "called endpoint", "cast", time.Since(start))
			defer func(start time.Time) {
				logger.Log("cast", time.Since(start))
			}(time.Now())
			return next(ctx, request)
		}
	}
}

var (
	ErrRateLimit = errors.New("rate limit exceeded")
)

func rateMiddleware(limit *rate.Limiter) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request any) (any, error) {
			// 限流逻辑
			if limit.Allow() {
				return next(ctx, request)
			} else {
				return nil, connect.NewError(connect.CodeInternal, ErrRateLimit)
			}
		}
	}
}
