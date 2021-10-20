package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/configor"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/virago/homework/internal/api"
	pb "github.com/virago/homework/pkg/protobuf/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type config struct {
	DSN string `env:"DSN" required:"true"`
}

func main() {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("who watches the watchmen (%v)?", err))
	}
	logger := zapLog.Sugar()

	if err := run(logger); err != nil {
		logger.Errorf("%v", err)
		os.Exit(1)
	}
}

func run(logger *zap.SugaredLogger) error {
	defer logger.Sync()
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return nil
	}
	var cfg config
	if err := configor.Load(&cfg); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	db, err := sqlx.Connect("pgx", cfg.DSN)
	if err != nil {
		return err
	}
	repo := api.New(db)
	handler, err := api.NewHandler(repo)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	httpServer := &http.Server{
		Addr:        ":8080",
		Handler:     handler,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	s := grpc.NewServer()
	pb.RegisterTaskServer(s, api.NewGRPCServer(repo))
	log.Printf("server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	httpServer.RegisterOnShutdown(cancel)
	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	<-signalChan
	log.Print("os.Interrupt - shutting down...\n")

	go func() {
		<-signalChan
		log.Fatal("os.Kill - terminating...\n")
	}()

	gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(gracefullCtx); err != nil {
		log.Printf("shutdown error: %v\n", err)
		defer os.Exit(1)
		return nil
	} else {
		log.Printf("gracefully stopped\n")
	}

	return nil
}
