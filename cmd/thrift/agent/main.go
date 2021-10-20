package main

import (
	"context"
	"fmt"
	"github.com/jinzhu/configor"
	"os"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/virago/homework/internal/agent"
	"github.com/virago/homework/pkg/thrift/task"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
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

	var transport thrift.TTransport
	transport = thrift.NewTSocketConf(cfg.ServerAddress, nil)

	transport, err := thrift.NewTBufferedTransportFactory(8192).GetTransport(transport)
	if err != nil {
		return err
	}
	defer transport.Close()
	if err := transport.Open(); err != nil {
		panic(err)
	}
	protocolFactory := thrift.NewTCompactProtocolFactoryConf(&thrift.TConfiguration{})

	ag := agent.NewThriftAgent(
		task.NewTaskServiceClientFactory(transport, protocolFactory),
		logger,
		ratelimit.New(1, ratelimit.Per(4*time.Second)),
	)

	return ag.Loop(context.Background())
}
