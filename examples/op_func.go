package main

import js "github.com/realint/monkey"

func assert(c bool) bool {
	if !c {
		panic("assert failed")
	}
	return c
}

func main() {
	// Create Script Runtime
	runtime, err1 := js.NewRuntime()
	if err1 != nil {
		panic(err1)
	}

	// Function
	if value, err := runtime.Eval("function(a,b){ return a+b; }"); assert(err == nil) {
		// Type Check
		assert(value.IsFunction())

		// Call
		value1, ok1 := value.Call([]js.Value{
			runtime.Int(10),
			runtime.Int(20),
		})

		// Result Check
		assert(ok1)
		assert(value1.IsNumber())
		assert(value1.Int() == 30)
	}

	runtime.Dispose()
}
