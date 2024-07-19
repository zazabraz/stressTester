package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"stress-tester/internal/app"
	"stress-tester/internal/domain/worker/quiz"
	"stress-tester/internal/infra/tester"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	rootLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	workers := flag.Int("wrk", 1, "number of concurrent workers")
	flag.Parse()

	testerInst := tester.New(*rootLogger, *workers, quiz.New(*rootLogger))

	appInst := app.New(*rootLogger, testerInst)

	err := appInst.RunApp(ctx)
	if err != nil {
		return
	}
}
