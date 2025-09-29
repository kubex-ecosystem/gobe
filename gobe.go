// Package gobe provides the core functionality for the GoBE framework.
package gobe

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	gdbf "github.com/kubex-ecosystem/gdbase/factory"
	"github.com/kubex-ecosystem/gdbase/services"
	"github.com/kubex-ecosystem/gdbase/types"
	crp "github.com/kubex-ecosystem/gobe/factory/security"
	rts "github.com/kubex-ecosystem/gobe/internal/app/router"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	is "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	cm "github.com/kubex-ecosystem/gobe/internal/commons"
	cfg "github.com/kubex-ecosystem/gobe/internal/config"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	"github.com/kubex-ecosystem/gobe/internal/utils"
	l "github.com/kubex-ecosystem/logz"

	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
)

type GoBECertData struct {
	Cert string `json:"cert" yaml:"cert" xml:"cert" csv:"cert" toml:"cert" gorm:"cert"`
	Key  string `json:"key" yaml:"key" xml:"key" csv:"key" toml:"key" gorm:"key"`
}

type GoBE struct {
	InitArgs    ci.InitArgs
	Logger      l.Logger
	environment ci.IEnvironment

	*t.Mutexes
	*t.Reference

	SignalManager ci.ISignalManager[chan string]

	requestWindow   time.Duration
	requestLimit    int
	requestsTracers map[string]ci.IRequestsTracer

	// Configuration paths

	configDir  string
	configFile string
	LogFile    string

	chanCtl    chan string
	emailQueue chan ci.ContactForm

	Properties  map[string]any
	Metadata    map[string]any
	Routes      map[string]map[string]any
	Middlewares map[string]any
}

func NewGoBE(name, port, bind, logFile, configFile string, isConfidential bool, logger l.Logger, debug, releaseMode bool) (ci.IGoBE, error) {
	if logger == nil {
		logger = l.GetLogger("GoBE")
	}
	if debug {
		gl.SetDebug(debug)
	}
	if releaseMode {
		os.Setenv("GIN_MODE", "release")
		gin.SetMode(gin.ReleaseMode)
	}

	chanCtl := make(chan string, 3)
	signamManager := t.NewSignalManager(chanCtl, logger)

	defaultDir := filepath.Dir(os.ExpandEnv(cm.DefaultGodoBaseConfigPath))
	if _, err := os.Stat(defaultDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(defaultDir, 0755); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error creating default config directory: %v", err))
			}
		}
	}

	if configFile == "" {
		configFile = os.ExpandEnv(cm.DefaultGoBEConfigPath)
		if _, err := os.Stat(configFile); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error creating directory: %v", err))
				}
				if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error creating config file: %v", err))
				}
			}
		}
	}
	if logFile == "" {
		logFile = filepath.Join(defaultDir, "request_tracer.json")
		if _, err := os.Stat(logFile); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error creating directory: %v", err))
				}
				if err := os.WriteFile(logFile, []byte(""), 0644); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error creating log file: %v", err))
				}
			}
		}
	}

	initArgs := ci.InitArgs{
		ConfigFile:     configFile,
		IsConfidential: isConfidential,
		Port:           port,
		Bind:           bind,
		Address:        net.JoinHostPort(bind, port),
		PubCertKeyPath: os.ExpandEnv(cm.DefaultGoBECertPath),
		PubKeyPath:     os.ExpandEnv(cm.DefaultGoBEKeyPath),
		Pwd:            "",
	}

	gbm := &GoBE{
		InitArgs: initArgs,

		Logger:    logger,
		Mutexes:   t.NewMutexesType(),
		Reference: t.NewReference(name).GetReference(),

		SignalManager: signamManager,
		Properties:    make(map[string]any),
		Metadata:      make(map[string]any),
		Middlewares:   make(map[string]any),

		configFile: configFile,
		LogFile:    logFile,
		configDir:  filepath.Dir(configFile),

		chanCtl:    chanCtl,
		emailQueue: make(chan ci.ContactForm, 20),

		requestWindow:   t.RequestWindow,
		requestLimit:    t.RequestLimit,
		requestsTracers: make(map[string]ci.IRequestsTracer),
	}

	if err := cfg.BootstrapMainConfig(gbm.InitArgs.ConfigFile, &gbm.InitArgs); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to bootstrap config file: %v", err))
	}

	var err error
	gbm.environment, err = t.NewEnvironment(configFile, isConfidential, logger)
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error creating environment: %v", err))
	}
	if gbm.environment == nil {
		gl.Log("fatal", fmt.Sprintf("Error creating environment: %v", fmt.Errorf("environment is nil")))
	}

	gbm.Properties["env"] = t.NewProperty("env", &gbm.environment, true, nil)
	gbm.Properties["port"] = t.NewProperty("port", &port, true, nil)
	gbm.Properties["bind"] = t.NewProperty("bind", &bind, true, nil)
	address := net.JoinHostPort(bind, port)
	gbm.Properties["address"] = t.NewProperty("address", &address, true, nil)

	pubCertKeyPath := gbm.environment.Getenv("CERT_FILE_PATH")
	if pubCertKeyPath == "" {
		pubCertKeyPath = os.ExpandEnv(cm.DefaultGoBECertPath)
	}
	pubKeyPath := gbm.environment.Getenv("KEY_FILE_PATH")
	if pubKeyPath == "" {
		pubKeyPath = os.ExpandEnv(cm.DefaultGoBEKeyPath)
	}

	var pwd string

	pwd = gbm.environment.Getenv("KEYRING_PASS")
	if pwd == "" {
		var pwdErr error
		// THIS SECRET WILL BE PASSED AS A PASSWORD TO ENCRYPT THE PRIVATE KEY
		// (jwt_secret is just a temporary fix) AND IT WILL BE STORED IN THE KEYRING
		// FOR FUTURE USE. TO DECRYPT THE PRIVATE KEY, THE SAME PASSWORD MUST BE USED!
		pwd, pwdErr = crt.GetOrGenPasswordKeyringPass("jwt_secret")
		if pwdErr != nil {
			gl.Log("fatal", fmt.Sprintf("Error reading keyring password: %v", pwdErr))
		}
	}

	gbm.Properties["initArgs"] = t.NewProperty("initArgs", &initArgs, false, nil)

	crptService := crp.NewCryptoService()
	crtService := crt.NewCertService(pubKeyPath, pubCertKeyPath)
	if _, err := os.Stat(pubKeyPath); err != nil {
		decodedPwd, decodeErr := crptService.DecodeBase64(pwd)
		if decodeErr != nil {
			gl.Log("error", fmt.Sprintf("Error decoding keyring password: %v", decodeErr))
			return nil, decodeErr
		}
		certBytes, keyBytes, err := crtService.GenerateCertificate(pubCertKeyPath, pubKeyPath, decodedPwd)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error generating certificate: %v", err))
			return nil, err
		}

		var keyEncodedBytes, certEncodedBytes []byte
		var keyString, certString string

		isEncoded := crptService.IsBase64String(string(bytes.TrimSpace(certBytes)))
		if !isEncoded {
			certEncodedBytes = bytes.TrimSpace([]byte(crptService.EncodeBase64(certBytes)))
		}
		isEncoded = crptService.IsBase64String(string(bytes.TrimSpace(keyBytes)))
		if !isEncoded {
			keyEncodedBytes = bytes.TrimSpace([]byte(crptService.EncodeBase64(keyBytes)))
		}
		certObj := GoBECertData{Cert: certString, Key: keyString}

		gl.Log("info", fmt.Sprintf("Certificate generated at %s", pubCertKeyPath))
		gl.Log("info", fmt.Sprintf("Private key generated at %s", pubKeyPath))
		gl.Log("debug", fmt.Sprintf("Certificate: %s", certString))
		gl.Log("debug", fmt.Sprintf("Private key: %s", keyString))
		certObj.Cert = string(certEncodedBytes)
		certObj.Key = string(keyEncodedBytes)
		gbm.Properties["cert"] = t.NewProperty("cert", &certObj.Cert, true, nil)

		mapper := t.NewMapper(&certObj, filepath.Join(gbm.configDir, "cert.json"))
		mapper.SerializeToFile("json")
		gl.Log("debug", fmt.Sprintf("Certificate generated at %s", pubCertKeyPath))
		gbm.Properties["privKey"] = t.NewProperty("privKey", &keyEncodedBytes, true, nil)
	} else {
		certObj := &GoBECertData{}
		mapper := t.NewMapper(&certObj, filepath.Join(gbm.configDir, "cert.json"))
		if _, err := mapper.DeserializeFromFile("json"); err != nil {
			gl.Log("error", fmt.Sprintf("Error reading certificate: %v", err))
			return nil, err
		}
		key := certObj.Key
		gbm.Properties["privKey"] = t.NewProperty("privKey", &key, true, nil)
	}
	if _, err := os.Stat(pubKeyPath); err != nil {
		gl.Log("error", fmt.Sprintf("Error generating certificate: %v", err))
		return nil, err
	}

	gbm.Properties["certPath"] = t.NewProperty("certPath", &pubCertKeyPath, true, nil)
	gbm.Properties["keyPath"] = t.NewProperty("keyPath", &pubKeyPath, true, nil)
	gbm.Properties["certService"] = crtService

	// Start listening for signals since the beginning, so we can handle them
	// gracefully even if the server is not started yet.

	go func(chan string, ci.ISignalManager[chan string], *GoBE) {
		signamManager.ListenForSignals()
		gl.Log("debug", "Listening for signals...")
		for msg := range chanCtl {
			switch msg {
			case "reload":
				gl.Log("info", "Received reload signal, reloading server...")
				gbm.StopGoBE()
				gbm.StartGoBE()
			default:
				gl.Log("info", "Received stop signal, stopping server...")
				gbm.StopGoBE()
				gl.Log("info", "Server stopped gracefully")
				return
			}
		}
	}(chanCtl, signamManager, gbm)

	return gbm, nil
}

func (g *GoBE) GetReference() ci.IReference {
	return g.Reference
}
func (g *GoBE) Environment() ci.IEnvironment {
	return g.environment
}

func (g *GoBE) InitializeResources() error {
	gl.Log("notice", "Initializing GoBE...")

	if g.Logger == nil {
		g.Logger = l.GetLogger("GoBE")
	}
	envT := g.Properties["env"].(*t.Property[ci.IEnvironment])
	env := envT.GetValue()
	var err error
	if env == nil {
		env, err = t.NewEnvironment(g.configFile, false, g.Logger)
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error creating environment: %v", err))
			return err
		}
		g.Properties["env"] = t.NewProperty("env", &env, true, nil)
	}

	dbService, initResourcesErr := is.InitializeAllServices(g.environment, g.Logger, g.environment.Getenv("DEBUG") == "true")
	if initResourcesErr != nil {
		return initResourcesErr
	}

	if dbService == nil {
		gl.Log("error", "Database service is nil")
		return errors.New("database service is nil")
	}
	g.Properties["dbService"] = t.NewProperty("dbService", &dbService, true, nil)

	g.SetDatabaseService(dbService)

	return nil
}
func (g *GoBE) InitializeServer() (ci.IRouter, error) {
	gl.Log("notice", "Initializing server...")

	portT := g.Properties["port"].(*t.Property[string])
	port := portT.GetValue()
	bindT := g.Properties["bind"].(*t.Property[string])
	bind := bindT.GetValue()
	if !reflect.ValueOf(port).IsValid() {
		gl.Log("warn", "No port specified, using default port 8666")
		port = "8666"
		portT.SetValue(&port)
	}
	if !reflect.ValueOf(bind).IsValid() {
		gl.Log("warn", "Binding to all interfaces (default/IPv4)")
		bind = "0.0.0.0"
		bindT.SetValue(&bind)
	}
	addressT := g.Properties["address"].(*t.Property[string])
	address := addressT.GetValue()
	if !reflect.ValueOf(address).IsValid() {
		address = net.JoinHostPort(bind, port)
		gl.Log("warn", "No address specified, using default address %s", address)
		addressT.SetValue(&address)
	}

	if g.configFile == "" {
		var err error
		g.configFile, err = utils.GetDefaultConfigPath()
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error getting default config path: %v", err))
			return nil, err
		}
	}

	gobeminConfig := t.NewGoBEConfig(g.Name, g.configFile, "json", bind, port)
	if _, err := os.Stat(g.configFile); err != nil {
		if os.IsNotExist(err) {
			// if err := ut.EnsureDir(filepath.Dir(g.configFile), 0644, []string{}); err != nil {
			// 	gl.Log("error", fmt.Sprintf("Error creating directory: %v", err))
			// 	return nil, err
			// }
			if err := os.MkdirAll(filepath.Dir(g.configFile), 0755); err != nil {
				gl.Log("error", fmt.Sprintf("Error creating directory: %v", err))
				return nil, err
			}
			// if err := ut.EnsureFile(g.configFile, 0644, []string{}); err != nil {
			// 	gl.Log("error", fmt.Sprintf("Error creating config file: %v", err))
			// 	return nil, err
			// }
			if err := os.WriteFile(g.configFile, []byte(""), 0644); err != nil {
				gl.Log("error", fmt.Sprintf("Error creating config file: %v", err))
				return nil, err
			}
			mapper := t.NewMapper(gobeminConfig, g.configFile)
			mapper.SerializeToFile("json")
		} else {
			gl.Log("error", fmt.Sprintf("Error reading config file: %v", err))
			return nil, err
		}
	}
	if gobeminConfig == nil {
		gl.Log("error", "Failed to create config file")
		return nil, fmt.Errorf("failed to create config file")
	}

	if gobeminConfig.GetJWTSecretKey() == "" {
		jwtSecret, jwtSecretErr := crt.GetOrGenPasswordKeyringPass("jwt_secret")
		if jwtSecretErr != nil {
			gl.Log("fatal", fmt.Sprintf("Error reading JWT secret key: %v", jwtSecretErr))
		}
		if jwtSecret == "" {
			gl.Log("error", "JWT secret key is empty")
			return nil, fmt.Errorf("jwt secret key is empty")
		}
		gobeminConfig.SetJWTSecretKey(jwtSecret)
	}

	rateLimitLimit := gobeminConfig.RateLimitLimit
	rateLimitBurst := gobeminConfig.RateLimitBurst
	requestWindow := gobeminConfig.RequestWindow
	if rateLimitLimit <= 0 {
		rateLimitLimit = cm.DefaultRateLimitLimit
	}
	if rateLimitBurst <= 0 {
		rateLimitBurst = cm.DefaultRateLimitBurst
	}
	if requestWindow <= 0 {
		requestWindow = time.Duration(cm.DefaultRequestWindow) * time.Millisecond
	}
	gobeminConfig.SetRateLimitLimit(rateLimitLimit)
	gobeminConfig.SetRateLimitBurst(rateLimitBurst)
	gobeminConfig.SetRequestWindow(requestWindow)

	dbServiceT := g.Properties["dbService"].(*t.Property[gdbf.DBService])
	dbService := dbServiceT.GetValue()
	if dbService == nil {
		gl.Log("error", "Database service is nil")
		return nil, errors.New("database service is nil")
	}

	_, kubexErr := crt.GetOrGenPasswordKeyringPass(cm.KeyringService)
	if kubexErr != nil {
		gl.Log("error", fmt.Sprintf("Error reading kubex keyring password: %v", kubexErr))
		return nil, kubexErr
	}

	//gobeminConfig.Set

	router, err := rts.NewRouter(gobeminConfig, dbService, g.Logger, g.environment.Getenv("DEBUG") == "true")
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error initializing router: %v", err))
		return nil, err
	}

	g.Properties["router"] = t.NewProperty("router", &router, true, nil)
	if router == nil {
		gl.Log("error", "Router is nil")
		return nil, errors.New("router is nil")
	}

	return router, nil
}
func (g *GoBE) GetLogger() l.Logger {
	return g.Logger
}
func (g *GoBE) StartGoBE() {
	gl.Log("info", "Starting server...")

	if err := g.InitializeResources(); err != nil {
		gl.Log("fatal", fmt.Sprintf("Error initializing GoBE: %v", err))
		return
	}

	router, err := g.InitializeServer()
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error initializing server: %v", err))
		return
	}
	if router == nil {
		gl.Log("fatal", "Router is nil")
		return
	}
	gl.Log("debug", "Loading request tracers...")
	g.Mutexes.MuAdd(1)
	go func(g *GoBE) {
		if g == nil {
			gl.Log("fatal", "GoBE instance is nil")
			// g.Mutexes.MuDone()
			return
		}
		defer g.Mutexes.MuDone()
		var err error
		//initArgs := g.Properties["initArgs"].(*t.Property[interfaces.InitArgs]).GetValue()
		// requestsTracers := t.NewRequestTracers(g)
		var requestsTracers ci.IRequestTracers

		requestsTracers, err = t.LoadRequestsTracerFromFile(g)
		if requestsTracers == nil {
			gl.Log("warn", "No persisted request tracers found, creating a new one")
			requestsTracers = t.NewRequestTracers(g)
		}
		g.requestsTracers = requestsTracers.GetRequestTracers()

		if err != nil {
			gl.Log("error", "Error loading request tracers: %v", err.Error())
		}
	}(g)
	gl.Log("notice", "Waiting for persisted request tracers to load...")
	g.Mutexes.MuWait()

	gl.Log("debug", fmt.Sprintf("Server started on port %s", g.Properties["port"].(*t.Property[string]).GetValue()))

	if err := router.Start(); err != nil {
		gl.Log("fatal", "Error starting server: %v", err.Error())
	}
}
func (g *GoBE) StopGoBE() {
	gl.Log("info", "Stopping server...")

	g.Mutexes.MuAdd(1)
	defer g.Mutexes.MuDone()

	routerT := g.Properties["router"].(*t.Property[ci.IRouter])
	router := routerT.GetValue()
	if router == nil {
		gl.Log("error", "Router is nil")
		return
	}

	router.ShutdownServerGracefully()
}
func (g *GoBE) GetChanCtl() chan string {
	//g.Mutexes.MuRLock()
	//defer g.Mutexes.MuRUnlock()
	return g.chanCtl
}

func (g *GoBE) GetLogFilePath() string {
	return g.LogFile
}
func (g *GoBE) GetConfigFilePath() string {
	return g.configFile
}
func (g *GoBE) SetDatabaseService(dbService gdbf.DBService) {
	//g.Mutexes.MuAdd(1)
	//defer g.Mutexes.MuDone()
	g.Properties["dbService"] = t.NewProperty("dbService", &dbService, true, nil)
}
func (g *GoBE) GetDatabaseService() gdbf.DBService {
	//g.Mutexes.MuRLock()
	//defer g.Mutexes.MuRUnlock()
	if dbT, ok := g.Properties["dbService"].(*t.Property[gdbf.DBService]); ok {
		return dbT.GetValue()
	} else if dbT, ok := g.Properties["dbService"].(*t.Property[services.IDBService]); ok {
		return dbT.GetValue()
	} else if dbT, ok := g.Properties["dbService"].(*t.Property[types.DBService]); ok {
		return dbT.GetValue()
	} else {
		gl.Log("error", "Database service is nil")
		return nil
	}
}
func (g *GoBE) LogsGoBE() (*io.OffsetWriter, error) {
	//g.Mutexes.MuRLock()
	//defer g.Mutexes.MuRUnlock()
	if loggerProp, ok := g.Properties["logger"].(*t.Property[l.Logger]); ok {
		if loggerProp == nil {
			gl.Log("error", "Logger is nil")
			return nil, errors.New("logger is nil")
		}
		gl.Log("info", "Retrieving logs...")
		logger := loggerProp.GetValue()
		if logger == nil {
			gl.Log("error", "Logger is nil")
			return nil, errors.New("logger is nil")
		}
		logsWriterInt := logger.GetWriter()
		if logsWriterInt == nil {
			gl.Log("error", "Logs writer is nil")
			return nil, errors.New("logs writer is nil")
		}
		logsWriter, ok := logsWriterInt.(io.Writer)
		if !ok {
			gl.Log("error", "Logs writer is not an io.Writer")
			return nil, errors.New("logs writer is not an io.Writer")
		}
		logsWriter.Write([]byte("Retrieving logs...\n"))
		if offsetWriter, ok := logsWriter.(*io.OffsetWriter); ok {
			return offsetWriter, nil
		}
	}
	gl.Log("error", "Logger is nil")
	return nil, errors.New("logger is nil")
}
