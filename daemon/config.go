package daemon

import (
	"pgdmpres/pkg/config"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

type AppCfg struct {
	Logger config.Logger `json:"logger" yaml:"logger" env-prefix:"LOGGER_"`
	S3     config.S3     `json:"s3" yaml:"s3" env-prefix:"S3_"`
	Dump   struct {
		Enabled  bool            `json:"enabled" yaml:"enabled" env:"ENABLED" env-default:"true"`
		Postgres config.Postgres `json:"postgres" yaml:"postgres" env-prefix:"POSTGRES_"`
	} `json:"dump" yaml:"dump" env-prefix:"DUMP_"`
	Restore struct {
		Enabled  bool            `json:"enabled" yaml:"enabled" env:"ENABLED" env-default:"false"`
		Postgres config.Postgres `json:"postgres" yaml:"postgres" env-prefix:"POSTGRES_"`
	} `json:"restore" yaml:"restore" env-prefix:"RESTORE_"`
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
