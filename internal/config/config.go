package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServiceName     string        `yaml:"service_name" env-default:"proxy-service"`
	Env             string        `yaml:"env" env-default:"local"`
	Version         string        `yaml:"version" env-default:"dev"`
	LogLevel        string        `yaml:"log_level" env-default:"info"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"10s"`

	HTTP   HTTPConfig   `yaml:"http"`
	Proxy  ProxyConfig  `yaml:"proxy"`
	Faults FaultsConfig `yaml:"faults"`
}

type HTTPConfig struct {
	Addr    string        `yaml:"addr" env-default:":8085"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type ProxyConfig struct {
	ListenAddr   string        `yaml:"listen_addr" env-default:":9095"`
	UpstreamAddr string        `yaml:"upstream_addr" env-default:"catalog-service:9091"`
	Timeout      time.Duration `yaml:"timeout" env-default:"5s"`
}

type FaultsConfig struct {
	Delay time.Duration `yaml:"delay" env-default:"0s"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	return mustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	return mustLoadByPath(path)
}

func mustLoadByPath(path string) *Config {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	if err := cfg.Validate(); err != nil {
		panic("invalid config: " + err.Error())
	}

	return &cfg
}

func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}
	if c.Env == "" {
		return fmt.Errorf("env is required")
	}
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if c.LogLevel == "" {
		return fmt.Errorf("log_level is required")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown_timeout must be > 0")
	}

	if c.HTTP.Addr == "" {
		return fmt.Errorf("http.addr is required")
	}
	if c.HTTP.Timeout <= 0 {
		return fmt.Errorf("http.timeout must be > 0")
	}

	if c.Proxy.ListenAddr == "" {
		return fmt.Errorf("proxy.listen_addr is required")
	}
	if c.Proxy.UpstreamAddr == "" {
		return fmt.Errorf("proxy.upstream_addr is required")
	}
	if c.Proxy.Timeout <= 0 {
		return fmt.Errorf("proxy.timeout must be > 0")
	}

	if c.Faults.Delay < 0 {
		return fmt.Errorf("faults.delay must be >= 0")
	}

	return nil
}

func fetchConfigPath() string {
	var path string

	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}
