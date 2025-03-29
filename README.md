# hostd

A Go-based host daemon project that monitors specified processes and reports their status to Redis.

## Features

- Process monitoring with configurable parameters
- Redis integration for status reporting
- JSON-based configuration
- Redis pub/sub for process control
- Graceful shutdown handling

## Prerequisites

- Go 1.21 or higher
- Redis server
- Processes to monitor (as specified in processes.json)

## Configuration

The application uses two configuration files:

### config.json
```json
{
    "redis": {
        "host": "localhost",
        "port": 6379,
        "password": "",
        "db": 0
    }
}
```

### processes.json
```json
{
    "processes": [
        {
            "name": "nginx",
            "restart": true,
            "maxRetries": 3
        },
        {
            "name": "redis-server",
            "restart": true,
            "maxRetries": 3
        }
    ]
}
```

## Running

```bash
go run main.go
```

## Redis Keys

The application stores process status in Redis using the following key pattern:
- `process:{process_name}:status` - Contains either "up" or "down"

## Redis Pub/Sub Commands

The application subscribes to the `hostd:commands` channel for process control. Send commands in JSON format:

```json
{
    "action": "start|stop|restart",
    "process": "process_name"
}
```

### Example Commands

Start a process:
```bash
redis-cli PUBLISH hostd:commands '{"action":"start","process":"nginx"}'
```

Stop a process:
```bash
redis-cli PUBLISH hostd:commands '{"action":"stop","process":"nginx"}'
```

Restart a process:
```bash
redis-cli PUBLISH hostd:commands '{"action":"restart","process":"nginx"}'
``` 
