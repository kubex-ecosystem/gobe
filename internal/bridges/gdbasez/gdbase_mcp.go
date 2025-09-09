package gdbasez

import (
	models "github.com/kubex-ecosystem/gdbase/factory/models/mcp"
)

// LLM aliases

type LLMService = models.LLMService
type LLMModel = models.LLMModel
type LLMRepo = models.LLMRepo

func NewLLMService(repo LLMRepo) LLMService {
	return models.NewLLMService(repo)
}

// Preferences aliases

type PreferencesService = models.PreferencesService
type PreferencesModel = models.PreferencesModel
type PreferencesRepo = models.PreferencesRepo

func NewPreferencesService(repo PreferencesRepo) PreferencesService {
	return models.NewPreferencesService(repo)
}

// Providers aliases

type ProvidersService = models.ProvidersService
type ProvidersModel = models.ProvidersModel
type ProvidersRepo = models.ProvidersRepo

func NewProvidersService(repo ProvidersRepo) ProvidersService {
	return models.NewProvidersService(repo)
}

// Tasks aliases

type TasksService = models.TasksService
type TasksModel = models.TasksModel
type TasksRepo = models.TasksRepo

func NewTasksService(repo TasksRepo) TasksService {
	return models.NewTasksService(repo)
}
