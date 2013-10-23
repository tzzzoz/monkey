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

	// Array
	if value, err := runtime.Eval("[123, 456]"); err == nil {
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

	runtime.Dispose()
}
