# workerpool

Golang worker pool

## How to use

``` go
package main

import (
	"fmt"
	"time"

	"github.com/shoyo10/workerpool"
)

func main() {
	wp := workerpool.New(10, 100)

	for i := 0; i < 100; i++ {
		n := i
		wp.AddTask(func() {
			fmt.Println("Task:", n)
		})
	}

	time.Sleep(1 * time.Millisecond)

	wp.Stop()
}

```
