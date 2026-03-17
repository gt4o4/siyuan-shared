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
	"sync"
	"time"
)

type TunnelStatus struct {
	Running   bool   `json:"running"`
	URL       string `json:"url"`
	Error     string `json:"error"`
	StartedAt int64  `json:"startedAt"`
}

var (
	tailscaleStatus   TunnelStatus
	cloudflaredStatus TunnelStatus
	tsMu              sync.RWMutex
	cfMu              sync.RWMutex
)

func GetTailscaleStatus() TunnelStatus {
	tsMu.RLock()
	defer tsMu.RUnlock()
	return tailscaleStatus
}

func setTailscaleStatus(running bool, url, errMsg string) {
	tsMu.Lock()
	defer tsMu.Unlock()
	tailscaleStatus.Running = running
	tailscaleStatus.URL = url
	tailscaleStatus.Error = errMsg
	if running && tailscaleStatus.StartedAt == 0 {
		tailscaleStatus.StartedAt = time.Now().Unix()
	}
	if !running {
		tailscaleStatus.StartedAt = 0
	}
}

func GetCloudflaredStatus() TunnelStatus {
	cfMu.RLock()
	defer cfMu.RUnlock()
	return cloudflaredStatus
}

func setCloudflaredStatus(running bool, url, errMsg string) {
	cfMu.Lock()
	defer cfMu.Unlock()
	cloudflaredStatus.Running = running
	cloudflaredStatus.URL = url
	cloudflaredStatus.Error = errMsg
	if running && cloudflaredStatus.StartedAt == 0 {
		cloudflaredStatus.StartedAt = time.Now().Unix()
	}
	if !running {
		cloudflaredStatus.StartedAt = 0
	}
}
