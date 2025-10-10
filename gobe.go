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
	"time"

	"github.com/gin-gonic/gin"
	crp "github.com/kubex-ecosystem/gobe/factory/security"
	rts "github.com/kubex-ecosystem/gobe/internal/app/router"
	crt "github.com/kubex-ecosystem/gobe/internal/app/security/certificates"
	cf "github.com/kubex-ecosystem/gobe/internal/bootstrap"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	l "github.com/kubex-ecosystem/logz"

	"github.com/kubex-ecosystem/gobe/internal/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	is "github.com/kubex-ecosystem/gdbase/factory"

	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type GoBECertData struct {
	Cert string `json:"cert" yaml:"cert" xml:"cert" csv:"cert" toml:"cert" gorm:"cert"`
	Key  string `json:"key" yaml:"key" xml:"key" csv:"key" toml:"key" gorm:"key"`
}

type GoBE struct {
	InitArgs    *gl.InitArgs
	Logger      l.Logger
	environment *is.EnvironmentType

	*t.Mutexes
	*t.Reference

	router *rts.Router

	SignalManager ci.ISignalManager[chan string]

	requestWindow   time.Duration
	requestLimit    int
	requestsTracers map[string]ci.IRequestsTracer

	dbName string

	// Configuration paths

	configDir    string
	configFile   string
	configDBFile string
	LogFile      string

	privKey  string
	certPath string
	keyPath  string

	chanCtl    chan string
	emailQueue chan ci.ContactForm

	certService *crt.CertService
	dbService   *is.DBServiceImpl

	Metadata    map[string]any
	Middlewares map[string]any
	Routes      map[string]map[string]any
}

func NewGoBE(args *gl.InitArgs, logger gl.Logger) (ci.IGoBE, error) {
	if logger == nil {
		logger = gl.GetLogger("GoBE")
	}

	chanCtl := make(chan string, 3)
	signamManager := t.NewSignalManager(chanCtl, logger.GetLogger())

	cfg, err := validateInitArgs(args)
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error validating init args: %v", err))
		return nil, err
	}
	if cfg == nil {
		gl.Log("fatal", "Main config is nil")
		return nil, fmt.Errorf("main config is nil")
	}

	dbName := gl.GetEnvOrDefault("GOBE_DB_NAME", "kubex_db")

	gbm := &GoBE{
		InitArgs:  args,
		Logger:    logger.GetLogger(),
		Mutexes:   t.NewMutexesType(),
		Reference: t.NewReference(args.Name).GetReference(),

		SignalManager: signamManager,

		Metadata:    make(map[string]any),
		Middlewares: make(map[string]any),

		dbName: dbName,

		configFile:   args.ConfigFile,
		configDBFile: args.ConfigDBFile,
		LogFile:      args.LogFile,
		configDir:    filepath.Dir(args.ConfigFile),

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
		pubCertKeyPath = os.ExpandEnv(gl.DefaultGoBECertPath)
	}
	pubKeyPath := gbm.environment.Getenv("KEY_FILE_PATH")
	if pubKeyPath == "" {
		pubKeyPath = os.ExpandEnv(gl.DefaultGoBEKeyPath)
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
		mapper := t.NewMapper(&certObj, filepath.Join(gbm.configDir, "cert.json"))
		mapper.SerializeToFile("json")
		gl.Log("debug", fmt.Sprintf("Certificate generated at %s", pubCertKeyPath))
	} else {
		certObj := &GoBECertData{}
		mapper := t.NewMapper(&certObj, filepath.Join(gbm.configDir, "cert.json"))
		if _, err := mapper.DeserializeFromFile("json"); err != nil {
			gl.Log("error", fmt.Sprintf("Error reading certificate: %v", err))
			return nil, err
		}
	}
	if _, err := os.Stat(pubKeyPath); err != nil {
		gl.Log("error", fmt.Sprintf("Error generating certificate: %v", err))
		return nil, err
	}

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
	dbService, initResourcesErr := g.initializeAllServices()
	if initResourcesErr != nil {
		return initResourcesErr
	}

	if dbService == nil {
		gl.Log("error", "Database service is nil")
		return errors.New("database service is nil")
	}
	g.dbService = dbService

	g.SetDatabaseService(dbService)

	return nil
}
func (g *GoBE) InitializeServer() (ci.IRouter, error) {
	gl.Log("notice", "Initializing server...")

	if g.InitArgs.Port == "" {
		gl.Log("warn", "No port specified, using default port 8666")
		g.InitArgs.Port = "8666"
	}
	if g.InitArgs.Bind == "" {
		gl.Log("warn", "Binding to all interfaces (default/IPv4)")
		g.InitArgs.Bind = "0.0.0.0"
	}
	if g.InitArgs.Address == "" {
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
		rateLimitLimit = gl.DefaultRateLimitLimit
	}
	if rateLimitBurst <= 0 {
		rateLimitBurst = gl.DefaultRateLimitBurst
	}
	if requestWindow <= 0 {
		requestWindow = time.Duration(gl.DefaultRequestWindow) * time.Millisecond
	}
	gobeminConfig.SetRateLimitLimit(rateLimitLimit)
	gobeminConfig.SetRateLimitBurst(rateLimitBurst)
	gobeminConfig.SetRequestWindow(requestWindow)

	if g.dbService == nil {
		gl.Log("error", "Database service is nil")
		return nil, errors.New("database service is nil")
	}

	// Ensure database is fully ready before proceeding
	ctx := context.Background()
	if !g.dbService.IsReady(ctx) {
		gl.Log("warn", "⏳ Database service is initializing, waiting...")
		// Wait a moment for DB to be ready
		time.Sleep(1 * time.Second)
		if !g.dbService.IsReady(ctx) {
			gl.Log("error", "❌ Database service failed to become ready")
			return nil, errors.New("database service not ready")
		}
	}
	gl.Log("info", "✅ Database service is ready")

	_, kubexErr := crt.GetOrGenPasswordKeyringPass(gl.KeyringService)
	if kubexErr != nil {
		gl.Log("error", fmt.Sprintf("Error reading kubex keyring password: %v", kubexErr))
		return nil, kubexErr
	}

	router, err := rts.NewRouter(gobeminConfig, g.dbService, g.InitArgs, g.Logger, g.environment.Getenv("DEBUG") == "true")
	if err != nil {
		gl.Log("error", fmt.Sprintf("Error initializing router: %v", err))
		return nil, err
	}
	g.router = router.(*rts.Router)
	if g.router == nil {
		gl.Log("error", "Router is nil")
		return nil, errors.New("router is nil")
	}
	return g.router, nil
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

	gl.Log("debug", "Initializing server...")
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

	// Register routes and middlewares
	if err := router.InitializeResources(); err != nil {
		gl.Log("fatal", fmt.Sprintf("Error initializing router resources: %v", err))
		return
	}

	gl.Log("debug", fmt.Sprintf("Server started on port %s", g.InitArgs.Port))

	if err := router.Start(); err != nil {
		gl.Log("fatal", "Error starting server: %v", err.Error())
	}
}
func (g *GoBE) StopGoBE() {
	gl.Log("info", "Stopping server...")

	g.Mutexes.MuAdd(1)
	defer g.Mutexes.MuDone()

	if g.router == nil {
		gl.Log("error", "Router is nil")
		return
	}

	g.router.ShutdownServerGracefully()
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
	g.dbService = dbService.(*is.DBServiceImpl)
}
func (g *GoBE) GetDatabaseService() is.DBService {
	//g.Mutexes.MuRLock()
	//defer g.Mutexes.MuRUnlock()
	return g.dbService
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

func validateInitArgs(args *gl.InitArgs) (*cf.Config, error) {
	if args == nil {
		args = &gl.InitArgs{}
	}

	if args.Debug {
		gl.SetDebugMode(args.Debug)
	}
	if args.ReleaseMode {
		os.Setenv("GIN_MODE", gl.GetEnvOrDefault("GIN_MODE", "release"))
		gin.SetMode(gin.ReleaseMode)
	}

	// Ensure default config directory exists
	kubexDefaultDir := os.ExpandEnv(gl.DefaultKubexConfigDir)
	if _, err := os.Stat(kubexDefaultDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(kubexDefaultDir, 0755); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error creating default kubex config directory: %v", err))
			}
		}
	}

	if args.ConfigDBFile == "" {
		args.ConfigDBFile = filepath.Join(kubexDefaultDir, "gdbase", "config", "config.json")
	}

	if args.PubCertKeyPath == "" {
		args.PubCertKeyPath = os.ExpandEnv(gl.DefaultGoBECertPath)
	}
	if args.PubKeyPath == "" {
		args.PubKeyPath = os.ExpandEnv(gl.DefaultGoBEKeyPath)
	}

	cfg, err := cf.BootstrapMainConfig[*cf.Config](args)
	if err != nil {
		gl.Log("fatal", fmt.Sprintf("Error loading main config: %v", err))
		return nil, err
	}
	if cfg == nil {
		gl.Log("fatal", "Main config is nil")
		return nil, fmt.Errorf("main config is nil")
	}

	return cfg, nil
}

// InitializeAllServices inicializa todos os serviços (Docker + Database) com context
func (g *GoBE) initializeAllServices() (*is.DBServiceImpl, error) {
	ctx := context.Background()

	// 🎯 NOVO SISTEMA: Usar DockerStackProvider com migrations programáticas
	gl.Log("info", "🚀 Initializing services with new DockerStackProvider...")

	// 1. Setup database config
	dbConfig, err := g.setupDatabase()
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar DBConfig: %v", err))
		return nil, fmt.Errorf("❌ Erro ao inicializar DBConfig: %w", err)
	}

	// 2. Initialize Docker service with existing DBConfig (legacy flow)
	dockerService, err := is.NewDockerService(dbConfig, g.Logger)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao criar DockerService: %v", err))
		return nil, fmt.Errorf("❌ Erro ao criar DockerService: %w", err)
	}

	// 3. Initialize containers (this creates/starts containers if needed)
	if err := dockerService.Initialize(); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar Docker containers: %v", err))
		return nil, fmt.Errorf("❌ Erro ao inicializar Docker containers: %w", err)
	}

	// 4. 🎯 NOVO: Run migrations programmatically using existing DBConfig
	pgConfig := dbConfig.Databases["kubex_db"]
	if pgConfig != nil && pgConfig.Enabled {
		gl.Log("info", "🎯 Running PostgreSQL migrations programmatically...")

		var port int
		switch p := pgConfig.Port.(type) {
		case int:
			port = p
		case string:
			fmt.Sscanf(p, "%d", &port)
		default:
			port = 5432
		}

		// pragma: allowlist nextline secret
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", // pragma: allowlist secret
			pgConfig.Username, pgConfig.Password, pgConfig.Host, port, pgConfig.Name) // pragma: allowlist secret

		migrationMgr := is.NewMigrationManager(dsn, g.Logger)
		if err := migrationMgr.WaitForPostgres(ctx, 30*time.Second); err != nil {
			gl.Log("error", fmt.Sprintf("❌ PostgreSQL não está pronto: %v", err))
			return nil, fmt.Errorf("❌ PostgreSQL não está pronto: %w", err)
		}

		results, err := migrationMgr.RunMigrations(ctx)
		if err != nil {
			gl.Log("warn", fmt.Sprintf("⚠️ Erro parcial nas migrations: %v", err))
		} else {
			totalSuccess := 0
			for _, r := range results {
				totalSuccess += r.SuccessfulStmts
			}
			gl.Log("info", fmt.Sprintf("✅ Migrations executadas com sucesso! (%d statements)", totalSuccess))
		}
	}

	// 5. Create and initialize Database Service
	dbService, err := is.NewDatabaseService(ctx, dbConfig, g.Logger)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar DatabapostgresqlseService: %v", err))
		return nil, fmt.Errorf("❌ Erro ao inicializar DatabaseService: %w", err)
	}

	// 6. Initialize Database Service connections
	if err := dbService.Initialize(ctx); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao conectar ao banco: %v", err))
		return nil, fmt.Errorf("❌ Erro ao conectar ao banco: %w", err)
	}

	gl.Log("info", "✅ All services initialized successfully!")
	return dbService, nil
}

func (g *GoBE) setupDatabase() (*is.DBConfigImpl, error) {
	if _, err := os.Stat(g.configDBFile); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(g.configDBFile), 0755); err != nil {
			gl.Log("error", fmt.Sprintf("❌ Erro ao criar o diretório do arquivo de configuração do banco de dados: %v", err))
			return nil, fmt.Errorf("❌ Erro ao criar o diretório do arquivo de configuração do banco de dados: %v", err)
		}
		if err := os.WriteFile(g.configDBFile, []byte(""), 0644); err != nil {
			gl.Log("error", fmt.Sprintf("❌ Erro ao criar o arquivo de configuração do banco de dados: %v", err))
			return nil, fmt.Errorf("❌ Erro ao criar o arquivo de configuração do banco de dados: %v", err)
		}
	}
	dbConfig := is.NewDBConfigWithArgs(context.Background(), g.dbName, g.configDBFile, true, g.Logger, g.environment.Getenv("DEBUG") == "true")
	// if dbConfig == nil {
	// 	gl.Log("error", "❌ Erro ao inicializar DBConfig")
	// 	return nil, fmt.Errorf("❌ Erro ao inicializar DBConfig")
	// }
	return dbConfig, nil
}

func WaitForDatabase(dbConfig is.DBConfig) (*gorm.DB, error) {
	if dbConfig == nil {
		return nil, fmt.Errorf("configuração do banco de dados não pode ser nula")
	}

	// Get PostgreSQL config using interface method
	pgConfigAny := dbConfig.GetPostgresConfig()
	if pgConfigAny == nil {
		return nil, fmt.Errorf("configuração PostgreSQL não encontrada")
	}

	pgConfig := pgConfigAny

	if pgConfig.Dsn == "" {
		pgConfig.Dsn = pgConfig.ConnectionString
	}
	if pgConfig.Dsn == "" {
		// pragma: allowlist secret
		pgConfig.Dsn = fmt.Sprintf("host=%s port=%v user=%s password=%s dbname=%s sslmode=disable",
			pgConfig.Host, pgConfig.Port, pgConfig.Username, pgConfig.Password, pgConfig.Name)
	}

	for index := 0; index < 10; index++ {
		db, err := gorm.Open(postgres.Open(pgConfig.Dsn), &gorm.Config{})
		if err == nil {
			return db, nil
		}
		fmt.Println("Aguardando banco de dados iniciar...")
		time.Sleep(5 * time.Second)
	}
	return nil, fmt.Errorf("tempo limite excedido ao esperar pelo banco de dados")
}
