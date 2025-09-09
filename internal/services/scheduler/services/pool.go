package services

import (
	"log"
	"sync"
	"time"

	m "github.com/kubex-ecosystem/gobe/internal/services/scheduler/monitor"

	tp "github.com/kubex-ecosystem/gobe/internal/services/scheduler/types"
)

// GoroutinePool gerencia a execução de tarefas usando um pool de goroutines.
// Ele suporta monitoramento avançado e ações automáticas para resiliência.
//
// Métodos principais:
// - Start: Inicia o pool de goroutines sem monitoramento.
// - StartWithMonitoring: Inicia o pool com monitoramento básico.
// - StartWithEnhancedMonitoring: Adiciona limites configuráveis e alertas.
// - StartWithResilientMonitoring: Inclui ações automáticas para reiniciar o pool.
// - Submit: Adiciona uma tarefa ao pool.
// - Stop: Para o pool de goroutines.
// - Restart: Reinicia o pool de forma segura.
//
// Exemplo de uso:
//
// pool := NewGoroutinePool(5)
// pool.StartWithResilientMonitoring(100, 500)
// pool.Submit(myJob)
// pool.Stop()
//
// Parâmetros de monitoramento:
// - maxGoroutines: Limite máximo de goroutines antes de acionar alertas.
// - maxHeapMB: Limite máximo de memória heap em MB antes de acionar alertas.
//
// Ações automáticas:
// - Reinício do pool ao exceder limites configurados.
// - Logs detalhados para rastreamento de ações e métricas.
type GoroutinePool struct {
	maxWorkers int
	jobs       chan tp.IJob
	wg         sync.WaitGroup
}

func NewGoroutinePool(maxWorkers int) *GoroutinePool {
	return &GoroutinePool{
		maxWorkers: maxWorkers,
		jobs:       make(chan tp.IJob),
	}
}

func (p *GoroutinePool) Start() {
	for i := 0; i < p.maxWorkers; i++ {
		go func() {
			for job := range p.jobs {
				if err := job.Run(); err != nil {
					log.Printf("Job failed: %v", err)
				}
				p.wg.Done()
			}
		}()
	}
}

func (p *GoroutinePool) StartWithMonitoring() {
	for i := 0; i < p.maxWorkers; i++ {
		go func(workerID int) {
			for job := range p.jobs {
				start := time.Now()
				if err := job.Run(); err != nil {
					log.Printf("Worker %d: Job failed: %v", workerID, err)
				}
				duration := time.Since(start)
				log.Printf("Worker %d: Job completed in %v", workerID, duration)
				p.wg.Done()
			}
		}(i)
	}

	// Monitorando goroutines e memória
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics := m.GetMetrics()
			log.Printf("Monitoring: Goroutines: %d, Heap: %.2f MB", metrics.Goroutines, metrics.HeapMB)
		}
	}()
}

// Aprimorando o monitoramento com limites configuráveis e alertas

func (p *GoroutinePool) StartWithEnhancedMonitoring(maxGoroutines int, maxHeapMB float64) {
	for i := 0; i < p.maxWorkers; i++ {
		go func(workerID int) {
			for job := range p.jobs {
				start := time.Now()
				if err := job.Run(); err != nil {
					log.Printf("Worker %d: Job failed: %v", workerID, err)
				}
				duration := time.Since(start)
				log.Printf("Worker %d: Job completed in %v", workerID, duration)
				p.wg.Done()
			}
		}(i)
	}

	// Monitorando goroutines e memória com limites configuráveis
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics := m.GetMetrics()
			if metrics.Goroutines > maxGoroutines {
				log.Printf("ALERT: Goroutines exceeded limit! Current: %d, Limit: %d", metrics.Goroutines, maxGoroutines)
			}
			if metrics.HeapMB > maxHeapMB {
				log.Printf("ALERT: Heap memory exceeded limit! Current: %.2f MB, Limit: %.2f MB", metrics.HeapMB, maxHeapMB)
			}
			log.Printf("Monitoring: Goroutines: %d, Heap: %.2f MB", metrics.Goroutines, metrics.HeapMB)
		}
	}()
}

// Aprimorando alertas com ações automáticas para resiliência

func (p *GoroutinePool) StartWithResilientMonitoring(maxGoroutines int, maxHeapMB float64) {
	for i := 0; i < p.maxWorkers; i++ {
		go func(workerID int) {
			for job := range p.jobs {
				start := time.Now()
				if err := job.Run(); err != nil {
					log.Printf("Worker %d: Job failed: %v", workerID, err)
				}
				duration := time.Since(start)
				log.Printf("Worker %d: Job completed in %v", workerID, duration)
				p.wg.Done()
			}
		}(i)
	}

	// Monitorando goroutines e memória com ações automáticas
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			metrics := m.GetMetrics()
			if metrics.Goroutines > maxGoroutines {
				log.Printf("ALERT: Goroutines exceeded limit! Current: %d, Limit: %d", metrics.Goroutines, maxGoroutines)
				log.Println("ACTION: Restarting GoroutinePool to recover...")
				p.Restart()
			}
			if metrics.HeapMB > maxHeapMB {
				log.Printf("ALERT: Heap memory exceeded limit! Current: %.2f MB, Limit: %.2f MB", metrics.HeapMB, maxHeapMB)
				log.Println("ACTION: Restarting GoroutinePool to recover...")
				p.Restart()
			}
			log.Printf("Monitoring: Goroutines: %d, Heap: %.2f MB", metrics.Goroutines, metrics.HeapMB)
		}
	}()
}

// Restart reinicia o pool de goroutines
func (p *GoroutinePool) Restart() {
	log.Println("Restarting GoroutinePool...")
	p.Stop()
	p.jobs = make(chan tp.IJob, cap(p.jobs)) // Recria o canal com o mesmo buffer
	p.StartWithResilientMonitoring(100, 500) // Valores padrão para reinício
}

func (p *GoroutinePool) Submit(job tp.IJob) {
	p.wg.Add(1)
	p.jobs <- job
}

func (p *GoroutinePool) Stop() {
	close(p.jobs)
	p.wg.Wait()
}
