// Package gdbasez provides database services for the application.
package gdbasez

import (
	"context"
	"reflect"

	svc "github.com/kubex-ecosystem/gdbase/factory"
	l "github.com/kubex-ecosystem/logz"
	_ "github.com/lib/pq"
)

// DefaultDBName is the default database name used throughout GoBE
const DefaultDBName = "kubex_db"

// Database service and config type aliases

type DBService = svc.DBService
type IDBService interface{ svc.DBService }
type DBServiceImpl = svc.DBServiceImpl

type DBConfig = svc.DBConfig
type IDBConfig = svc.IDBConfig
type DBConfigImpl = svc.DBConfigImpl

type DatabaseImpl = svc.DatabaseImpl
type MessageryImpl = svc.Messagery

// Additional type aliases from factory

type Database = svc.Database
type EnvironmentType = svc.EnvironmentType
type Environment = svc.Environment

func NewEnvironment(configFile string, isConfidential bool, logger l.Logger) (*EnvironmentType, error) {
	return svc.NewEnvironment(configFile, isConfidential, logger)
}

type MongoDB = svc.MongoDB
type Redis = svc.Redis
type RabbitMQ = svc.RabbitMQ

// Helper functions for JSONB

func NewJSONB() JSONB {
	return svc.NewJSONB()
}

func NewJSONBData() JSONBImpl {
	return svc.NewJSONBData()
}

func MapToJSONB(data map[string]interface{}) svc.JSONBImpl {
	jsonb := svc.JSONBImpl{}
	for k, v := range data {
		jsonb[k] = v
	}
	return jsonb
}

func JSONBToImpl(data interface{}) JSONBImpl {
	if data == nil {
		return svc.NewJSONBData()
	}
	// Try to convert to JSONBImpl
	if impl, ok := data.(JSONBImpl); ok {
		return impl
	}
	// Try to convert to map
	mapData := svc.NewJSONBData()
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			mapData.Set(k, v)
		}
		return mapData
	}
	return svc.NewJSONBData()
}

func NewDBService(ctx context.Context, config *DBConfigImpl, logger l.Logger) (DBService, error) {
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
