// N-API wrapper for the SiYuan kernel c-shared library.
// Exposes Go kernel functions to Node.js/Electron.
//
// Cross-compiled per platform alongside libkernel.so/dylib/dll.
// Each platform ships: kernel.node + libkernel.* (matched pair)

#include <node_api.h>
#include <string.h>
#include <stdlib.h>
#include "libkernel.h"

// Helper: extract a string argument from N-API
static char* get_string_arg(napi_env env, napi_value val) {
    size_t len;
    napi_get_value_string_utf8(env, val, NULL, 0, &len);
    char* buf = (char*)malloc(len + 1);
    napi_get_value_string_utf8(env, val, buf, len + 1, &len);
    return buf;
}

// startKernel(workspace: string, port: string, lang: string, wd: string) → number
static napi_value napi_start_kernel(napi_env env, napi_callback_info info) {
    size_t argc = 4;
    napi_value argv[4];
    napi_get_cb_info(env, info, &argc, argv, NULL, NULL);

    char* workspace = argc > 0 ? get_string_arg(env, argv[0]) : strdup("");
    char* port      = argc > 1 ? get_string_arg(env, argv[1]) : strdup("0");
    char* lang      = argc > 2 ? get_string_arg(env, argv[2]) : strdup("");
    char* wd        = argc > 3 ? get_string_arg(env, argv[3]) : strdup("");

    int rc = StartKernel(workspace, port, lang, wd);

    free(workspace);
    free(port);
    free(lang);
    free(wd);

    napi_value result;
    napi_create_int32(env, rc, &result);
    return result;
}

// isHttpServing() → boolean
static napi_value napi_is_http_serving(napi_env env, napi_callback_info info) {
    int serving = IsHttpServing();
    napi_value result;
    napi_get_boolean(env, serving != 0, &result);
    return result;
}

// stopKernel() → void
static napi_value napi_stop_kernel(napi_env env, napi_callback_info info) {
    StopKernel();
    napi_value undefined;
    napi_get_undefined(env, &undefined);
    return undefined;
}

// runCLI(args: string[], appDir: string) → number
// Runs a SiYuan CLI command in-process. Intended for a dedicated launcher process
// (electron-as-node), not the GUI. Returns the command's exit code.
static napi_value napi_run_cli(napi_env env, napi_callback_info info) {
    size_t argc = 2;
    napi_value argv[2];
    napi_get_cb_info(env, info, &argc, argv, NULL, NULL);

    uint32_t n = 0;
    if (argc > 0) {
        napi_get_array_length(env, argv[0], &n);
    }
    char** cargs = (char**)malloc(sizeof(char*) * (n ? n : 1));
    for (uint32_t i = 0; i < n; i++) {
        napi_value el;
        napi_get_element(env, argv[0], i, &el);
        cargs[i] = get_string_arg(env, el);
    }
    char* appDir = argc > 1 ? get_string_arg(env, argv[1]) : strdup("");

    int rc = RunCLI((int)n, cargs, appDir);

    for (uint32_t i = 0; i < n; i++) {
        free(cargs[i]);
    }
    free(cargs);
    free(appDir);

    napi_value result;
    napi_create_int32(env, rc, &result);
    return result;
}

// Module initialization
static napi_value init(napi_env env, napi_value exports) {
    napi_property_descriptor props[] = {
        {"startKernel",    NULL, napi_start_kernel,     NULL, NULL, NULL, napi_default, NULL},
        {"isHttpServing",  NULL, napi_is_http_serving,  NULL, NULL, NULL, napi_default, NULL},
        {"stopKernel",     NULL, napi_stop_kernel,      NULL, NULL, NULL, napi_default, NULL},
        {"runCLI",         NULL, napi_run_cli,          NULL, NULL, NULL, napi_default, NULL},
    };
    napi_define_properties(env, exports, sizeof(props) / sizeof(props[0]), props);
    return exports;
}

NAPI_MODULE(NODE_GYP_MODULE_NAME, init)
