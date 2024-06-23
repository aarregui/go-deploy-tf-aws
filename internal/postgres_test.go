package internal_test

import (
	"log"
	"testing"

	"github.com/aarregui/go-deploy-tf-aws/internal"
	"github.com/stretchr/testify/assert"
)

func getTestDBConfig() internal.Config {
	c, err := internal.NewConfig("../.env")
	if err != nil {
		log.Fatal(err)
	}

	c.DB.Host = "localhost"
	c.DB.MigrationsPath = "migrations" //todo: have separate test mgirations

	return *c
}

func Test_Postgres_Open(t *testing.T) {
	c := getTestDBConfig()
	db := internal.NewPostgres(c)

	err := db.Close()
	assert.ErrorIs(t, err, internal.ErrOpenNotCalled)

	err = db.Open()
	assert.Nil(t, err)

	assert.NotNil(t, db.GetCon())

	err = db.Open()
	assert.ErrorIs(t, err, internal.ErrConectionAlreadyEstablished)

	err = db.Close()
	assert.Nil(t, err)

	assert.Nil(t, db.GetCon())
}

func Test_Postgres_Open_Bad_Host(t *testing.T) {
	c := getTestConfig()
	c.DB.Host = "lorem"

	db := internal.NewPostgres(c)

	err := db.Open()

	assert.Error(t, err)
}
