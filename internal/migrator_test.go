package internal_test

import (
	"log"
	"testing"

	"github.com/aarregui/go-deploy-tf-aws/internal"
	"github.com/stretchr/testify/assert"
)

func Test_Migrator(t *testing.T) {
	c := getTestDBConfig()
	db := internal.NewPostgres(c)
	_ = db.Open()
	defer db.Close()

	_, err := internal.NewMigrator(nil, c.DB.MigrationsPath)
	assert.Error(t, err)

	migrator, err := internal.NewMigrator(db.GetCon(), c.DB.MigrationsPath)
	if err != nil {
		log.Fatal(err)
	}

	err = migrator.Up()
	assert.Nil(t, err)

	err = migrator.Steps(0)
	assert.Nil(t, err)

	c.DB.MigrationsPath = "abc"

	_, err = internal.NewMigrator(db.GetCon(), c.DB.MigrationsPath)
	assert.Error(t, err)
}
