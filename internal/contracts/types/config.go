package types

import (
	"errors"
	"fmt"
	"os"
	"time"

	crp "github.com/kubex-ecosystem/gobe/internal/app/security/crypto"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	"github.com/kubex-ecosystem/gobe/internal/module/kbx"
	gl "github.com/kubex-ecosystem/logz/logger"
)

// TLSConfig is a struct that holds the TLS configuration for the GoBE instance.
type TLSConfig struct {
	*Reference
	*Mutexes
	CertFile      string `json:"cert_file" yaml:"cert_file" env:"CERT_FILE" toml:"cert_file" xml:"cert_file" gorm:"cert_file"`
	KeyFile       string `json:"key_file" yaml:"key_file" env:"KEY_FILE" toml:"key_file" xml:"key_file" gorm:"key_file"`
	CAFile        string `json:"ca_file" yaml:"ca_file" env:"CA_FILE" toml:"ca_file" xml:"ca_file" gorm:"ca_file"`
	Enabled       bool   `json:"enabled" yaml:"enabled" env:"TLS_ENABLED" toml:"tls_enabled" xml:"tls_enabled" gorm:"tls_enabled"`
	SkipVerify    bool   `json:"skip_verify" yaml:"skip_verify" env:"TLS_SKIP_VERIFY" toml:"tls_skip_verify" xml:"tls_skip_verify" gorm:"tls_skip_verify"`
	StrictHostKey bool   `json:"strict_host_key" yaml:"strict_host_key" env:"TLS_STRICT_HOST_KEY" toml:"tls_strict_host_key" xml:"tls_strict_host_key" gorm:"tls_strict_host_key"`
	MinVersion    string `json:"min_version" yaml:"min_version" env:"TLS_MIN_VERSION" toml:"tls_min_version" xml:"tls_min_version" gorm:"tls_min_version"`
	Mapper        ci.IMapper[*TLSConfig]
}

// newTLSConfig creates a new TLSConfig instance with the provided name and file path.
func newTLSConfig(name, filePath string) *TLSConfig {
	tlsCfg := &TLSConfig{
		Reference:     newReference("TLSConfig").GetReference(),
		Mutexes:       NewMutexesType(),
		CertFile:      "",
		KeyFile:       "",
		CAFile:        "",
		Enabled:       false,
		SkipVerify:    false,
		StrictHostKey: false,
		MinVersion:    "TLS1.2",
	}

	tlsCfg.Mapper = NewMapper(&tlsCfg, filePath)

	return tlsCfg
}

// NewTLSConfig creates a new TLSConfig instance with the provided name and file path.
func NewTLSConfig(name, filePath string) ci.ITLSConfig {
	return newTLSConfig(name, filePath)
}

func (t *TLSConfig) GetCertFile() string                 { return t.CertFile }
func (t *TLSConfig) GetKeyFile() string                  { return t.KeyFile }
func (t *TLSConfig) GetCAFile() string                   { return t.CAFile }
func (t *TLSConfig) GetEnabled() bool                    { return t.Enabled }
func (t *TLSConfig) GetSkipVerify() bool                 { return t.SkipVerify }
func (t *TLSConfig) GetStrictHostKey() bool              { return t.StrictHostKey }
func (t *TLSConfig) GetMinVersion() string               { return t.MinVersion }
func (t *TLSConfig) SetCertFile(certFile string)         { t.CertFile = certFile }
func (t *TLSConfig) SetKeyFile(keyFile string)           { t.KeyFile = keyFile }
func (t *TLSConfig) SetCAFile(caFile string)             { t.CAFile = caFile }
func (t *TLSConfig) SetEnabled(enabled bool)             { t.Enabled = enabled }
func (t *TLSConfig) SetSkipVerify(skipVerify bool)       { t.SkipVerify = skipVerify }
func (t *TLSConfig) SetStrictHostKey(strictHostKey bool) { t.StrictHostKey = strictHostKey }
func (t *TLSConfig) SetMinVersion(minVersion string)     { t.MinVersion = minVersion }
func (t *TLSConfig) GetTLSConfig() ci.ITLSConfig         { return t }
func (t *TLSConfig) SetTLSConfig(tlsConfig ci.ITLSConfig) {
	t.CertFile = tlsConfig.GetCAFile()
	t.KeyFile = tlsConfig.GetKeyFile()
	t.CAFile = tlsConfig.GetCAFile()
	t.Enabled = tlsConfig.GetEnabled()
	t.SkipVerify = tlsConfig.GetSkipVerify()
	t.StrictHostKey = tlsConfig.GetStrictHostKey()
	t.MinVersion = tlsConfig.GetMinVersion()
}
func (t *TLSConfig) GetReference() ci.IReference { return nil }
func (t *TLSConfig) GetMutexes() ci.IMutexes     { return nil }
func (t *TLSConfig) SetReference(ref ci.IReference) {
	// No-op
}
func (t *TLSConfig) SetMutexes(mutexes ci.IMutexes) {
	// No-op
}
func (t *TLSConfig) Save() error {
	// No-op
	return nil
}
func (t *TLSConfig) Load() error {
	// No-op
	return nil
}

type GoBEConfig struct {
	*Reference
	*Mutexes

	FilePath string `json:"file_path" yaml:"file_path" env:"FILE_PATH" toml:"file_path" xml:"file_path" gorm:"file_path"`

	WorkerThreads int `json:"worker_threads" yaml:"worker_threads" env:"WORKER_THREADS" toml:"worker_threads" xml:"worker_threads" gorm:"worker_threads"`

	RateLimitLimit int           `json:"rate_limit_limit" yaml:"rate_limit_limit" env:"RATE_LIMIT_LIMIT" toml:"rate_limit_limit" xml:"rate_limit_limit" gorm:"rate_limit_limit"`
	RateLimitBurst int           `json:"rate_limit_burst" yaml:"rate_limit_burst" env:"RATE_LIMIT_BURST" toml:"rate_limit_burst" xml:"rate_limit_burst" gorm:"rate_limit_burst"`
	RequestWindow  time.Duration `json:"request_window" yaml:"request_window" env:"REQUEST_WINDOW" toml:"request_window" xml:"request_window" gorm:"request_window"`

	ProxyEnabled   bool          `json:"proxy_enabled" yaml:"proxy_enabled" env:"PROXY_ENABLED" toml:"proxy_enabled" xml:"proxy_enabled" gorm:"proxy_enabled"`
	ProxyHost      string        `json:"proxy_host" yaml:"proxy_host" env:"PROXY_HOST" toml:"proxy_host" xml:"proxy_host" gorm:"proxy_host"`
	ProxyPort      string        `json:"proxy_port" yaml:"proxy_port" env:"PROXY_PORT" toml:"proxy_port" xml:"proxy_port" gorm:"proxy_port"`
	ProxyBindAddr  string        `json:"proxy_bind_addr" yaml:"proxy_bind_addr" env:"PROXY_BIND_ADDR" toml:"proxy_bind_addr" xml:"proxy_bind_addr" gorm:"proxy_bind_addr"`
	BasePath       string        `json:"base_path" yaml:"base_path" env:"BASE_PATH" toml:"base_path" xml:"base_path" gorm:"base_path"`
	Port           string        `json:"port" yaml:"port" env:"PORT" toml:"port" xml:"port" gorm:"port"`
	BindAddress    string        `json:"bind_address" yaml:"bind_address" env:"BIND_ADDRESS" toml:"bind_address" xml:"bind_address" gorm:"bind_address"`
	Timeouts       time.Duration `json:"timeouts" yaml:"timeouts" env:"TIMEOUTS" toml:"timeouts" xml:"timeouts" gorm:"timeouts"`
	MaxConnections int           `json:"max_connections" yaml:"max_connections" env:"MAX_CONNECTIONS" toml:"max_connections" xml:"max_connections" gorm:"max_connections"`

	LogLevel       string `json:"log_level" yaml:"log_level" env:"LOG_LEVEL" toml:"log_level" xml:"log_level" gorm:"log_level"`
	LogFormat      string `json:"log_format" yaml:"log_format" env:"LOG_FORMAT" toml:"log_format" xml:"log_format" gorm:"log_format"`
	LogDir         string `json:"log_file" yaml:"log_file" env:"LOG_FILE" toml:"log_file" xml:"log_file" gorm:"log_file"`
	RequestLogging bool   `json:"request_logging" yaml:"request_logging" env:"REQUEST_LOGGING" toml:"request_logging" xml:"request_logging" gorm:"request_logging"`
	MetricsEnabled bool   `json:"metrics_enabled" yaml:"metrics_enabled" env:"METRICS_ENABLED" toml:"metrics_enabled" xml:"metrics_enabled" gorm:"metrics_enabled"`

	JWTSecretKey           string        `json:"jwt_secret_key" yaml:"jwt_secret"`
	RefreshTokenExpiration time.Duration `json:"refresh_token_expiration" yaml:"refresh_token_expiration" env:"REFRESH_TOKEN_EXPIRATION" toml:"refresh_token_expiration" xml:"refresh_token_expiration" gorm:"refresh_token_expiration"`
	AccessTokenExpiration  time.Duration `json:"access_token_expiration" yaml:"access_token_expiration" env:"ACCESS_TOKEN_EXPIRATION" toml:"access_token_expiration" xml:"access_token_expiration" gorm:"access_token_expiration"`
	TLSConfig              TLSConfig     `json:"tls_config" yaml:"tls_config" env:"TLS_CONFIG" toml:"tls_config" xml:"tls_config" gorm:"tls_config"`
	AllowedOrigins         []string      `json:"allowed_origins" yaml:"allowed_origins" env:"ALLOWED_ORIGINS" toml:"allowed_origins" xml:"allowed_origins" gorm:"allowed_origins"`
	APIKeyAuth             bool          `json:"api_key_auth" yaml:"api_key_auth" env:"API_KEY_AUTH" toml:"api_key_auth" xml:"api_key_auth" gorm:"api_key_auth"`
	APIKey                 string        `json:"api_key" yaml:"api_key" env:"API_KEY" toml:"api_key" xml:"api_key" gorm:"api_key"`

	ConfigFormat string `json:"config_format" yaml:"config_format" env:"CONFIG_FORMAT" toml:"config_format" xml:"config_format" gorm:"config_format"`

	Mapper ci.IMapper[*GoBEConfig]
}

func NewGoBEConfig(name, filePath, configFormat, bind, port string) *GoBEConfig {
	if configFormat == "" {
		configFormat = "yaml"
	}
	if filePath == "" {
		filePath = os.ExpandEnv(kbx.DefaultGoBEConfigPath)
	}
	if bind == "" {
		bind = "0.0.0.0"
	}
	if port == "" {
		port = "3666"
	}
	if name == "" {
		name = "GoBE"
	}

	gbmCfg := &GoBEConfig{
		Reference:              newReference(name).GetReference(),
		Mutexes:                NewMutexesType(),
		FilePath:               filePath,
		WorkerThreads:          2,
		RateLimitLimit:         0,
		RateLimitBurst:         0,
		RequestWindow:          time.Minute,
		ProxyEnabled:           false,
		ProxyHost:              "",
		ProxyPort:              "",
		BindAddress:            bind,
		BasePath:               "/",
		Port:                   port,
		Timeouts:               30 * time.Second,
		MaxConnections:         100,
		LogLevel:               "info",
		LogFormat:              "text",
		LogDir:                 "gobe.log",
		RequestLogging:         false,
		MetricsEnabled:         false,
		JWTSecretKey:           "",
		RefreshTokenExpiration: time.Hour * 24,
		AccessTokenExpiration:  time.Hour,
		TLSConfig: TLSConfig{
			CertFile:      "",
			KeyFile:       "",
			CAFile:        "",
			Enabled:       false,
			SkipVerify:    false,
			StrictHostKey: false,
			MinVersion:    "TLS1.2",
		},
		AllowedOrigins: []string{"*"},
		APIKeyAuth:     false,
		APIKey:         "",
		ConfigFormat:   "yaml",
	}

	gbmCfg.Mapper = NewMapper[*GoBEConfig](&gbmCfg, filePath)
	if _, statErr := os.Stat(filePath); statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			gbmCfg.Mapper.SerializeToFile(configFormat)
		} else {
			gl.Log("error", fmt.Sprintf("Failed to stat config file: %v", statErr))
		}
	} else {
		gbmCfg.Mapper.DeserializeFromFile(configFormat)
	}

	return gbmCfg
}

func (c *GoBEConfig) GetFilePath() string                      { return c.FilePath }
func (c *GoBEConfig) GetWorkerThreads() int                    { return c.WorkerThreads }
func (c *GoBEConfig) GetRateLimitLimit() int                   { return c.RateLimitLimit }
func (c *GoBEConfig) GetRateLimitBurst() int                   { return c.RateLimitBurst }
func (c *GoBEConfig) GetRequestWindow() time.Duration          { return c.RequestWindow }
func (c *GoBEConfig) GetProxyEnabled() bool                    { return c.ProxyEnabled }
func (c *GoBEConfig) GetProxyHost() string                     { return c.ProxyHost }
func (c *GoBEConfig) GetProxyPort() string                     { return c.ProxyPort }
func (c *GoBEConfig) GetBindAddress() string                   { return c.BindAddress }
func (c *GoBEConfig) GetPort() string                          { return c.Port }
func (c *GoBEConfig) GetTimeouts() time.Duration               { return c.Timeouts }
func (c *GoBEConfig) GetMaxConnections() int                   { return c.MaxConnections }
func (c *GoBEConfig) GetLogLevel() string                      { return c.LogLevel }
func (c *GoBEConfig) GetLogFormat() string                     { return c.LogFormat }
func (c *GoBEConfig) GetLogDir() string                        { return c.LogDir }
func (c *GoBEConfig) GetRequestLogging() bool                  { return c.RequestLogging }
func (c *GoBEConfig) GetMetricsEnabled() bool                  { return c.MetricsEnabled }
func (c *GoBEConfig) GetJWTSecretKey() string                  { return c.JWTSecretKey }
func (c *GoBEConfig) GetRefreshTokenExpiration() time.Duration { return c.RefreshTokenExpiration }
func (c *GoBEConfig) GetAccessTokenExpiration() time.Duration  { return c.AccessTokenExpiration }
func (c *GoBEConfig) GetTLSConfig() TLSConfig                  { return c.TLSConfig }
func (c *GoBEConfig) GetAllowedOrigins() []string              { return c.AllowedOrigins }
func (c *GoBEConfig) GetAPIKeyAuth() bool                      { return c.APIKeyAuth }
func (c *GoBEConfig) GetAPIKey() string                        { return c.APIKey }
func (c *GoBEConfig) GetConfigFormat() string                  { return c.ConfigFormat }
func (c *GoBEConfig) GetMapper() ci.IMapper[*GoBEConfig]       { return c.Mapper }

func (c *GoBEConfig) SetFilePath(filePath string)          { c.FilePath = filePath }
func (c *GoBEConfig) SetWorkerThreads(workerThreads int)   { c.WorkerThreads = workerThreads }
func (c *GoBEConfig) SetRateLimitLimit(rateLimitLimit int) { c.RateLimitLimit = rateLimitLimit }
func (c *GoBEConfig) SetRateLimitBurst(rateLimitBurst int) { c.RateLimitBurst = rateLimitBurst }
func (c *GoBEConfig) SetRequestWindow(requestWindow time.Duration) {
	c.RequestWindow = requestWindow
}
func (c *GoBEConfig) SetProxyEnabled(proxyEnabled bool)     { c.ProxyEnabled = proxyEnabled }
func (c *GoBEConfig) SetProxyHost(proxyHost string)         { c.ProxyHost = proxyHost }
func (c *GoBEConfig) SetProxyPort(proxyPort string)         { c.ProxyPort = proxyPort }
func (c *GoBEConfig) SetBindAddress(bindAddress string)     { c.BindAddress = bindAddress }
func (c *GoBEConfig) SetPort(port string)                   { c.Port = port }
func (c *GoBEConfig) SetTimeouts(timeouts time.Duration)    { c.Timeouts = timeouts }
func (c *GoBEConfig) SetMaxConnections(maxConnections int)  { c.MaxConnections = maxConnections }
func (c *GoBEConfig) SetLogLevel(logLevel string)           { c.LogLevel = logLevel }
func (c *GoBEConfig) SetLogFormat(logFormat string)         { c.LogFormat = logFormat }
func (c *GoBEConfig) SetLogFile(LogDir string)              { c.LogDir = LogDir }
func (c *GoBEConfig) SetRequestLogging(requestLogging bool) { c.RequestLogging = requestLogging }
func (c *GoBEConfig) SetMetricsEnabled(metricsEnabled bool) { c.MetricsEnabled = metricsEnabled }
func (c *GoBEConfig) SetJWTSecretKey(jwtSecretKey string) {
	cryptoService := crp.NewCryptoService()
	if jwtSecretKey == "" {
		gl.Log("error", "JWT secret key is empty")
		jwtSecretKeyByte, jwtSecretKeyByteErr := cryptoService.GenerateKeyWithLength(32)
		if jwtSecretKeyByteErr != nil {
			jwtSecretKey = ""
			gl.Log("fatal", fmt.Sprintf("Failed to generate JWT secret key: %v", jwtSecretKeyByteErr))
		} else {
			jwtSecretKey = cryptoService.EncodeBase64(jwtSecretKeyByte)
			if jwtSecretKey == "" {
				gl.Log("fatal", "Failed to generate JWT secret key")
			}
		}
	}
	c.JWTSecretKey = jwtSecretKey
}
func (c *GoBEConfig) SetRefreshTokenExpiration(refreshTokenExpiration time.Duration) {
	c.RefreshTokenExpiration = refreshTokenExpiration
}
func (c *GoBEConfig) SetAccessTokenExpiration(accessTokenExpiration time.Duration) {
	c.AccessTokenExpiration = accessTokenExpiration
}
func (c *GoBEConfig) SetTLSConfig(tlsConfig TLSConfig) { c.TLSConfig = tlsConfig }
func (c *GoBEConfig) SetAllowedOrigins(allowedOrigins []string) {
	c.AllowedOrigins = allowedOrigins
}
func (c *GoBEConfig) SetAPIKeyAuth(apiKeyAuth bool)       { c.APIKeyAuth = apiKeyAuth }
func (c *GoBEConfig) SetAPIKey(apiKey string)             { c.APIKey = apiKey }
func (c *GoBEConfig) SetConfigFormat(configFormat string) { c.ConfigFormat = configFormat }
func (c *GoBEConfig) SetMapper(mapper ci.IMapper[*GoBEConfig]) {
	c.Mapper = mapper
}
func (c *GoBEConfig) GetReference() ci.IReference { return c.Reference }

func (c *GoBEConfig) Save() error {
	if c.Mutexes == nil {
		c.Mutexes = NewMutexesType()
	}
	c.Mutexes.MuLock()
	defer c.Mutexes.MuUnlock()

	if c.Mapper == nil {
		c.Mapper = NewMapper[*GoBEConfig](&c, c.FilePath)
	}

	c.Mapper.SerializeToFile(c.ConfigFormat)

	return nil
}

func (c *GoBEConfig) Load() error {
	if c.Mutexes == nil {
		c.Mutexes = NewMutexesType()
	}
	c.Mutexes.MuLock()
	defer c.Mutexes.MuUnlock()

	if c.Mapper == nil {
		c.Mapper = NewMapper[*GoBEConfig](&c, c.FilePath)
	}

	_, err := c.Mapper.DeserializeFromFile(c.ConfigFormat)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Failed to load config: %v", err))
	}
	//c = *newCfg

	return nil
}

// ErrConfigFileEmpty indicates that the configuration file is empty.
var ErrConfigFileEmpty = errors.New("config file is empty")

// ErrConfigFileNotFound indicates that the configuration file was not found.
var ErrConfigFileNotFound = errors.New("config file not found")

// ErrConfigFileInvalidFormat indicates that the configuration file has an invalid format.
var ErrConfigFileInvalidFormat = errors.New("config file has invalid format")

// ErrConfigFileReadError indicates that there was an error reading the configuration file.
var ErrConfigFileReadError = errors.New("error reading config file")

// ErrConfigFileWriteError indicates that there was an error writing the configuration file.
var ErrConfigFileWriteError = errors.New("error writing config file")

// ErrConfigInvalid indicates that the configuration is invalid.
var ErrConfigInvalid = errors.New("config is invalid")
