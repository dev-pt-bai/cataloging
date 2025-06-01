package manager

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"runtime/debug"
	"sync"
	"time"

	"github.com/dev-pt-bai/cataloging/configs"
	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

type Broker interface {
	LPush(ctx context.Context, key string, values ...any) *redis.IntCmd
	BRPop(ctx context.Context, timeout time.Duration, keys ...string) *redis.StringSliceCmd
	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd
	ZRem(ctx context.Context, key string, members ...any) *redis.IntCmd
	RPush(ctx context.Context, key string, values ...any) *redis.IntCmd
	Close() error
}

type TaskManager struct {
	broker       Broker
	mu           sync.Mutex
	keyName      string
	retryKeyName string
	handlers     map[string]func(json.RawMessage) error
	numWorkers   int
	maxRetries   int
	tasks        chan *Task
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	once         sync.Once
	timeNow      func() time.Time
}

type Task struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Data       json.RawMessage `json:"data"`
	RetryCount int             `json:"retryCount,omitempty"`
}

func New(broker Broker, config *configs.Config) (*TaskManager, error) {
	if config == nil {
		return nil, fmt.Errorf("missing config")
	}

	if len(config.App.Async.KeyName) == 0 {
		return nil, fmt.Errorf("missing async task manager key name")
	}

	if len(config.App.Async.RetryKeyName) == 0 {
		return nil, fmt.Errorf("missing async task manager retry key name")
	}

	if broker == nil {
		return nil, fmt.Errorf("missing broker")
	}

	numWorkers := min(5, config.App.Async.NumWorkers)
	numWorkers = max(2, numWorkers)

	maxRetries := min(5, config.App.Async.MaxRetries)
	maxRetries = max(2, maxRetries)

	return &TaskManager{
		broker:       broker,
		keyName:      config.App.Async.KeyName,
		retryKeyName: config.App.Async.RetryKeyName,
		handlers:     make(map[string]func(json.RawMessage) error),
		numWorkers:   numWorkers,
		maxRetries:   maxRetries,
		tasks:        make(chan *Task, 2*numWorkers),
		timeNow:      time.Now,
	}, nil
}

func (tm *TaskManager) HandleFunc(taskType string, handler func(json.RawMessage) error) {
	if len(taskType) == 0 {
		panic("manager.HandleFunc: invalid taskType")
	}

	if handler == nil {
		panic("manager.HandleFunc: nil handler")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()
	_, exists := tm.handlers[taskType]
	if exists {
		panic(fmt.Sprintf("manager.HandleFunc: taskType %s conflicts", taskType))
	}

	tm.handlers[taskType] = handler
}

func (tm *TaskManager) ListenAndServe(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	tm.cancel = cancel

	eg, ctx := errgroup.WithContext(ctx)

	// using Go 1.24
	for i := range tm.numWorkers {
		tm.wg.Add(1)
		eg.Go(func() error {
			defer tm.wg.Done()
			tm.worker(ctx, i)
			return nil
		})
	}

	tm.wg.Add(1)
	eg.Go(func() error {
		defer tm.wg.Done()
		tm.dispatcher(ctx)
		return nil
	})

	tm.wg.Add(1)
	eg.Go(func() error {
		defer tm.wg.Done()
		tm.retryDispatcher(ctx)
		return nil
	})

	return eg.Wait()
}

func (tm *TaskManager) Enqueue(ctx context.Context, task *Task) *errors.Error {
	if task == nil || task.ID == "" || task.Type == "" {
		return errors.New(errors.InvalidTask)
	}

	data, err := json.Marshal(task)
	if err != nil {
		return errors.New(errors.JSONEncodeFailure).Wrap(err)
	}

	if err := tm.broker.LPush(ctx, tm.keyName, data).Err(); err != nil {
		return errors.New(errors.EnqueueTaskFailure).Wrap(err)
	}

	return nil
}

func (tm *TaskManager) dispatcher(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		result, err := tm.broker.BRPop(ctx, time.Second, tm.keyName).Result()
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return
			}

			if err == redis.Nil {
				continue
			}

			slog.Error("manager.dispatcher: failed to BRPop from Redis", slog.String("cause", err.Error()))
			continue
		}

		if len(result) != 2 {
			slog.Error("manager.dispatcher: invalid BRPop result", slog.Any("result", result))
			continue
		}

		task := new(Task)
		if err := json.Unmarshal([]byte(result[1]), task); err != nil {
			slog.Error("manager.dispatcher: failed to parse task", slog.String("cause", err.Error()))
			continue
		}

		if len(task.Type) == 0 {
			slog.Error("manager.dispatcher: invalid task type", slog.String("id", task.ID))
			continue
		}

		select {
		case tm.tasks <- task:
		case <-time.After(time.Second):
			slog.Warn("manager.dispatcher: enqueue timeout, attempting retry", slog.String("id", task.ID), slog.String("type", task.Type))
			tm.retrier(ctx, task, false)
		}
	}
}

func (tm *TaskManager) retryDispatcher(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := tm.timeNow().Unix()
			tasks, err := tm.broker.ZRangeByScore(ctx, tm.retryKeyName, &redis.ZRangeBy{
				Min: "0",
				Max: fmt.Sprintf("%d", now),
			}).Result()
			if err != nil {
				slog.Error("manager.retryDispatcher: failed to fetch retry tasks", slog.String("cause", err.Error()))
				continue
			}

			for _, raw := range tasks {
				var task Task
				if err := json.Unmarshal([]byte(raw), &task); err != nil {
					slog.Error("manager.retryDispatcher: failed to parse retry task", slog.String("cause", err.Error()), slog.String("type", task.Type), slog.String("id", task.ID))
					tm.broker.ZRem(ctx, tm.retryKeyName, raw)
					continue
				}

				tm.broker.ZRem(ctx, tm.retryKeyName, raw)

				if err := tm.broker.RPush(ctx, tm.keyName, raw).Err(); err != nil {
					slog.Error("manager.retryDispatcher: failed to requeue task", slog.String("cause", err.Error()), slog.String("type", task.Type), slog.String("id", task.ID))
				}
			}
		}
	}
}

func (tm *TaskManager) worker(ctx context.Context, id int) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("manager.worker: panic: ", slog.String("stack", string(debug.Stack())))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tm.tasks:
			if !ok {
				slog.Debug("manager.worker: task channel closed, exiting", slog.Int("worker", id))
				return
			}

			tm.mu.Lock()
			handler, exists := tm.handlers[task.Type]
			tm.mu.Unlock()
			if !exists {
				slog.Error("manager.worker: unknown task type", slog.String("type", task.Type), slog.String("id", task.ID))
				continue
			}

			if err := handler(task.Data); err != nil {
				slog.Error("manager.worker: failed to handle task, attempting retry", slog.String("cause", err.Error()), slog.String("type", task.Type), slog.String("id", task.ID))
				tm.retrier(ctx, task, true)
			} else {
				slog.Debug("manager.worker: task handled", slog.Int("worker", id), slog.String("type", task.Type), slog.String("id", task.ID))
			}
		}
	}
}

func (tm *TaskManager) retrier(ctx context.Context, task *Task, shouldIncrement bool) {
	if task == nil {
		return
	}

	if shouldIncrement {
		task.RetryCount++
		if task.RetryCount > tm.maxRetries {
			slog.Error("manager.retrier: max retries reached, dropping task", slog.Int("maxRetries", tm.maxRetries), slog.String("type", task.Type), slog.String("id", task.ID))
			return
		}
		slog.Info("manager.retrier: retry task", slog.Int("retryCount", task.RetryCount), slog.String("type", task.Type), slog.String("id", task.ID))
	} else {
		slog.Info("manager.retrier: retry task due to backpressure or channel timeout", slog.String("type", task.Type), slog.String("id", task.ID))
	}

	backoff := time.Duration(1<<task.RetryCount) * time.Second
	backoff = min(5*time.Minute, backoff)
	nBig, err := rand.Int(rand.Reader, big.NewInt(backoff.Nanoseconds()))
	if err == nil {
		backoff = time.Duration(nBig.Int64())
	}
	retryTime := tm.timeNow().Add(backoff).Unix()

	data, err := json.Marshal(task)
	if err != nil {
		slog.Error("manager.retrier: failed to serialize task for retry", slog.String("cause", err.Error()), slog.String("type", task.Type), slog.String("id", task.ID))
		return
	}

	if err := tm.broker.ZAdd(ctx, tm.retryKeyName, redis.Z{
		Score:  float64(retryTime),
		Member: data,
	}).Err(); err != nil {
		slog.Error("manager.retrier: failed to schedule retry", slog.String("cause", err.Error()), slog.String("type", task.Type), slog.String("id", task.ID))
	}
}

func (tm *TaskManager) Shutdown(ctx context.Context) error {
	if tm.cancel != nil {
		tm.cancel()
	}

	tm.once.Do(func() { close(tm.tasks) })

	done := make(chan struct{})
	go func() {
		tm.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return tm.broker.Close()
	}
}
