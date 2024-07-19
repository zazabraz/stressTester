package tester

import (
	"context"
	"fmt"
	"log/slog"
	"stress-tester/internal/domain/worker"
	"stress-tester/internal/infra"
)

type tester struct {
	log slog.Logger
	wrk testWorker.TestWorker
}

func New(log slog.Logger, wrk testWorker.TestWorker) infra.Controller {
	return &tester{log: log, wrk: wrk}
}

func (t *tester) Run(ctx context.Context) error {
	t.log = *t.log.With("tester.Run")
	err := t.wrk.TestWork(ctx)
	if err != nil {
		return fmt.Errorf("TestWork: %w", err)
	}
	return nil
}
