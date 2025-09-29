// Package types defines the Job structure and its interface for scheduling tasks.
package types

import (
	"fmt"

	"github.com/google/uuid"

	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
)

type IJob interface {
	Mu() *t.Mutexes
	Ref() *t.Reference
	GetUserID() uuid.UUID
	Run() error
	Retry() error
	Cancel() error
}

type Job struct {
	*t.Mutexes
	*t.Reference

	ID       int
	Name     string
	Schedule string
	Command  string

	userID uuid.UUID
	Status JobStatus // Adicionado para rastrear o status do job
}

func NewJob(id int, name, schedule, command string) IJob {
	return &Job{
		ID:       id,
		Name:     name,
		Schedule: schedule,
		Command:  command,
	}
}

func (j *Job) Mu() *t.Mutexes {
	return j.Mutexes
}
func (j *Job) Ref() *t.Reference {
	return j.Reference
}
func (j *Job) GetUserID() uuid.UUID {
	return j.userID
}
func (j *Job) Run() error {
	gl.Log("info", fmt.Sprintf("Running job: %s (ID: %d)", j.Name, j.ID))
	// Implement the logic to execute the command.
	return nil
}
func (j *Job) Retry() error {
	gl.Log("info", fmt.Sprintf("Retrying job: %s (ID: %d)", j.Name, j.ID))

	return nil
}
func (j *Job) Cancel() error {
	gl.Log("info", fmt.Sprintf("Cancelling job: %s (ID: %d)", j.Name, j.ID))
	// Implement cancel logic.
	return nil
}
