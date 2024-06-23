package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	m "github.com/aarregui/go-deploy-tf-aws/internal/middleware"
)

type App interface {
	Serve() error
	Migrate() MigratorClient
}

type app struct {
	config   Config
	rdbms    RDBMSClient
	migrator MigratorClient
}

func NewApp(
	config Config,
	rdbms RDBMSClient,
	migrator MigratorClient,
) *app {
	return &app{config, rdbms, migrator}
}

func (a app) Serve() error {
	r := a.setRoutes()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    ":" + a.config.App.Port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("server failed to start")
		}
	}()
	log.Info().Msg(fmt.Sprintf("server version [%s] listening on post: %s", a.config.Version, a.config.App.Port))

	<-done
	log.Info().Msg("server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	log.Info().Msg("server exited properly")

	return nil
}

func (a app) Migrate() MigratorClient {
	return a.migrator
}

func (a app) setRoutes() *chi.Mux {
	r := chi.NewRouter()
	setMiddlewares(r)

	r.Get("/", func(rw http.ResponseWriter, r *http.Request) { rw.WriteHeader(http.StatusOK) })

	return r
}

func setMiddlewares(r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(m.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.NoCache)
}
