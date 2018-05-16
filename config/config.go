package config

import (
	"io/ioutil"
)

// Config info
type Config struct {
	reader func(filename string) ([]byte, error)
}

// New bla bla
func New() *Config {
	return &Config{reader: ioutil.ReadFile}
}

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

func (cfg *Config) Load(configPath string) (string, error) {
	rawData, err := cfg.reader(configPath)
	if err != nil {
		return "", err
	}
	return string(rawData), nil
}

// Reader interface is used to easily stub ioutil package
// type Reader interface {
// 	ReadFile(filename string) ([]byte, error)
// }
