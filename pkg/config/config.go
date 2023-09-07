package config

import (
	"gopkg.in/ini.v1"
)

type Config struct {
	cfg *ini.File
}

func NewConfig(path string) (*Config, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}
	return &Config{cfg: cfg}, nil
}

func NewStaticConfig(content map[string]map[string]string) (*Config, error) {
	cfg := ini.Empty()
	for section, m := range content {
		sec, err := cfg.NewSection(section)
		if err != nil {
			return nil, err
		}
		for key, value := range m {
			if _, err := sec.NewKey(key, value); err != nil {
				return nil, err
			}
		}
	}
	return &Config{cfg: cfg}, nil
}

func (c *Config) Get(section string, key string) (string, bool) {
	if sec := c.cfg.Section(section); sec != nil {
		if sec.HasKey(key) {
			k := sec.Key(key)
			return k.Value(), true
		}
	}
	return "", false
}

func (c *Config) GetAll(section string) map[string]string {
	if sec := c.cfg.Section(section); sec != nil {
		m := make(map[string]string)
		for _, k := range sec.Keys() {
			m[k.Name()] = k.Value()
		}
		return m
	}
	return nil
}
