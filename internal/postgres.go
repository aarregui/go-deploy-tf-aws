package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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

	log := log.Logger.Level(zerolog.Level(c.LogLevel))
	db, err := gorm.Open(driverPostgres.Open(dataSource), &gorm.Config{
		Logger:               zLogger{zerolog: log},
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

type zLogger struct {
	zerolog zerolog.Logger
}

func (l zLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := l
	switch level {
	case logger.Silent:
		newlogger.zerolog = l.zerolog.Level(zerolog.Disabled)
	case logger.Error:
		newlogger.zerolog = l.zerolog.Level(zerolog.ErrorLevel)
	case logger.Warn:
		newlogger.zerolog = l.zerolog.Level(zerolog.WarnLevel)
	case logger.Info:
		newlogger.zerolog = l.zerolog.Level(zerolog.InfoLevel)
	default:
		newlogger.zerolog = l.zerolog
	}
	return newlogger
}

func (l zLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.zerolog.Info().Msg(fmt.Sprintf(msg, data...))
}

func (l zLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.zerolog.Warn().Msg(fmt.Sprintf(msg, data...))
}

func (l zLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.zerolog.Error().Msg(fmt.Sprintf(msg, data...))
}

func (l zLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		if rows == -1 {
			l.zerolog.Error().Err(err).Dur("elapsed", elapsed).Str("sql", sql).Msg("trace")
		} else {
			l.zerolog.Error().Err(err).Dur("elapsed", elapsed).Str("sql", sql).Int64("rows", rows).Msg("trace")
		}
	case elapsed > 200*time.Millisecond:
		sql, rows := fc()
		if rows == -1 {
			l.zerolog.Warn().Dur("elapsed", elapsed).Str("sql", sql).Msg("trace")
		} else {
			l.zerolog.Warn().Dur("elapsed", elapsed).Str("sql", sql).Int64("rows", rows).Msg("trace")
		}
	default:
		sql, rows := fc()
		if rows == -1 {
			l.zerolog.Info().Dur("elapsed", elapsed).Str("sql", sql).Msg("trace")
		} else {
			l.zerolog.Info().Dur("elapsed", elapsed).Str("sql", sql).Int64("rows", rows).Msg("trace")
		}
	}
}
