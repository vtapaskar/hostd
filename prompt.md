# Hardware Monitoring System Tasks

## 1. Hardware Interface Implementation
Created a `HardwareInterface` in `hardwareIf.go` with methods:
- `getName() string`: Returns component name with instance
- `getStatus(ctx context.Context) (FruStatus, error)`: Returns operational status
- `updateMetrics(ctx context.Context) error`: Updates and stores metrics
- `available() bool`: Checks component availability
- `setInstance(instance int)`: Sets instance number

## 2. Component Implementations

### Network Processing Unit (NPU)
- Monitors network processing metrics:
  - Packet rate (packets per second)
  - Throughput (Gbps)
  - Buffer usage (%)
  - Processor usage (%)
- Status thresholds:
  - Red: Buffer or processor usage > 95%
  - Yellow: Buffer usage > 80% or processor usage > 85%

### Power Supply Unit (PSU)
- Monitors power metrics:
  - Voltage (V)
  - Current (A)
  - Power (W)
- Status thresholds:
  - Red: Voltage < 10.8V or > 13.2V
  - Yellow: Power > 800W

### Fan
- Monitors cooling metrics:
  - Speed (RPM)
  - Duty cycle (%)
- Status thresholds:
  - Red: Speed < 100 RPM
  - Yellow: Duty > 90%

## 3. Redis Integration
- Each component stores metrics in Redis
- Key format: `hardware:{type}:{instance}:metrics`
- Metrics include timestamp in RFC3339 format
- JSON format for each component type

### Example Redis Data Structures
```json
// NPU Metrics
{
  "packet_rate": 1000000.0,
  "throughput": 40.0,
  "buffer_usage": 60.0,
  "processor_usage": 70.0,
  "timestamp": "2025-03-24T04:39:59-07:00"
}

// PSU Metrics
{
  "voltage": 12.0,
  "current": 50.0,
  "power": 600.0,
  "timestamp": "2025-03-24T04:39:59-07:00"
}

// Fan Metrics
{
  "speed": 2000,
  "duty": 60,
  "timestamp": "2025-03-24T04:39:59-07:00"
}
```

## 4. Instance Support
- Added instance tracking to all components
- Components are identified by name-instance format (e.g., "NPU-0", "PSU-1")
- Instance number included in:
  - Component names
  - Error messages
  - Log messages
  - Redis keys

## 5. Original Task Prompts

1. "Create a main.go file"
   - Initial setup of the host daemon project

2. "Create a hardware interface"
   - Request to create the base hardware monitoring interface
   - Added methods for monitoring hardware components

3. "Create NPU implementation"
   - Initially created as Neural Processing Unit
   - Later corrected to Network Processing Unit
   - Added metrics and status monitoring

4. "Create PSU implementation"
   - Added Power Supply Unit monitoring
   - Implemented voltage, current, and power metrics

5. "Create Fan implementation"
   - Added cooling fan monitoring
   - Implemented speed and duty cycle metrics

6. "Add instance support"
   - Request to add instance tracking to hardware components
   - Modified constructors and methods to handle instances

7. "Store metrics in Redis"
   - Request to update metrics storage to use Redis
   - Added Redis client and JSON serialization

8. "NPU stands for Network Processing Unit"
   - Correction of NPU implementation
   - Updated metrics to focus on network processing

9. "go run"
   - Attempt to run the implementation
   - Found missing Go installation

## Next Steps
1. Install Go development environment
2. Run and test the implementation
3. Consider adding:
   - More hardware types
   - Metrics history
   - Alerting system
   - Web interface for monitoring
