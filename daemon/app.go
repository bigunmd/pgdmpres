package daemon

import (
	"context"
	"os"
	"os/signal"

	"github.com/bigunmd/pgdmpres/pkg/util"

	"github.com/go-co-op/gocron/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

const appName = "PG Dmp & Res"

func Run() {
	// ________________________________________________________________________
	// Parse cli args to config
	filepath := pflag.StringP("config", "c", "", "configuration filepath (default: None)")
	pflag.Parse()
	if err := initCfg(*filepath); err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// ________________________________________________________________________
	// Banner
	util.PrintBanner(appName)

	// ________________________________________________________________________
	// Setup logger
	lvl, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse logger level")
	}
	log.Logger = log.Level(lvl)
	log.Info().Msgf("Application logger level is set to '%s'", log.Logger.GetLevel().String())

	// ________________________________________________________________________
	// Create S3 client
	mc, err := minio.New(cfg.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3.AccessID, cfg.S3.AccessSecret, cfg.S3.Token),
		Secure: cfg.S3.UseSSL,
		Region: cfg.S3.Region,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create S3 client")
	}
	ctx, cancel := context.WithTimeout(context.TODO(), cfg.S3.Timeout)
	defer cancel()
	ok, err := mc.BucketExists(ctx, cfg.S3.Bucket)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to check S3 bucket")
	}
	if !ok {
		log.Fatal().Msgf("Bucket '%s' does not exists at '%s'", cfg.S3.Bucket, cfg.S3.Endpoint)
	}
	log.Info().Msgf("Successfully connected with S3 client to '%s'", cfg.S3.Endpoint)

	// ________________________________________________________________________
	// Create cron jobs
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create cron scheduler")
	}
	s.Start()
	log.Info().Msg("Successfully started cron scheduler")
	if cfg.Dump.Enabled {
		var jd gocron.JobDefinition
		if cfg.Dump.Crontab != "" {
			jd = gocron.CronJob(cfg.Dump.Crontab, cfg.Dump.CrontabWithSeconds)
		} else {
			jd = gocron.DurationJob(cfg.Dump.Interval)
		}
		job, err := s.NewJob(jd, gocron.NewTask(dmp, mc))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create dump job")
		}
		nr, err := job.NextRun()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get dump job next run")
		}
		log.Info().Str("job_id", job.ID().String()).Time("next_run", nr).Msg("Successfully created dump job")
	}
	if cfg.Restore.Enabled {
		var jd gocron.JobDefinition
		if cfg.Restore.Crontab != "" {
			jd = gocron.CronJob(cfg.Restore.Crontab, cfg.Restore.CrontabWithSeconds)
		} else {
			jd = gocron.DurationJob(cfg.Restore.Interval)
		}
		job, err := s.NewJob(jd, gocron.NewTask(res, mc))
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create restore job")
		}
		nr, err := job.NextRun()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get restore job next run")
		}
		log.Info().Str("job_id", job.ID().String()).Time("next_run", nr).Msg("Successfully created restore job")
	}

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// This blocks the main thread until an interrupt is received
	<-quit

	log.Info().Msg("Gracefully shutting down")
	if err = s.Shutdown(); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown cron scheduler")
	}
	log.Info().Msg("Gracefull shutdown completed")
}
