package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// PSUMetrics represents the metrics for a PSU
type PSUMetrics struct {
	Voltage   float64 `json:"voltage"`
	Current   float64 `json:"current"`
	Power     float64 `json:"power"`
	Timestamp string  `json:"timestamp"`
}

// PSU represents a Power Supply Unit
type PSU struct {
	name      string
	logger    *Logger
	redis     *RedisClient
	voltage   float64
	current   float64
	power     float64
	isPresent bool
	instance  int
}

// NewPSU creates a new PSU instance
func NewPSU(name string, instance int, logger *Logger, redis *RedisClient) *PSU {
	return &PSU{
		name:      name,
		logger:    logger,
		redis:     redis,
		isPresent: true, // Initially assume PSU is present
		instance:  instance,
	}
}

func (p *PSU) getName() string {
	return fmt.Sprintf("%s-%d", p.name, p.instance)
}

func (p *PSU) getStatus(ctx context.Context) (FruStatus, error) {
	if !p.isPresent {
		return FruStatusRed, fmt.Errorf("PSU %d not present", p.instance)
	}

	if err := p.updateMetrics(ctx); err != nil {
		return FruStatusRed, fmt.Errorf("failed to update PSU %d metrics", p.instance)
	}

	// Example thresholds - adjust based on actual requirements
	if p.voltage < 10.8 || p.voltage > 13.2 { // ±10% of 12V
		return FruStatusRed, nil
	}
	if p.power > 800 { // Example: 800W threshold
		return FruStatusYellow, nil
	}
	return FruStatusGreen, nil
}

func (p *PSU) updateMetrics(ctx context.Context) error {
	// In a real implementation, this would read from hardware
	// For now, using example values
	p.voltage = 12.0  // 12V
	p.current = 50.0  // 50A
	p.power = 600.0   // 600W

	// Create metrics structure
	metrics := PSUMetrics{
		Voltage:   p.voltage,
		Current:   p.current,
		Power:     p.power,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Convert metrics to JSON
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		p.logger.Error("Failed to marshal PSU %d metrics: %v", p.instance, err)
		return err
	}

	// Store metrics in Redis
	key := fmt.Sprintf("hardware:psu:%d:metrics", p.instance)
	if err := p.redis.client.Set(ctx, key, string(metricsJSON), 0).Err(); err != nil {
		p.logger.Error("Failed to store PSU %d metrics in Redis: %v", p.instance, err)
		return err
	}

	p.logger.Info("Updated PSU %d metrics: Voltage=%.2fV, Current=%.2fA, Power=%.2fW",
		p.instance, p.voltage, p.current, p.power)
	return nil
}

func (p *PSU) available() bool {
	return p.isPresent
}

func (p *PSU) setInstance(instance int) {
	p.instance = instance
	p.logger.Info("Set PSU instance to %d", instance)
}
