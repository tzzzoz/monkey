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

	// String
	if value, err := runtime.Eval("'abc'"); err == nil {
		assert(value.IsString())
		assert(value.String() == "abc")
	}

	// Int
	if value, err := runtime.Eval("123456789"); err == nil {
		assert(value.IsInt())
		assert(value.Int() == 123456789)
	}

	// Number
	if value, err := runtime.Eval("12345.6789"); err == nil {
		assert(value.IsNumber())
		assert(value.Number() == 12345.6789)
	}

	runtime.Dispose()
}
