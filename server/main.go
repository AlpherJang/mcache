package main

import (
	"context"
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
	// @todo init base config
	// @todo init server config
	// @todo shutdown server with notify
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
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
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
