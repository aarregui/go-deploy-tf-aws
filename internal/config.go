package internal

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

const (
	ENV_LOCAL = "local"
)

var (
	ErrDBPasswordNotSet = errors.New("DB_PASSWORD or AWS_RDS_MASTER_PASSWORD_SECRET_ID must be set")
)

type Config struct {
	Version string `envconfig:"VERSION" default:"13.37"`
	App     AppCfg
	DB      DBCfg
	AWS     AWSCfg
}

type AppCfg struct {
	Env      string `envconfig:"APP_ENV" required:"true"`
	Port     string `envconfig:"APP_PORT" required:"true"`
	LogLevel int    `envconfig:"APP_LOG_LEVEL" default:"0"` // panic=5 fatal=4 error=3 warn=2 info=1 debug=0
}

type DBCfg struct {
	Host           string `envconfig:"DB_HOST" required:"true"`
	Port           string `envconfig:"DB_PORT" required:"true"`
	Database       string `envconfig:"DB_DATABASE" required:"true"`
	Username       string `envconfig:"DB_USERNAME" required:"true"`
	Password       string `envconfig:"DB_PASSWORD"`
	LogLevel       int    `envconfig:"DB_LOG_LEVEL" default:"0"` // panic=5 fatal=4 error=3 warn=2 info=1 debug=0
	MigrationsPath string `envconfig:"DB_MIGRATIONS_PATH" default:"internal/migrations"`
}

type AWSCfg struct {
	RDSMasterPasswordSecretID string `envconfig:"AWS_RDS_MASTER_PASSWORD_SECRET_ID"`
}

func NewConfig(dotEnvPath string) (*Config, error) {
	_, err := os.Stat(dotEnvPath)
	if !errors.Is(err, os.ErrNotExist) {
		err := godotenv.Load(dotEnvPath)
		if err != nil {
			return nil, err
		}
	}

	c := &Config{}
	err = envconfig.Process("", c)
	if err != nil {
		return nil, err
	}

	if c.DB.Password == "" && c.AWS.RDSMasterPasswordSecretID == "" {
		return nil, ErrDBPasswordNotSet
	}

	return c, nil
}
