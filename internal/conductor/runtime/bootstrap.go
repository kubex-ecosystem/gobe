// Package runtime provides the runtime environment for the conductor.
package runtime

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kubex-ecosystem/gobe/internal/module/kbx"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler"
	logz "github.com/kubex-ecosystem/logz/api/notifiers"
	gl "github.com/kubex-ecosystem/logz/logger"
)

var wsNotifierConfig *logz.NotifierWebSocketConfig
var notifier *logz.WebSocketNotifier

func init() {
	// inicia métricas/WS do logz (Prometheus + hub WS, se habilitado)
	wsNotifierConfig = logz.NewNotifierWebSocketConfig(
		nil, //*tls.Config{},
		2*time.Second,
		nil,
		nil, //websocket.BufferPool,
		0,
		20,
		[]string{},
		false,
		func(network string, addr string) (net.Conn, error) {
			return net.Dial(network, addr)
		},
		func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return net.Dial(network, addr)
		},
		func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return net.Dial(network, addr)
		},
		func(*http.Request) (*url.URL, error) { return nil, nil },
	)

	// notifier = logz.NewWebSocketNotifier(wsNotifierConfig)
}

func Init(router *gin.Engine, s scheduler.IScheduler) {
	// log.Println("[Conductor] initializing runtime...")
	gl.Log("info", "[Conductor] initializing runtime...")

	if notifier.EnabledFlag {
		notifier.Enable()
		notifier.AuthToken = kbx.GetEnvOrDefault("LOGZ_AUTH_TOKEN", "")
		gl.Log("info", "[Conductor] WebSocket notifier enabled.")
	}

	d := NewDispatcher(s)
	BindDispatcher(d)
	router.Use(RouteIntentMiddleware)

	gl.Log("info", "[Conductor] middleware attached, DCL runtime (logz-backed) ativo.")
}
