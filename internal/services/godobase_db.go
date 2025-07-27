// Package services provides database services for the application.
package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	f "github.com/rafa-mori/gdbase/factory"
	sc "github.com/rafa-mori/gdbase/types"
	ut "github.com/rafa-mori/gdbase/utils"
	cm "github.com/rafa-mori/gobe/internal/common"
	ci "github.com/rafa-mori/gobe/internal/interfaces"
	fcs "github.com/rafa-mori/gobe/internal/security/certificates"
	t "github.com/rafa-mori/gobe/internal/types"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBService = sc.IDBService

func NewDBService(config *sc.DBConfig, logger l.Logger) (DBService, error) {
	return f.NewDatabaseService(config, logger)
}

type IDBConfig = sc.DBConfig

func SetupDatabase(environment ci.IEnvironment, dbConfigFilePath string, logger l.Logger, debug bool) (*sc.DBConfig, error) {
	dbName := environment.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "GoBE-DB"
	}
	if _, err := os.Stat(dbConfigFilePath); err != nil && os.IsNotExist(err) {
		if err := ut.EnsureDir(filepath.Dir(dbConfigFilePath), 0644, []string{}); err != nil {
			gl.Log("error", fmt.Sprintf("❌ Erro ao criar o diretório do arquivo de configuração do banco de dados: %v", err))
			return nil, fmt.Errorf("❌ Erro ao criar o diretório do arquivo de configuração do banco de dados: %v", err)
		}
		if err := os.WriteFile(dbConfigFilePath, []byte(""), 0644); err != nil {
			gl.Log("error", fmt.Sprintf("❌ Erro ao criar o arquivo de configuração do banco de dados: %v", err))
			return nil, fmt.Errorf("❌ Erro ao criar o arquivo de configuração do banco de dados: %v", err)
		}
	}
	dbConfig := sc.NewDBConfigWithArgs(dbName, dbConfigFilePath, true, logger, debug)
	if dbConfig == nil {
		gl.Log("error", "❌ Erro ao inicializar DBConfig")
		return nil, fmt.Errorf("❌ Erro ao inicializar DBConfig")
	}
	if len(dbConfig.Databases) == 0 {
		gl.Log("error", "❌ Erro ao inicializar DBConfig: Nenhum banco de dados encontrado")
		return nil, fmt.Errorf("❌ Erro ao inicializar DBConfig: Nenhum banco de dados encontrado")
	}
	gl.Log("success", fmt.Sprintf("Banco de dados encontrado: %v", dbConfig.Databases))
	return dbConfig, nil
}

func WaitForDatabase(dbConfig *sc.DBConfig) (*gorm.DB, error) {
	if dbConfig == nil {
		return nil, fmt.Errorf("configuração do banco de dados não pode ser nula")
	}
	if len(dbConfig.Databases) == 0 {
		return nil, fmt.Errorf("nenhum banco de dados encontrado na configuração")
	}
	var pgConfig *sc.Database
	for _, db := range dbConfig.Databases {
		if db.Type == "postgresql" {
			pgConfig = db
			break
		}
	}
	if pgConfig == nil {
		return nil, fmt.Errorf("configuração do banco de dados não pode ser nula")
	}
	if pgConfig.Dsn == "" {
		pgConfig.Dsn = pgConfig.ConnectionString
	}
	if pgConfig.Dsn == "" {
		pgConfig.Dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
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

func InitializeAllServices(environment ci.IEnvironment, logger l.Logger, debug bool) (DBService, error) {
	if logger == nil {
		logger = l.NewLogger("GoBE")
	}
	var err error
	if environment == nil {
		environment, err = t.NewEnvironment(os.ExpandEnv(cm.DefaultGoBEConfigPath), false, logger)
		if err != nil {
			gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar o ambiente: %v", err))
			return nil, fmt.Errorf("❌ Erro ao inicializar o ambiente: %v", err)
		}
	}

	// 1. Inicializar Certificados
	keyPath := environment.Getenv("GOBE_KEY_PATH")
	certPath := environment.Getenv("GOBE_CERT_PATH")
	if keyPath == "" {
		keyPath = os.ExpandEnv(cm.DefaultGoBEKeyPath)
	}
	if certPath == "" {
		certPath = os.ExpandEnv(cm.DefaultGoBECertPath)
	}
	certService := fcs.NewCertService(keyPath, certPath)
	if certService == nil {
		gl.Log("error", "❌ Erro ao inicializar CertService")
		return nil, fmt.Errorf("❌ Erro ao inicializar CertService")
	}

	dbConfigFile := environment.Getenv("DB_CONFIG_FILE")
	if dbConfigFile == "" {
		dbConfigFile = os.ExpandEnv(cm.DefaultGodoBaseConfigPath)
	}
	dbConfig, dbConfigErr := SetupDatabase(environment, dbConfigFile, logger, debug)
	if dbConfigErr != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar DBConfig: %v", dbConfigErr))
		return nil, fmt.Errorf("❌ Erro ao inicializar DBConfig: %v", dbConfigErr)
	}
	if dbConfig == nil {
		gl.Log("error", "❌ Erro ao inicializar DBConfig")
		return nil, fmt.Errorf("❌ Erro ao inicializar DBConfig")
	}

	// 2. Inicializar Docker
	dockerService, dockerServiceErr := f.NewDockerService(dbConfig, logger)
	if dockerServiceErr != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar DockerService: %v", dockerServiceErr))
		return nil, fmt.Errorf("❌ Erro ao inicializar DockerService: %v", dockerServiceErr)
	}

	err = f.SetupDatabaseServices(dockerService, dbConfig)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao configurar Docker: %v", err))
		return nil, err
	}

	err = dockerService.Initialize()
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar Docker: %v", err))
		return nil, err
	}
	if err := f.SetupDatabaseServices(dockerService, dbConfig); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao configurar Docker: %v", err))
		return nil, fmt.Errorf("❌ Erro ao configurar Docker: %v", err)
	}

	// 3. Inicializar Banco de Dados --- TA PÁRANDOA QUI ATÉ CAIR POR TIMEOUT.. O DOCKER NÃO ESTÁ SUBINDO O PG
	if _, err = WaitForDatabase(dbConfig); err != nil {
		return nil, err
	}
	dbService, err := f.NewDatabaseService(dbConfig, logger)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar DatabaseService: %v", err))
		return nil, fmt.Errorf("❌ Erro ao inicializar DatabaseService: %v", err)
	}
	if err := dbService.Initialize(); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao conectar ao banco: %v", err))
		return nil, fmt.Errorf("❌ Erro ao conectar ao banco: %v", err)
	}

	fmt.Println("✅ Todos os serviços rodando corretamente!")

	// Retorno o DB para o BE
	return dbService, nil
}
