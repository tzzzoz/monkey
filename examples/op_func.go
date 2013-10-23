package main

import js "github.com/realint/monkey"

func assert(c bool) {
	if !c {
		panic("assert failed")
	}
}

func main() {
	// Create Script Runtime
	runtime, err1 := js.NewRuntime()
	if err1 != nil {
		panic(err1)
	}

	// Function
	if value, err := runtime.Eval("function(a,b,c){ return a+b+c; }"); err == nil {
		// Type Check
		assert(value.IsFunction())

		// Call
		value1, ok1 := value.Call([]js.Value{
			runtime.Int(10),
			runtime.Int(20),
			runtime.Int(30),
		})

		// Result Check
		assert(ok1)
		assert(value1.IsNumber())
		assert(value1.Int() == 60)
	}

	runtime.Dispose()
}
