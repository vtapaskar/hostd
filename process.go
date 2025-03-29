package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ProcessStatus represents the current status of a process
type ProcessStatus struct {
	Name          string     `json:"name"`
	CurrentPID    int        `json:"current_pid"`
	PreviousPID   *int       `json:"previous_pid,omitempty"`
	Status        string     `json:"status"`
	LastChange    time.Time  `json:"last_change"`
	MemoryStats   MemoryStats `json:"memory_stats"`
	CurrentMemory int64      `json:"current_memory"` // in bytes
}

// MemoryStats tracks memory usage statistics
type MemoryStats struct {
	MinMemory    int64     `json:"min_memory"`    // in bytes
	MaxMemory    int64     `json:"max_memory"`    // in bytes
	MinTimestamp time.Time `json:"min_timestamp"`
	MaxTimestamp time.Time `json:"max_timestamp"`
}

// ProcessMonitor handles process monitoring
type ProcessMonitor struct {
	processes []Process
	redis     *RedisClient
	logger    *Logger
}

// NewProcessMonitor creates a new process monitor
func NewProcessMonitor(processes []Process, redis *RedisClient, logger *Logger) *ProcessMonitor {
	return &ProcessMonitor{
		processes: processes,
		redis:     redis,
		logger:    logger,
	}
}

// getProcessPID gets the PID of a running process, returns 0 if not running
func (pm *ProcessMonitor) getProcessPID(processName string) (int, error) {
	cmd := exec.Command("pgrep", "-f", processName)
	output, err := cmd.Output()
	if err != nil {
		return 0, nil // Process not running
	}

	// Get first PID if multiple instances are running
	pids := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(pids) == 0 {
		return 0, nil
	}

	pid, err := strconv.Atoi(pids[0])
	if err != nil {
		return 0, fmt.Errorf("invalid PID format: %v", err)
	}

	return pid, nil
}

// getProcessMemory gets the current memory usage of a process in bytes
func (pm *ProcessMonitor) getProcessMemory(pid int) (int64, error) {
	cmd := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error getting memory usage: %v", err)
	}

	// Convert KB to bytes (ps outputs in KB)
	memKB, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing memory value: %v", err)
	}

	return memKB * 1024, nil // Convert KB to bytes
}

// getProcStatus gets the current status from Redis
func (pm *ProcessMonitor) getProcStatus(ctx context.Context, processName string) (*ProcessStatus, error) {
	data, err := pm.redis.GetProcessStatus(ctx, processName)
	if err != nil {
		return &ProcessStatus{
			Name:       processName,
			Status:     "unknown",
			LastChange: time.Now(),
			MemoryStats: MemoryStats{
				MinMemory:    0,
				MaxMemory:    0,
				MinTimestamp: time.Now(),
				MaxTimestamp: time.Now(),
			},
		}, nil
	}

	var status ProcessStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, fmt.Errorf("error parsing status: %v", err)
	}

	return &status, nil
}

// updateProcStatus checks process status and updates Redis
func (pm *ProcessMonitor) updateProcStatus(ctx context.Context, proc Process) {
	currentPID, err := pm.getProcessPID(proc.Name)
	if err != nil {
		pm.logger.Error("Error getting PID for process %s: %v", proc.Name, err)
		return
	}

	// Get current status from Redis
	currentStatus, err := pm.getProcStatus(ctx, proc.Name)
	if err != nil {
		pm.logger.Error("Error getting current status for process %s: %v", proc.Name, err)
		return
	}

	// Determine if status has changed
	status := "down"
	var currentMemory int64 = 0

	if currentPID > 0 {
		status = "up"
		// Get memory usage if process is running
		mem, err := pm.getProcessMemory(currentPID)
		if err != nil {
			pm.logger.Error("Error getting memory usage for process %s: %v", proc.Name, err)
		} else {
			currentMemory = mem
		}
	}

	newStatus := &ProcessStatus{
		Name:          proc.Name,
		CurrentPID:    currentPID,
		Status:        status,
		LastChange:    currentStatus.LastChange,
		MemoryStats:   currentStatus.MemoryStats,
		CurrentMemory: currentMemory,
	}

	// Update status if PID has changed
	if currentPID != currentStatus.CurrentPID {
		if currentStatus.CurrentPID > 0 && currentPID == 0 {
			pm.logger.Critical("Process %s has stopped (previous PID: %d)", proc.Name, currentStatus.CurrentPID)
		} else if currentStatus.CurrentPID == 0 && currentPID > 0 {
			pm.logger.Info("Process %s has started (PID: %d)", proc.Name, currentPID)
		} else {
			pm.logger.Info("Process %s PID changed: %d -> %d", proc.Name, currentStatus.CurrentPID, currentPID)
		}
		newStatus.PreviousPID = &currentStatus.CurrentPID
		newStatus.LastChange = time.Now()
	} else {
		newStatus.PreviousPID = currentStatus.PreviousPID
	}

	// Update memory stats if process is running
	if currentMemory > 0 {
		now := time.Now()
		
		// Initialize memory stats if needed
		if newStatus.MemoryStats.MinMemory == 0 || currentMemory < newStatus.MemoryStats.MinMemory {
			newStatus.MemoryStats.MinMemory = currentMemory
			newStatus.MemoryStats.MinTimestamp = now
			pm.logger.Info("New minimum memory for process %s: %.2f MB", proc.Name, float64(currentMemory)/(1024*1024))
		}
		if currentMemory > newStatus.MemoryStats.MaxMemory {
			newStatus.MemoryStats.MaxMemory = currentMemory
			newStatus.MemoryStats.MaxTimestamp = now
			pm.logger.Info("New maximum memory for process %s: %.2f MB", proc.Name, float64(currentMemory)/(1024*1024))
		}
	}

	// Convert to JSON and update Redis
	statusJSON, err := json.Marshal(newStatus)
	if err != nil {
		pm.logger.Error("Error marshaling status for process %s: %v", proc.Name, err)
		return
	}

	if err := pm.redis.UpdateProcessStatus(ctx, proc.Name, string(statusJSON)); err != nil {
		pm.logger.Error("Error updating Redis for process %s: %v", proc.Name, err)
		return
	}

	pm.logger.Info("Process %s status: %s (PID: %d, Memory: %.2f MB)", 
		proc.Name, status, currentPID, float64(currentMemory)/(1024*1024))
}
