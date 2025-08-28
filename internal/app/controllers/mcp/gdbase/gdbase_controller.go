// Package gdbase provides the MetricsController for handling system metrics and related operations in the GDBASE module.
package gdbase

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafa-mori/gobe/internal/bridges/gdbasez"
	"github.com/rafa-mori/gobe/internal/module/logger"
	"github.com/rafa-mori/gobe/internal/services/mcp/hooks"
	"github.com/rafa-mori/gobe/internal/services/mcp/system"
	"gorm.io/gorm"

	l "github.com/rafa-mori/logz"
)

// type TunnelMode string

// const (
// 	ModeOff   TunnelMode = "off"
// 	ModeQuick TunnelMode = "quick"
// 	ModeNamed TunnelMode = "named"
// )

// type TunnelManager interface {
// 	Up(ctx context.Context, mode TunnelMode, args any) (public string, err error)
// 	Down(ctx context.Context) error
// 	Status() (mode TunnelMode, public string, running bool)
// }

// // --- args “genéricos” (ajuste aos teus bridges/SDK) ---
// type QuickArgs struct {
// 	Network string        `json:"network"`
// 	Target  string        `json:"target"` // service DNS no docker (ex.: "pgadmin")
// 	Port    int           `json:"port"`   // ex.: 80
// 	Timeout time.Duration `json:"timeout,omitempty"`
// }

// type NamedArgs struct {
// 	Network string `json:"network"`
// 	Token   string `json:"token"` // TUNNEL_TOKEN
// }

// // --- Controller ---
// type TunnelController struct {
//
// }

// POST /_admin/tunnel/up
// body: { "mode":"quick","target":"pgadmin","port":80,"network":"gdbase_net" }

// }

// // POST /_admin/tunnel/down

// }

// // GET /_admin/tunnel/status
// func (c *TunnelController) Status(w http.ResponseWriter, r *http.Request) {

// }

// func NewTunnelController(tm TunnelManager) *TunnelController { return &TunnelController{TM: tm} }

var (
	gl = logger.GetLogger[l.Logger](nil)
)

type GDBaseController struct {
	dbConn   *gorm.DB
	mcpState *hooks.Bitstate[uint64, system.SystemDomain]
	TM       *gdbasez.TunnelHandle
}

func NewGDBaseController(db *gorm.DB) *GDBaseController {
	if db == nil {
		gl.Log("warn", "Database connection is nil")
	}

	// We allow the system service to be nil, as it can be set later.
	return &GDBaseController{
		dbConn: db,
	}
}

func (g *GDBaseController) PostGDBaseTunnelUp(c *gin.Context) {
	//
	//	{ "mode":"named","token":"***","network":"gdbase_net" }
	//
	// func (c *TunnelController) Up(w http.ResponseWriter, r *http.Request) {
	// 	type req struct {
	// 		Mode    string `json:"mode"`
	// 		Network string `json:"network,omitempty"`
	// 		Target  string `json:"target,omitempty"`
	// 		Port    int    `json:"port,omitempty"`
	// 		Token   string `json:"token,omitempty"`
	// 		Timeout string `json:"timeout,omitempty"` // "10s"
	// 	}
	// 	var in req
	// 	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
	// 		writeErr(w, http.StatusBadRequest, "invalid json: "+err.Error())
	// 		return
	// 	}

	// 	mode := TunnelMode(in.Mode)
	// 	switch mode {
	// 	case ModeQuick:
	// 		to := 10 * time.Second
	// 		if in.Timeout != "" {
	// 			if d, err := time.ParseDuration(in.Timeout); err == nil && d > 0 {
	// 				to = d
	// 			}
	// 		}
	// 		if in.Target == "" || in.Port <= 0 {
	// 			writeErr(w, http.StatusBadRequest, "quick mode requires target and port")
	// 			return
	// 		}
	// 		pub, err := c.TM.Up(r.Context(), ModeQuick, QuickArgs{
	// 			Network: in.Network,
	// 			Target:  in.Target,
	// 			Port:    in.Port,
	// 			Timeout: to,
	// 		})
	// 		if err != nil {
	// 			writeErr(w, http.StatusInternalServerError, err.Error())
	// 			return
	// 		}
	// 		writeJSON(w, http.StatusOK, map[string]any{
	// 			"mode":    ModeQuick,
	// 			"public":  pub, // *.trycloudflare.com
	// 			"network": in.Network,
	// 			"target":  in.Target + ":" + strconv.Itoa(in.Port),
	// 		})
	// 		return

	// case ModeNamed:
	//
	//	if in.Token == "" {
	//		writeErr(w, http.StatusBadRequest, "named mode requires token")
	//		return
	//	}
	//	_, err := c.TM.Up(r.Context(), ModeNamed, NamedArgs{
	//		Network: in.Network,
	//		Token:   in.Token,
	//	})
	//	if err != nil {
	//		writeErr(w, http.StatusInternalServerError, err.Error())
	//		return
	//	}
	//	writeJSON(w, http.StatusOK, map[string]any{
	//		"mode":    ModeNamed,
	//		"public":  "(use seus hostnames do tunnel)",
	//		"network": in.Network,
	//	})
	//	return
	//
	// default:
	//
	//		writeErr(w, http.StatusBadRequest, "mode must be quick or named")
	//		return
	//	}
}

func (g *GDBaseController) PostGDBaseTunnelDown(c *gin.Context) {
	//	func (c *TunnelController) Down(w http.ResponseWriter, r *http.Request) {
	//		if err := c.TM.Down(r.Context()); err != nil {
	//			writeErr(w, http.StatusInternalServerError, err.Error())
	//			return
	//		}
	//		w.WriteHeader(http.StatusNoContent)
}
func (g *GDBaseController) GetGDBaseTunnelStatus(c *gin.Context) {
	// mode, pub, running := c.TM.Status()
	//
	//	writeJSON(w, http.StatusOK, map[string]any{
	//		"mode":    mode,
	//		"public":  pub,
	//		"running": running,
	//	})
}

// --- helpers ---
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{
		"error":   http.StatusText(code),
		"message": msg,
	})
}
