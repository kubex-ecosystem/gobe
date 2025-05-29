package routes

import (
	l "github.com/faelmori/logz"
	c "github.com/rafa-mori/gobe/internal/controllers/cron"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
)

type CronRoutes struct {
	ar.IRouter
}

// NewCronRoutes cria novas rotas para o serviço de cron jobs.
func NewCronRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		l.ErrorCtx("Router is nil for CronRoute", nil)
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}

	cronJobController := c.NewCronJobController(dbGorm)
	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]any)

	routesMap["CreateCronJobRoute"] = NewRoute("POST", "/cronjobs", "application/json", cronJobController.CreateCronJob, middlewaresMap, dbService)
	routesMap["GetCronJobRoute"] = NewRoute("GET", "/cronjobs/:id", "application/json", cronJobController.GetCronJobByID, middlewaresMap, dbService)
	routesMap["ListCronJobsRoute"] = NewRoute("GET", "/cronjobs", "application/json", cronJobController.ListCronJobs, middlewaresMap, dbService)
	routesMap["UpdateCronJobRoute"] = NewRoute("PUT", "/cronjobs/:id", "application/json", cronJobController.UpdateCronJob, middlewaresMap, dbService)
	routesMap["DeleteCronJobRoute"] = NewRoute("DELETE", "/cronjobs/:id", "application/json", cronJobController.DeleteCronJob, middlewaresMap, dbService)
	routesMap["EnableCronJobRoute"] = NewRoute("POST", "/cronjobs/:id/enable", "application/json", cronJobController.EnableCronJob, middlewaresMap, dbService)
	routesMap["DisableCronJobRoute"] = NewRoute("POST", "/cronjobs/:id/disable", "application/json", cronJobController.DisableCronJob, middlewaresMap, dbService)
	routesMap["ExecuteCronJobManuallyRoute"] = NewRoute("POST", "/cronjobs/:id/execute", "application/json", cronJobController.ExecuteCronJobManually, middlewaresMap, dbService)
	routesMap["ListActiveCronJobsRoute"] = NewRoute("GET", "/cronjobs/active", "application/json", cronJobController.ListActiveCronJobs, middlewaresMap, dbService)
	routesMap["RescheduleCronJobRoute"] = NewRoute("PUT", "/cronjobs/:id/reschedule", "application/json", cronJobController.RescheduleCronJob, middlewaresMap, dbService)
	routesMap["ValidateCronExpressionRoute"] = NewRoute("POST", "/cronjobs/validate", "application/json", cronJobController.ValidateCronExpression, middlewaresMap, dbService)

	return routesMap
}
