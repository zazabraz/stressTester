package testWorker

import (
	"context"
)

type TestWorker interface {
	TestWork(ctx context.Context) error
}
