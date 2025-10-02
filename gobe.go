// Package gobe provides the core functionality for the GoBE framework.
package gobe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	crp "github.com/kubex-ecosystem/gobe/factory/security"
	rts "github.com/kubex-ecosystem/gobe/internal/app/router"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	is "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	cm "github.com/kubex-ecosystem/gobe/internal/commons"
	"github.com/kubex-ecosystem/gobe/internal/config"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	"github.com/kubex-ecosystem/gobe/internal/utils"
	l "github.com/kubex-ecosystem/logz"

	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type GoBECertData struct {
	Cert string `json:"cert" yaml:"cert" xml:"cert" csv:"cert" toml:"cert" gorm:"cert"`
	Key  string `json:"key" yaml:"key" xml:"key" csv:"key" toml:"key" gorm:"key"`
}

type GoBE struct {
	InitArgs    gl.InitArgs
	Logger      l.Logger
	environment *is.EnvironmentType

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
	Middlewares map[string]any
	Routes      map[string]map[string]any
}

func NewGoBE(args gl.InitArgs, logger gl.Logger) (ci.IGoBE, error) {
	if logger == nil {
		logger = gl.GetLogger("GoBE")
	}

	chanCtl := make(chan string, 3)
	signamManager := t.NewSignalManager(chanCtl, logger.GetLogger())
	var err error

	args, err = validateIniArgs(args)
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error validating init args: %v", err))
		return nil, err
	}

	gbm := &GoBE{
		InitArgs:  args,
		Logger:    logger.GetLogger(),
		Mutexes:   t.NewMutexesType(),
		Reference: t.NewReference(args.Name).GetReference(),

		SignalManager: signamManager,
		Properties:    make(map[string]any),
		Metadata:      make(map[string]any),
		Middlewares:   make(map[string]any),

		configFile: args.ConfigFile,
		LogFile:    args.LogFile,
		configDir:  filepath.Dir(args.ConfigFile),

		chanCtl:    chanCtl,
		emailQueue: make(chan ci.ContactForm, 20),

		requestWindow:   t.RequestWindow,
		requestLimit:    t.RequestLimit,
		requestsTracers: make(map[string]ci.IRequestsTracer),
	}

	gbm.environment, err = is.NewEnvironment(args.EnvFile, args.IsConfidential, nil)
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error creating environment: %v", err))
	}
	if gbm.environment == nil {
		gl.Log("fatal", fmt.Sprintf("Error creating environment: %v", fmt.Errorf("environment is nil")))
	}

	args.Address = net.JoinHostPort(args.Bind, args.Port)
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

	return gbm, nil
}

func (g *GoBE) GetReference() ci.IReference {
	return g.Reference
}
func (g *GoBE) Environment() is.Environment {
	return g.environment
}

func (g *GoBE) InitializeResources() error {
	gl.Log("notice", "Initializing GoBE...")

	if g.Logger == nil {
		g.Logger = l.GetLogger("GoBE")
	}
	// Initialize the environment
	ctx := context.Background()
	dbService, initResourcesErr := is.InitializeAllServices(ctx, g.environment, g.Logger, g.environment.Getenv("DEBUG") == "true")
	if initResourcesErr != nil {
		return initResourcesErr
	}

	if dbService == nil {
		gl.Log("error", "Database service is nil")
		return errors.New("database service is nil")
	}
	g.Properties["dbService"] = t.NewProperty[is.DBService]("dbService", &dbService, true, nil)

	g.SetDatabaseService(dbService)

	return nil
}
func (g *GoBE) InitializeServer() (ci.IRouter, error) {
	gl.Log("notice", "Initializing server...")

	if !reflect.ValueOf(g.InitArgs.Port).IsValid() {
		gl.Log("warn", "No port specified, using default port 8666")
		g.InitArgs.Port = "8666"
	}
	if !reflect.ValueOf(g.InitArgs.Bind).IsValid() {
		gl.Log("warn", "Binding to all interfaces (default/IPv4)")
		g.InitArgs.Bind = "0.0.0.0"
	}
	if !reflect.ValueOf(g.InitArgs.Address).IsValid() {
		g.InitArgs.Address = net.JoinHostPort(g.InitArgs.Bind, g.InitArgs.Port)
		gl.Log("warn", "No address specified, using default address %s", g.InitArgs.Address)
	}

	if g.configFile == "" {
		var err error
		g.configFile, err = utils.GetDefaultConfigPath()
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error getting default config path: %v", err))
			return nil, err
		}
	}

	gobeminConfig := t.NewGoBEConfig(g.Name, g.configFile, "json", g.InitArgs.Bind, g.InitArgs.Port)
	if _, err := os.Stat(g.configFile); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(g.configFile), 0755); err != nil {
				gl.Log("error", fmt.Sprintf("Error creating directory: %v", err))
				return nil, err
			}
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

	dbServiceT := g.Properties["dbService"].(*t.Property[is.DBService])
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

	gl.Log("debug", fmt.Sprintf("Server started on port %s", g.InitArgs.Port))

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
func (g *GoBE) SetDatabaseService(dbService is.DBService) {
	//g.Mutexes.MuAdd(1)
	//defer g.Mutexes.MuDone()
	g.Properties["dbService"] = t.NewProperty[is.DBService]("dbService", &dbService, true, nil)
}
func (g *GoBE) GetDatabaseService() is.DBService {
	//g.Mutexes.MuRLock()
	//defer g.Mutexes.MuRUnlock()
	if dbT, ok := g.Properties["dbService"].(*t.Property[is.DBService]); ok {
		db := dbT.GetValue()
		return db
	} else if dbT, ok := g.Properties["dbService"].(*t.Property[is.DBService]); ok {
		dbS := dbT.GetValue()
		return dbS
	} else if dbT, ok := g.Properties["dbService"].(*t.Property[is.DBService]); ok {
		dbS := dbT.GetValue()
		return dbS
	} else {
		gl.Log("error", "Database service is nil")
		return nil
	}
}
func (g *GoBE) LogsGoBE() (*io.OffsetWriter, error) {
	//g.Mutexes.MuRLock()
	//defer g.Mutexes.MuRUnlock()
	logger := g.Logger
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
	gl.Log("error", "Logger is nil")
	return nil, errors.New("logger is nil")
}
func (g *GoBE) GetProperty(name string) (any, reflect.Type, error) {
	if prop, ok := g.Properties[name]; ok {
		return prop, reflect.TypeOf(prop), nil
	}
	return nil, nil, fmt.Errorf("property %q not found", name)
}

// Auxiliary functions for GoBE

func chanRoutineCtl[T any](v ci.IChannelCtl[T], chCtl chan string, ch chan T) {
	select {
	case cmd := <-chCtl:
		switch cmd {
		case "stop":
			gl.Log("info", "Received stop command for:", v.GetName(), "ID:", v.GetID().String())
			// If we receive a stop command, we need to close the channels.
			if ch != nil {
				close(ch)
				ch = nil
			}
			if chCtl != nil {
				close(chCtl)
				chCtl = nil
			}
		case "reload":
			gl.Log("info", "Received reload command for:", v.GetName(), "ID:", v.GetID().String())
			// If we receive a reload command, we need to close the channels and create new ones.
			if ch != nil {
				close(ch)
				ch = nil
			}
			if chCtl != nil {
				close(chCtl)
				chCtl = nil
			}
		default:
			gl.Log("warn", "Received unknown command for:", v.GetName(), "ID:", v.GetID().String(), "Command:", cmd)
		}
	case data := <-ch:
		gl.Log("debug", "Received data for:", v.GetName(), "ID:", v.GetID().String(), "Data:", data)
		// Process the data received from the main channel.
		action := func() string {
			if reflect.ValueOf(data).Kind() == reflect.Ptr && !reflect.ValueOf(data).IsNil() {
				return fmt.Sprintf("%v", reflect.ValueOf(data).Elem().Interface())
			}
			return fmt.Sprintf("%v", data)
		}
		v.ProcessData(action())
	case <-time.After(5 * time.Second):
		// Timeout after 5 seconds to check if we need to exit the routine.
	}
}

func listenSystemSignals[T any](gbm *GoBE) {
	chCtl := gbm.GetChanCtl()
	if chCtl == nil {
		gl.Log("error", "Control channel is nil, cannot listen for system signals")
		return
	}
	signamManager := gbm.SignalManager
	if signamManager == nil {
		gl.Log("error", "Signal manager is nil, cannot listen for system signals")
		return
	}
	if reflect.ValueOf(signamManager).IsNil() {
		gl.Log("error", "Signal manager is nil, cannot listen for system signals")
		return
	}
	go func(chan string, ci.ISignalManager[chan string], *GoBE) {
		signamManager.ListenForSignals()
		gl.Log("debug", "Listening for signals...")
		for msg := range chCtl {
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
	}(chCtl, signamManager, gbm)
}
func validateIniArgs(args gl.InitArgs) (gl.InitArgs, error) {
	if args.Debug {
		gl.SetDebugMode(args.Debug)
	}
	if args.ReleaseMode {
		os.Setenv("GIN_MODE", "release")
		gin.SetMode(gin.ReleaseMode)
	}

	defaultDir := filepath.Dir(os.ExpandEnv(cm.DefaultGodoBaseConfigPath))
	if _, err := os.Stat(defaultDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(defaultDir, 0755); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error creating default config directory: %v", err))
			}
		}
	}

	if args.ConfigFile == "" {
		args.ConfigFile = os.ExpandEnv(cm.DefaultGoBEConfigPath)
		if _, err := os.Stat(args.ConfigFile); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(args.ConfigFile), 0755); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error creating directory: %v", err))
				}
				if err := os.WriteFile(args.ConfigFile, []byte(""), 0644); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error creating config file: %v", err))
				}
			}
		}
	}
	if args.LogFile == "" {
		args.LogFile = filepath.Join(defaultDir, "request_tracer.json")
		if _, err := os.Stat(args.LogFile); err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(args.LogFile), 0755); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error creating directory: %v", err))
				}
				if err := os.WriteFile(args.LogFile, []byte(""), 0644); err != nil {
					gl.Log("fatal", fmt.Sprintf("Error creating log file: %v", err))
				}
			}
		}
	}

	if args.PubCertKeyPath == "" {
		args.PubCertKeyPath = os.ExpandEnv(cm.DefaultGoBECertPath)
	}
	if args.PubKeyPath == "" {
		args.PubKeyPath = os.ExpandEnv(cm.DefaultGoBEKeyPath)
	}

	initArgs := gl.InitArgs{
		ConfigFile:     args.ConfigFile,
		IsConfidential: args.IsConfidential,
		Port:           args.Port,
		Bind:           args.Bind,
		Address:        net.JoinHostPort(args.Bind, args.Port),
		PubCertKeyPath: args.PubCertKeyPath,
		EnvFile:        args.EnvFile,
		Name:           args.Name,
		Debug:          args.Debug,
		ReleaseMode:    args.ReleaseMode,
		LogFile:        args.LogFile,
		PubKeyPath:     args.PubKeyPath,
		Pwd:            args.Pwd,
	}

	if err := config.BootstrapMainConfig(&initArgs); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to bootstrap config file: %v", err))
	}

	// We don't return a pointer, this allows us to
	// work with more than one instance of GoBE with different
	// purposes.
	return initArgs, nil
}
