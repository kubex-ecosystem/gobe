package interfaces

import (
	"time"

	"github.com/gin-gonic/gin"
)

type IRequestsTracer interface {
	GetIP() string
	GetPort() string
	GetLastUserAgent() string
	GetUserAgents() []string
	GetEndpoint() string
	GetMethod() string
	GetTimeList() []time.Time
	GetCount() int
	GetError() error
	GetMutexes() IMutexes
	IsValid() bool
	GetOldFilePath() string

	GetFilePath() string
	SetFilePath(filePath string)
	GetMapper() IMapper[IRequestsTracer]
	SetMapper(mapper IMapper[IRequestsTracer])
	GetRequestWindow() time.Duration
	SetRequestWindow(window time.Duration)
	GetRequestLimit() int
	SetRequestLimit(limit int)
	Mu() IMutexes
}

type IRequestTracers interface {
	RequestsTracerMiddleware() gin.HandlerFunc
	GetRequestTracers() map[string]IRequestsTracer
	SetRequestTracers(tracers map[string]IRequestsTracer)
	AddRequestTracer(name string, tracer IRequestsTracer)
	GetRequestTracer(name string) (IRequestsTracer, bool)
	RemoveRequestTracer(name string)
}
