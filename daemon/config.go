package daemon

import (
	"pgdmpres/pkg/config"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

type AppCfg struct {
	Logger config.Logger `json:"logger" yaml:"logger" env-prefix:"LOGGER_"`
	S3     config.S3     `json:"s3" yaml:"s3" env-prefix:"S3_"`
}

var cfg AppCfg

func initCfg(filepath string) error {
	if filepath != "" {
		if err := cleanenv.ReadConfig(filepath, &cfg); err != nil {
			return errors.Wrapf(err, "cannot read config file '%s'", filepath)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return errors.Wrap(err, "cannot read environment to config")
		}
	}
	return nil
}
