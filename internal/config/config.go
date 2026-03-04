package config

import (
	"fmt"
	"os"
	"time"
)

const (
	defaultPort                  = "8080"
	defaultHeartbeatTickInterval = 30 * time.Second
	defaultHeartbeatTimeout      = 90 * time.Second
)

type Config struct {
	Port                  string
	HeartbeatTickInterval time.Duration
	HeartbeatTimeout      time.Duration
}

func NewConfig() (Config, error) {
	heartbeatTickInterval, err := envDuration(
		"PUBSUB_HEARTBEAT_TICK_INTERVAL",
		defaultHeartbeatTickInterval,
	)
	if err != nil {
		return Config{}, err
	}

	heartbeatTimeout, err := envDuration(
		"PUBSUB_HEARTBEAT_TIMEOUT",
		defaultHeartbeatTimeout,
	)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:                  envString("PUBSUB_PORT", defaultPort),
		HeartbeatTickInterval: heartbeatTickInterval,
		HeartbeatTimeout:      heartbeatTimeout,
	}, nil
}

func envString(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return duration, nil
}
