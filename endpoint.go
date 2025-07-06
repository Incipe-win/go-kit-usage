package main

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// 请求和响应
type SumRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

type SumResponse struct {
	Result int    `json:"result"`
	Err    string `json:"error,omitempty"`
}

type ConcatRequest struct {
	A string `json:"a"`
	B string `json:"b"`
}

type ConcatResponse struct {
	Result string `json:"result"`
	Err    string `json:"error,omitempty"`
}

// 2 Endpoints
func makeSumEndpoint(s AddService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(SumRequest)
		v, err := s.Sum(ctx, req.A, req.B)
		if err != nil {
			return SumResponse{Result: v, Err: err.Error()}, nil
		}
		return SumResponse{Result: v}, nil
	}
}

func makeConcatEndpoint(s AddService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ConcatRequest)
		v, err := s.Concat(ctx, req.A, req.B)
		if err != nil {
			return ConcatResponse{Result: v, Err: err.Error()}, nil
		}
		return ConcatResponse{Result: v}, nil
	}
}
