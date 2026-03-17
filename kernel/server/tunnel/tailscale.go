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
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/siyuan-note/logging"
	"github.com/siyuan-note/siyuan/kernel/model"
	"github.com/siyuan-note/siyuan/kernel/util"
	"tailscale.com/tsnet"
)

var (
	tsServer   *tsnet.Server
	tsListener net.Listener
	tsHTTP     *http.Server
	tsMutex    sync.Mutex
)

func InitTailscale() {
	tsMutex.Lock()
	defer tsMutex.Unlock()

	conf := model.Conf.Tunnel
	if conf == nil || conf.Tailscale == nil || !conf.Tailscale.Enable {
		return
	}

	ts := conf.Tailscale
	logging.LogInfof("starting tailscale tunnel [hostname=%s, port=%d, tls=%v]", ts.Hostname, ts.Port, ts.UseTLS)

	stateDir := ts.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(util.ConfDir, "tsnet")
	}
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		logging.LogErrorf("create tsnet state dir failed: %s", err)
		setTailscaleStatus(false, "", err.Error())
		return
	}

	tsServer = &tsnet.Server{
		Dir:      stateDir,
		Hostname: ts.Hostname,
		AuthKey:  ts.AuthKey,
		Logf: func(format string, args ...any) {
			msg := fmt.Sprintf(format, args...)
			// Capture auth URL and push to UI
			if strings.Contains(msg, "https://login.tailscale.com/") {
				for _, word := range strings.Fields(msg) {
					if strings.HasPrefix(word, "https://login.tailscale.com/") {
						util.PushMsg(fmt.Sprintf("Tailscale login required: %s", word), 0)
						break
					}
				}
			}
			logging.LogInfof("tsnet: %s", msg)
		},
	}

	if err := tsServer.Start(); err != nil {
		logging.LogErrorf("start tsnet server failed: %s", err)
		setTailscaleStatus(false, "", err.Error())
		tsServer = nil
		return
	}

	addr := fmt.Sprintf(":%d", ts.Port)
	var ln net.Listener
	var err error
	if ts.UseTLS {
		ln, err = tsServer.ListenTLS("tcp", addr)
	} else {
		ln, err = tsServer.Listen("tcp", addr)
	}
	if err != nil {
		logging.LogErrorf("tsnet listen failed: %s", err)
		setTailscaleStatus(false, "", err.Error())
		tsServer.Close()
		tsServer = nil
		return
	}
	tsListener = ln

	// Build tunnel URL from cert domains
	url := ""
	domains := tsServer.CertDomains()
	if len(domains) > 0 {
		scheme := "http"
		if ts.UseTLS {
			scheme = "https"
		}
		if ts.Port == 443 || ts.Port == 80 {
			url = fmt.Sprintf("%s://%s", scheme, domains[0])
		} else {
			url = fmt.Sprintf("%s://%s:%d", scheme, domains[0], ts.Port)
		}
	}

	tsHTTP = &http.Server{
		Handler: &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(util.ServerURL)
				r.SetXForwarded()
			},
		},
	}

	setTailscaleStatus(true, url, "")
	logging.LogInfof("tailscale tunnel is running [url=%s]", url)
	if url != "" {
		util.PushMsg(fmt.Sprintf("Tailscale tunnel: %s", url), 7000)
	}

	go func() {
		if err := tsHTTP.Serve(tsListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logging.LogErrorf("tailscale http serve failed: %s", err)
			setTailscaleStatus(false, "", err.Error())
		}
	}()
}

func StopTailscale() {
	tsMutex.Lock()
	defer tsMutex.Unlock()

	if tsHTTP != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tsHTTP.Shutdown(ctx)
		tsHTTP = nil
	}

	if tsListener != nil {
		tsListener.Close()
		tsListener = nil
	}

	if tsServer != nil {
		tsServer.Close()
		tsServer = nil
	}

	setTailscaleStatus(false, "", "")
	logging.LogInfof("tailscale tunnel stopped")
}

func RestartTailscale() {
	StopTailscale()
	InitTailscale()
}
