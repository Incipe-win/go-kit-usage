package main

import (
	"addsrv3/proto"
	"addsrv3/proto/protoconnect"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type server struct {
	protoconnect.UnimplementedTrimHandler
}

func (s *server) TrimSpace(ctx context.Context, request *connect.Request[proto.TrimRequest]) (*connect.Response[proto.TrimResponse], error) {
	ov := request.Msg.S
	v := strings.ReplaceAll(ov, " ", "")
	fmt.Printf("Received TrimSpace request: %q, returning: %q\n", ov, v)
	return connect.NewResponse(&proto.TrimResponse{S: v}), nil
}

func TestTrimSpace_Integration(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle(protoconnect.NewTrimHandler(&server{}))

	srv := &http.Server{
		Addr:    ":9999",
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	fmt.Printf("Starting server on %s\n", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	t.Cleanup(func() {
		srv.Close()
	})
}
