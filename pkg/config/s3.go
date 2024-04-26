package config

import "time"

type S3 struct {
	Endpoint string        `json:"endpoint" yaml:"endpoint" env:"ENDPOINT" env-default:"127.0.0.1:9000"`
	UseSSL   bool          `json:"useSSL" yaml:"useSSL" env:"USE_SSL" env-default:"false"`
	Region   string        `json:"region" yaml:"region" env:"REGION" env-default:""`
	Bucket   string        `json:"bucket" yaml:"bucket" env:"BUCKET" env-default:""`
	ID       string        `json:"id" yaml:"id" env:"ID" env-default:""`
	Secret   string        `json:"secret" yaml:"secret" env:"SECRET" env-default:""`
	Token    string        `json:"token" yaml:"token" env:"TOKEN" env-default:""`
	Timeout  time.Duration `json:"timeout" yaml:"timeout" env:"TIMEOUT" env-default:"4s"`
}
