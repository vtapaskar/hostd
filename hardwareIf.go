package main

import "context"

// FruStatus represents the operational status of hardware
type FruStatus string

const (
	// FruStatusGreen indicates normal operation
	FruStatusGreen FruStatus = "green"
	// FruStatusYellow indicates warning condition
	FruStatusYellow FruStatus = "yellow"
	// FruStatusRed indicates critical condition
	FruStatusRed FruStatus = "red"
)

// HardwareInterface defines methods for hardware monitoring
type HardwareInterface interface {
	// getName returns the name of the hardware component
	getName() string

	// getStatus returns the current operational status of the hardware
	// Returns: green (normal), yellow (warning), or red (critical)
	getStatus(ctx context.Context) (FruStatus, error)

	// updateMetrics updates the hardware metrics
	// Returns: true if metrics were successfully updated, false otherwise
	updateMetrics(ctx context.Context) bool

	// available checks if the hardware is available for monitoring
	// Returns: true if hardware is available, false otherwise
	available() bool

	// setInstance sets the instance number for the hardware component
	setInstance(instance int)
}
