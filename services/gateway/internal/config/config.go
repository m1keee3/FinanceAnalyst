package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env     string           `yaml:"env" env-default:"local"`
	HTTP    HTTPConfig       `yaml:"http"`
	Scanner GrpcClientConfig `yaml:"scanner"`
	Watcher GrpcClientConfig `yaml:"watcher"`
}

type HTTPConfig struct {
	Port int `yaml:"port" env-default:"8090"`
}

type GrpcClientConfig struct {
	Address string        `yaml:"address"`
	Timeout time.Duration `yaml:"timeout" env-default:"10s"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config file path is empty")
	}

	var cfg Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "config path")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
