// SiYuan - https://github.com/siyuan-note/siyuan
// Copyright (c) 2020-present, b3log.org
// SPDX-License-Identifier: AGPL-3.0-or-later

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/siyuan-note/siyuan/kernel/apicontract"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/server/tunnel"
)

func settingTunnelPayload(value *conf.Tunnel) *apicontract.SettingTunnel {
	if value == nil {
		return nil
	}
	result := &apicontract.SettingTunnel{}
	if value.Tailscale != nil {
		result.Tailscale = &apicontract.SettingTailscaleConf{
			Enable:   value.Tailscale.Enable,
			Hostname: value.Tailscale.Hostname,
			AuthKey:  value.Tailscale.AuthKey,
			Port:     value.Tailscale.Port,
			UseTLS:   value.Tailscale.UseTLS,
			StateDir: value.Tailscale.StateDir,
		}
	}
	if value.Cloudflared != nil {
		result.Cloudflared = &apicontract.SettingCloudflaredConf{
			Enable:      value.Cloudflared.Enable,
			TunnelType:  value.Cloudflared.TunnelType,
			TunnelToken: value.Cloudflared.TunnelToken,
		}
	}
	return result
}

func tunnelConfFromContract(request apicontract.SettingTunnel) *conf.Tunnel {
	tunnelConf := &conf.Tunnel{}
	if request.Tailscale != nil {
		tunnelConf.Tailscale = &conf.TailscaleConf{
			Enable:   request.Tailscale.Enable,
			Hostname: request.Tailscale.Hostname,
			AuthKey:  request.Tailscale.AuthKey,
			Port:     request.Tailscale.Port,
			UseTLS:   request.Tailscale.UseTLS,
			StateDir: request.Tailscale.StateDir,
		}
	}
	if request.Cloudflared != nil {
		tunnelConf.Cloudflared = &conf.CloudflaredConf{
			Enable:      request.Cloudflared.Enable,
			TunnelType:  request.Cloudflared.TunnelType,
			TunnelToken: request.Cloudflared.TunnelToken,
		}
	}
	return tunnelConf
}

func tunnelRuntimeStatusPayload(status tunnel.TunnelStatus) apicontract.TunnelRuntimeStatus {
	return apicontract.TunnelRuntimeStatus{
		Running:   status.Running,
		URL:       status.URL,
		Error:     status.Error,
		StartedAt: status.StartedAt,
	}
}

var setTunnel = contractHandler(apicontract.SetTunnel, func(c *gin.Context, request apicontract.SettingTunnel) apicontract.Response[apicontract.SetTunnelData] {
	tunnelConf := tunnelConfFromContract(request)

	oldTailscaleEnable := model.Conf.Tunnel != nil && model.Conf.Tunnel.Tailscale != nil && model.Conf.Tunnel.Tailscale.Enable
	oldCloudflaredEnable := model.Conf.Tunnel != nil && model.Conf.Tunnel.Cloudflared != nil && model.Conf.Tunnel.Cloudflared.Enable

	model.Conf.Tunnel = tunnelConf
	model.Conf.Save()

	newTailscaleEnable := tunnelConf.Tailscale != nil && tunnelConf.Tailscale.Enable
	newCloudflaredEnable := tunnelConf.Cloudflared != nil && tunnelConf.Cloudflared.Enable

	// 隧道开关变化或启用时重启对应隧道。
	if oldTailscaleEnable != newTailscaleEnable || newTailscaleEnable {
		go tunnel.RestartTailscale()
	}
	if oldCloudflaredEnable != newCloudflaredEnable || newCloudflaredEnable {
		go tunnel.RestartCloudflared()
	}

	return apicontract.Success(apicontract.SetTunnelData{
		Tunnel:               settingTunnelPayload(model.Conf.Tunnel),
		CloudflaredAvailable: tunnel.CloudflaredAvailable,
	})
})

var getTunnel = contractHandler(apicontract.GetTunnel, func(c *gin.Context, request apicontract.EmptyRequest) apicontract.Response[apicontract.GetTunnelData] {
	return apicontract.Success(apicontract.GetTunnelData{
		Tunnel:               settingTunnelPayload(model.Conf.Tunnel),
		TailscaleStatus:      tunnelRuntimeStatusPayload(tunnel.GetTailscaleStatus()),
		CloudflaredStatus:    tunnelRuntimeStatusPayload(tunnel.GetCloudflaredStatus()),
		CloudflaredAvailable: tunnel.CloudflaredAvailable,
	})
})

var tunnelStatus = contractHandler(apicontract.GetTunnelStatus, func(c *gin.Context, request apicontract.EmptyRequest) apicontract.Response[apicontract.TunnelStatusData] {
	return apicontract.Success(apicontract.TunnelStatusData{
		TailscaleStatus:      tunnelRuntimeStatusPayload(tunnel.GetTailscaleStatus()),
		CloudflaredStatus:    tunnelRuntimeStatusPayload(tunnel.GetCloudflaredStatus()),
		CloudflaredAvailable: tunnel.CloudflaredAvailable,
	})
})
