#include <pthread.h>
#include <math.h>

#define NUM_THREADS 10

/*
 * This define is for Windows only, it is a work-around for bug 661663.
 */
#ifdef _MSC_VER
# define XP_WIN
#endif

/* Include the JSAPI header file to get access to SpiderMonkey. */
#include "jsapi.h"

/* The class of the global object. */
static JSClass global_class = {
    "global", JSCLASS_GLOBAL_FLAGS,
    JS_PropertyStub, JS_PropertyStub, JS_PropertyStub, JS_StrictPropertyStub,
    JS_EnumerateStub, JS_ResolveStub, JS_ConvertStub, JS_FinalizeStub,
    JSCLASS_NO_OPTIONAL_MEMBERS
};

/* JSAPI variables. */
JSRuntime *rt;
JSContext *cx;
JSObject  *global;

/* The error reporter callback. */
void reportError(JSContext *cx, const char *message, JSErrorReport *report)
{
    fprintf(stderr, "%s:%u:%s\n",
            report->filename ? report->filename : "<no filename=\"filename\">",
            (unsigned int) report->lineno,
            message);
}

void *PrintHello(void *t)
{
    int i;

    for (i=0; i<10000; i++) {
        JS_Lock(rt);

        const char *script = "'Hello ' + 'Thread!'";
        jsval rval;

        JSBool ok = JS_EvaluateScript(cx, global, script, strlen(script), "noname", 0, &rval);
        if (rval != JSVAL_NULL && rval != JS_FALSE){
            JSString *str = JS_ValueToString(cx, rval);
            printf("%s\n", JS_EncodeString(cx, str));
        }

        JS_Unlock(rt);
    }

    pthread_exit((void*)t);
}

int main(int argc, const char *argv[])
{
    /* Create a JS runtime. You always need at least one runtime per process. */
    rt = JS_NewRuntime(8 * 1024 * 1024);
    if (rt == NULL)
        return 1;

    /* 
     * Create a context. You always need a context per thread.
     * Note that this program is not multi-threaded.
     */
    cx = JS_NewContext(rt, 8192);
    if (cx == NULL)
        return 1;
    JS_SetOptions(cx, JSOPTION_VAROBJFIX | JSOPTION_JIT | JSOPTION_METHODJIT);
    JS_SetVersion(cx, JSVERSION_LATEST);
    JS_SetErrorReporter(cx, reportError);

    /*
     * Create the global object in a new compartment.
     * You always need a global object per context.
     */
    global = JS_NewCompartmentAndGlobalObject(cx, &global_class, NULL);
    if (global == NULL)
        return 1;

    /*
     * Populate the global object with the standard JavaScript
     * function and object classes, such as Object, Array, Date.
     */
    if (!JS_InitStandardClasses(cx, global))
        return 1;

    pthread_t threads[NUM_THREADS];
    pthread_attr_t attr;
    void *status;

    /* Initialize and set thread detached attribute */
    pthread_attr_init(&attr);
    pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_JOINABLE);

    int rc;
    long t;
    for(t=0; t<NUM_THREADS; t++){
        printf("Creating thread %ld\n", t);
        rc = pthread_create(&threads[t], &attr, PrintHello, (void *)t);
        if (rc){
            printf("ERROR: return code from pthread_create() is %d\n", rc);
            exit(-1);
        }
    }

    pthread_attr_destroy(&attr);
    for(t=0; t<NUM_THREADS; t++) {
        rc = pthread_join(threads[t], &status);
        if (rc) {
            printf("ERROR: return code from pthread_join() is %d\n", rc);
            exit(-1);
        }
    }

    // pthread_exit(NULL);

    /* Clean things up and shut down SpiderMonkey. */
    JS_DestroyContext(cx);
    JS_DestroyRuntime(rt);
    JS_ShutDown();

    return 0;
}
