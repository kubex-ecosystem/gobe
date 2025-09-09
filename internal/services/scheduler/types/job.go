package types

import (
	"log"

	"github.com/google/uuid"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
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
	log.Printf("Running job: %s (ID: %d)", j.Name, j.ID)
	// Implement the logic to execute the command.
	return nil
}
func (j *Job) Retry() error {
	log.Printf("Retrying job: %s (ID: %d)", j.Name, j.ID)

	return nil
}
func (j *Job) Cancel() error {
	log.Printf("Cancelling job: %s (ID: %d)", j.Name, j.ID)
	// Implement cancel logic.
	return nil
}
