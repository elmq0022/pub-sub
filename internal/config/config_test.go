package config

import (
	"testing"
	"time"
)

func TestNewConfigUsesDefaults(t *testing.T) {
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", cfg.Port)
	}
	if cfg.HeartbeatTickInterval != 30*time.Second {
		t.Fatalf(
			"expected default heartbeat tick interval %v, got %v",
			30*time.Second,
			cfg.HeartbeatTickInterval,
		)
	}
	if cfg.HeartbeatTimeout != 90*time.Second {
		t.Fatalf(
			"expected default heartbeat timeout %v, got %v",
			90*time.Second,
			cfg.HeartbeatTimeout,
		)
	}
}

func TestNewConfigUsesEnvOverrides(t *testing.T) {
	t.Setenv("PUBSUB_PORT", "9090")
	t.Setenv("PUBSUB_HEARTBEAT_TICK_INTERVAL", "5s")
	t.Setenv("PUBSUB_HEARTBEAT_TIMEOUT", "12s")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Fatalf("expected overridden port 9090, got %q", cfg.Port)
	}
	if cfg.HeartbeatTickInterval != 5*time.Second {
		t.Fatalf(
			"expected overridden heartbeat tick interval %v, got %v",
			5*time.Second,
			cfg.HeartbeatTickInterval,
		)
	}
	if cfg.HeartbeatTimeout != 12*time.Second {
		t.Fatalf(
			"expected overridden heartbeat timeout %v, got %v",
			12*time.Second,
			cfg.HeartbeatTimeout,
		)
	}
}

func TestNewConfigReturnsErrorForInvalidDuration(t *testing.T) {
	t.Setenv("PUBSUB_HEARTBEAT_TICK_INTERVAL", "not-a-duration")

	_, err := NewConfig()
	if err == nil {
		t.Fatal("expected NewConfig to fail for invalid duration")
	}
}
