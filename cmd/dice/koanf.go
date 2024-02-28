package main

import (
	"foxdice/utils"

	"github.com/knadh/koanf/v2"
)

func NewConfig(path string) utils.IConfig {
	return &Config{Koanf: koanf.New(path)}
}

type Config struct {
	*koanf.Koanf
}

func (c *Config) Sub(path string) utils.IConfig {
	return &Config{Koanf: koanf.New(path)}
}
