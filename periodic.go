package main

import (
	"context"
	"sync"
	"time"
)

// PeriodicRunner handles periodic tasks
type PeriodicRunner struct {
	monitor    *ProcessMonitor
	logger     *Logger
	wg         sync.WaitGroup
	lastCheck  time.Time
	checkMutex sync.Mutex
}

// NewPeriodicRunner creates a new periodic runner
func NewPeriodicRunner(monitor *ProcessMonitor, logger *Logger) *PeriodicRunner {
	return &PeriodicRunner{
		monitor: monitor,
		logger:  logger,
	}
}

// Start begins the periodic execution
func (pr *PeriodicRunner) Start(ctx context.Context) {
	pr.wg.Add(1)
	go pr.run(ctx)
}

// Wait waits for all periodic tasks to complete
func (pr *PeriodicRunner) Wait() {
	pr.wg.Wait()
}

// run executes the periodic tasks
func (pr *PeriodicRunner) run(ctx context.Context) {
	defer pr.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case currentTime := <-ticker.C:
			pr.checkMutex.Lock()
			if currentTime.Sub(pr.lastCheck) >= time.Minute {
				pr.logger.Info("Running periodic process check at %v", currentTime.Format(time.RFC3339))
				
				// Run process monitoring
				for _, proc := range pr.monitor.processes {
					pr.monitor.updateProcStatus(ctx, proc)
				}
				
				pr.lastCheck = currentTime
			}
			pr.checkMutex.Unlock()
		}
	}
}
