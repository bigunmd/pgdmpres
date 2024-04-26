package config

type Logger struct {
	Level string `json:"level" yaml:"level" env:"LEVEL" env-default:"info"`
}
