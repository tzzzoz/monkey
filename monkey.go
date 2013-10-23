package gomonkey

// CGO_LDFLAGS="-L /usr/local/Cellar/spidermonkey/1.8.5/lib/ -lmozjs185.1.0" go install github.com/realint/gomonkey

/*
#include "js/jsapi.h"

extern JSClass global_class;
extern JSNative the_go_func_callback;
extern JSErrorReporter the_error_callback;
extern const char* eval_filename;

extern void _JS_SET_RVAL(JSContext *cx, jsval* vp, jsval v);
extern jsval JS_GET_ARGV(JSContext *cx, jsval* vp, int n);

extern jsval GET_JS_NULL();
extern jsval GET_JS_VOID();
*/
import "C"
import "sync"
import "errors"
import "unsafe"
import "reflect"
import "github.com/realint/monkey/goid"

type ErrorReporter func(report *ErrorReport)

type ErrorReportFlags uint

const (
	JSREPORT_WARNING   = ErrorReportFlags(C.JSREPORT_WARNING)
	JSREPORT_EXCEPTION = ErrorReportFlags(C.JSREPORT_EXCEPTION)
	JSREPORT_STRICT    = ErrorReportFlags(C.JSREPORT_STRICT)
)

type ErrorReport struct {
	Message    string
	FileName   string
	LineBuf    string
	LineNum    int
	ErrorNum   int
	TokenIndex int
	Flags      ErrorReportFlags
}

//export call_error_func
func call_error_func(r unsafe.Pointer, message *C.char, report *C.JSErrorReport) {
	if (*Runtime)(r).errorReporter != nil {
		(*Runtime)(r).errorReporter(&ErrorReport{
			Message:    C.GoString(message),
			FileName:   C.GoString(report.filename),
			LineNum:    int(report.lineno),
			ErrorNum:   int(report.errorNumber),
			LineBuf:    C.GoString(report.linebuf),
			TokenIndex: int(uintptr(unsafe.Pointer(report.tokenptr)) - uintptr(unsafe.Pointer(report.linebuf))),
		})
	}
}

type JsFunc func(argv []Value) (Value, bool)

//export call_go_func
func call_go_func(r unsafe.Pointer, name *C.char, argc C.uintN, vp *C.jsval) C.JSBool {
	var runtime = (*Runtime)(r)

	var argv = make([]Value, int(argc))

	for i := 0; i < len(argv); i++ {
		argv[i] = Value{runtime, C.JS_GET_ARGV(runtime.cx, vp, C.int(i))}
	}

	var result, ok = runtime.callbacks[C.GoString(name)](argv)

	if ok {
		C._JS_SET_RVAL(runtime.cx, vp, result.val)
		return C.JS_TRUE
	}

	return C.JS_FALSE
}

// JavaScript Runtime
type Runtime struct {
	rt            *C.JSRuntime
	cx            *C.JSContext
	global        *C.JSObject
	callbacks     map[string]JsFunc
	errorReporter ErrorReporter
	lockBy        int32
	lockLevel     int
	mutex         sync.Mutex
}

func printCall(argv []Value, newline bool) bool {
	for i := 0; i < len(argv); i++ {
		print(argv[i].ToString())
		if i < len(argv)-1 {
			print(", ")
		}
	}
	if newline {
		println()
	}
	return true
}

func NewRuntime() (*Runtime, error) {
	r := new(Runtime)
	r.callbacks = make(map[string]JsFunc)

	r.rt = C.JS_NewRuntime(8 * 1024 * 1024)
	if r.rt == nil {
		return nil, errors.New("Could't create JSRuntime")
	}

	r.cx = C.JS_NewContext(r.rt, 8192)
	if r.cx == nil {
		return nil, errors.New("Could't create JSContext")
	}

	C.JS_SetOptions(r.cx, C.JSOPTION_VAROBJFIX|C.JSOPTION_JIT|C.JSOPTION_METHODJIT)
	C.JS_SetVersion(r.cx, C.JSVERSION_LATEST)
	C.JS_SetErrorReporter(r.cx, C.the_error_callback)

	r.global = C.JS_NewCompartmentAndGlobalObject(r.cx, &C.global_class, nil)

	if C.JS_InitStandardClasses(r.cx, r.global) != C.JS_TRUE {
		return nil, errors.New("Could't init global class")
	}

	C.JS_SetRuntimePrivate(r.rt, unsafe.Pointer(r))

	r.DefineFunction("print", func(argv []Value) (Value, bool) {
		return r.Null(), printCall(argv, false)
	})

	r.DefineFunction("println", func(argv []Value) (Value, bool) {
		return r.Null(), printCall(argv, true)
	})

	return r, nil
}

func (r *Runtime) Dispose() {
	C.JS_DestroyContext(r.cx)
	C.JS_DestroyRuntime(r.rt)
}

func (r *Runtime) lock() {
	id := goid.Get()
	if r.lockBy != id {
		r.mutex.Lock()
		r.lockBy = id
	} else {
		r.lockLevel += 1
	}
}

func (r *Runtime) unlock() {
	r.lockLevel -= 1
	if r.lockLevel < 0 {
		r.lockLevel = 0
		r.lockBy = -1
		r.mutex.Unlock()
	}
}

func (r *Runtime) SetErrorReporter(reporter ErrorReporter) {
	r.errorReporter = reporter
}

func (r *Runtime) Eval(script string) (Value, error) {
	r.lock()
	defer r.unlock()

	cscript := C.CString(script)
	defer C.free(unsafe.Pointer(cscript))

	var rval C.jsval
	if C.JS_EvaluateScript(r.cx, r.global, cscript, C.uintN(len(script)), C.eval_filename, 0, &rval) == C.JS_TRUE {
		return Value{r, rval}, nil
	}

	return r.Null(), errors.New("Could't evaluate script")
}

func (r *Runtime) Compile(script, filename string, lineno int) (*Script, error) {
	r.lock()
	defer r.unlock()

	cscript := C.CString(script)
	defer C.free(unsafe.Pointer(cscript))

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	var scriptObj = C.JS_CompileScript(r.cx, r.global, cscript, C.size_t(len(script)), cfilename, C.uintN(lineno))

	if scriptObj != nil {
		return &Script{r, scriptObj}, nil
	}

	return nil, errors.New("Could't compile script")
}

func (r *Runtime) DefineFunction(name string, callback JsFunc) error {
	r.lock()
	defer r.unlock()

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if C.JS_DefineFunction(r.cx, r.global, cname, C.the_go_func_callback, 0, 0) == nil {
		return errors.New("Could't define function")
	}

	r.callbacks[name] = callback

	return nil
}

func (r *Runtime) Int(v int32) Value {
	return Value{r, C.INT_TO_JSVAL(C.int32(v))}
}

func (r *Runtime) Null() Value {
	return Value{r, C.GET_JS_NULL()}
}

func (r *Runtime) Void() Value {
	return Value{r, C.GET_JS_VOID()}
}

func (r *Runtime) Boolean(v bool) Value {
	if v {
		return Value{r, C.JS_TRUE}
	}
	return Value{r, C.JS_FALSE}
}

func (r *Runtime) String(v string) Value {
	cv := C.CString(v)
	defer C.free(unsafe.Pointer(cv))
	return Value{r, C.STRING_TO_JSVAL(C.JS_NewStringCopyN(r.cx, cv, C.size_t(len(v))))}
}

func (r *Runtime) NewArray() Object {
	return Object{r, C.JS_NewArrayObject(r.cx, 0, nil)}
}

func (r *Runtime) NewObject() Object {
	return Object{r, C.JS_NewObject(r.cx, nil, nil, nil)}
}

// Compiled Script
type Script struct {
	runtime   *Runtime
	scriptObj *C.JSObject
}

func (s *Script) Execute() (*Value, error) {
	s.runtime.lock()
	defer s.runtime.unlock()

	var rval C.jsval
	if C.JS_ExecuteScript(s.runtime.cx, s.runtime.global, s.scriptObj, &rval) == C.JS_TRUE {
		return &Value{s.runtime, rval}, nil
	}

	return nil, errors.New("Could't execute script")
}

// JavaScript Value
type Value struct {
	runtime *Runtime
	val     C.jsval
}

func (v Value) IsNull() bool {
	if C.JSVAL_IS_NULL(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v Value) IsVoid() bool {
	if C.JSVAL_IS_VOID(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v Value) IsInt() bool {
	if C.JSVAL_IS_INT(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v Value) IsString() bool {
	if C.JSVAL_IS_STRING(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v Value) IsNumber() bool {
	if C.JSVAL_IS_NUMBER(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v Value) IsBoolean() bool {
	if C.JSVAL_IS_BOOLEAN(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v Value) IsObject() bool {
	if C.JSVAL_IS_OBJECT(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v Value) IsFunction() bool {
	if !v.IsObject() {
		return false
	}

	if C.JS_ObjectIsFunction(v.runtime.cx, C.JSVAL_TO_OBJECT(v.val)) == C.JS_TRUE {
		return true
	}

	return false
}

func (v Value) ToString() string {
	cstring := C.JS_EncodeString(v.runtime.cx, C.JS_ValueToString(v.runtime.cx, v.val))
	gostring := C.GoString(cstring)
	C.JS_free(v.runtime.cx, unsafe.Pointer(cstring))
	return gostring
}

func (v Value) ToInt() (int32, bool) {
	var r C.int32
	if C.JS_ValueToInt32(v.runtime.cx, v.val, &r) == C.JS_TRUE {
		return int32(r), true
	}
	return 0, false
}

func (v Value) ToNumber() (float64, bool) {
	var r C.jsdouble
	if C.JS_ValueToNumber(v.runtime.cx, v.val, &r) == C.JS_TRUE {
		return float64(r), true
	}
	return 0, false
}

func (v Value) ToBoolean() (bool, bool) {
	var r C.JSBool
	if C.JS_ValueToBoolean(v.runtime.cx, v.val, &r) == C.JS_TRUE {
		if r == C.JS_TRUE {
			return true, true
		}
		return false, true
	}
	return false, false
}

func (v Value) ToObject() (Object, bool) {
	var obj *C.JSObject

	if C.JS_ValueToObject(v.runtime.cx, v.val, &obj) == C.JS_TRUE {
		return Object{v.runtime, obj}, true
	}

	return Object{}, false
}

func (v Value) String() string {
	cstring := C.JS_EncodeString(v.runtime.cx, C.JSVAL_TO_STRING(v.val))
	gostring := C.GoString(cstring)
	C.JS_free(v.runtime.cx, unsafe.Pointer(cstring))

	return gostring
}

func (v Value) Int() int32 {
	return int32(C.JSVAL_TO_INT(v.val))
}

func (v Value) Number() float64 {
	return float64(C.JSVAL_TO_DOUBLE(v.val))
}

func (v Value) Boolean() bool {
	if C.JSVAL_TO_BOOLEAN(v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (v Value) Object() Object {
	return Object{v.runtime, C.JSVAL_TO_OBJECT(v.val)}
}

func (v Value) Call(argv []Value) (Value, bool) {
	argv2 := make([]C.jsval, len(argv))

	for i := 0; i < len(argv); i++ {
		argv2[i] = argv[i].val
	}

	argv3 := unsafe.Pointer(&argv2)
	argv4 := (*reflect.SliceHeader)(argv3).Data
	argv5 := (*C.jsval)(unsafe.Pointer(argv4))

	r := v.runtime.Void()
	if C.JS_CallFunctionValue(v.runtime.cx, nil, v.val,
		C.uintN(len(argv)), argv5, &r.val,
	) == C.JS_TRUE {
		return r, true
	}

	return r, false
}

func (v Value) TypeName() string {
	jstype := C.JS_TypeOfValue(v.runtime.cx, v.val)
	return C.GoString(C.JS_GetTypeName(v.runtime.cx, jstype))
}

// JavaScript Object
type Object struct {
	runtime *Runtime
	obj     *C.JSObject
}

func (o Object) IsArray() bool {
	if C.JS_IsArrayObject(o.runtime.cx, o.obj) == C.JS_TRUE {
		return true
	}
	return false
}

func (o Object) GetArrayLength() int {
	var l C.jsuint
	C.JS_GetArrayLength(o.runtime.cx, o.obj, &l)
	return int(l)
}

func (o Object) SetArrayLength(length int) bool {
	if C.JS_SetArrayLength(o.runtime.cx, o.obj, C.jsuint(length)) == C.JS_TRUE {
		return true
	}
	return false
}

func (o Object) GetElement(index int) (Value, bool) {
	r := o.runtime.Void()
	if C.JS_GetElement(o.runtime.cx, o.obj, C.jsint(index), &r.val) == C.JS_TRUE {
		return r, true
	}
	return r, false
}

func (o Object) SetElement(index int, v Value) bool {
	if C.JS_SetElement(o.runtime.cx, o.obj, C.jsint(index), &v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (o Object) GetProperty(name string) (Value, bool) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	r := o.runtime.Void()
	if C.JS_GetProperty(o.runtime.cx, o.obj, cname, &r.val) == C.JS_TRUE {
		return r, true
	}
	return r, false
}

func (o Object) SetProperty(name string, v Value) bool {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if C.JS_SetProperty(o.runtime.cx, o.obj, cname, &v.val) == C.JS_TRUE {
		return true
	}
	return false
}

func (o Object) ToValue() Value {
	return Value{o.runtime, C.OBJECT_TO_JSVAL(o.obj)}
}
