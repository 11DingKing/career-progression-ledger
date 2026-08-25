package main

import (
	"careerprogression/internal/config"
	"careerprogression/internal/handler"
	"careerprogression/internal/migrations"
	"careerprogression/internal/repository"
	"careerprogression/internal/service"
	"careerprogression/internal/store"
	"careerprogression/internal/worker"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	db, e := store.Open(ctx, cfg.DBPath)
	if e != nil {
		panic(e)
	}
	defer db.Close()
	if e = migrations.Apply(ctx, db.SQL); e != nil {
		panic(e)
	}
	repo := repository.New(db.SQL)
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	authSvc := &service.AuthService{DB: db, Repo: repo, TTL: cfg.SessionTTL}
	career := &service.CareerService{DB: db, Repo: repo}
	w := &worker.Worker{DB: db, Every: cfg.WorkerInterval, Log: log}
	go w.Run(ctx)
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: handler.Logging(log, (&handler.Handler{Auth: authSvc, Career: career, DB: db, Log: log}).Routes()), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_ = srv.Shutdown(shutdown)
	}()
	if e = srv.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
		panic(e)
	}
}
