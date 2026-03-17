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

package conf

type Tunnel struct {
	Tailscale   *TailscaleConf   `json:"tailscale"`
	Cloudflared *CloudflaredConf `json:"cloudflared"`
}

type TailscaleConf struct {
	Enable   bool   `json:"enable"`   // 是否启用 Tailscale 隧道
	Hostname string `json:"hostname"` // Tailnet 主机名，默认 "siyuan"
	AuthKey  string `json:"authKey"`  // 预认证密钥（可选，为空则需交互登录）
	Port     uint16 `json:"port"`     // Tailnet 监听端口，默认 443
	UseTLS   bool   `json:"useTLS"`   // 是否使用 tsnet 内置 Let's Encrypt TLS
	StateDir string `json:"stateDir"` // 状态目录（默认：ConfDir/tsnet）
}

type CloudflaredConf struct {
	Enable      bool   `json:"enable"`      // 是否启用 Cloudflare 隧道
	TunnelType  string `json:"tunnelType"`  // "quick"（无需账号）或 "named"
	TunnelToken string `json:"tunnelToken"` // 命名隧道的 Token
}

func NewTunnel() *Tunnel {
	return &Tunnel{
		Tailscale: &TailscaleConf{
			Enable:   false,
			Hostname: "siyuan",
			Port:     443,
			UseTLS:   true,
		},
		Cloudflared: &CloudflaredConf{
			Enable:     false,
			TunnelType: "quick",
		},
	}
}
