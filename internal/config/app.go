package config

import "time"

// App holds application-level config.
type App struct {
	Name    string        `env:"NAME" envDefault:"silicon-brief"`
	Env     string        `env:"ENV" envDefault:"development"`
	Timeout time.Duration `env:"TIMEOUT" envDefault:"30s"`
}
