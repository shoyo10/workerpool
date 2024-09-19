package workerpool

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
)

type WorkerPool struct {
	workerNums         int
	taskQueueSize      int
	taskQueue          chan TaskFunc
	taskQueueCloseOnce sync.Once
	runningWorkerCount int
	mutex              sync.Mutex
	ctx                context.Context
	ctxCancel          context.CancelFunc
	workerWg           sync.WaitGroup
	stopped            int32
}

type TaskFunc func()

func New(workerNums int, taskQueueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	wp := &WorkerPool{
		workerNums:    workerNums,
		taskQueueSize: taskQueueSize,
		taskQueue:     make(chan TaskFunc, taskQueueSize),
		ctx:           ctx,
		ctxCancel:     cancel,
	}

	for i := 0; i < workerNums; i++ {
		wp.workerWg.Add(1)
		wp.runningWorkerCount++
		go wp.worker()
	}

	return wp
}

func (wp *WorkerPool) AddTask(task TaskFunc) {
	if task == nil {
		return
	}

	if wp.Stopped() {
		return
	}

	wp.mutex.Lock()
	if wp.runningWorkerCount < wp.workerNums {
		n := wp.workerNums - wp.runningWorkerCount
		for i := 0; i < n; i++ {
			wp.workerWg.Add(1)
			wp.runningWorkerCount++
			go wp.worker()
		}
	}
	wp.mutex.Unlock()

	select {
	case <-wp.ctx.Done():
		return
	case wp.taskQueue <- task:
	}
}

// Stop stops the worker pool.
func (wp *WorkerPool) Stop() {
	atomic.StoreInt32(&wp.stopped, 1)
	wp.ctxCancel()
	wp.workerWg.Wait()
	wp.taskQueueCloseOnce.Do(func() {
		close(wp.taskQueue)
	})
}

// Stopped returns true if the worker pool has been stopped.
func (wp *WorkerPool) Stopped() bool {
	return atomic.LoadInt32(&wp.stopped) == 1
}

func (wp *WorkerPool) worker() {
	defer func() {
		if r := recover(); r != nil {
			var msg string
			for i := 2; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				msg += fmt.Sprintf("%s:%d\n", file, line)
			}
			log.Printf("========== PANIC Start ==========\n%s\n%s========== PANIC End ==========", r, msg)
		}

		wp.workerWg.Done()

		wp.mutex.Lock()
		wp.runningWorkerCount--
		wp.mutex.Unlock()
	}()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case task, ok := <-wp.taskQueue:
			select {
			case <-wp.ctx.Done():
				return
			default:
				if task == nil || !ok {
					return
				}
				task()
			}
		}
	}
}
