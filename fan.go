package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// FanMetrics represents the metrics for a fan
type FanMetrics struct {
	Speed     int    `json:"speed"`
	Duty      int    `json:"duty"`
	Timestamp string `json:"timestamp"`
}

// Fan represents a cooling fan
type Fan struct {
	name      string
	logger    *Logger
	redis     *RedisClient
	speed     int // RPM
	duty      int // Percentage
	isPresent bool
	instance  int
}

// NewFan creates a new Fan instance
func NewFan(name string, instance int, logger *Logger, redis *RedisClient) *Fan {
	return &Fan{
		name:      name,
		logger:    logger,
		redis:     redis,
		isPresent: true, // Initially assume fan is present
		instance:  instance,
	}
}

func (f *Fan) getName() string {
	return fmt.Sprintf("%s-%d", f.name, f.instance)
}

func (f *Fan) getStatus(ctx context.Context) (FruStatus, error) {
	if !f.isPresent {
		return FruStatusRed, fmt.Errorf("fan %d not present", f.instance)
	}

	if err := f.updateMetrics(ctx); err != nil {
		return FruStatusRed, fmt.Errorf("failed to update fan %d metrics", f.instance)
	}

	// Example thresholds - adjust based on actual requirements
	if f.speed < 100 { // Fan almost stopped
		return FruStatusRed, nil
	}
	if f.duty > 90 { // Fan working too hard
		return FruStatusYellow, nil
	}
	return FruStatusGreen, nil
}

func (f *Fan) updateMetrics(ctx context.Context) error {
	// In a real implementation, this would read from hardware
	// For now, using example values
	f.speed = 2000 // 2000 RPM
	f.duty = 60    // 60% duty cycle

	// Create metrics structure
	metrics := FanMetrics{
		Speed:     f.speed,
		Duty:      f.duty,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Convert metrics to JSON
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		f.logger.Error("Failed to marshal fan %d metrics: %v", f.instance, err)
		return err
	}

	// Store metrics in Redis
	key := fmt.Sprintf("hardware:fan:%d:metrics", f.instance)
	if err := f.redis.client.Set(ctx, key, string(metricsJSON), 0).Err(); err != nil {
		f.logger.Error("Failed to store fan %d metrics in Redis: %v", f.instance, err)
		return err
	}

	f.logger.Info("Updated fan %d metrics: Speed=%dRPM, Duty=%d%%",
		f.instance, f.speed, f.duty)
	return nil
}

func (f *Fan) available() bool {
	return f.isPresent
}

func (f *Fan) setInstance(instance int) {
	f.instance = instance
	f.logger.Info("Set fan instance to %d", instance)
}
