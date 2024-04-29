package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
)

const (
	tmpPath              = "/tmp"
	dmpFilename          = "dmp.tar"
	dmpFilenameEncrypted = "dmp.tar.gpg"
	dmpFilepath          = tmpPath + "/" + dmpFilename
	dmpFilepathEncrypted = tmpPath + "/" + dmpFilenameEncrypted
	resFilename          = "res.tar"
	resFilenameDecrypted = "decrypted_res.tar"
	resFilepath          = tmpPath + "/" + resFilename
	resFilepathDecrypted = tmpPath + "/" + resFilenameDecrypted
)

// const tmpFilepath = "/tmp/" + tmpFilename
// const tmpFilepathEncrypted = "/tmp/" + tmpFilenameEncrypted
// const tmpFilepathRestore = "/tmp/res"
// const tmpFilepathRestoreDecrypted = "/tmp/res.dec"

const (
	cmdPgDump    = "pg_dump"
	cmdPgRestore = "pg_restore"
	cmdGPG       = "gpg"

	argNoPassword = "--no-password"
	argVerbose    = "--verbose"
	argFormat     = "--format=t"
	// argFile       = "--file=" + tmpFilepath

	argSymmetric = "--symmetric"
	argBatch     = "--batch"
	argDecrypt   = "--decrypt"
	// argOutputEncrypted = "--output=" + tmpFilepathEncrypted
	// argOutputDecrypted = "--output=" + tmpFilepathRestoreDecrypted
)

type pgDumpStd struct{}

func (o *pgDumpStd) Write(p []byte) (n int, err error) {
	log.Debug().Str("command", cmdPgDump).Msg(string(p))
	return len(p), nil
}

type pgRestoreStd struct{}

func (o *pgRestoreStd) Write(p []byte) (n int, err error) {
	log.Debug().Str("command", cmdPgRestore).Msg(string(p))
	return len(p), nil
}

type gpgStd struct{}

func (o *gpgStd) Write(p []byte) (n int, err error) {
	log.Debug().Str("command", cmdGPG).Msg(string(p))
	return len(p), nil
}

func dmp(mc *minio.Client) {
	dmpLog := log.With().Str("job", cmdPgDump).Logger()
	dmpLog.Info().Msgf("Started '%s' job", cmdPgDump)
	dmpLog.Info().
		Str("addr", cfg.Dump.Postgres.Addr()).
		Str("db", cfg.Dump.Postgres.DB).
		Msg("Creating backup file")
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
		fmt.Sprintf("--file=%s", dmpFilepath),
		argVerbose,
	)
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", cfg.Dump.Postgres.Password))
	cmd.Stderr = &pgDumpStd{}
	cmd.Stdout = &pgDumpStd{}
	if err := cmd.Run(); err != nil {
		dmpLog.Error().Err(err).Msgf("Failed to run '%s' command", cmd.String())
		return
	}
	dmpLog.Info().
		Str("addr", cfg.Dump.Postgres.Addr()).
		Str("db", cfg.Dump.Postgres.DB).
		Msg("Successfully created backup file")

	var f *os.File
	var err error
	objectName := fmt.Sprintf("%s_%s.tar", cfg.Dump.Postgres.DB, time.Now().Format(time.RFC3339))
	if cfg.S3.Prefix != "" {
		objectName = cfg.S3.Prefix + "/" + objectName
	}
	if cfg.Dump.GPG.Passphrase != "" {
		dmpLog.Info().Msg("Encrypting backup file")
		if err := deleteFile(dmpFilepathEncrypted); err != nil {
			return
		}
		cmd = exec.CommandContext(
			ctx,
			cmdGPG,
			argSymmetric,
			argBatch,
			fmt.Sprintf("--passphrase=%s", cfg.Dump.GPG.Passphrase),
			fmt.Sprintf("--output=%s", dmpFilepathEncrypted),
			argVerbose,
			dmpFilepath,
		)
		cmd.Stderr = &gpgStd{}
		cmd.Stdout = &gpgStd{}
		if err := cmd.Run(); err != nil {
			dmpLog.Error().Err(err).Msgf("Failed to run '%s' command", cmdGPG)
			return
		}
		if err := deleteFile(dmpFilepath); err != nil {
			return
		}
		dmpLog.Info().Msg("Successfully encrypted backup file")
		f, err = os.Open(dmpFilepathEncrypted)
		if err != nil {
			dmpLog.Error().Err(err).Msg("Failed to open encrypted backup file")
			return
		}
		defer deleteFile(dmpFilepathEncrypted)
		objectName = objectName + ".gpg"
	} else {
		f, err = os.Open(dmpFilepath)
		if err != nil {
			dmpLog.Error().Err(err).Msg("Failed to open backup file")
			return
		}
		defer deleteFile(dmpFilepath)
	}
	defer f.Close()

	ctx, cancel = context.WithTimeout(ctx, cfg.S3.Timeout)
	defer cancel()
	info, err := mc.PutObject(ctx, cfg.S3.Bucket, objectName, f, -1, minio.PutObjectOptions{})
	if err != nil {
		dmpLog.Error().Err(err).Msgf("Failed to upload backup file to S3")
		return
	}
	dmpLog.Info().Str("key", info.Key).Msg("Successfully uploaded backup file to S3")

	if cfg.Dump.Rotate != 0 {
		dmpLog.Info().Msg("Removing old backup files from S3...")
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
					dmpLog.Error().Err(err).Str("key", object.Key).Msg("Failed to remove object")
				} else {
					dmpLog.Debug().Str("key", object.Key).Msg("Removed object")
				}
			}
		}
	}
	dmpLog.Info().Msgf("Finished '%s' job", cmdPgDump)
}

func res(mc *minio.Client) {
	resLog := log.With().Str("job", cmdPgRestore).Logger()
	resLog.Info().Msgf("Started '%s' job", cmdPgRestore)
	ctx, cancel := context.WithTimeout(context.TODO(), cfg.S3.Timeout)
	defer cancel()
	var objects []minio.ObjectInfo
	for object := range mc.ListObjects(ctx, cfg.S3.Bucket, minio.ListObjectsOptions{
		// WithVersions: true,
		// WithMetadata: true,
		Prefix:    cfg.S3.Prefix,
		Recursive: true,
	}) {
		objects = append(objects, object)
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(objects[j].LastModified)
	})
	if err := mc.FGetObject(ctx, cfg.S3.Bucket, objects[0].Key, resFilepath, minio.GetObjectOptions{}); err != nil {
		resLog.Error().Err(err).Msgf("Failed to get latest backup file from S3")
		return
	}
	defer deleteFile(resFilepath)
	resLog.Info().Str("key", objects[0].Key).Msg("Successfully downloaded latest backup file from S3")

	ctx, cancel = context.WithTimeout(ctx, cfg.Restore.Timeout)
	defer cancel()
	var bkpFilepath string
	if cfg.Restore.GPG.Passphrase != "" {
		resLog.Info().Msg("Decrypting backup file")
		cmd := exec.CommandContext(
			ctx,
			cmdGPG,
			argDecrypt,
			argBatch,
			fmt.Sprintf("--passphrase=%s", cfg.Dump.GPG.Passphrase),
			fmt.Sprintf("--output=%s", resFilepathDecrypted),
			argVerbose,
			resFilepath,
		)
		cmd.Stderr = &gpgStd{}
		cmd.Stdout = &gpgStd{}
		if err := cmd.Run(); err != nil {
			resLog.Error().Err(err).Msgf("Failed to run '%s' command", cmdGPG)
			return
		}
		defer deleteFile(resFilepathDecrypted)
		resLog.Info().Msg("Successfully decrypted backup file")
		bkpFilepath = resFilepathDecrypted
	} else {
		bkpFilepath = resFilepath
	}

	resLog.Info().
		Str("addr", cfg.Restore.Postgres.Addr()).
		Str("db", cfg.Restore.Postgres.DB).
		Msg("Restoring database from backup file")
	cmd := exec.CommandContext(
		ctx,
		cmdPgRestore,
		fmt.Sprintf("--host=%s", cfg.Dump.Postgres.Host),
		fmt.Sprintf("--port=%v", cfg.Dump.Postgres.Port),
		fmt.Sprintf("--dbname=%s", cfg.Dump.Postgres.DB),
		fmt.Sprintf("--username=%s", cfg.Dump.Postgres.User),
		argNoPassword,
		argFormat,
		argVerbose,
		bkpFilepath,
	)
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", cfg.Restore.Postgres.Password))
	cmd.Stderr = &pgRestoreStd{}
	cmd.Stdout = &pgRestoreStd{}
	if err := cmd.Run(); err != nil {
		resLog.Error().Err(err).Msgf("Failed to run '%s' command", cmd.String())
		return
	}
	resLog.Info().
		Str("addr", cfg.Restore.Postgres.Addr()).
		Str("db", cfg.Restore.Postgres.DB).
		Msg("Successfully restored database from backup file")
	resLog.Info().Msgf("Finished '%s' job", cmdPgRestore)
}

func deleteFile(filepath string) error {
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		log.Error().Err(err).Msgf("Failed to remove file '%s'", filepath)
		return err
	}
	return nil
}
