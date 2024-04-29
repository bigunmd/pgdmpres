package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
)

const tmpFilename = "dmp.tar"
const tmpFilenameEncrypted = "dmp.tar.gpg"
const tmpFilepath = "/tmp/" + tmpFilename
const tmpFilepathEncrypted = "/tmp/" + tmpFilenameEncrypted

const (
	cmdPgDump    = "pg_dump"
	cmdPgRestore = "pg_restore"
	cmdGPG       = "gpg"

	argNoPassword = "--no-password"
	argVerbose    = "--verbose"
	argFormat     = "--format=t"
	argFile       = "--file=" + tmpFilepath

	argSymmetric = "--symmetric"
	argBatch     = "--batch"
	argOutput    = "--output=" + tmpFilepathEncrypted
)

type pgDumpStd struct{}

func (o *pgDumpStd) Write(p []byte) (n int, err error) {
	log.Debug().Str("command", cmdPgDump).Msg(string(p))
	return len(p), nil
}

type pgRestoreOutput struct{}

func (o *pgRestoreOutput) Write(p []byte) (n int, err error) {
	log.Debug().Str("command", cmdPgRestore).Msg(string(p))
	return len(p), nil
}

type gpgStd struct{}

func (o *gpgStd) Write(p []byte) (n int, err error) {
	log.Debug().Str("command", cmdGPG).Msg(string(p))
	return len(p), nil
}

func dmp(mc *minio.Client) {
	log.Info().Msg("Starting dump job")
	log.Info().Str("addr", cfg.Dump.Postgres.Addr()).Str("db", cfg.Dump.Postgres.DB).Msgf("Creating dump file with '%s' command", cmdPgDump)
	ctx, cancel := context.WithTimeout(context.TODO(), cfg.Dump.Timeout)
	defer cancel()
	cmd := exec.CommandContext(
		ctx,
		cmdPgDump,
		fmt.Sprintf("--host=%s", cfg.Dump.Postgres.Host),
		fmt.Sprintf("--port=%v", cfg.Dump.Postgres.Port),
		fmt.Sprintf("--dbname=%s", cfg.Dump.Postgres.DB),
		fmt.Sprintf("--username=%s", cfg.Dump.Postgres.User),
		argNoPassword,
		argFormat,
		argFile,
		argVerbose,
	)
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", cfg.Dump.Postgres.Password))
	cmd.Stderr = &pgDumpStd{}
	cmd.Stdout = &pgDumpStd{}
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Msgf("Failed to run '%s' command", cmd.String())
		return
	}
	log.Info().Str("addr", cfg.Dump.Postgres.Addr()).Str("db", cfg.Dump.Postgres.DB).Msgf("Successfully completed dump with '%s' command", cmdPgDump)

	var f *os.File
	var err error
	objectName := fmt.Sprintf("%s_%s.tar", cfg.Dump.Postgres.DB, time.Now().Format(time.RFC3339))
	if cfg.S3.Prefix != "" {
		objectName = cfg.S3.Prefix + "/" + objectName
	}
	if cfg.Dump.GPG.Passphrase != "" {
		log.Info().Msgf("Encrypting dump file with '%s' command", cmdGPG)
		if err := os.Remove(tmpFilepathEncrypted); err != nil && !os.IsNotExist(err) {
			log.Error().Err(err).Msgf("Failed to remove file '%s'", tmpFilepathEncrypted)
			return
		}
		cmd = exec.CommandContext(
			ctx,
			cmdGPG,
			argSymmetric,
			argBatch,
			fmt.Sprintf("--passphrase=%s", cfg.Dump.GPG.Passphrase),
			argOutput,
			argVerbose,
			tmpFilepath,
		)
		cmd.Stderr = &gpgStd{}
		cmd.Stdout = &gpgStd{}
		if err := cmd.Run(); err != nil {
			log.Error().Err(err).Msgf("Failed to run '%s' command", cmdGPG)
			return
		}
		if err := os.Remove(tmpFilepath); err != nil && !os.IsNotExist(err) {
			log.Error().Err(err).Msgf("Failed to remove file '%s'", tmpFilepath)
			return
		}
		log.Info().Msgf("Successfully completed encryption with '%s' command", cmdGPG)
		f, err = os.Open(tmpFilepathEncrypted)
		if err != nil {
			log.Error().Err(err).Msg("Failed to open encrypted dump file")
			return
		}
		objectName = objectName + ".gpg"
	} else {
		f, err = os.Open(tmpFilepathEncrypted)
		if err != nil {
			log.Error().Err(err).Msg("Failed to open dump file")
			return
		}
	}

	ctx, cancel = context.WithTimeout(ctx, cfg.S3.Timeout)
	defer cancel()
	info, err := mc.PutObject(ctx, cfg.S3.Bucket, objectName, f, -1, minio.PutObjectOptions{})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to upload dump to S3")
		return
	}
	log.Info().Str("key", info.Key).Msg("Successfully uploaded dump to S3")

	if cfg.Dump.Rotate != 0 {
		log.Info().Msg("Removing older versions...")
		ctx, cancel = context.WithTimeout(ctx, cfg.S3.Timeout)
		defer cancel()
		for object := range mc.ListObjects(ctx, cfg.S3.Bucket, minio.ListObjectsOptions{
			// WithVersions: true,
			// WithMetadata: true,
			Prefix:    cfg.S3.Prefix,
			Recursive: true,
		}) {
			if time.Since(object.LastModified) > cfg.Dump.Rotate {
				if err := mc.RemoveObject(ctx, cfg.S3.Bucket, object.Key, minio.RemoveObjectOptions{ForceDelete: true}); err != nil {
					log.Error().Err(err).Str("key", object.Key).Msg("Failed to remove object")
				} else {
					log.Debug().Str("key", object.Key).Msg("Removed object")
				}
			}
		}
	}
	log.Info().Msg("Finished dump job")
}

func res() {
	log.Info().Str("addr", cfg.Restore.Postgres.Addr()).Str("db", cfg.Restore.Postgres.DB).Msgf("Initiating '%s' command", cmdPgRestore)
}
