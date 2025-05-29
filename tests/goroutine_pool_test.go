package tests

// import (
// 	"log"
// 	"sync/atomic"
// 	"testing"
// 	"time"

// 	"github.com/rafa-mori/gobe/internal/scheduler/pipeline"
// )

// type MockJob struct {
// 	executed *int32
// }

// func (m *MockJob) Run() error {
// 	atomic.AddInt32(m.executed, 1)
// 	time.Sleep(100 * time.Millisecond) // Simula trabalho
// 	return nil
// }

// func (m *MockJob) Retry() error {
// 	return m.Run()
// }

// func (m *MockJob) Cancel() error {
// 	return nil
// }

// func TestGoroutinePool(t *testing.T) {
// 	executed := int32(0)
// 	pool := pipeline.NewGoroutinePool(5)
// 	mockJob := &MockJob{executed: &executed}

// 	// Inicia o pool com monitoramento resiliente
// 	go pool.StartWithResilientMonitoring(10, 50)

// 	// Submete 20 jobs
// 	for i := 0; i < 20; i++ {
// 		pool.Submit(mockJob)
// 	}

// 	// Aguarda a conclusão dos jobs
// 	pool.Stop()

// 	if executed != 20 {
// 		t.Errorf("Expected 20 jobs to be executed, but got %d", executed)
// 	}

// 	log.Println("Test completed successfully")
// }
