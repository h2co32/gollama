package config

import "time"

type ConfigProfile struct {
	MaxRetries    int
	Timeout       time.Duration
	RateLimit     int
	ModelSettings map[string]interface{}
}

var DefaultProfile = ConfigProfile{
	MaxRetries: 3,
	Timeout:    5 * time.Second,
	RateLimit:  10,
	ModelSettings: map[string]interface{}{
		"temperature": 0.7,
		"max_tokens":  1024,
	},
}

var ProductionProfile = ConfigProfile{
	MaxRetries: 5,
	Timeout:    10 * time.Second,
	RateLimit:  100,
	ModelSettings: map[string]interface{}{
		"temperature": 0.5,
		"max_tokens":  2048,
	},
}
