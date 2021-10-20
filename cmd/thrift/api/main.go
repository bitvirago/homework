package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jinzhu/configor"
	"github.com/jmoiron/sqlx"
	"github.com/virago/homework/internal/api"
	"github.com/virago/homework/pkg/thrift/task"
	"go.uber.org/zap"
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

	transportFactory := thrift.NewTBufferedTransportFactory(8192)
	protocolFactory := thrift.NewTCompactProtocolFactoryConf(&thrift.TConfiguration{})

	transport, err := thrift.NewTServerSocket(":9092")
	processor := task.NewTaskServiceProcessor(api.NewThriftServer(repo))
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	log.Printf("Starting simple thrift server on: %s\n", ":9092")
	go func() {
		if err := server.Serve(); err != nil {
			logger.Fatalf("thrift server ListenAndServe: %v", err)
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
