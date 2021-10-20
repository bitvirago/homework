package agent

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/virago/homework/pkg/protobuf/v1"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

type GRPCAgent struct {
	client      pb.TaskClient
	logger      *zap.SugaredLogger
	rateLimiter ratelimit.Limiter
}

func NewGRPCAgent(client pb.TaskClient, logger *zap.SugaredLogger, rateLimiter ratelimit.Limiter) *GRPCAgent {
	return &GRPCAgent{client: client, logger: logger, rateLimiter: rateLimiter}
}

func (a GRPCAgent) Loop(ctx context.Context) error {
	for {
		a.rateLimiter.Take()
		response, err := a.client.GetNextTask(ctx, nil)
		if err != nil {
			a.logger.Errorf("client fetch next task: %v", err)
			continue
		}

		res, err := commandExec(uuid.MustParse(response.ID), response.Command)
		if err != nil {
			a.logger.Errorf("run job: %v", err)
		}

		_, rErr := a.client.UpdateTask(ctx, &pb.TaskUpdateRequest{
			ID:         res.ID.String(),
			StartedAt:  res.StartedAt.String(),
			FinishedAt: res.FinishedAt.String(),
			StdErr:     res.StdErr,
			StdOut:     res.StdOut,
			ExitCode:   res.ExitCode,
			Status:     "finished",
		})
		if rErr != nil {
			a.logger.Errorf("update task: %v", err)
		}
	}
}
