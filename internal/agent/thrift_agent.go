package agent

import (
	"context"

	"github.com/google/uuid"
	"github.com/virago/homework/pkg/thrift/task"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

type ThriftAgent struct {
	client      task.TaskService
	logger      *zap.SugaredLogger
	rateLimiter ratelimit.Limiter
}

func NewThriftAgent(client task.TaskService, logger *zap.SugaredLogger, rateLimiter ratelimit.Limiter) *ThriftAgent {
	return &ThriftAgent{client: client, logger: logger, rateLimiter: rateLimiter}
}

func (a ThriftAgent) Loop(ctx context.Context) error {
	for {
		a.rateLimiter.Take()
		response, err := a.client.GetNextTask(ctx)
		if err != nil {
			a.logger.Errorf("client fetch next task: %v", err)
			continue
		}

		res, err := commandExec(uuid.MustParse(response.ID), response.Command)
		if err != nil {
			a.logger.Errorf("run job: %v", err)
		}

		rErr := a.client.UpdateTask(ctx, &task.TaskUpdateRequest{
			ID:         res.ID.String(),
			StartedAt:  res.StartedAt.String(),
			FinishedAt: res.FinishedAt.String(),
			Stderr:     res.StdErr,
			Stdout:     res.StdOut,
			ExitCode:   res.ExitCode,
			Status:     "finished",
		})
		if rErr != nil {
			a.logger.Errorf("update task: %v", err)
		}
	}
}
