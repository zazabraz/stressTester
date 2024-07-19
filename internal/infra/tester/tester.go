package tester

import (
	"context"
	"fmt"
	"log/slog"
	"stress-tester/internal/domain/worker"
	"stress-tester/internal/infra"
	"sync"
)

type tester struct {
	log      slog.Logger
	workersN int
	wrk      worker.Worker
}

func New(log slog.Logger, workersN int, wrk worker.Worker) infra.Controller {
	return &tester{log: log, workersN: workersN, wrk: wrk}
}

func (t *tester) Run(ctx context.Context) error {
	t.log = *t.log.With("tester.Run")

	errCh := make(chan error)
	workDone := make(chan struct{})

	wg := sync.WaitGroup{}
	for i := 0; i <= t.workersN; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := t.wrk.DoWork(ctx)
			if err != nil {
				errCh <- err
			}
		}()
	}

	go func() {
		wg.Wait()
		workDone <- struct{}{}
	}()

	for {
		select {
		case <-workDone:
			return nil
		case err := <-errCh:
			t.log.Error(fmt.Errorf("tester err: %w", err).Error())
		}
	}
}
