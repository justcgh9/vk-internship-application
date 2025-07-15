package config

import (
	"time"

	config "github.com/justcgh9/go-config"
)

type Config struct {
	
	Server struct {
		Host 	string 			`yaml:"host"`
		Port 	string 			`yaml:"port"`
		Timeout time.Duration	`yaml:"timeout"`
		IdleTimeout time.Duration	`yaml:"idle_timeout"`
	} `yaml:"server"`

	DatabaseURI string `yaml:"db_uri"`
}

func MustLoad() *Config {
	return config.MustLoad[Config]()
}