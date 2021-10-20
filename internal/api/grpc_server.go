package api

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	pb "github.com/virago/homework/pkg/protobuf/v1"
	"time"
)

type repository interface {
	GetNextTask(ctx context.Context) (GetNextTaskRow, error)
	FinishTask(ctx context.Context, arg FinishTaskParams) error
}

type GRPCServer struct {
	pb.UnimplementedTaskServer
	db repository
}

func NewGRPCServer(db repository) *GRPCServer {
	return &GRPCServer{db: db}
}

func (s GRPCServer) GetNextTask(ctx context.Context, _ *pb.NextTaskRequest) (*pb.NextTaskResponse, error) {
	task, err := s.db.GetNextTask(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.NextTaskResponse{ID: task.ID.String(), Command: task.Command.String}, nil
}

func (s GRPCServer) UpdateTask(ctx context.Context, in *pb.TaskUpdateRequest) (*pb.TaskUpdateResponse, error) {
	layout := "2006-01-02 15:04:05.999999999 -0700 MST"
	startedAt, err := time.Parse(layout, in.StartedAt)
	if err != nil {
		return nil, err
	}
	finishAt, err := time.Parse(layout, in.FinishedAt)
	if err != nil {
		return nil, err
	}

	return &pb.TaskUpdateResponse{}, s.db.FinishTask(ctx, FinishTaskParams{
		ID:         uuid.MustParse(in.ID),
		StartedAt:  sql.NullTime{Time: startedAt, Valid: true},
		FinishedAt: sql.NullTime{Time: finishAt, Valid: true},
		Stdout:     sql.NullString{String: in.StdOut, Valid: true},
		Stderr:     sql.NullString{String: in.StdErr, Valid: true},
		ExitCode:   sql.NullInt32{Int32: in.ExitCode, Valid: true},
		Status:     sql.NullString{String: in.Status, Valid: true},
	})
}
