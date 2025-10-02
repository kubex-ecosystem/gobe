// Package gdbasez provides database services for the application.
package gdbasez

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	l "github.com/kubex-ecosystem/logz"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed sql_migrations/*.sql
var migrationFiles embed.FS

// Database service and config type aliases

type DBService = svc.DBService
type IDBService = svc.IDBService
type DBConfig = svc.DBConfig
type IDBConfig = svc.IDBConfig
type DatabaseType = svc.DatabaseType
type JSONB = svc.JSONB
type IJSONBData = svc.IJSONBData
type JSONBData = svc.JSONBData
type JSONBImpl = svc.JSONBImpl

// Additional type aliases from factory

type Database = svc.Database
type EnvironmentType = svc.Environment
type Environment = svc.Environment

func NewEnvironment(configFile string, isConfidential bool, logger l.Logger) (Environment, error) {
	return svc.NewEnvironment(configFile, isConfidential, logger)
}

type MongoDB = svc.MongoDB
type Redis = svc.Redis
type RabbitMQ = svc.RabbitMQ

// Helper functions for JSONB

func NewJSONB() JSONB {
	return svc.NewJSONB()
}

func NewJSONBData() IJSONBData {
	return svc.NewJSONBData()
}

func MapToJSONB(data map[string]interface{}) JSONBImpl {
	jsonb := svc.JSONBImpl{}
	for k, v := range data {
		jsonb[k] = v
	}
	return jsonb
}

func JSONBToImpl(data interface{}) JSONBImpl {
	if data == nil {
		return JSONBImpl{}
	}
	// Try to convert to JSONBImpl
	if impl, ok := data.(JSONBImpl); ok {
		return impl
	}
	// Try to convert to map
	if m, ok := data.(map[string]interface{}); ok {
		return MapToJSONB(m)
	}
	return JSONBImpl{}
}

func NewDBService(ctx context.Context, config DBConfig, logger l.Logger) (DBService, error) {
	return svc.NewDatabaseService(ctx, config, logger)
}

func getEnvOrDefault[T string | int | bool](environment svc.Environment, key string, defaultValue T) T {
	value := environment.Getenv(key)
	if value == "" {
		return defaultValue
	} else {
		valInterface := reflect.ValueOf(value)
		if valInterface.Type().ConvertibleTo(reflect.TypeFor[T]()) {
			return valInterface.Convert(reflect.TypeFor[T]()).Interface().(T)
		}
	}
	return defaultValue
}

func SetupDatabase(ctx context.Context, environment svc.Environment, dbConfigFilePath string, logger l.Logger, debug bool) (DBConfig, error) {
	dbName := getEnvOrDefault(environment, "DB_NAME", "kubex_db")
	if _, err := os.Stat(dbConfigFilePath); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dbConfigFilePath), 0755); err != nil {
			gl.Log("error", fmt.Sprintf("❌ Erro ao criar o diretório do arquivo de configuração do banco de dados: %v", err))
			return nil, fmt.Errorf("❌ Erro ao criar o diretório do arquivo de configuração do banco de dados: %v", err)
		}
		if err := os.WriteFile(dbConfigFilePath, []byte(""), 0644); err != nil {
			gl.Log("error", fmt.Sprintf("❌ Erro ao criar o arquivo de configuração do banco de dados: %v", err))
			return nil, fmt.Errorf("❌ Erro ao criar o arquivo de configuração do banco de dados: %v", err)
		}
	}
	dbConfig := svc.NewDBConfigWithArgs(ctx, dbName, dbConfigFilePath, true, logger, debug)
	// if dbConfig == nil {
	// 	gl.Log("error", "❌ Erro ao inicializar DBConfig")
	// 	return nil, fmt.Errorf("❌ Erro ao inicializar DBConfig")
	// }
	return dbConfig, nil
}

func WaitForDatabase(dbConfig DBConfig) (*gorm.DB, error) {
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

// InitializeAllServices inicializa todos os serviços (Docker + Database) com context
func InitializeAllServices(ctx context.Context, environment svc.Environment, logger l.Logger, debug bool) (DBService, error) {
	if logger == nil {
		logger = l.NewLogger("GoBE")
	}

	// 1. Setup database config
	dbConfigFile := getEnvOrDefault(environment, "DB_CONFIG_FILE", os.ExpandEnv("$HOME/.kubex/gdbase/config/config.json"))
	dbConfig, err := SetupDatabase(ctx, environment, dbConfigFile, logger, debug)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar DBConfig: %v", err))
		return nil, fmt.Errorf("❌ Erro ao inicializar DBConfig: %w", err)
	}

	// 2. Initialize Docker Service (usando factory wrapper)
	dockerService := svc.IDockerService(nil)
	// TODO: Implementar wrapper do DockerService via factory se necessário
	// Por enquanto, vamos aguardar o DB sem Docker setup

	// 3. Setup Database Services via factory
	// if dockerService == nil {
	// 	gl.Log("info", "⚠️ DockerService é nil, pulando configuração do Docker")
	// 	return nil, fmt.Errorf("⚠️ DockerService é nil, pulando configuração do Docker")
	// }

	if err := svc.SetupDatabaseServices(ctx, dockerService, dbConfig); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao configurar Docker: %v", err))
		return nil, fmt.Errorf("❌ Erro ao configurar Docker: %w", err)
	}

	// 4. Wait for Database
	if _, err := WaitForDatabase(dbConfig); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao aguardar database: %v", err))
		return nil, err
	}

	// 5. Create Database Service
	dbService, err := svc.NewDatabaseService(ctx, dbConfig, logger)
	if err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao inicializar DatabaseService: %v", err))
		return nil, fmt.Errorf("❌ Erro ao inicializar DatabaseService: %w", err)
	}

	// 6. Initialize Database Service
	if err := dbService.Initialize(ctx); err != nil {
		gl.Log("error", fmt.Sprintf("❌ Erro ao conectar ao banco: %v", err))
		return nil, fmt.Errorf("❌ Erro ao conectar ao banco: %w", err)
	}

	return dbService, nil
}
