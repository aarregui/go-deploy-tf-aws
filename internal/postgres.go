package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/lmittmann/tint"

	_ "github.com/golang-migrate/migrate/source/file"
	driverPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ErrConectionAlreadyEstablished = errors.New("connection already established")
	ErrOpenNotCalled               = errors.New("Open() has not been called")
)

type RDBMSClient interface {
	Open() error
	Close() error
	GetCon() *gorm.DB
}

type rdbms struct {
	gorm   *gorm.DB
	config Config
}

func NewPostgres(config Config) *rdbms {
	return &rdbms{config: config}
}

func (r *rdbms) Open() error {
	if r.gorm != nil {
		return ErrConectionAlreadyEstablished
	}

	password, err := r.getPassword()
	if err != nil {
		return fmt.Errorf("failed to password: %w", err)
	}

	c := r.config.DB
	dataSource := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", c.Host, c.Username, password, c.Database, c.Port)

	handler := tint.NewHandler(os.Stdout, &tint.Options{
		Level: slog.Level(c.LogLevel),
	})
	log := slog.New(handler)
	db, err := gorm.Open(driverPostgres.Open(dataSource), &gorm.Config{
		Logger:               slogLogger{slog: log},
		FullSaveAssociations: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return err
	}

	r.gorm = db

	return nil
}

func (r rdbms) getPassword() (string, error) {
	if r.config.App.Env == ENV_LOCAL {
		return r.config.DB.Password, nil
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}

	client := secretsmanager.NewFromConfig(cfg)
	out, err := client.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(r.config.AWS.RDSMasterPasswordSecretID),
	})

	if err != nil {
		return "", err
	}

	return *out.SecretString, nil
}

func (r *rdbms) Close() error {
	if r.gorm == nil {
		return ErrOpenNotCalled
	}

	db, err := r.gorm.DB()
	if err != nil {
		return errors.New(fmt.Sprint("failed to get db", err))
	}

	err = db.Close()
	if err != nil {
		return errors.New(fmt.Sprint("failed to close db", err))
	}

	r.gorm = nil

	return nil
}

func (r rdbms) GetCon() *gorm.DB {
	return r.gorm
}

type slogLogger struct {
	slog *slog.Logger
}

func (l slogLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := l
	switch level {
	case logger.Silent:
		newLogger.slog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError + 1}))
	case logger.Error:
		newLogger.slog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	case logger.Warn:
		newLogger.slog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
	case logger.Info:
		newLogger.slog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		newLogger.slog = l.slog
	}
	return newLogger
}

func (l slogLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.slog.Info(fmt.Sprintf(msg, data...))
}

func (l slogLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.slog.Warn(fmt.Sprintf(msg, data...))
}

func (l slogLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.slog.Error(fmt.Sprintf(msg, data...))
}

func (l slogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		if rows == -1 {
			l.slog.Error("SQL execution error", "elapsed", elapsed, "sql", sql, "error", err)
		} else {
			l.slog.Error("SQL execution error", "elapsed", elapsed, "sql", sql, "rows", rows, "error", err)
		}
	case elapsed > 200*time.Millisecond:
		if rows == -1 {
			l.slog.Warn("Slow SQL query", "elapsed", elapsed, "sql", sql)
		} else {
			l.slog.Warn("Slow SQL query", "elapsed", elapsed, "sql", sql, "rows", rows)
		}
	default:
		if rows == -1 {
			l.slog.Info("SQL query executed", "elapsed", elapsed, "sql", sql)
		} else {
			l.slog.Info("SQL query executed", "elapsed", elapsed, "sql", sql, "rows", rows)
		}
	}
}
