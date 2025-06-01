package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type task struct {
	ticker  *time.Ticker
	handler func() error
}

type TaskScheduler struct {
	mu       sync.Mutex
	handlers map[string]task
}

func New() *TaskScheduler {
	return &TaskScheduler{handlers: make(map[string]task)}
}

func (ts *TaskScheduler) HandleFunc(taskType string, duration time.Duration, handler func() error) {
	if duration <= 0 {
		panic("scheduler.HandleFunc: invalid duration")
	}

	if handler == nil {
		panic("scheduler.HandleFunc: nil handler")
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	_, exists := ts.handlers[taskType]
	if exists {
		panic(fmt.Sprintf("scheduler.HandleFunc: taskType %s conflicts", taskType))
	}

	ts.handlers[taskType] = task{
		ticker:  time.NewTicker(duration * time.Second),
		handler: handler,
	}
}

func (ts *TaskScheduler) Start(ctx context.Context) error {
	for i := range ts.handlers {
		handler := ts.handlers[i]
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-handler.ticker.C:
					if err := handler.handler(); err != nil {
						slog.ErrorContext(ctx, "scheduler.Start: running handler returns error", slog.String("cause", err.Error()))
					}
				}
			}
		}()
	}

	return ctx.Err()
}

func (ts *TaskScheduler) Stop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for i := range ts.handlers {
			ts.handlers[i].ticker.Stop()
		}
		return nil
	}
}
