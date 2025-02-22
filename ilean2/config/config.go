package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Serial         int  `yaml:"serial"`
}

func (c Config) Validate() error {
	if c.Serial == 0 {
		return errors.New("rabbitmq_uri is empty")
	}

	return nil
}

func LoadConfig(configPath string) (Config, error) {
	configYAML, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s file: %w", configPath, err)
	}

	var c Config

	err = yaml.Unmarshal(configYAML, &c)
	if err != nil {
		return Config{}, fmt.Errorf("YAML unmarshal config: %w", err)
	}

	return c, nil
}
