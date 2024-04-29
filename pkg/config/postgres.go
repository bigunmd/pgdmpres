package config

import "fmt"

type Postgres struct {
	Host     string `json:"host" yaml:"host" env:"HOST" env-default:"127.0.0.1"`
	Port     uint32 `json:"port" yaml:"port" env:"PORT" env-default:"5432"`
	DB       string `json:"db" yaml:"db" env:"DB" env-default:"postgres"`
	User     string `json:"user" yaml:"user" env:"USER" env-default:"postgres"`
	Password string `json:"password" yaml:"password" env:"PASSWORD" env-default:"postgres"`
}

func (p Postgres) Addr() string {
	return fmt.Sprintf("%s:%v", p.Host, p.Port)
}
