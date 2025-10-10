// Package services provides implementations for various services in the application.
package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	tp "github.com/kubex-ecosystem/gobe/internal/services/scheduler/types"
)

// ICronService define os métodos necessários para interagir com o serviço de cronjobs.
type ICronService interface {
	// GetScheduledCronJobs retorna os cronjobs agendados para execução.
	GetScheduledCronJobs(context.Context) ([]tp.Job, error)
}

// CronServiceImpl implements the ICronService interface to fetch scheduled cronjobs from the database.
type CronServiceImpl struct {
	db tp.DirectDatabase // Assume a Database interface is defined elsewhere for database operations.
}

// NewCronService creates a new instance of CronService.
func NewCronService(db tp.DirectDatabase) ICronService {
	return &CronServiceImpl{db: db}
}

// GetScheduledCronJobs fetches the scheduled cronjobs from the database.
func (s *CronServiceImpl) GetScheduledCronJobs(ctx context.Context) ([]tp.Job, error) {
	// Example query to fetch cronjobs. Adjust the query and mapping as per your database schema.
	rws, err := s.db.Query(ctx, "SELECT id, name, schedule, command FROM cronjobs WHERE active = true")
	if err != nil {
		return nil, err
	}
	rows, ok := rws.(tp.Rows)
	if !ok {
		return nil, fmt.Errorf("failed to assert rows type")
	}

	defer rows.Close()

	var jobs []tp.Job
	for rows.Next() {
		var jobID uuid.UUID
		var name, schedule, command string
		if err := rows.Scan(&jobID, &name, &schedule, &command); err != nil {
			return nil, err
		}

		// Create a concrete implementation of IJob for each row.
		job := tp.NewJob(jobID, name, schedule, command)
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}
