/*
	Copyright © 2021 RockHopper OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/
package specs

import (
	v "github.com/spf13/viper"

	"gopkg.in/yaml.v2"
)

const (
	RHCTL_CONFIGNAME = "rhos"
	RHCTL_ENV_PREFIX = "RHCTL"
)

type RhCtlConfig struct {
	Viper *v.Viper `yaml:"-" json:"-"`

	General RhCtlGeneral `mapstructure:"general" json:"general,omitempty" yaml:"general,omitempty"`
	Logging RhCtlLogging `mapstructure:"logging" json:"logging,omitempty" yaml:"logging,omitempty"`
}

type RhCtlGeneral struct {
	Debug bool `mapstructure:"debug,omitempty" json:"debug,omitempty" yaml:"debug,omitempty"`
}

type RhCtlLogging struct {
	// Path of the logfile
	Path string `mapstructure:"path,omitempty" json:"path,omitempty" yaml:"path,omitempty"`
	// Enable/Disable logging to file
	EnableLogFile bool `mapstructure:"enable_logfile,omitempty" json:"enable_logfile,omitempty" yaml:"enable_logfile,omitempty"`
	// Enable JSON format logging in file
	JsonFormat bool `mapstructure:"json_format,omitempty" json:"json_format,omitempty" yaml:"json_format,omitempty"`

	// Log level
	Level string `mapstructure:"level,omitempty" json:"level,omitempty" yaml:"level,omitempty"`

	// Enable emoji
	EnableEmoji bool `mapstructure:"enable_emoji,omitempty" json:"enable_emoji,omitempty" yaml:"enable_emoji,omitempty"`
	// Enable/Disable color in logging
	Color bool `mapstructure:"color,omitempty" json:"color,omitempty" yaml:"color,omitempty"`
}

func NewRhCtlConfig(viper *v.Viper) *RhCtlConfig {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &RhCtlConfig{Viper: viper}
}

func (c *RhCtlConfig) GetGeneral() *RhCtlGeneral {
	return &c.General
}

func (c *RhCtlConfig) GetLogging() *RhCtlLogging {
	return &c.Logging
}

func (c *RhCtlConfig) Unmarshal() error {
	var err error

	if c.Viper.InConfig("etcd-config") &&
		c.Viper.GetBool("etcd-config") {
		err = c.Viper.ReadRemoteConfig()
	} else {
		err = c.Viper.ReadInConfig()
	}

	if err != nil {
		if _, ok := err.(v.ConfigFileNotFoundError); !ok {
			return err
		}
		// else: Config file not found; ignore error
	}

	err = c.Viper.Unmarshal(&c)

	return err
}

func (c *RhCtlConfig) Yaml() ([]byte, error) {
	return yaml.Marshal(c)
}

func GenDefault(viper *v.Viper) {
	viper.SetDefault("general.debug", false)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.enable_logfile", false)
	viper.SetDefault("logging.path", "/var/log/rhos/rhctl.log")
	viper.SetDefault("logging.json_format", false)
	viper.SetDefault("logging.enable_emoji", true)
	viper.SetDefault("logging.color", true)
}

func (g *RhCtlGeneral) HasDebug() bool {
	return g.Debug
}
