// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package apicontract

type SettingTailscaleConf struct {
	Enable   bool   `json:"enable" api:"optional,nullable"`
	Hostname string `json:"hostname" api:"optional,nullable"`
	AuthKey  string `json:"authKey" api:"optional,nullable"`
	Port     uint16 `json:"port" api:"optional,nullable"`
	UseTLS   bool   `json:"useTLS" api:"optional,nullable"`
	StateDir string `json:"stateDir" api:"optional,nullable"`
}

type SettingCloudflaredConf struct {
	Enable      bool   `json:"enable" api:"optional,nullable"`
	TunnelType  string `json:"tunnelType" api:"optional,nullable"`
	TunnelToken string `json:"tunnelToken" api:"optional,nullable"`
}

type SettingTunnel struct {
	Tailscale   *SettingTailscaleConf   `json:"tailscale" api:"optional,nullable"`
	Cloudflared *SettingCloudflaredConf `json:"cloudflared" api:"optional,nullable"`
}

type TunnelRuntimeStatus struct {
	Running   bool   `json:"running"`
	URL       string `json:"url"`
	Error     string `json:"error"`
	StartedAt int64  `json:"startedAt"`
}

type SetTunnelData struct {
	Tunnel               *SettingTunnel `json:"tunnel" api:"nullable"`
	CloudflaredAvailable bool           `json:"cloudflaredAvailable"`
}

type GetTunnelData struct {
	Tunnel               *SettingTunnel      `json:"tunnel" api:"nullable"`
	TailscaleStatus      TunnelRuntimeStatus `json:"tailscaleStatus"`
	CloudflaredStatus    TunnelRuntimeStatus `json:"cloudflaredStatus"`
	CloudflaredAvailable bool                `json:"cloudflaredAvailable"`
}

type TunnelStatusData struct {
	TailscaleStatus      TunnelRuntimeStatus `json:"tailscaleStatus"`
	CloudflaredStatus    TunnelRuntimeStatus `json:"cloudflaredStatus"`
	CloudflaredAvailable bool                `json:"cloudflaredAvailable"`
}
