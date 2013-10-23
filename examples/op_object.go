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

	// Return Object From JavaScript
	if value, err := runtime.Eval("x={a:123}"); assert(err == nil) {
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

	// Return Object From Go
	if err := runtime.DefineFunction("get_data",
		func(argv []js.Value) (js.Value, bool) {
			obj := runtime.NewObject()
			obj.SetProperty("abc", runtime.Int(100))
			obj.SetProperty("def", runtime.Int(200))
			return obj.ToValue(), true
		},
	); err == nil {
		if value, err := runtime.Eval("get_data()"); assert(err == nil) {
			// Type Check
			assert(value.IsObject())
			obj := value.Object()

			// Get Property 'abc'
			value1, ok1 := obj.GetProperty("abc")
			assert(ok1)
			assert(value1.IsInt())
			assert(value1.Int() == 100)

			// Get Property 'def'
			value2, ok2 := obj.GetProperty("def")
			assert(ok2)
			assert(value2.IsInt())
			assert(value2.Int() == 200)
		}
	}

	runtime.Dispose()
}
