package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type taskHandler interface {
	GetTask(ctx context.Context, id uuid.UUID) (Task, error)
	GetTasks(ctx context.Context) ([]Task, error)
	CreateTask(ctx context.Context, arg CreateTaskParams) error
}

type api struct {
	db taskHandler
}

type apiError struct {
}

func NewHandler(db taskHandler) (http.Handler, error) {
	api := api{db}
	r := chi.NewRouter()
	r.Route("/tasks", func(r chi.Router) {
		r.Get("/", api.listTasks)
		r.Post("/", api.createTasks)
		r.Get("/{taskID}", api.getTasksByID)

	})

	return r, nil
}

func (a api) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := a.db.GetTasks(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (a api) createTasks(w http.ResponseWriter, r *http.Request) {
	var command struct {
		Command string `json:"command"`
	}

	err := json.NewDecoder(r.Body).Decode(&command)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if command.Command == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = a.db.CreateTask(
		r.Context(),
		CreateTaskParams{
			ID:      uuid.New(),
			Command: sql.NullString{String: command.Command, Valid: true},
		},
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a api) getTasksByID(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	task, err := a.db.GetTask(r.Context(), uuid.MustParse(taskID))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
