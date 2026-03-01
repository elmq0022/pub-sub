package broker

import "time"

type BrokerConfig struct {
	HeartbeatTickInterval time.Duration
	HeartbeatTimeout      time.Duration
}

func NewBrokerConfig(heartbeatTickInterval, heartbeatTimeout time.Duration) BrokerConfig {
	return BrokerConfig{
		HeartbeatTickInterval: heartbeatTickInterval,
		HeartbeatTimeout:      heartbeatTimeout,
	}
}
