package app

import (
	"context"
	"log/slog"
	"stress-tester/internal/infra"

	"golang.org/x/sync/errgroup"
)

type Runner interface {
	RunApp(ctx context.Context) error
}

type app struct {
	log    slog.Logger
	tester infra.Controller
}

func New(log slog.Logger, tester infra.Controller) Runner {
	return &app{log: log, tester: tester}
}

func (a *app) RunApp(ctx context.Context) error {
	grp, ctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return a.tester.Run(ctx)
	})

	return grp.Wait()
}
