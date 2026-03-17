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

package tunnel

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cloudflare/cloudflared/cmd/cloudflared"
	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/util"
)

const CloudflaredAvailable = true

var cfMutex sync.Mutex

func InitCloudflared() {
	cfMutex.Lock()
	defer cfMutex.Unlock()

	conf := model.Conf.Tunnel
	if conf == nil || conf.Cloudflared == nil || !conf.Cloudflared.Enable {
		return
	}

	cf := conf.Cloudflared
	logging.LogInfof("starting cloudflared tunnel [type=%s]", cf.TunnelType)

	cloudflared.Init()
	setCloudflaredStatus(false, "", "starting")

	switch cf.TunnelType {
	case "quick":
		port, _ := strconv.Atoi(util.ServerPort)
		cloudflared.StartQuickTunnel(port)
	case "named":
		if cf.TunnelToken == "" {
			setCloudflaredStatus(false, "", "tunnel token is required for named tunnels")
			logging.LogErrorf("cloudflared named tunnel requires a token")
			return
		}
		cloudflared.RunNamed(cf.TunnelToken)
	default:
		setCloudflaredStatus(false, "", fmt.Sprintf("unknown tunnel type: %s", cf.TunnelType))
		return
	}

	// Poll for tunnel URL readiness (up to 60 seconds)
	go func() {
		for i := 0; i < 60; i++ {
			time.Sleep(1 * time.Second)
			if cloudflared.IsReady() {
				url := cloudflared.GetURL()
				setCloudflaredStatus(true, url, "")
				logging.LogInfof("cloudflared tunnel is running [url=%s]", url)
				util.PushMsg(fmt.Sprintf("Cloudflare tunnel: %s", url), 7000)
				return
			}
		}
		setCloudflaredStatus(false, "", "timeout waiting for tunnel URL")
		logging.LogErrorf("cloudflared tunnel did not become ready within 60 seconds")
	}()
}

func StopCloudflared() {
	cfMutex.Lock()
	defer cfMutex.Unlock()

	cloudflared.Stop()
	setCloudflaredStatus(false, "", "")
	logging.LogInfof("cloudflared tunnel stopped")
}

func RestartCloudflared() {
	StopCloudflared()
	InitCloudflared()
}
