package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

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
			slog.Error("server failed to start", "error", err)
		}
	}()
	slog.Info(fmt.Sprintf("server version [%s] listening on post: %s", a.config.Version, a.config.App.Port))

	<-done
	slog.Info("server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}

	slog.Info("server exited properly")

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
