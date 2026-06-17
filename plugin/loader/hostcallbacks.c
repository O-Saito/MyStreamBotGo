#include "_cgo_export.h"
#include "hostapi.h"
#include <stdlib.h>
#include <stdint.h>

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

/* ── Plugin function calling wrappers (ull = unsigned long long, avoids unsafe.Pointer in Go) ── */

const char* call_plugin_name_ull(unsigned long long fn) {
	typedef const char* (*name_fn_t)(void);
	return ((name_fn_t)(void*)(uintptr_t)fn)();
}

uint32_t call_plugin_api_ver_ull(unsigned long long fn) {
	typedef uint32_t (*ver_fn_t)(void);
	return ((ver_fn_t)(void*)(uintptr_t)fn)();
}

void call_plugin_set_host_ull(unsigned long long fn, const bot_host_api_t* api) {
	typedef void (*set_host_fn_t)(const bot_host_api_t*);
	((set_host_fn_t)(void*)(uintptr_t)fn)(api);
}

void call_plugin_free_ull(unsigned long long fn, void* ptr) {
	typedef void (*free_fn_t)(void*);
	((free_fn_t)(void*)(uintptr_t)fn)(ptr);
}

const char* call_plugin_init_ull(unsigned long long fn, const char* json) {
	typedef const char* (*init_fn_t)(const char*);
	return ((init_fn_t)(void*)(uintptr_t)fn)(json);
}

void call_plugin_start_ull(unsigned long long fn) {
	typedef void (*start_fn_t)(void);
	((start_fn_t)(void*)(uintptr_t)fn)();
}

void call_plugin_stop_ull(unsigned long long fn) {
	typedef void (*stop_fn_t)(void);
	((stop_fn_t)(void*)(uintptr_t)fn)();
}

void call_plugin_on_chat_ull(unsigned long long fn, const char* json) {
	typedef void (*chat_fn_t)(const char*);
	((chat_fn_t)(void*)(uintptr_t)fn)(json);
}

void call_plugin_on_event_ull(unsigned long long fn, const char* json) {
	typedef void (*evt_fn_t)(const char*);
	((evt_fn_t)(void*)(uintptr_t)fn)(json);
}

const char* call_plugin_get_actions_ull(unsigned long long fn) {
	typedef const char* (*get_fn_t)(void);
	return ((get_fn_t)(void*)(uintptr_t)fn)();
}

const char* call_plugin_call_action_ull(unsigned long long fn, const char* name, const char* json) {
	typedef const char* (*action_fn_t)(const char*, const char*);
	return ((action_fn_t)(void*)(uintptr_t)fn)(name, json);
}
