package services

import (
	tp "github.com/rafa-mori/gobe/internal/scheduler/types"
)

// ICronService define os métodos necessários para interagir com o serviço de cronjobs.
type ICronService interface {
	// GetScheduledCronJobs retorna os cronjobs agendados para execução.
	GetScheduledCronJobs() ([]tp.IJob, error)
}

// CronService implements the ICronService interface to fetch scheduled cronjobs from the database.
type CronService struct {
	db tp.Database // Assume a Database interface is defined elsewhere for database operations.
}

// NewCronService creates a new instance of CronService.
func NewCronService(db tp.Database) ICronService {
	return &CronService{db: db}
}

// GetScheduledCronJobs fetches the scheduled cronjobs from the database.
func (s *CronService) GetScheduledCronJobs() ([]tp.IJob, error) {
	// Example query to fetch cronjobs. Adjust the query and mapping as per your database schema.
	rows, err := s.db.Query("SELECT id, name, schedule, command FROM cronjobs WHERE active = true")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []tp.IJob
	for rows.Next() {
		var jobID int
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
