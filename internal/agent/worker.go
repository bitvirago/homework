package agent

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Result struct {
	ID         uuid.UUID
	StdErr     string
	StdOut     string
	StartedAt  time.Time
	FinishedAt time.Time
	ExitCode   int32
}

func commandExec(ID uuid.UUID, Command string) (*Result, error) {
	command := strings.Split(Command, " ")
	if len(command) < 1 {
		return nil, errors.New(fmt.Sprintf("Invalid command: %s", Command))
	}
	startAt := time.Now().UTC()
	cmd := exec.Command(command[0], command[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	var exitCode int
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	return &Result{
		ID:         ID,
		StdErr:     stderr.String(),
		StdOut:     stdout.String(),
		ExitCode:   int32(exitCode),
		FinishedAt: time.Now().UTC(),
		StartedAt:  startAt,
	}, nil
}
