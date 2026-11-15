// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !mobile

package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"os"
	"sync"
	"unsafe"

	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/cache"
	"github.com/siyuan-note/siyuan/kernel/cli/cmd"
	"github.com/siyuan-note/siyuan/kernel/job"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/plugin"
	"github.com/siyuan-note/siyuan/kernel/server"
	"github.com/siyuan-note/siyuan/kernel/server/tunnel"
	"github.com/siyuan-note/siyuan/kernel/sql"
	"github.com/siyuan-note/siyuan/kernel/util"
)

var (
	kernelMu      sync.Mutex
	kernelRunning bool
)

// StartKernel 启动进程内内核，失败时返回退出码而不终止宿主进程，以便锁冲突时可继续重试。
//
//export StartKernel
func StartKernel(workspace, port, lang, wd *C.char) (retCode C.int) {
	defer func() {
		if r := recover(); r != nil {
			logging.LogErrorf("kernel startup panic: %v", r)
			if retCode == 0 {
				retCode = C.int(1)
			}
		}
	}()

	kernelMu.Lock()
	defer kernelMu.Unlock()
	if kernelRunning {
		return C.int(0)
	}

	goWorkspace := C.GoString(workspace)
	goPort := C.GoString(port)
	goLang := C.GoString(lang)
	goWd := C.GoString(wd)

	if goPort != "" && goPort != "0" {
		if !util.CheckPortAvailable(goPort) {
			return C.int(logging.ExitCodeUnavailablePort)
		}
	}

	if code := util.BootDesktopLib(goWorkspace, goPort, goLang, goWd); code != 0 {
		return C.int(code)
	}
	kernelRunning = true

	model.InitJwtKey()
	model.InitConf()
	model.InitAppearance()
	go server.Serve(false, model.Conf.CookieKey)

	go func() {
		sql.InitDatabase(false)
		sql.InitHistoryDatabase(false)
		sql.InitAssetContentDatabase(false)
		model.ApplySearchRuntime()

		model.BootSyncData()
		model.CleanupEmptyPluginStorageDirs()
		model.InitBoxes()
		model.LoadFlashcards()
		util.LoadAssetsTexts()

		util.SetBooted()
		util.PushClearAllMsg()

		job.StartCron()
		go model.AutoGenerateFileHistory()
		go cache.LoadAssets()
		go util.CheckFileSysStatus()
		go plugin.InitManager()
		go model.StartEmbeddingIndexer()

		model.WatchAssets()
		model.WatchEmojis()
		model.WatchThemes()
	}()

	return C.int(0)
}

// IsHttpServing 返回 HTTP 服务是否已就绪。
//
//export IsHttpServing
func IsHttpServing() C.int {
	if util.HttpServing {
		return C.int(1)
	}
	return C.int(0)
}

// StopKernel 优雅关闭进程内内核与隧道。
//
//export StopKernel
func StopKernel() {
	defer func() {
		if r := recover(); r != nil {
			logging.LogErrorf("kernel stop panic: %v", r)
		}
	}()

	tunnel.StopTailscale()
	tunnel.StopCloudflared()
	model.Close(false, true, 0)
	model.ExitKernel()

	kernelMu.Lock()
	kernelRunning = false
	kernelMu.Unlock()
}

// RunCLI 在独立的命令行模式进程（ELECTRON_RUN_AS_NODE=1）中运行 Cobra CLI。
//
//export RunCLI
func RunCLI(argc C.int, argv **C.char, appDir *C.char) (retCode C.int) {
	defer func() {
		if r := recover(); r != nil {
			logging.LogErrorf("cli panic: %v", r)
			if retCode == 0 {
				retCode = C.int(1)
			}
		}
	}()

	n := int(argc)
	args := make([]string, 0, n+1)
	args = append(args, "siyuan-cli")
	cargs := unsafe.Slice(argv, n)
	for i := 0; i < n; i++ {
		args = append(args, C.GoString(cargs[i]))
	}

	if n > 0 && C.GoString(cargs[0]) == "serve" {
		logging.LogErrorf("serve is not supported via RunCLI; the GUI uses StartKernel")
		return C.int(2)
	}

	if ad := C.GoString(appDir); "" != ad {
		os.Setenv("SIYUAN_WORKING_DIR", ad)
	}
	os.Args = args

	if err := cmd.Execute(); err != nil {
		return C.int(1)
	}
	return C.int(0)
}

func main() {}
