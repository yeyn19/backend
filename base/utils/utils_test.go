package utils

import (
	"bytes"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/leoleoasd/EduOJBackend/base"
	"github.com/leoleoasd/EduOJBackend/base/config"
	"github.com/leoleoasd/EduOJBackend/base/log"
	"github.com/leoleoasd/EduOJBackend/database"
	"github.com/minio/minio-go"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	defer database.SetupDatabaseForTest()()
	PanicIfDBError(base.DB.AutoMigrate(&TestObject{}), "could not create table for test object")

	configFile := bytes.NewBufferString(`debug: true
server:
  port: 8080
  origin:
    - http://127.0.0.1:8000
`)

	if err := config.ReadConfig(configFile); err != nil {
		panic(err)
	}

	// fake s3
	faker := gofakes3.New(s3mem.New()) // in-memory s3 server.
	ts := httptest.NewServer(faker.Server())
	defer ts.Close()
	var err error
	base.Storage, err = minio.NewWithRegion(ts.URL[7:], "", "", false, "us-east-1")
	if err != nil {
		panic(err)
	}
	log.Disable()

	os.Exit(m.Run())
}