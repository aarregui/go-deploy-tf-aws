package cli

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/aarregui/go-deploy-tf-aws/internal"
)

type CLIClient interface {
	Execute()
}

type CLI struct {
	app      internal.App
	cfg      internal.Config
	db       internal.RDBMSClient
	migrator internal.MigratorClient
}

func New() *CLI {
	return &CLI{}
}

var cli CLI

var rootCmd = &cobra.Command{
	Use:           "go-deploy-tf-aws",
	Short:         "Deploy a go webserver on AWS",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func (c *CLI) Execute() {
	var hasError bool
	err := rootCmd.Execute()
	if err != nil {
		hasError = true
		log.Error().Err(err).Msg("")
	}

	err = cli.db.Close()
	if err != nil {
		hasError = true
		log.Error().Err(err).Msg("")
	}

	if hasError {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogger, initDB, initApp)

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
}

func initLogger() {
	zerolog.SetGlobalLevel(zerolog.Level(cli.cfg.App.LogLevel))
	if cli.cfg.App.Env == internal.ENV_LOCAL {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		}).With().Caller().Logger()
	} else {
		log.Logger = log.With().Caller().Logger()
	}
}

func initConfig() {
	c, err := internal.NewConfig(".env")
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	cli.cfg = *c
}

func initDB() {
	rdbms := internal.NewPostgres(cli.cfg)
	if err := rdbms.Open(); err != nil {
		log.Fatal().Err(err).Msg("")
	}

	cli.db = rdbms

	m, err := internal.NewMigrator(rdbms.GetCon(), cli.cfg.DB.MigrationsPath)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	cli.migrator = m
}

func initApp() {
	cli.app = internal.NewApp(cli.cfg, cli.db, cli.migrator)
}
