What is
============

SpiderMonkey binding for go.

Just newborn, use in production enviroment at your own risk!

Install
=======

Mac OS X:

    brew install spidermonkey
    export CGO_LDFLAGS="-L /usr/local/Cellar/spidermonkey/1.8.5/lib/ -lmozjs185.1.0"
    go get github.com/realint/monkey

Examples
========

All the example codes can be found in examples folder.

Maybe you need this to fix CGO problem:

    export CGO_LDFLAGS="-L /usr/local/Cellar/spidermonkey/1.8.5/lib/ -lmozjs185.1.0"

The "hello\_world.go" shows what Monkey can do.

```go
package main

import "fmt"
import js "github.com/realint/monkey"

func main() {
	// Create Script Runtime
	runtime, err1 := js.NewRuntime()
	if err1 != nil {
		panic(err1)
	}

	// Evaluate Script
	if value, err := runtime.Eval("'Hello ' + 'World!'"); err == nil {
		println(value.ToString())
	}

	// Built-in Function
	runtime.Eval("println('Hello Built-in Function!')")

	// Compile Once, Run Many Times
	if script, err := runtime.Compile(
		"println('Hello Compiler!')",
		"<no name>", 0,
	); err == nil {
		script.Execute()
		script.Execute()
		script.Execute()
	}

	// Define Function
	if err := runtime.DefineFunction("add",
		func(argv []js.Value) (js.Value, bool) {
			if len(argv) != 2 {
				return runtime.Null(), false
			}
			return runtime.Int(argv[0].Int() + argv[1].Int()), true
		},
	); err == nil {
		if value, err := runtime.Eval("add(100, 200)"); err == nil {
			println(value.Int())
		}
	}

	// Error Handle
	runtime.SetErrorReporter(func(report *js.ErrorReport) {
		println(fmt.Sprintf(
			"%s:%d: %s",
			report.FileName, report.LineNum, report.Message,
		))
		if report.LineBuf != "" {
			println("\t", report.LineBuf)
		}
	})

	// Trigger An Error
	runtime.Eval("abc()")

	// Say Good Bye
	runtime.Dispose()
}
```
This code will output:

```
Hello World!
Hello Built-in Function!
Hello Compiler!
Hello Compiler!
Hello Compiler!
300
Eval():0: ReferenceError: abc is not defined
```

The "op_value.go" shows how to convert JS value to Go value.

```go
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
```

The "op_object.go" shows how to play with JS object.

```go
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
```
The "op_array.go" shows how to play with JS array.

```go
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
```

The "op_func.go" shows how to play with JS function value.

```go
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
```

The "many\_many.go" shows Monkey is thread safe.

```go
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

	// One Runtime Instance Used By Many Goroutines
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
```
