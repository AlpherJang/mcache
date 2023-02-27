package main

import (
	"context"
	"github.com/AlpherJang/mcache/pkg/proto"
	"github.com/AlpherJang/mcache/pkg/rpc"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"log"

	"github.com/AlpherJang/mcache/pkg/handler"
	"github.com/gin-gonic/gin"
)

// main function provide an endpoint for cache service
func main() {
	gin.ForceConsoleColor()
	router := gin.Default()
	handler.Register(router)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	go func() {
		lis, err := net.Listen("tcp", ":8081")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		var opts []grpc.ServerOption
		grpcServer := grpc.NewServer(opts...)
		proto.RegisterCacheRpcServiceServer(grpcServer, rpc.NewServer())
		grpcServer.Serve(lis)
	}()
	// close server with signal, set timeout context about 5 seconds
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
