package interfaces

import "time"

type ITLSConfig interface {
	GetCertFile() string
	GetKeyFile() string
	GetCAFile() string
	GetEnabled() bool
	GetSkipVerify() bool
	GetStrictHostKey() bool
	GetMinVersion() string
	SetCertFile(string)
	SetKeyFile(string)
	SetCAFile(string)
	SetEnabled(bool)
	SetSkipVerify(bool)
	SetStrictHostKey(bool)
	SetMinVersion(string)
	GetTLSConfig() ITLSConfig
	SetTLSConfig(ITLSConfig)
	GetReference() IReference
	GetMutexes() IMutexes
	SetReference(IReference)
	SetMutexes(IMutexes)
	Save() error
	Load() error
}

type IGoBEConfig interface {
	GetFilePath() string
	GetWorkerThreads() int
	GetRateLimitLimit() int
	GetRateLimitBurst() int
	GetRequestWindow() time.Duration
	GetProxyEnabled() bool
	GetProxyHost() string
	GetProxyPort() string
	GetBindAddress() string
	GetPort() string
	GetTimeouts() time.Duration
	GetMaxConnections() int
	GetLogLevel() string
	GetLogFormat() string
	GetLogDir() string
	GetRequestLogging() bool
	GetMetricsEnabled() bool
	GetJWTSecretKey() string
	GetRefreshTokenExpiration() time.Duration
	GetAccessTokenExpiration() time.Duration
	GetTLSConfig() ITLSConfig
	GetAllowedOrigins() []string
	GetAPIKeyAuth() bool
	GetAPIKey() string
	GetConfigFormat() string
	GetMapper() IMapper[IGoBEConfig]
	SetFilePath(string)
	SetWorkerThreads(int)
	SetRateLimitLimit(int)
	SetRateLimitBurst(int)
	SetRequestWindow(time.Duration)
	SetProxyEnabled(bool)
	SetProxyHost(string)
	SetProxyPort(string)
	SetBindAddress(string)
	SetPort(string)
	SetTimeouts(time.Duration)
	SetMaxConnections(int)
	SetLogLevel(string)
	SetLogFormat(string)
	SetLogFile(string)
	SetRequestLogging(bool)
	SetMetricsEnabled(bool)
	SetJWTSecretKey(string)
	SetRefreshTokenExpiration(time.Duration)
	SetAccessTokenExpiration(time.Duration)
	SetTLSConfig(ITLSConfig)
	SetAllowedOrigins([]string)
	SetAPIKeyAuth(bool)
	SetAPIKey(string)
	SetConfigFormat(string)
	SetMapper(IMapper[IGoBEConfig])
	Save() error
	Load() error
}
