// SiYuan - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Build as c-shared library for Electron N-API integration.
// This provides the same kernel functionality as kernel/main.go but
// callable via C function exports instead of running as a separate process.
//
// Build: go build -buildmode=c-shared -tags fts5 -o libkernel.so ./kernel/cshared/

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
	"github.com/siyuan-note/siyuan/kernel/server"
	"github.com/siyuan-note/siyuan/kernel/server/tunnel"
	"github.com/siyuan-note/siyuan/kernel/sql"
	"github.com/siyuan-note/siyuan/kernel/util"
)

var kernelOnce sync.Once

// StartKernel starts the SiYuan kernel with the given parameters.
// Uses BootDesktopLib() which returns error codes instead of calling os.Exit().
//
// Returns 0 on success, or a logging.ExitCode* value on failure:
//
//	20 = database unavailable
//	21 = port bind failed
//	24 = workspace locked
//	25 = workspace init failed
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

	kernelOnce.Do(func() {
		goWorkspace := C.GoString(workspace)
		goPort := C.GoString(port)
		goLang := C.GoString(lang)
		goWd := C.GoString(wd)

		// Pre-check port availability
		if goPort != "" && goPort != "0" {
			if !util.CheckPortAvailable(goPort) {
				retCode = C.int(logging.ExitCodeUnavailablePort)
				return
			}
		}

		// Boot kernel (returns error code instead of os.Exit)
		if code := util.BootDesktopLib(goWorkspace, goPort, goLang, goWd); code != 0 {
			retCode = C.int(code)
			return
		}

		// Same init sequence as kernel/main.go
		model.InitConf()
		go server.Serve(false, model.Conf.CookieKey)

		// Run the rest async (like mobile) so we don't block the caller
		go func() {
			model.InitAppearance()
			sql.InitDatabase(false)
			sql.InitHistoryDatabase(false)
			sql.InitAssetContentDatabase(false)
			sql.SetCaseSensitive(model.Conf.Search.CaseSensitive)
			sql.SetIndexAssetPath(model.Conf.Search.IndexAssetPath)

			model.BootSyncData()
			model.InitBoxes()
			model.LoadFlashcards()
			util.LoadAssetsTexts()

			util.SetBooted()
			util.PushClearAllMsg()

			job.StartCron()
			go model.AutoGenerateFileHistory()
			go cache.LoadAssets()
			go util.CheckFileSysStatus()

			model.WatchAssets()
			model.WatchEmojis()
			model.WatchThemes()
		}()
	})

	return
}

// IsHttpServing returns 1 if the HTTP server is ready, 0 otherwise.
//
//export IsHttpServing
func IsHttpServing() C.int {
	if util.HttpServing {
		return C.int(1)
	}
	return C.int(0)
}

// StopKernel gracefully shuts down the kernel.
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
}

// RunCLI runs a SiYuan CLI command (the cobra command tree from kernel/cli/cmd)
// in-process. It is invoked from a dedicated launcher process (electron-as-node),
// never from the GUI process, so stdout/exit-code semantics behave normally.
//
// argv holds the CLI arguments (without argv[0]); appDir is the directory that
// contains the bundled "appearance/langs" (the GUI's appDir = resources/). The
// "serve" subcommand is rejected because it blocks on HandleSignal — the GUI uses
// StartKernel instead.
//
// Returns 0 on success, 1 on command error/panic, 2 if "serve" was requested.
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
	args = append(args, "siyuan-cli") // argv[0]; root.go init() derives Use from base(os.Args[0])
	cargs := unsafe.Slice(argv, n)
	for i := 0; i < n; i++ {
		args = append(args, C.GoString(cargs[i]))
	}

	// "serve" blocks on HandleSignal — not valid for a one-shot launcher; the GUI uses StartKernel.
	if n > 0 && C.GoString(cargs[0]) == "serve" {
		logging.LogErrorf("serve is not supported via RunCLI; the GUI uses StartKernel")
		return C.int(2)
	}

	if ad := C.GoString(appDir); "" != ad {
		os.Setenv("SIYUAN_WORKING_DIR", ad) // consumed by kernel/cli/cmd/root.go
	}
	os.Args = args

	if err := cmd.Execute(); err != nil {
		return C.int(1)
	}
	return C.int(0)
}

// required for c-shared build mode
func main() {}
