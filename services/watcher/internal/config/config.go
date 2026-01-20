package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env        string           `yaml:"env" env-default:"development"`
	App        AppConfig        `yaml:"app"`
	Grpc       GrpcConfig       `yaml:"grpc"`
	GrpcClient GrpcClientConfig `yaml:"grpc_client"`
	Kafka      KafkaConfig      `yaml:"kafka"`
	Scheduler  SchedulerConfig  `yaml:"scheduler"`
}

type AppConfig struct {
	CandleOptions ScanCandleOptions `yaml:"scan_candle_opts"`
	ChartOptions  ScanChartOptions  `yaml:"scan_chart_opts"`
	SearchFrom    time.Time         `yaml:"search_from"`
	SearchTo      time.Time         `yaml:"search_to"`
	Tickers       []string          `yaml:"tickers"`
	CandleMinLen  int               `yaml:"candle_min_len" env-default:"3"`
	CandleMaxLen  int               `yaml:"candle_max_len" env-default:"5"`
	ChartMinLen   int               `yaml:"chart_min_len" env-default:"15"`
	ChartMaxLen   int               `yaml:"chart_max_len" env-default:"30"`
}

type ScanCandleOptions struct {
	MinTailLen      int     `yaml:"min_tail_len"`
	MaxTailLen      int     `yaml:"max_tail_len"`
	ShadowTolerance float64 `yaml:"shadow_tolerance"`
	BodyTolerance   float64 `yaml:"body_tolerance"`
	DaysToWatch     int     `yaml:"days_to_watch" env-default:"3"`
	MinMatches      int     `yaml:"min_matches" env-default:"10"`
}

type ScanChartOptions struct {
	MinScale    float64 `yaml:"min_scale"`
	MaxScale    float64 `yaml:"max_scale"`
	Tolerance   float64 `yaml:"tolerance"`
	DaysToWatch int     `yaml:"days_to_watch" env-default:"5"`
	MinMatches  int     `yaml:"min_matches" env-default:"10"`
}

type GrpcConfig struct {
	Port           int           `yaml:"port" env-default:"8080"`
	RequestTimeout time.Duration `yaml:"request_timeout" env:"GRPC_TIMEOUT" env-default:"10s"`
}

type GrpcClientConfig struct {
	Address string        `yaml:"address" env-default:"localhost:8080"`
	Timeout time.Duration `yaml:"timeout" env-default:"10s"`
}

type KafkaConfig struct {
	BootstrapServers string `yaml:"bootstrap_servers" env-default:"localhost:9092"`
	Topic            string `yaml:"topic" env-default:"watcher"`
}

type SchedulerConfig struct {
	Cron     string `yaml:"cron" env-default:"0 10 * * *"`
	Timezone string `yaml:"timezone" env-default:"UTC"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config file path is empty")
	}

	var cfg Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist" + configPath)
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
