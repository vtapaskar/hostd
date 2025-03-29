package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// NPUMetrics represents the metrics for a Network Processing Unit
type NPUMetrics struct {
	PacketRate     float64 `json:"packet_rate"`     // Packets per second
	Throughput     float64 `json:"throughput"`      // Gbps
	BufferUsage    float64 `json:"buffer_usage"`    // Percentage of buffer usage
	ProcessorUsage float64 `json:"processor_usage"` // NPU processor utilization
	Timestamp      string  `json:"timestamp"`
}

// NPU represents a Network Processing Unit
type NPU struct {
	name           string
	logger         *Logger
	redis          *RedisClient
	packetRate     float64 // Packets per second
	throughput     float64 // Gbps
	bufferUsage    float64 // Percentage
	processorUsage float64 // Percentage
	isPresent      bool
	instance       int
}

// NewNPU creates a new Network Processing Unit instance
func NewNPU(name string, instance int, logger *Logger, redis *RedisClient) *NPU {
	return &NPU{
		name:      name,
		logger:    logger,
		redis:     redis,
		isPresent: true, // Initially assume NPU is present
		instance:  instance,
	}
}

func (n *NPU) getName() string {
	return fmt.Sprintf("%s-%d", n.name, n.instance)
}

func (n *NPU) getStatus(ctx context.Context) (FruStatus, error) {
	if !n.isPresent {
		return FruStatusRed, fmt.Errorf("NPU %d not present", n.instance)
	}

	if err := n.updateMetrics(ctx); err != nil {
		return FruStatusRed, fmt.Errorf("failed to update NPU %d metrics", n.instance)
	}

	// Example thresholds for network processing metrics
	if n.bufferUsage > 95 || n.processorUsage > 95 { // Critical resource exhaustion
		return FruStatusRed, nil
	}
	if n.bufferUsage > 80 || n.processorUsage > 85 { // High resource utilization
		return FruStatusYellow, nil
	}
	return FruStatusGreen, nil
}

func (n *NPU) updateMetrics(ctx context.Context) error {
	// In a real implementation, this would read from hardware
	// For now, using example values
	n.packetRate = 1000000.0  // 1M packets per second
	n.throughput = 40.0       // 40 Gbps
	n.bufferUsage = 60.0      // 60% buffer usage
	n.processorUsage = 70.0   // 70% NPU processor utilization

	// Create metrics structure
	metrics := NPUMetrics{
		PacketRate:     n.packetRate,
		Throughput:     n.throughput,
		BufferUsage:    n.bufferUsage,
		ProcessorUsage: n.processorUsage,
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	// Convert metrics to JSON
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		n.logger.Error("Failed to marshal NPU %d metrics: %v", n.instance, err)
		return err
	}

	// Store metrics in Redis
	key := fmt.Sprintf("hardware:npu:%d:metrics", n.instance)
	if err := n.redis.client.Set(ctx, key, string(metricsJSON), 0).Err(); err != nil {
		n.logger.Error("Failed to store NPU %d metrics in Redis: %v", n.instance, err)
		return err
	}

	n.logger.Info("Updated NPU %d metrics: PacketRate=%.1f pps, Throughput=%.1f Gbps, BufferUsage=%.1f%%, ProcessorUsage=%.1f%%",
		n.instance, n.packetRate, n.throughput, n.bufferUsage, n.processorUsage)
	return nil
}

func (n *NPU) available() bool {
	return n.isPresent
}

func (n *NPU) setInstance(instance int) {
	n.instance = instance
	n.logger.Info("Set NPU instance to %d", instance)
}
