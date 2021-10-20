package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_listTasks(t *testing.T) {
	repo := &mockTaskHandler{}
	id := uuid.New()
	r, err := NewHandler(repo)
	require.NoError(t, err)
	ts := httptest.NewServer(r)
	defer ts.Close()
	data := []Task{
		{
			ID:      id,
			Command: sql.NullString{String: "ls -la", Valid: true},
		},
		{
			ID:      id,
			Command: sql.NullString{String: "ps aux", Valid: true},
		},
	}

	repo.On("GetTasks", mock.Anything).Return(
		data,
		nil,
	).Times(1)

	j, err := json.Marshal(data)
	require.NoError(t, err)
	apitest.New().
		Handler(r).
		Get("/tasks").
		Expect(t).
		Body(string(j)).
		Status(http.StatusOK).
		End()

	repo.On("GetTasks", mock.Anything).Return(
		nil,
		errors.New("some error"),
	).Times(1)
	apitest.New().
		Handler(r).
		Get("/tasks").
		Expect(t).
		Body(`some error`).
		Status(http.StatusOK).
		End()
}
