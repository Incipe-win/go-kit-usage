package main

import (
	"addsrv3/proto"
	"addsrv3/proto/protoconnect"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/go-kit/log"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc/test/bufconn"
	"gopkg.in/go-playground/assert.v1"
)

const bufSize = 1024 * 1024

func TestProtoService(t *testing.T) {
	listener := bufconn.Listen(bufSize)

	srv := NewService()

	// 初始化带logger的service
	logger := log.NewJSONLogger(os.Stdout)
	srv = NewLogMiddleware(logger, srv)

	server := NewGRPCServer(srv, logger)

	path, handler := protoconnect.NewAddHandler(server)
	mux := http.NewServeMux()
	mux.Handle(path, handler)

	httpServer := &http.Server{
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}
	go func() {
		if err := httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
			return
		}
	}()
	t.Cleanup(func() {
		httpServer.Shutdown(context.Background())
		listener.Close()
	})

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return listener.Dial()
			},
		},
	}

	// http://localhost 随便写都无所谓的
	client := protoconnect.NewAddClient(httpClient, "http://localhost")

	t.Run("Sum", func(t *testing.T) {
		request := &proto.SumRequest{
			A: 3,
			B: 5,
		}
		resp, err := client.Sum(context.Background(), connect.NewRequest(request))
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(8), resp.Msg.Result)
	})

	t.Run("Concat", func(t *testing.T) {
		request := &proto.ConcatRequest{
			A: "Hello, ",
			B: "Connect!",
		}
		resp, err := client.Concat(context.Background(), connect.NewRequest(request))
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "Hello, Connect!", resp.Msg.Result)
	})
}
