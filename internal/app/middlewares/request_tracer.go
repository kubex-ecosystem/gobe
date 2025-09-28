package middlewares

import (
	"github.com/gin-gonic/gin"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

type RequestTracerMiddleware struct {
	*t.RequestTracers
}

func NewRequestTracerMiddlewareType(g ci.IGoBE) *RequestTracerMiddleware {
	requestTracers, ok := t.NewRequestTracers(g).(*t.RequestTracers)
	if !ok {
		gl.Log("error", "Failed to create RequestTracers instance")
		return nil
	}

	return &RequestTracerMiddleware{
		RequestTracers: requestTracers,
	}
}

func NewRequestTracerMiddleware(g ci.IGoBE) gin.HandlerFunc {
	RequestsTracerMiddleware := NewRequestTracerMiddlewareType(g)
	return RequestsTracerMiddleware.RequestsTracerMiddleware()
}

func (gm *RequestTracerMiddleware) GetRequestTracers() map[string]ci.IRequestsTracer {
	//g.Mutexes.MuRLock()
	//defer g.Mutexes.MuRUnlock()
	return gm.RequestTracers.GetRequestTracers()
}
func (gm *RequestTracerMiddleware) SetRequestTracers(tracers map[string]ci.IRequestsTracer) {
	/*g.Mutexes.MuAdd(1)
	defer g.Mutexes.MuDone()*/
	gm.RequestTracers.SetRequestTracers(tracers)
}
func (gm *RequestTracerMiddleware) AddRequestTracer(name string, tracer ci.IRequestsTracer) {
	//g.Mutexes.MuAdd(1)
	//defer g.Mutexes.MuDone()
	gm.RequestTracers.AddRequestTracer(name, tracer)
}
func (gm *RequestTracerMiddleware) GetRequestTracer(name string) (ci.IRequestsTracer, bool) {
	//g.Mutexes.MuRLock()
	//defer g.Mutexes.MuRUnlock()
	tracer, ok := gm.RequestTracers.GetRequestTracer(name)
	return tracer, ok
}
func (gm *RequestTracerMiddleware) RemoveRequestTracer(name string) {
	//g.Mutexes.MuAdd(1)
	//defer g.Mutexes.MuDone()
	gm.RequestTracers.RemoveRequestTracer(name)
}
