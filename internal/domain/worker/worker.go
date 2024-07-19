package worker

import "context"

type Worker interface {
	DoWork(ctx context.Context) error
}
