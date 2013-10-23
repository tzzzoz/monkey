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

	// Return Array From JavaScript
	if value, err := runtime.Eval("[123, 456]"); assert(err == nil) {
		// Type Check
		assert(value.IsObject())
		obj := value.Object()
		assert(obj.IsArray())
		assert(obj.GetArrayLength() == 2)

		// Get First Item
		value1, ok1 := obj.GetElement(0)
		assert(ok1)
		assert(value1.IsInt())
		assert(value1.Int() == 123)

		// Get Second Item
		value2, ok2 := obj.GetElement(1)
		assert(ok2)
		assert(value2.IsInt())
		assert(value2.Int() == 456)

		// Set First Item
		assert(obj.SetElement(0, runtime.Int(789)))
		value3, ok3 := obj.GetElement(0)
		assert(ok3)
		assert(value3.IsInt())
		assert(value3.Int() == 789)

		// Grows
		assert(obj.SetArrayLength(3))
		assert(obj.GetArrayLength() == 3)
	}

	// Return Array From Go
	if err := runtime.DefineFunction("get_data",
		func(argv []js.Value) (js.Value, bool) {
			array := runtime.NewArray()
			array.SetElement(0, runtime.Int(100))
			array.SetElement(1, runtime.Int(200))
			return array.ToValue(), true
		},
	); err == nil {
		if value, err := runtime.Eval("get_data()"); assert(err == nil) {
			// Type Check
			assert(value.IsObject())
			obj := value.Object()
			assert(obj.IsArray())
			assert(obj.GetArrayLength() == 2)

			// Get First Item
			value1, ok1 := obj.GetElement(0)
			assert(ok1)
			assert(value1.IsInt())
			assert(value1.Int() == 100)

			// Get Second Item
			value2, ok2 := obj.GetElement(1)
			assert(ok2)
			assert(value2.IsInt())
			assert(value2.Int() == 200)
		}
	}

	runtime.Dispose()
}
