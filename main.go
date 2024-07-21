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

var (
	workers = flag.Int("w", 10, "number of concurrent workers")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()
	rootLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	flag.Parse()

	testerInst := tester.New(*rootLogger, quiz.New(*rootLogger, *workers))

	appInst := app.New(*rootLogger, testerInst)

	err := appInst.RunApp(ctx)
	if err != nil {
		return
	}
}
