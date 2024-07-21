package quiz

import (
	"context"
	"fmt"
	"log/slog"
	"stress-tester/internal/domain/worker"
	"sync"
	"time"
)

type quiz struct {
	log      slog.Logger
	workersN int
}

func New(log slog.Logger, workersN int) testWorker.TestWorker {
	return &quiz{log: log, workersN: workersN}
}

func (q *quiz) TestWork(ctx context.Context) error {
	errCh := make(chan error)
	workDone := make(chan struct{})

	q.log.Info(fmt.Errorf("start testing quiz with %v workers", q.workersN).Error())
	wg := sync.WaitGroup{}

	rateLimiter := make(chan struct{}, 3)

	for i := 1; i <= q.workersN; i++ {
		wg.Add(1)
		go func() {
			wrk := newWorker(rateLimiter)
			defer wg.Done()
			start := time.Now()
			res, err := wrk.doWork(ctx)
			if err != nil {
				errCh <- err
			}
			q.log.Info("Quiz worker ends",
				slog.Int("worker", i),
				slog.Duration("duration", time.Since(start)),
				slog.String("final_page", res),
			)
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
			q.log.Error("quiz error",
				slog.String("error", err.Error()),
			)
			return err
		}
	}

}
