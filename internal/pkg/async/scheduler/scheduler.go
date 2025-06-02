package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
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
	wg       sync.WaitGroup
}

func New() *TaskScheduler {
	return &TaskScheduler{handlers: make(map[string]task)}
}

func (ts *TaskScheduler) HandleFunc(taskType string, intervalSec int, handler func() error) {
	if intervalSec <= 0 {
		panic("scheduler.HandleFunc: invalid interval")
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
		ticker:  time.NewTicker(time.Duration(intervalSec) * time.Second),
		handler: handler,
	}
}

func (ts *TaskScheduler) Start(ctx context.Context) error {
	ts.mu.Lock()
	handlers := make(map[string]task, len(ts.handlers))
	maps.Copy(handlers, ts.handlers)
	ts.mu.Unlock()

	for _, handler := range handlers {
		ts.wg.Add(1)
		go func(h task) {
			defer ts.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case <-h.ticker.C:
					if err := h.handler(); err != nil {
						slog.ErrorContext(ctx, "scheduler.Start: running handler returns error", slog.String("cause", err.Error()))
					}
				}
			}
		}(handler)
	}

	<-ctx.Done()
	return nil
}

func (ts *TaskScheduler) Stop(ctx context.Context) error {
	ts.mu.Lock()
	handlers := make(map[string]task, len(ts.handlers))
	maps.Copy(handlers, ts.handlers)
	ts.mu.Unlock()

	for _, h := range handlers {
		h.ticker.Stop()
	}

	done := make(chan struct{})
	go func() {
		ts.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
