package pool

import (
	"context"
)

type (
	IJob interface {
		Do() error
	}

	Result struct {
		Url       string
		RequestID string
		Error     error
	}

	Pool struct {
		taskQueue   chan IJob
		resultChan  chan Result
		workerCount int
	}

	Worker struct {
		id        int
		taskQueue <-chan IJob
		result    chan<- Result
	}
)

func NewWorkerPool(workerCount int) *Pool {
	return &Pool{
		taskQueue:   make(chan IJob),
		resultChan:  make(chan Result),
		workerCount: workerCount,
	}
}

func (wp *Pool) Start(ctx context.Context) {
	for i := 0; i < wp.workerCount; i++ {
		worker := Worker{id: i, taskQueue: wp.taskQueue, result: wp.resultChan}
		worker.Start(ctx)
	}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
	sender:
		for task := range w.taskQueue {
			select {
			case <-ctx.Done():
				break sender
			default:
			}
			err := task.Do()
			if err != nil {
				//w.result <- Result{Url: task.Url, Error: err}
			}

		}
	}()
}

func (wp *Pool) Submit(job IJob) {
	wp.taskQueue <- job
}

func (wp *Pool) GetResult() chan Result {
	return wp.resultChan
}
