package main

import "fmt"
import js "github.com/realint/monkey"

func main() {
	// Create Script Runtime
	runtime, err1 := js.NewRuntime()
	if err1 != nil {
		panic(err1)
	}

	// Register Error Reporter
	runtime.SetErrorReporter(func(report *js.ErrorReport) {
		println(fmt.Sprintf(
			"%s:%d: %s",
			report.FileName, report.LineNum, report.Message,
		))
		if report.LineBuf != "" {
			println("\t", report.LineBuf)
		}
	})

	// Evaluate Script
	if value, err := runtime.Eval("'Hello ' + 'World!'"); err == nil {
		println(value.ToString())
	}

	// Built-in Functions
	runtime.Eval("println('Hello World!')")

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

	// Compile Script
	if script, err := runtime.Compile("add(1,2)", "<no name>", 0); err == nil {
		script.Execute()
	}

	// Script Error
	runtime.Eval("abc()")

	// Say Good Bye
	runtime.Dispose()
}
