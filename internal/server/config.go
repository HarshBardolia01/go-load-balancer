package server

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Schema     string `json:"schema" yaml:"schema" validate:"required,oneof=http https"`
	Host       string `json:"host" yaml:"host" validate:"required"`
	Port       int    `json:"port" yaml:"port" validate:"required,min=1,max=65535"`
	InstanceId string `json:"instanceId" yaml:"instanceId" validate:"required"`

	Weight     int    `json:"weight" yaml:"weight" validate:"required,gte=0"`
	ServerType string `json:"serverType" yaml:"serverType" validate:"required,oneof=fast slow medium"`

	ServingEp string `json:"servingEp" yaml:"servingEp" validate:"required,startswith=/"`

	LbSchema string `json:"lbSchema" yaml:"lbSchema" validate:"required,oneof=http https"`
	LbHost   string `json:"lbHost" yaml:"lbHost" validate:"required"`
	LbPort   int    `json:"lbPort" yaml:"lbPort" validate:"required,min=1,max=65535"`

	LbRegisterEp   string `json:"lbRegisterEp" yaml:"lbRegisterEp" validate:"required,startswith=/"`
	LbDeregisterEp string `json:"lbDeregisterEp" yaml:"lbDeregisterEp" validate:"required,startswith=/"`
	LbHeartbeatEp  string `json:"lbHeartbeatEp" yaml:"lbHeartbeatEp" validate:"required,startswith=/"`
}

func LoadConfig(filePath string) (*Config, error) {
	var config Config
	viper.SetConfigFile(filePath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := config.validateConfig(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) validateConfig() error {
	validate := validator.New()
	return validate.Struct(c)
}

func ParseArguments() (string, error) {
	if len(os.Args) != 2 {
		return "", fmt.Errorf("expected exactly one argument for config file path, got %d", len(os.Args)-1)
	}

	return os.Args[1], nil
}
