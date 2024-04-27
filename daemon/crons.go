package daemon

import "github.com/rs/zerolog/log"

func dmp() {
	log.Info().Msgf("Initiating 'pg_dump' operation on `%s` and database `%s`", cfg.Dump.Postgres.Addr, cfg.Dump.Postgres.DB)
}

func res() {
	log.Info().Msgf("Initiating 'pg_restore' operation on `%s` and database `%s`", cfg.Dump.Postgres.Addr, cfg.Dump.Postgres.DB)
}
