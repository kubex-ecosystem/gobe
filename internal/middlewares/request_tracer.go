package middlewares

import (
	ci "github.com/rafa-mori/gobe/internal/interfaces"
)

type RequestTracerMiddleware struct {
	requestsTracers map[string]ci.IRequestsTracer
}

func NewRequestTracerMiddleware() *RequestTracerMiddleware {
	return &RequestTracerMiddleware{
		requestsTracers: make(map[string]ci.IRequestsTracer),
	}
}

func (g *RequestTracerMiddleware) GetRequestTracers() map[string]ci.IRequestsTracer {
	//g.Mutexes.MuRLock()
	//defer g.Mutexes.MuRUnlock()
	return g.requestsTracers
}
func (g *RequestTracerMiddleware) SetRequestTracers(tracers map[string]ci.IRequestsTracer) {
	/*g.Mutexes.MuAdd(1)
	defer g.Mutexes.MuDone()*/
	g.requestsTracers = tracers
}
