package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Redis RedisConfig `json:"redis"`
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type ProcessConfig struct {
	Processes []Process `json:"processes"`
}

type Process struct {
	Name       string `json:"name"`
	Restart    bool   `json:"restart"`
	MaxRetries int    `json:"maxRetries"`
}

type Command struct {
	Action  string `json:"action"`  // start, stop, restart
	Process string `json:"process"` // process name
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

func loadProcessConfig(filename string) (*ProcessConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading process config file: %v", err)
	}

	var config ProcessConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing process config file: %v", err)
	}

	return &config, nil
}

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(config *RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %v", err)
	}

	return &RedisClient{client: client}, nil
}

func (r *RedisClient) Close() {
	r.client.Close()
}

func (r *RedisClient) UpdateProcessStatus(ctx context.Context, processName string, status string) error {
	return r.client.Set(ctx, fmt.Sprintf("process:%s:status", processName), status, 0).Err()
}

func (r *RedisClient) SubscribeToCommands(ctx context.Context, handler func(ctx context.Context, cmd Command) error) {
	pubsub := r.client.Subscribe(ctx, "hostd:commands")
	defer pubsub.Close()

	// Wait for confirmation that subscription is created before publishing anything
	_, err := pubsub.Receive(ctx)
	if err != nil {
		log.Printf("Error receiving subscription confirmation: %v", err)
		return
	}

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			var cmd Command
			if err := json.Unmarshal([]byte(msg.Payload), &cmd); err != nil {
				log.Printf("Error parsing command: %v", err)
				continue
			}

			if err := handler(ctx, cmd); err != nil {
				log.Printf("Error handling command: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	// Initialize logger
	logger, err := NewLogger()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Load configurations
	config, err := loadConfig("config.json")
	if err != nil {
		logger.Critical("Failed to load config: %v", err)
		os.Exit(1)
	}

	processConfig, err := loadProcessConfig("processes.json")
	if err != nil {
		logger.Critical("Failed to load process config: %v", err)
		os.Exit(1)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to Redis
	redisClient, err := NewRedisClient(&config.Redis)
	if err != nil {
		logger.Critical("Failed to connect to Redis: %v", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Create process monitor
	processMonitor := NewProcessMonitor(processConfig.Processes, redisClient, logger)

	// Create and start periodic runner
	periodicRunner := NewPeriodicRunner(processMonitor, logger)
	periodicRunner.Start(ctx)

	logger.Info("Host daemon started")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Cancel context to stop all goroutines
	logger.Info("Shutting down...")
	cancel()
	
	// Wait for periodic tasks to complete
	periodicRunner.Wait()
	
	logger.Info("Shutdown complete")
}
