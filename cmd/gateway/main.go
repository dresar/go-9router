package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dresar/go-9router/internal/config"
	"github.com/dresar/go-9router/internal/handlers"
	"github.com/dresar/go-9router/internal/logging"
	"github.com/dresar/go-9router/internal/router"
	"github.com/dresar/go-9router/internal/server"
	"github.com/dresar/go-9router/internal/storage"
)

func main() {
	cfg := config.Load()

	if cfg.JWTSecret == "" {
		logging.Warn("BOOT", "JWT_SECRET is not set — dashboard auth will fail. Set JWT_SECRET in .env")
	}

	db, err := storage.Open(cfg.DataDir)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer storage.Close()

	if err := storage.Migrate(db); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	h := &handlers.Handler{DB: db, Cfg: cfg}
	mux := router.New(h, cfg.JWTSecret)
	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	srv := server.New(addr, mux, cfg.ReadHeaderTimeout, cfg.IdleTimeout)

	go func() {
		if err := srv.Start(); err != nil {
			if err.Error() != "http: Server closed" {
				log.Fatalf("server error: %v", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logging.Info("SERVER", "shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logging.Error("SERVER", "shutdown error", "err", err)
	}
	logging.Info("SERVER", "stopped")
}
