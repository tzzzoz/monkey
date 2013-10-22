package main

import "sync"
import "runtime"
import js "github.com/realint/monkey"

func main() {
	runtime.GOMAXPROCS(20)

	// Create Script Runtime
	runtime, err1 := js.NewRuntime()
	if err1 != nil {
		panic(err1)
	}

	wg := new(sync.WaitGroup)

	// Evaluate Script By Many Goroutines
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000; j++ {
				runtime.Eval("println('Hello World!')")
			}
			wg.Done()
		}()
	}

	wg.Wait()

	// Say Good Bye
	runtime.Dispose()
}
