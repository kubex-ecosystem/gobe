// Package types defines the Job structure and its interface for scheduling tasks.
package types

import (
	"fmt"

	"github.com/google/uuid"

	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type Job interface {
	Mu() *t.Mutexes
	Ref() *t.Reference
	GetUserID() uuid.UUID
	Run() error
	Retry() error
	Cancel() error
}

type JobImpl struct {
	*t.Mutexes
	*t.Reference

	ID       uuid.UUID
	Name     string
	Schedule string
	Command  string

	userID uuid.UUID
	Status JobStatus // Adicionado para rastrear o status do job
}

func NewJobImpl(id uuid.UUID, name, schedule, command string) *JobImpl {
	return &JobImpl{
		ID:       id,
		Name:     name,
		Schedule: schedule,
		Command:  command,
	}
}

func NewJob(id uuid.UUID, name, schedule, command string) Job {
	return NewJobImpl(id, name, schedule, command)
}

func (j *JobImpl) Mu() *t.Mutexes {
	return j.Mutexes
}
func (j *JobImpl) Ref() *t.Reference {
	return j.Reference
}
func (j *JobImpl) GetUserID() uuid.UUID {
	return j.userID
}
func (j *JobImpl) Run() error {
	gl.Log("info", fmt.Sprintf("Running job: %s (ID: %d)", j.Name, j.ID))
	// Implement the logic to execute the command.
	return nil
}
func (j *JobImpl) Retry() error {
	gl.Log("info", fmt.Sprintf("Retrying job: %s (ID: %d)", j.Name, j.ID))

	return nil
}
func (j *JobImpl) Cancel() error {
	gl.Log("info", fmt.Sprintf("Cancelling job: %s (ID: %d)", j.Name, j.ID))
	// Implement cancel logic.
	return nil
}
