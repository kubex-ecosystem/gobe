package manager

import (
	"log"
	"time"

	pl "github.com/rafa-mori/gobe/internal/scheduler/services"
)

// CronJobScheduler gerencia a execução de cronjobs usando o GoroutinePool.
type CronJobScheduler struct {
	pool         *pl.GoroutinePool
	ICronService pl.ICronService // Interface para interagir com o serviço de cronjobs
}

// NewCronJobScheduler cria uma nova instância do CronJobScheduler.
func NewCronJobScheduler(pool *pl.GoroutinePool, ICronService pl.ICronService) *CronJobScheduler {
	return &CronJobScheduler{
		pool:         pool,
		ICronService: ICronService,
	}
}

// Start inicia o loop de verificação e execução de cronjobs.
func (s *CronJobScheduler) Start() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Verifica os cronjobs a cada minuto
		defer ticker.Stop()
		for range ticker.C {
			cronJobs, err := s.ICronService.GetScheduledCronJobs()
			if err != nil {
				log.Printf("Error fetching scheduled cronjobs: %v", err)
				continue
			}
			for _, job := range cronJobs {
				s.pool.Submit(job)
			}
		}
	}()
}
