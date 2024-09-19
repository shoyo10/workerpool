package workerpool

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	beforeGoNums := runtime.NumGoroutine()
	workerNums := 5
	taskQueueSize := 10
	wp := New(workerNums, taskQueueSize)
	time.Sleep(1 * time.Millisecond)
	afterGoNums := runtime.NumGoroutine()

	assert.Equal(t, workerNums, wp.workerNums)
	assert.Equal(t, taskQueueSize, wp.taskQueueSize)
	assert.Equal(t, wp.runningWorkerCount, workerNums)
	assert.Equal(t, workerNums, afterGoNums-beforeGoNums)
}

func TestAddTask(t *testing.T) {
	workerNums := 2
	taskQueueSize := 2
	wp := New(workerNums, taskQueueSize)

	var count int32
	wp.AddTask(func() {
		atomic.AddInt32(&count, 1)
	})
	wp.AddTask(func() {
		atomic.AddInt32(&count, 1)
	})
	time.Sleep(1 * time.Millisecond)

	assert.Equal(t, int32(2), atomic.LoadInt32(&count))

	wp.Stop()
}

func TestAddTaskWithPanic(t *testing.T) {
	workerNums := 2
	taskQueueSize := 2
	wp := New(workerNums, taskQueueSize)

	var count int32
	// add a task that will panic
	wp.AddTask(func() {
		s := make([]int, 1)
		s[1] = 1
		atomic.AddInt32(&count, 1)
	})
	wp.AddTask(func() {
		atomic.AddInt32(&count, 1)
	})
	time.Sleep(1 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&count))
	assert.Equal(t, 1, wp.runningWorkerCount)

	wp.AddTask(func() {
		atomic.AddInt32(&count, 1)
	})
	assert.Equal(t, 2, wp.runningWorkerCount)

	wp.Stop()
}

func TestStop(t *testing.T) {
	beforeGoNums := runtime.NumGoroutine()
	workerNums := 2
	taskQueueSize := 2
	wp := New(workerNums, taskQueueSize)

	time.Sleep(1 * time.Millisecond)
	afterGoNums := runtime.NumGoroutine()
	assert.Equal(t, 2, afterGoNums-beforeGoNums)

	ok := wp.Stopped()
	assert.False(t, ok)

	wp.Stop()

	afterStopGoNums := runtime.NumGoroutine()
	assert.Equal(t, 0, afterStopGoNums-beforeGoNums)
	var count int32
	wp.AddTask(func() {
		atomic.AddInt32(&count, 1)
	})
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(&count))
	ok = wp.Stopped()
	assert.True(t, ok)

	_, ok = <-wp.taskQueue
	assert.False(t, ok)
}
