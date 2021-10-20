package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/configor"
	"github.com/virago/homework/internal/agent"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	pb "github.com/virago/homework/pkg/protobuf/v1"
)

type config struct {
	ServerAddress string `env:"SERVER_ADDRESS" required:"true"`
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
	var cfg config
	if err := configor.Load(&cfg); err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	conn, err := grpc.Dial(cfg.ServerAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	ag := agent.NewGRPCAgent(
		pb.NewTaskClient(conn),
		logger,
		ratelimit.New(1, ratelimit.Per(4*time.Second)),
	)

	return ag.Loop(context.Background())
}
