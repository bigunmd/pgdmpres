package daemon

import (
	"time"

	"github.com/bigunmd/pgdmpres/pkg/config"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

type GPG struct {
	Passphrase string `json:"passphrase" yaml:"passphrase" env:"PASSPHRASE" env-default:""`
}

type Cron struct {
	Interval           time.Duration `json:"interval" yaml:"interval" env:"INTERVAL" env-default:"1h"`
	Crontab            string        `json:"crontab" yaml:"crontab" env:"CRONTAB" env-default:""`
	CrontabWithSeconds bool          `json:"crontabWithSeconds" yaml:"crontabWithSeconds" env:"CRONTAB_WITH_SECONDS" env-default:"false"`
}

type AppCfg struct {
	Logger config.Logger `json:"logger" yaml:"logger" env-prefix:"LOGGER_"`
	S3     config.S3     `json:"s3" yaml:"s3" env-prefix:"S3_"`
	Dump   struct {
		Enabled  bool            `json:"enabled" yaml:"enabled" env:"ENABLED" env-default:"true"`
		Postgres config.Postgres `json:"postgres" yaml:"postgres" env-prefix:"POSTGRES_"`
		Cron
		Timeout   time.Duration `json:"timeout" yaml:"timeout" env:"TIMEOUT" env-default:"4s"`
		GPG       `json:"gpg" yaml:"gpg" env-prefix:"GPG_"`
		Rotate    time.Duration `json:"rotate" yaml:"rotate" env:"ROTATE"`
		ExtraArgs []string      `json:"extraArgs" yaml:"extraArgs" env:"EXTRA_ARGS" env-default:""`
	} `json:"dump" yaml:"dump" env-prefix:"DUMP_"`
	Restore struct {
		Enabled  bool            `json:"enabled" yaml:"enabled" env:"ENABLED" env-default:"false"`
		Postgres config.Postgres `json:"postgres" yaml:"postgres" env-prefix:"POSTGRES_"`
		Cron
		Timeout   time.Duration `json:"timeout" yaml:"timeout" env:"TIMEOUT" env-default:"4s"`
		GPG       `json:"gpg" yaml:"gpg" env-prefix:"GPG_"`
		ExtraArgs []string `json:"extraArgs" yaml:"extraArgs" env:"EXTRA_ARGS" env-default:""`
	} `json:"restore" yaml:"restore" env-prefix:"RESTORE_"`
	DataPath string `json:"dataPath" yaml:"dataPath" env:"DATA_PATH" env-default:"/tmp"`
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
