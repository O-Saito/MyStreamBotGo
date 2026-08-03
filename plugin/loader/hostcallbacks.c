#include "_cgo_export.h"
#include "hostapi.h"
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

/* ── Platform-specific crash protection headers ── */

#if defined(_WIN32) || defined(_WIN64)
#include <windows.h>
#include <setjmp.h>
#elif defined(__linux__) || defined(__APPLE__)
#include <signal.h>
#include <setjmp.h>
#endif

/* ── Crash guard helpers ──
 *
 *  Each wrapper installs platform-specific crash protection before calling
 *  the plugin function pointer and restores the old handlers afterward.  If
 *  the plugin faults (access violation, null deref, SEGV, etc.), the handler
 *  longjmps back to the call site, logs via hostLog, and returns a safe zero
 *  value.  The Go-side defer/recover in registry.go marks the plugin
 *  stopped=true.
 */

#if defined(_WIN32) || defined(_WIN64)

/* Windows: Vectored Exception Handling + setjmp/longjmp */
static __thread jmp_buf plugin_jmp_buf;
static __thread int plugin_in_call = 0;

static LONG CALLBACK plugin_veh(PEXCEPTION_POINTERS ExceptionInfo) {
    if (plugin_in_call) {
        plugin_in_call = 0;
        longjmp(plugin_jmp_buf, 1);
    }
    return EXCEPTION_CONTINUE_SEARCH;
}

static void* veh_handle = NULL;
static int veh_installed = 0;

static void install_crash_guards(void) {
    plugin_in_call = 0;
    if (!veh_installed) {
        veh_handle = AddVectoredExceptionHandler(1, plugin_veh);
        veh_installed = 1;
    }
    plugin_in_call = 1;
}

static void restore_crash_guards(void) {
    plugin_in_call = 0;
}

#elif defined(__linux__) || defined(__APPLE__)

/* POSIX: sigaction + sigsetjmp/siglongjmp (saves signal mask) */
static __thread sigjmp_buf plugin_jmp_buf;
static __thread int plugin_in_call = 0;

static void plugin_fault_handler(int sig) {
    if (plugin_in_call) siglongjmp(plugin_jmp_buf, 1);
}

static void install_crash_guards(struct sigaction *old_segv,
                                 struct sigaction *old_bus) {
    struct sigaction sa;
    memset(&sa, 0, sizeof(sa));
    sa.sa_handler = plugin_fault_handler;
    sa.sa_flags   = SA_RESETHAND | SA_NODEFER;
    sigaction(SIGSEGV, NULL, old_segv);
    sigaction(SIGSEGV, &sa,  NULL);
#if defined(SIGBUS)
    sigaction(SIGBUS, NULL, old_bus);
    sigaction(SIGBUS, &sa,  NULL);
#else
    (void)old_bus;
#endif
    plugin_in_call = 1;
}

static void restore_crash_guards(const struct sigaction *old_segv,
                                 const struct sigaction *old_bus) {
    plugin_in_call = 0;
    sigaction(SIGSEGV, old_segv, NULL);
#if defined(SIGBUS)
    sigaction(SIGBUS,  old_bus,  NULL);
#else
    (void)old_bus;
#endif
}

#endif

/* ── Macro to reduce boilerplate ── */

#if defined(_WIN32) || defined(_WIN64)
#define CRASH_GUARD(fn_call)                                                  \
    do {                                                                      \
        install_crash_guards();                                                \
        if (setjmp(plugin_jmp_buf) == 0) {                                    \
            fn_call;                                                          \
        } else {                                                              \
            crashed = 1;                                                      \
        }                                                                     \
        restore_crash_guards();                                                \
    } while (0)
#elif defined(__linux__) || defined(__APPLE__)
#define CRASH_GUARD(fn_call)                                                  \
    do {                                                                      \
        struct sigaction old_segv, old_bus;                                    \
        install_crash_guards(&old_segv, &old_bus);                            \
        if (sigsetjmp(plugin_jmp_buf, 1) == 0) {                              \
            fn_call;                                                          \
        } else {                                                              \
            crashed = 1;                                                      \
        }                                                                     \
        restore_crash_guards(&old_segv, &old_bus);                            \
    } while (0)
#endif

/* ── Host API struct construction ── */

bot_host_api_t* create_host_api(void) {
	bot_host_api_t* api = (bot_host_api_t*)malloc(sizeof(bot_host_api_t));
	if (!api) return NULL;
	api->api_version = MYSTREAM_BOT_API_VERSION;
	api->log = hostLog;
	api->emit_event = hostEmitEvent;
	api->free = free;
	return api;
}

void destroy_host_api(bot_host_api_t* api) {
	if (api) free(api);
}

/* ── Crash-protected plugin function calling wrappers ──
 *     (ull = unsigned long long, avoids unsafe.Pointer in Go)
 *
 *  Each wrapper protects the plugin function call via CRASH_GUARD.  On crash
 *  the wrapper logs via hostLog and returns a safe zero value so the host
 *  process survives.
 */

const char* call_plugin_name_ull(unsigned long long fn) {
	typedef const char* (*name_fn_t)(void);
	const char* result = NULL;
	int crashed = 0;
	CRASH_GUARD(result = ((name_fn_t)(void*)(uintptr_t)fn)());
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_name_ull");
	}
	return result;
}

uint32_t call_plugin_api_ver_ull(unsigned long long fn) {
	typedef uint32_t (*ver_fn_t)(void);
	uint32_t result = 0;
	int crashed = 0;
	CRASH_GUARD(result = ((ver_fn_t)(void*)(uintptr_t)fn)());
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_api_ver_ull");
	}
	return result;
}

void call_plugin_set_host_ull(unsigned long long fn, const bot_host_api_t* api) {
	typedef void (*set_host_fn_t)(const bot_host_api_t*);
	int crashed = 0;
	CRASH_GUARD(((set_host_fn_t)(void*)(uintptr_t)fn)(api));
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_set_host_ull");
	}
}

void call_plugin_free_ull(unsigned long long fn, void* ptr) {
	typedef void (*free_fn_t)(void*);
	int crashed = 0;
	CRASH_GUARD(((free_fn_t)(void*)(uintptr_t)fn)(ptr));
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_free_ull");
	}
}

const char* call_plugin_init_ull(unsigned long long fn, const char* json) {
	typedef const char* (*init_fn_t)(const char*);
	const char* result = NULL;
	int crashed = 0;
	CRASH_GUARD(result = ((init_fn_t)(void*)(uintptr_t)fn)(json));
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_init_ull");
	}
	return result;
}

void call_plugin_start_ull(unsigned long long fn) {
	typedef void (*start_fn_t)(void);
	int crashed = 0;
	CRASH_GUARD(((start_fn_t)(void*)(uintptr_t)fn)());
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_start_ull");
	}
}

void call_plugin_stop_ull(unsigned long long fn) {
	typedef void (*stop_fn_t)(void);
	int crashed = 0;
	CRASH_GUARD(((stop_fn_t)(void*)(uintptr_t)fn)());
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_stop_ull");
	}
}

void call_plugin_on_chat_ull(unsigned long long fn, const char* json) {
	typedef void (*chat_fn_t)(const char*);
	int crashed = 0;
	CRASH_GUARD(((chat_fn_t)(void*)(uintptr_t)fn)(json));
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_on_chat_ull");
	}
}

void call_plugin_on_event_ull(unsigned long long fn, const char* json) {
	typedef void (*evt_fn_t)(const char*);
	int crashed = 0;
	CRASH_GUARD(((evt_fn_t)(void*)(uintptr_t)fn)(json));
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_on_event_ull");
	}
}

const char* call_plugin_get_actions_ull(unsigned long long fn) {
	typedef const char* (*get_fn_t)(void);
	const char* result = NULL;
	int crashed = 0;
	CRASH_GUARD(result = ((get_fn_t)(void*)(uintptr_t)fn)());
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_get_actions_ull");
	}
	return result;
}

const char* call_plugin_call_action_ull(unsigned long long fn,
                                        const char* name,
                                        const char* json,
                                        const char* meta) {
	typedef const char* (*action_fn_t)(const char*, const char*, const char*);
	const char* result = NULL;
	int crashed = 0;
	CRASH_GUARD(result = ((action_fn_t)(void*)(uintptr_t)fn)(name, json, meta));
	if (crashed) {
		hostLog(3, "[PLUGIN_CRASH] crash caught in call_plugin_call_action_ull");
	}
	return result;
}
