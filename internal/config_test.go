package internal_test

import (
	"os"
	"testing"

	"github.com/aarregui/go-deploy-tf-aws/internal"
	"github.com/stretchr/testify/assert"
)

func Test_Load_Success(t *testing.T) {
	override := "lorem"
	os.Setenv("VERSION", override)
	c, err := internal.NewConfig("../.env")

	assert.Nil(t, err)
	assert.Equal(t, override, c.Version)
}

func Test_Load_Missing_DB_Password(t *testing.T) {
	cfg, err := internal.NewConfig("../.env")
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.Setenv("DB_PASSWORD", cfg.DB.Password)
		os.Setenv("AWS_RDS_MASTER_PASSWORD_SECRET_ID", cfg.AWS.RDSMasterPasswordSecretID)
	})

	os.Setenv("DB_PASSWORD", "")
	os.Setenv("AWS_RDS_MASTER_PASSWORD_SECRET_ID", "")
	_, err = internal.NewConfig("../.env")

	assert.Equal(t, internal.ErrDBPasswordNotSet, err)
}
