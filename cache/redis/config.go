package redis

import (
	"time"
)

// Config .
type Config struct {
	Address   string `yaml:"addr"`
	Password  string `yaml:"pass"`
	MaxActive int    `yaml:"active"`
	MaxIdle   int    `yaml:"idle"`
	DBNumber  int    `yaml:"db_number"`

	IdleTimeout time.Duration `yaml:"idleTimeout"`
}
