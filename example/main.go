package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"connectrpc.com/connect"
	pingv1 "connectrpc.com/connect/internal/gen/connect/ping/v1"
	"connectrpc.com/connect/internal/gen/simple/connect/ping/v1/pingv1connect"
)

type PingServer struct {
	pingv1connect.UnimplementedPingServiceHandler // returns errors from all methods
}

func (ps *PingServer) Ping(ctx context.Context, req *pingv1.PingRequest) (*pingv1.PingResponse, error) {
	callInfo, _ := connect.CallInfoForHandlerContext(ctx)
	callInfo.ResponseHeader().Set("x-custom-key", "value")

	if req.Number < 0 {
		return nil, connect.NewError(connect.CodeUnimplemented, nil)
	}

	return &pingv1.PingResponse{Number: req.Number}, nil
}

func serve() {
	mux := http.NewServeMux()
	mux.Handle(pingv1connect.NewPingServiceHandler(&PingServer{}))
	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)
	s := &http.Server{
		Addr:      "localhost:8080",
		Handler:   mux,
		Protocols: p,
	}

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}

	go s.Serve(ln)
}

func request(ctx context.Context, number int64) string {
	client := pingv1connect.NewPingServiceClient(
		http.DefaultClient,
		"http://localhost:8080/",
	)

	ctx, callInfo := connect.NewClientContext(ctx)
	req := &pingv1.PingRequest{
		Number: number,
	}
	_, _ = client.Ping(ctx, req)

	return callInfo.ResponseHeader().Get("x-custom-key")
}

func main() {
	serve()

	ctx := context.Background()

	// found x-custom-key in response header
	log.Println("response header  ok:", request(ctx, 1))
	// not found
	log.Println("response header err:", request(ctx, -1))
}
