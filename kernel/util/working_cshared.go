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

package util

import (
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/88250/gulu"
	figure "github.com/common-nighthawk/go-figure"
	"github.com/gofrs/flock"
	"github.com/siyuan-note/httpclient"
	"github.com/siyuan-note/logging"
)

// BootDesktopLib initializes the kernel for c-shared library mode.
// Unlike Boot(), it takes parameters directly (no CLI flag parsing) and
// returns an error code instead of calling os.Exit().
// Unlike BootMobile(), it supports desktop features like dynamic port
// and workspace locking.
//
// Returns 0 on success, or a logging.ExitCode* value on failure.
func BootDesktopLib(workspace, port, lang, wd string) int {
	initEnvVars()
	IncBootProgress(3, "Booting kernel...")
	rand.Seed(time.Now().UTC().UnixNano())
	initMime()
	initHttpClient()

	// Set parameters directly (like BootMobile)
	if wd != "" {
		WorkingDir = wd
	}
	if lang != "" {
		Lang = lang
	}
	ServerPort = port
	Mode = "prod"
	Container = ContainerCShared

	msStoreFilePath := filepath.Join(WorkingDir, "ms-store")
	ISMicrosoftStore = gulu.File.IsExist(msStoreFilePath)

	UserAgent = UserAgent + " " + Container + "/" + runtime.GOOS
	httpclient.SetUserAgent(UserAgent)

	// Initialize workspace (same as initWorkspaceDir but returns error)
	if code := initWorkspaceDirLib(workspace); code != 0 {
		return code
	}

	LogPath = filepath.Join(TempDir, "siyuan.log")
	logging.SetLogPath(LogPath)

	// Try to lock workspace (same as tryLockWorkspace but returns error)
	if code := tryLockWorkspaceLib(); code != 0 {
		return code
	}

	AppearancePath = filepath.Join(ConfDir, "appearance")
	ThemesPath = filepath.Join(AppearancePath, "themes")
	IconsPath = filepath.Join(AppearancePath, "icons")

	// Create standard directories (conf, data, temp, etc.)
	if code := initPathDirLib(); code != 0 {
		return code
	}

	bootBanner := figure.NewColorFigure("SiYuan", "isometric3", "green", true)
	logging.LogInfof("\n" + bootBanner.String())
	logBootInfo()

	return 0
}

// initWorkspaceDirLib is like initWorkspaceDir but returns error code
// instead of calling os.Exit.
func initWorkspaceDirLib(workspaceArg string) int {
	userHomeConfDir := filepath.Join(HomeDir, ".config", "siyuan")
	workspaceConf := filepath.Join(userHomeConfDir, "workspace.json")
	logging.SetLogPath(filepath.Join(userHomeConfDir, "kernel.log"))

	if !gulu.File.IsExist(workspaceConf) {
		if err := os.MkdirAll(userHomeConfDir, 0755); err != nil && !os.IsExist(err) {
			logging.LogErrorf("create user home conf folder [%s] failed: %s", userHomeConfDir, err)
			return logging.ExitCodeInitWorkspaceErr
		}
	}

	defaultWorkspaceDir := filepath.Join(HomeDir, "SiYuan")
	if gulu.OS.IsWindows() {
		if userProfile := os.Getenv("USERPROFILE"); "" != userProfile {
			defaultWorkspaceDir = filepath.Join(userProfile, "SiYuan")
		}
	}

	var workspacePaths []string
	if !gulu.File.IsExist(workspaceConf) {
		WorkspaceDir = defaultWorkspaceDir
	} else {
		workspacePaths, _ = ReadWorkspacePaths()
		if 0 < len(workspacePaths) {
			WorkspaceDir = workspacePaths[len(workspacePaths)-1]
		} else {
			WorkspaceDir = defaultWorkspaceDir
		}
	}

	if "" != workspaceArg {
		WorkspaceDir = workspaceArg
	}

	if !gulu.File.IsDir(WorkspaceDir) {
		logging.LogWarnf("use the default workspace [%s] since the specified workspace [%s] is not a dir", defaultWorkspaceDir, WorkspaceDir)
		if err := os.MkdirAll(defaultWorkspaceDir, 0755); err != nil && !os.IsExist(err) {
			logging.LogErrorf("create default workspace folder [%s] failed: %s", defaultWorkspaceDir, err)
			return logging.ExitCodeInitWorkspaceErr
		}
		WorkspaceDir = defaultWorkspaceDir
	}
	workspacePaths = append(workspacePaths, WorkspaceDir)

	if err := WriteWorkspacePaths(workspacePaths); err != nil {
		logging.LogErrorf("write workspace conf [%s] failed: %s", workspaceConf, err)
		return logging.ExitCodeInitWorkspaceErr
	}

	WorkspaceName = filepath.Base(WorkspaceDir)
	ConfDir = filepath.Join(WorkspaceDir, "conf")
	DataDir = filepath.Join(WorkspaceDir, "data")
	RepoDir = filepath.Join(WorkspaceDir, "repo")
	HistoryDir = filepath.Join(WorkspaceDir, "history")
	TempDir = filepath.Join(WorkspaceDir, "temp")
	osTmpDir := filepath.Join(TempDir, "os")
	os.RemoveAll(osTmpDir)
	if err := os.MkdirAll(osTmpDir, 0755); err != nil {
		logging.LogErrorf("create os tmp dir [%s] failed: %s", osTmpDir, err)
		return logging.ExitCodeInitWorkspaceErr
	}
	os.RemoveAll(filepath.Join(TempDir, "repo"))
	os.Setenv("TMPDIR", osTmpDir)
	os.Setenv("TEMP", osTmpDir)
	os.Setenv("TMP", osTmpDir)
	DBPath = filepath.Join(TempDir, DBName)
	HistoryDBPath = filepath.Join(TempDir, "history.db")
	AssetContentDBPath = filepath.Join(TempDir, "asset_content.db")
	BlockTreeDBPath = filepath.Join(TempDir, "blocktree.db")
	SnippetsPath = filepath.Join(DataDir, "snippets")
	ShortcutsPath = filepath.Join(userHomeConfDir, "shortcuts")

	return 0
}

// tryLockWorkspaceLib is like tryLockWorkspace but returns error code.
func tryLockWorkspaceLib() int {
	WorkspaceLock = flock.New(filepath.Join(WorkspaceDir, ".lock"))
	ok, err := WorkspaceLock.TryLock()
	if ok {
		return 0
	}
	if err != nil {
		logging.LogErrorf("lock workspace [%s] failed: %s", WorkspaceDir, err)
	} else {
		logging.LogErrorf("lock workspace [%s] failed", WorkspaceDir)
	}
	return logging.ExitCodeWorkspaceLocked
}

// initPathDirLib is like initPathDir but returns error code.
func initPathDirLib() int {
	dirs := []struct {
		path string
		name string
	}{
		{ConfDir, "conf"},
		{DataDir, "data"},
		{TempDir, "temp"},
		{filepath.Join(DataDir, "assets"), "data/assets"},
		{filepath.Join(DataDir, "templates"), "data/templates"},
		{filepath.Join(DataDir, "widgets"), "data/widgets"},
		{filepath.Join(DataDir, "plugins"), "data/plugins"},
		{filepath.Join(DataDir, "emojis"), "data/emojis"},
		{filepath.Join(DataDir, "public"), "data/public"},
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d.path, 0755); err != nil && !os.IsExist(err) {
			logging.LogErrorf("create %s folder [%s] failed: %s", d.name, d.path, err)
			return logging.ExitCodeInitWorkspaceErr
		}
	}
	return 0
}

// CheckPortAvailable tests if a port can be bound. Used by c-shared
// callers to pre-validate port before starting the HTTP server.
func CheckPortAvailable(port string) bool {
	ln, err := net.Listen("tcp", LocalHost+":"+port)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
