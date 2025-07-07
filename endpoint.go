package main

import (
	"addsrv3/proto"
	"addsrv3/proto/protoconnect"
	"context"

	"connectrpc.com/connect"
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

type TrimRequest struct {
	S string `json:"trim_request"`
}

type TrimResponse struct {
	S string `json:"trim_response"`
}

// 2 Endpoints
func makeSumEndpoint(s AddService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(SumRequest)
		v, err := s.Sum(ctx, req.A, req.B)
		if err != nil {
			return SumResponse{Result: v, Err: err.Error()}, err
		}
		return SumResponse{Result: v}, nil
	}
}

func makeConcatEndpoint(s AddService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(ConcatRequest)
		v, err := s.Concat(ctx, req.A, req.B)
		if err != nil {
			return ConcatResponse{Result: v, Err: err.Error()}, err
		}
		return ConcatResponse{Result: v}, nil
	}
}

// makeTrimEndpoint 客户端endpoint
// 不是直接的提供服务，而是请求其他服务
func makeTrimEndpoint(client protoconnect.TrimClient) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(TrimRequest)
		resp, err := client.TrimSpace(ctx, connect.NewRequest(&proto.TrimRequest{S: req.S}))
		if err != nil {
			return nil, err
		}
		return TrimResponse{S: resp.Msg.S}, nil
	}
}
