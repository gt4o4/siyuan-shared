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

package api

import (
	"net/http"

	"github.com/88250/gulu"
	"github.com/gin-gonic/gin"
	"github.com/siyuan-note/siyuan/kernel/conf"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/server/tunnel"
	"github.com/siyuan-note/siyuan/kernel/util"
)

func setTunnel(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	arg, ok := util.JsonArg(c, ret)
	if !ok {
		return
	}

	param, err := gulu.JSON.MarshalJSON(arg)
	if err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	tunnelConf := &conf.Tunnel{}
	if err = gulu.JSON.UnmarshalJSON(param, tunnelConf); err != nil {
		ret.Code = -1
		ret.Msg = err.Error()
		return
	}

	oldTailscaleEnable := model.Conf.Tunnel.Tailscale != nil && model.Conf.Tunnel.Tailscale.Enable
	oldCloudflaredEnable := model.Conf.Tunnel.Cloudflared != nil && model.Conf.Tunnel.Cloudflared.Enable

	model.Conf.Tunnel = tunnelConf
	model.Conf.Save()

	newTailscaleEnable := tunnelConf.Tailscale != nil && tunnelConf.Tailscale.Enable
	newCloudflaredEnable := tunnelConf.Cloudflared != nil && tunnelConf.Cloudflared.Enable

	// Restart tunnels if config changed
	if oldTailscaleEnable != newTailscaleEnable || newTailscaleEnable {
		go tunnel.RestartTailscale()
	}
	if oldCloudflaredEnable != newCloudflaredEnable || newCloudflaredEnable {
		go tunnel.RestartCloudflared()
	}

	ret.Data = map[string]any{
		"tunnel":              model.Conf.Tunnel,
		"cloudflaredAvailable": tunnel.CloudflaredAvailable,
	}
}

func getTunnel(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	ret.Data = map[string]any{
		"tunnel":              model.Conf.Tunnel,
		"tailscaleStatus":    tunnel.GetTailscaleStatus(),
		"cloudflaredStatus":  tunnel.GetCloudflaredStatus(),
		"cloudflaredAvailable": tunnel.CloudflaredAvailable,
	}
}

func tunnelStatus(c *gin.Context) {
	ret := gulu.Ret.NewResult()
	defer c.JSON(http.StatusOK, ret)

	ret.Data = map[string]any{
		"tailscaleStatus":    tunnel.GetTailscaleStatus(),
		"cloudflaredStatus":  tunnel.GetCloudflaredStatus(),
		"cloudflaredAvailable": tunnel.CloudflaredAvailable,
	}
}
