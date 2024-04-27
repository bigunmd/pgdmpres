package config

type Postgres struct {
	Addr     string `json:"addr" yaml:"addr" env:"ADDR" env-default:"127.0.0.1:5432"`
	DB       string `json:"db" yaml:"db" env:"DB" env-default:"postgres"`
	User     string `json:"user" yaml:"user" env:"USER" env-default:"postgres"`
	Password string `json:"password" yaml:"password" env:"PASSWORD" env-default:"postgres"`
}
