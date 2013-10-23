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

	// Object
	if value, err := runtime.Eval("x={a:123}"); err == nil {
		// Type Check
		assert(value.IsObject())
		obj := value.Object()

		// Get Property
		value1, ok1 := obj.GetProperty("a")
		assert(ok1)
		assert(value1.IsInt())
		assert(value1.Int() == 123)

		// Set Property
		assert(obj.SetProperty("b", runtime.Int(456)))
		value2, ok2 := obj.GetProperty("b")
		assert(ok2)
		assert(value2.IsInt())
		assert(value2.Int() == 456)
	}

	runtime.Dispose()
}
