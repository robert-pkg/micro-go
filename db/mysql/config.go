package mysql

import (
	"time"
)

// Config mysql config.
type Config struct {
	DSN         string        `yaml:"dsn"`         // data source name.
	Active      int           `yaml:"active"`      // pool
	Idle        int           `yaml:"idle"`        // pool
	MaxLiefTime time.Duration `yaml:"maxLiefTime"` // connect max life time.
}
