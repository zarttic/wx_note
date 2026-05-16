package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nefu-dev/wx-note/internal/api"
	"github.com/nefu-dev/wx-note/internal/repository"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8100"
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("create data directory: %v", err)
	}

	db, err := repository.InitDB(dataDir)
	if err != nil {
		log.Fatalf("database init failed: %v", err)
	}

	handler := api.NewHandler(db)
	router := handler.Setup()

	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动 server（goroutine 中运行，不阻塞）
	go func() {
		log.Printf("wx_note server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 等待中断信号，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("received signal %v — shutting down...", sig)

	// 给正在处理的请求最多 10 秒完成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}

	// 关闭数据库连接
	if err := db.Close(); err != nil {
		log.Printf("database close error: %v", err)
	}

	log.Println("server exited cleanly")
}
