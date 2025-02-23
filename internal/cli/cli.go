package cli

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
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
		slog.Error("failed to execute", "error", err)
	}

	err = cli.db.Close()
	if err != nil {
		hasError = true
		slog.Error("failed to close db", "error", err)
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
	var handler slog.Handler
	level := slog.Level(cli.cfg.App.LogLevel)

	if cli.cfg.App.Env == internal.ENV_LOCAL {
		handler = tint.NewHandler(os.Stderr, &tint.Options{
			Level:     level,
			AddSource: true,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func initConfig() {
	c, err := internal.NewConfig(".env")
	if err != nil {
		slog.Error("failed to initConfig", "error", err)
		os.Exit(1)
	}

	cli.cfg = *c
}

func initDB() {
	rdbms := internal.NewPostgres(cli.cfg)
	if err := rdbms.Open(); err != nil {
		slog.Error("failed to open db", "error", err)
		os.Exit(1)
	}

	cli.db = rdbms

	m, err := internal.NewMigrator(rdbms.GetCon(), cli.cfg.DB.MigrationsPath)
	if err != nil {
		slog.Error("failed to init migrator", "error", err)
		os.Exit(1)
	}
	cli.migrator = m
}

func initApp() {
	cli.app = internal.NewApp(cli.cfg, cli.db, cli.migrator)
}
