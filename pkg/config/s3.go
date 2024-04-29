package config

import "time"

type S3 struct {
	Endpoint     string        `json:"endpoint" yaml:"endpoint" env:"ENDPOINT" env-default:"127.0.0.1:9000"`
	UseSSL       bool          `json:"useSSL" yaml:"useSSL" env:"USE_SSL" env-default:"false"`
	Region       string        `json:"region" yaml:"region" env:"REGION" env-default:""`
	Bucket       string        `json:"bucket" yaml:"bucket" env:"BUCKET" env-default:""`
	Prefix       string        `json:"prefix" yaml:"prefix" env:"PREFIX" env-default:""`
	AccessID     string        `json:"accessID" yaml:"accessID" env:"ACCESS_ID" env-default:""`
	AccessSecret string        `json:"accessSecret" yaml:"accessSecret" env:"ACCESS_SECRET" env-default:""`
	Token        string        `json:"token" yaml:"token" env:"TOKEN" env-default:""`
	Timeout      time.Duration `json:"timeout" yaml:"timeout" env:"TIMEOUT" env-default:"4s"`
}
