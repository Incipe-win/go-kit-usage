package main

import (
	"addsrv3/proto"
	protogateway "addsrv3/proto/gateway"
	"addsrv3/proto/protoconnect"
	"context"
	"net/http"

	"github.com/go-kit/log"
	"golang.org/x/time/rate"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"connectrpc.com/connect"
	grpctransport "github.com/go-kit/kit/transport/grpc"
)

type grpcServer struct {
	protoconnect.UnimplementedAddHandler
	sum    grpctransport.Handler
	concat grpctransport.Handler
}

func NewGRPCServer(srv AddService, logger log.Logger) protoconnect.AddHandler {
	sum := makeSumEndpoint(srv)
	sum = loggingMiddleware(log.With(logger, "method", "sum"))(sum)
	sum = rateMiddleware(rate.NewLimiter(1, 1))(sum)

	concat := makeConcatEndpoint(srv)
	concat = loggingMiddleware(log.With(logger, "method", "concat"))(concat)
	concat = rateMiddleware(rate.NewLimiter(1, 1))(concat)
	return &grpcServer{
		sum: grpctransport.NewServer(
			sum,
			decodeGRPCSumRequest,
			encodeGRPCSumResponse,
		),
		concat: grpctransport.NewServer(
			concat,
			decodeGRPCConcatRequest,
			encodeGRPCConcatResponse,
		),
	}
}

func (s *grpcServer) Sum(ctx context.Context, request *connect.Request[proto.SumRequest]) (*connect.Response[proto.SumResponse], error) {
	_, resp, err := s.sum.ServeGRPC(ctx, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return resp.(*connect.Response[proto.SumResponse]), nil
}

func (s *grpcServer) Concat(ctx context.Context, request *connect.Request[proto.ConcatRequest]) (*connect.Response[proto.ConcatResponse], error) {
	_, resp, err := s.concat.ServeGRPC(ctx, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return resp.(*connect.Response[proto.ConcatResponse]), nil
}

// gRPC的请求与响应
// decodeGRPCSumRequest 将Sum方法的gRPC请求参数转为内部的SumRequest
func decodeGRPCSumRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*connect.Request[proto.SumRequest])
	return SumRequest{
		A: int(req.Msg.A),
		B: int(req.Msg.B),
	}, nil
}

// decodeGRPCConcatRequest 将Concat方法的gRPC请求参数转为内部的ConcatRequest
func decodeGRPCConcatRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*connect.Request[proto.ConcatRequest])
	return ConcatRequest{
		A: req.Msg.A,
		B: req.Msg.B,
	}, nil
}

// encodeGRPCSumResponse 封装Sum的gRPC响应
func encodeGRPCSumResponse(_ context.Context, response any) (any, error) {
	resp := response.(SumResponse)
	return connect.NewResponse(&proto.SumResponse{Result: int64(resp.Result), Error: resp.Err}), nil
}

// encodeGRPCConcatResponse 封装Concat的gRPC响应
func encodeGRPCConcatResponse(_ context.Context, response any) (any, error) {
	resp := response.(ConcatResponse)
	return connect.NewResponse(&proto.ConcatResponse{Result: resp.Result, Error: resp.Err}), nil
}

func NewHTTPServer(gatewayAddr string) http.Handler {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption("*", &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{UseProtoNames: true},
			},
		}),
	)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err := protogateway.RegisterAddHandlerFromEndpoint(context.Background(), mux, gatewayAddr, opts)
	if err != nil {
		panic("failed to register gRPC gateway: " + err.Error())
	}
	return mux
}

// // encodeTrimRequest 将内部结构体转为protobuf中的结构体
// // 对外发起gRPC请求
// func encodeTrimRequest(_ context.Context, request any) (any, error) {
// 	req := request.(TrimRequest)
// 	return connect.NewRequest(&proto.TrimRequest{S: req.S}), nil
// }

// // decodeTrimResponse 将收到的gRPC响应转为内部的响应结构体
// func decodeTrimResponse(_ context.Context, response any) (any, error) {
// 	resp := response.(*connect.Response[proto.TrimResponse])
// 	return &TrimResponse{S: resp.Msg.S}, nil
// }
