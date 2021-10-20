package api

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	"github.com/virago/homework/pkg/thrift/task"
	"time"
)

type ThriftServer struct {
	db repository
}

func NewThriftServer(db repository) *ThriftServer {
	return &ThriftServer{db: db}
}

func (t ThriftServer) GetNextTask(ctx context.Context) (*task.TaskResponse, error) {
	ta, err := t.db.GetNextTask(ctx)
	if err != nil {
		return nil, err
	}

	return &task.TaskResponse{ID: ta.ID.String(), Command: ta.Command.String}, nil
}

func (t ThriftServer) UpdateTask(ctx context.Context, updateRequest *task.TaskUpdateRequest) error {
	layout := "2006-01-02 15:04:05.999999999 -0700 MST"
	startedAt, err := time.Parse(layout, updateRequest.StartedAt)
	if err != nil {
		return err
	}
	finishAt, err := time.Parse(layout, updateRequest.FinishedAt)
	if err != nil {
		return err
	}

	return t.db.FinishTask(ctx, FinishTaskParams{
		ID:         uuid.MustParse(updateRequest.ID),
		StartedAt:  sql.NullTime{Time: startedAt, Valid: true},
		FinishedAt: sql.NullTime{Time: finishAt, Valid: true},
		Stdout:     sql.NullString{String: updateRequest.Stdout, Valid: true},
		Stderr:     sql.NullString{String: updateRequest.Stderr, Valid: true},
		ExitCode:   sql.NullInt32{Int32: updateRequest.ExitCode, Valid: true},
		Status:     sql.NullString{String: updateRequest.Status, Valid: true},
	})
}
