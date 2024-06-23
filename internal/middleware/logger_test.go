package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aarregui/go-deploy-tf-aws/internal/middleware"
	"github.com/go-chi/chi/v5"
	m "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func Test_Logger(t *testing.T) {
	method := http.MethodGet
	uri := "/"
	req := httptest.NewRequest(method, uri, nil)
	req.Host = "localhost:8080"
	buf := new(bytes.Buffer)

	log.Logger = log.Output(buf)

	testRequest(t, req, middleware.Logger)

	dec := json.NewDecoder(buf)
	m := make(map[string]interface{})
	if err := dec.Decode(&m); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, method, m["method"])
	assert.Equal(t, uri, m["uri"])
	assert.Equal(t, float64(http.StatusOK), m["status_code"])
	assert.Greater(t, m["bytes"], float64(0))
	assert.Greater(t, m["duration"], float64(0))
	assert.NotEmpty(t, m["duration_display"])
	assert.NotEmpty(t, m["remote_ip"])
	assert.Equal(t, "HTTP/1.1", m["proto"])
}

func Test_Logger_RequestID(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "localhost:8080"
	buf := new(bytes.Buffer)

	log.Logger = log.Output(buf)

	testRequest(t, req, m.RequestID, middleware.Logger)

	dec := json.NewDecoder(buf)
	m := make(map[string]interface{})
	if err := dec.Decode(&m); err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, m["request_id"])
}

// func Test_Logger_HTTPS(t *testing.T) {
// 	logger, hook := test.NewNullLogger()
// 	r := chi.NewRouter()
// 	r.Use(middleware.Logger(logger))
// 	r.Get("/", func(w http.ResponseWriter, r *http.Request) {})
// 	server := httptest.NewUnstartedServer(r)
// 	server.TLS = &tls.Config{
// 		InsecureSkipVerify: true,
// 	}
// 	server.StartTLS()

// 	log.Println(server.URL)
// 	res, err := http.Get(server.URL)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	assert.Equal(t, http.StatusOK, res.StatusCode)
// 	assert.Equal(t, 1, len(hook.Entries))
// 	assert.Equal(t, "/", hook.LastEntry().Data["uri"])
// }

func testRequest(t *testing.T, req *http.Request, middleware ...func(h http.Handler) http.Handler) chi.Router {
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Use(middleware...)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello World"))
	})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Response status code")

	return r
}
