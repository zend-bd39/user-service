package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-service/bootstrap"
	"user-service/config"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()
	lis, err := net.Listen("tcp", ":" + cfg.AppConfig.Port)
	if err != nil {
		log.Fatal(err)
	}
	app := bootstrap.NewApp(cfg)
	if err:= app.SetupGRPC(ctx); err != nil {
		log.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Printf("gRPC server running on port:%s", cfg.AppConfig.Port)
		if err := app.Run(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("failed to run gRPC server: %v", err)
		}
		
	}()
	s := <-sig
	log.Printf("gRPC server stop due to %s", s.String())

	
	shutdownCtx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()

	stopChan := make(chan struct{})
	go func() {
		log.Println("gRPC server graceful stopped")
		app.GracefulStop()
		close(stopChan)
	}()
	select {
	case <-shutdownCtx.Done():
		log.Println("gRPC server stopped because time out")
		app.Stop()
	case <- stopChan:
		log.Println("gRPC server and resources successfully stopped gracefully")
	}
	log.Println("gRPC server closed.. bye")
}