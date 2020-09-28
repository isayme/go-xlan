package conf

import (
	"sync"

	"github.com/isayme/go-config"
	"github.com/isayme/go-logger"
)

type Logger struct {
	Level string `json:"level" yaml:"level"`
}

type ServerConfig struct {
	Addr   string `json:"addr" yaml:"addr"`
	Port   int    `json:"port" yaml:"port"`
	Secret string `json:"secret" yaml:"secret"`
}

type ClientConfig struct {
	Server ServerConfig `json:"server" yaml:"server"`
}

type ServiceConfig struct {
	Name       string `json:"name" yaml:"name"`
	LocalIP    string `json:"local_ip" yaml:"local_ip"`
	LocalPort  int    `json:"local_port" yaml:"local_port"`
	RemotePort int    `json:"remote_port" yaml:"remote_port"`
}

type Config struct {
	Logger Logger `json:"logger" yaml:"logger"`

	Server ServerConfig `json:"server" yaml:"server"`
	Client ClientConfig `json:"client" yaml:"client"`

	Services []ServiceConfig `json:"services" yaml:"services"`
}

var once sync.Once
var globalConfig Config

// Get parse config
func Get() *Config {
	config.Parse(&globalConfig)

	once.Do(func() {
		logger.SetLevel(globalConfig.Logger.Level)
	})

	return &globalConfig
}
