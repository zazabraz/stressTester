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

	q.log.Info(fmt.Errorf("workers:%v", q.workersN).Error())
	wg := sync.WaitGroup{}
	for i := 1; i <= q.workersN; i++ {
		wg.Add(1)
		go func() {
			wrk := newWorker(q.log)
			defer wg.Done()
			start := time.Now()
			res, err := wrk.doWork(ctx)
			if err != nil {
				errCh <- err
			}
			q.log.Info(
				fmt.Sprintf(
					"quiz worker %v ends with duration:%s final page: \n %s",
					i,
					time.Now().Sub(start),
					res,
				),
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
			q.log.Error(fmt.Errorf("tester err: %w", err).Error())
		}
	}

}
