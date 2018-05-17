// Package config implements config loader from .yml file
// into struct datatype
package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config struct define the config data structure
type Config struct {
	reader func(filename string) ([]byte, error)
}

// New used to initiate config struct for thread safe instance
func New() *Config {
	return &Config{reader: ioutil.ReadFile}
}

// StubReader is used to be able to stub ioutil.ReadFile function
// for testing purpose
func (cfg *Config) StubReader(stub func(filename string) ([]byte, error)) {
	cfg.reader = stub
}

// ISPConfig retains all information regarding and ISP, interface it is connected to
// the gateway IP address and the address it can check for basic connectivity
type ISPConfig struct {
	Name    string `yaml:"name"`
	Eth     string `yaml:"eth"`
	Gateway string `yaml:"gateway"`
	CheckIP string `yaml:"checkip"`
}

// Load will load a config from a yml file
func (cfg *Config) Load(configPath string) ([]ISPConfig, error) {
	rawData, err := cfg.reader(configPath)
	if err != nil {
		return nil, err
	}
	buffer := []ISPConfig{}
	errs := yaml.Unmarshal(rawData, &buffer)
	if errs != nil {
		return nil, errs
	}
	return buffer, nil
}
