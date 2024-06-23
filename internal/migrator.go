package internal

import (
	"errors"
	"fmt"

	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"

	"github.com/golang-migrate/migrate/v4"
)

type MigratorClient interface {
	Up() error
	Steps(n int) error
}

type migrator struct {
	client *migrate.Migrate
}

func NewMigrator(db *gorm.DB, migrationsPath string) (*migrator, error) {
	if db == nil {
		return nil, errors.New("missing db connection")
	}

	mainDB, _ := db.DB()
	driver, err := migratePostgres.WithInstance(mainDB, &migratePostgres.Config{})
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("file://%s", migrationsPath)
	client, err := migrate.NewWithDatabaseInstance(
		path,
		"postgres",
		driver,
	)
	if err != nil {
		return nil, err
	}

	return &migrator{client}, nil
}

func (m migrator) Up() error {
	err := m.client.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (m migrator) Steps(n int) error {
	err := m.client.Steps(n)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
