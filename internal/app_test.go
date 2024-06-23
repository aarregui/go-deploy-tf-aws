package internal_test

import (
	"log"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/aarregui/go-deploy-tf-aws/internal"
	"github.com/stretchr/testify/assert"
)

func getTestConfig() internal.Config {
	c, err := internal.NewConfig("../.env")
	if err != nil {
		log.Fatal(err)
	}

	c.DB.Host = "localhost"
	c.DB.MigrationsPath = "migrations" //todo: have separate test mgirations

	return *c
}

func Test_App_Serve(t *testing.T) {
	c := getTestConfig()
	db := internal.NewPostgres(c)
	t.Cleanup(func() {
		db.Close()
	})

	app := internal.NewApp(c, db, nil)

	go func() {
		_ = app.Serve()
	}()

	time.Sleep(time.Second)

	res, err := http.Get("http://localhost:" + c.App.Port)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(time.Second)
}
