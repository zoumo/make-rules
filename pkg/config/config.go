package config

import (
	"fmt"
	"io/ioutil"
	"runtime"

	"sigs.k8s.io/yaml"
)

const (
	ConfigPath = "make-rules.yaml"
)

var (
	DefaultPlatforms = []string{fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)}
)

func New() *Config {
	c := &Config{}
	c.SetDefaults()
	return c
}

func (c *Config) SetDefaults() {
	if len(c.Go.Build.Platforms) == 0 {
		c.Go.Build.Platforms = DefaultPlatforms
	}
}

func Load() (*Config, error) {
	data, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadOrDie() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}
