#include <stdio.h>
#include "js/jsapi.h"
#include "_cgo_export.h"

JSClass global_class = {
    "global", JSCLASS_GLOBAL_FLAGS,
    JS_PropertyStub, JS_PropertyStub, JS_PropertyStub, JS_StrictPropertyStub,
    JS_EnumerateStub, JS_ResolveStub, JS_ConvertStub, JS_FinalizeStub,
    JSCLASS_NO_OPTIONAL_MEMBERS
};

/* The go function callback. */
JSBool go_func_callback(JSContext *cx, uintN argc, jsval *vp) {
	JSObject *callee = JSVAL_TO_OBJECT(JS_CALLEE(cx, vp));

	jsval name;
	JS_GetProperty(cx, callee, "name", &name);

	return call_go_func(
		JS_GetRuntimePrivate(JS_GetRuntime(cx)), 
		JS_EncodeString(cx, JS_ValueToString(cx, name)), 
		argc,
		vp
	);
}

/* The error reporter callback. */
void error_callback(JSContext *cx, const char *message, JSErrorReport *report) {
	call_error_func(JS_GetRuntimePrivate(JS_GetRuntime(cx)), (char*)message, report);
}

/* Function pointers to avoid CGO warnning. */
JSNative the_go_func_callback = &go_func_callback;
JSErrorReporter the_error_callback = &error_callback;

/* File name for evaluate script. */
const char* eval_filename = "Eval()";

void _JS_SET_RVAL(JSContext *cx, jsval* vp, jsval v) {
	JS_SET_RVAL(cx, vp, v);
}

jsval JS_GET_ARGV(JSContext *cx, jsval* vp, int n) {
	return JS_ARGV(cx, vp)[n];
}

jsval GET_JS_NULL() {
	return JSVAL_NULL;
}

jsval GET_JS_VOID() {
	return JSVAL_VOID;
}