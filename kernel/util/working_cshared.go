// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

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

// BootDesktopLib 初始化桌面端进程内内核（c-shared N-API），返回退出码而不调用 os.Exit。
func BootDesktopLib(workspace, port, lang, wd string) int {
	initEnvVars()
	IncBootProgress(3, "Booting kernel...")
	rand.Seed(time.Now().UTC().UnixNano())
	initMime()
	initHttpClient()

	if wd != "" {
		absPath, err := filepath.Abs(wd)
		if err != nil {
			logging.LogErrorf("resolve wd [%s] failed: %s", wd, err)
			return logging.ExitCodeInitWorkspaceErr
		}
		WorkingDir = absPath
	}
	if lang != "" {
		Lang = LangToBCP47(lang)
	}
	ServerPort = port
	Mode = "prod"
	InProcess = true
	AttachUI = false
	Container = ContainerStd

	UserAgent = UserAgent + " " + Container + "/" + runtime.GOOS
	httpclient.SetUserAgent(UserAgent)

	if code := initWorkspaceDirLib(workspace); code != 0 {
		return code
	}
	if code := initPathDirLib(); code != 0 {
		return code
	}

	AppearancePath = filepath.Join(ConfDir, "appearance")
	ThemesPath = filepath.Join(DataDir, "themes")
	IconsPath = filepath.Join(DataDir, "icons")
	LogPath = filepath.Join(TempDir, "siyuan.log")

	msStoreFilePath := filepath.Join(WorkingDir, "ms-store")
	ISMicrosoftStore = gulu.File.IsExist(msStoreFilePath)

	logging.SetLogPath(LogPath)

	if code := tryLockWorkspaceLib(); code != 0 {
		return code
	}
	if code := clearWorkspaceStartupTempLib(); code != 0 {
		return code
	}

	bootBanner := figure.NewColorFigure("SiYuan", "isometric3", "green", true)
	logging.LogInfo("\n" + bootBanner.String())
	logBootInfo()

	return 0
}

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

	defaultWorkspaceDir := defaultWorkspacePath(HomeDir, runtime.GOOS, os.Getenv("USERPROFILE"), homeDirOverridden)

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
	WorkspaceDir = filepath.Clean(WorkspaceDir)

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
	QueueDir = filepath.Join(TempDir, "queue")
	osTmpDir := filepath.Join(TempDir, "os")
	if err := os.MkdirAll(osTmpDir, 0755); err != nil {
		logging.LogErrorf("create os tmp dir [%s] failed: %s", osTmpDir, err)
		return logging.ExitCodeInitWorkspaceErr
	}
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

func clearWorkspaceStartupTempLib() int {
	for _, dir := range []string{"os", "repo", "export", "clipboard"} {
		os.RemoveAll(filepath.Join(TempDir, dir))
	}
	osTmpDir := filepath.Join(TempDir, "os")
	if err := os.MkdirAll(osTmpDir, 0755); err != nil {
		logging.LogErrorf("create os tmp dir [%s] failed: %s", osTmpDir, err)
		return logging.ExitCodeInitWorkspaceErr
	}
	return 0
}

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
		{filepath.Join(DataDir, "themes"), "data/themes"},
		{filepath.Join(DataDir, "icons"), "data/icons"},
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d.path, 0755); err != nil && !os.IsExist(err) {
			logging.LogErrorf("create %s folder [%s] failed: %s", d.name, d.path, err)
			return logging.ExitCodeInitWorkspaceErr
		}
	}
	return 0
}

// CheckPortAvailable 在启动 HTTP 服务前探测端口是否可用。
func CheckPortAvailable(port string) bool {
	ln, err := net.Listen("tcp", LocalHost+":"+port)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
