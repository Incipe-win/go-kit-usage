package main

import (
	"addsrv3/proto"
	"addsrv3/proto/protoconnect"
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	apiconsul "github.com/hashicorp/consul/api"
)

const serviceName = "trim_service"

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
	cc, err := NewConsulClient("localhost:8500")
	if err != nil {
		t.Fatalf("Failed to create Consul client: %v", err)
	}
	ipInfo, err := getOutboundIP()
	if err != nil {
		t.Fatalf("Failed to get outbound IP: %v", err)
	}
	if err := cc.RegisterService(serviceName, ipInfo.String(), 9999); err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

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
		cc.DeregisterService(serviceName + "-" + ipInfo.String() + "-9999")
	})
}

type consulClient struct {
	client *apiconsul.Client
}

func NewConsulClient(consulAddr string) (*consulClient, error) {
	cfg := apiconsul.DefaultConfig()
	cfg.Address = consulAddr
	client, err := apiconsul.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &consulClient{client: client}, nil
}

func (c *consulClient) RegisterService(serviceName, ip string, port int) error {
	srv := &apiconsul.AgentServiceRegistration{
		ID:      serviceName + "-" + ip + "-" + strconv.Itoa(port),
		Name:    serviceName,
		Tags:    []string{"hchao", "trim"},
		Address: ip,
		Port:    port,
	}
	return c.client.Agent().ServiceRegister(srv)
}

func (c *consulClient) DeregisterService(serviceID string) error {
	return c.client.Agent().ServiceDeregister(serviceID)
}

// getOutboundIP 获取本机的出口IP
func getOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
