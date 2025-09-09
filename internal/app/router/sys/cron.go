package sys

import (
	"github.com/gin-gonic/gin"
	c "github.com/kubex-ecosystem/gobe/internal/app/controllers/sys/cron"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	l "github.com/kubex-ecosystem/logz"
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
	if dbService == nil {
		gl.Log("error", "Database service is nil for CronRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}

	cronJobController := c.NewCronJobController(dbGorm)
	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["CreateCronJobRoute"] = proto.NewRoute("POST", "/api/v1/cronjobs", "application/json", cronJobController.CreateCronJob, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetCronJobRoute"] = proto.NewRoute("GET", "/api/v1/cronjobs/:id", "application/json", cronJobController.GetCronJobByID, middlewaresMap, dbService, secureProperties, nil)
	routesMap["ListCronJobsRoute"] = proto.NewRoute("GET", "/api/v1/cronjobs", "application/json", cronJobController.ListCronJobs, middlewaresMap, dbService, secureProperties, nil)

	// Define the routes for cron jobs
	routesMap["UpdateCronJobRoute"] = proto.NewRoute("PUT", "/api/v1/cronjobs/:id", "application/json", cronJobController.UpdateCronJob, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeleteCronJobRoute"] = proto.NewRoute("DELETE", "/api/v1/cronjobs/:id", "application/json", cronJobController.DeleteCronJob, middlewaresMap, dbService, secureProperties, nil)
	routesMap["EnableCronJobRoute"] = proto.NewRoute("POST", "/api/v1/cronjobs/:id/enable", "application/json", cronJobController.EnableCronJob, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DisableCronJobRoute"] = proto.NewRoute("POST", "/api/v1/cronjobs/:id/disable", "application/json", cronJobController.DisableCronJob, middlewaresMap, dbService, secureProperties, nil)
	routesMap["ExecuteCronJobManuallyRoute"] = proto.NewRoute("POST", "/api/v1/cronjobs/:id/execute", "application/json", cronJobController.ExecuteCronJobManually, middlewaresMap, dbService, secureProperties, nil)
	routesMap["ListActiveCronJobsRoute"] = proto.NewRoute("GET", "/api/v1/cronjobs/active", "application/json", cronJobController.ListActiveCronJobs, middlewaresMap, dbService, secureProperties, nil)
	routesMap["RescheduleCronJobRoute"] = proto.NewRoute("PUT", "/api/v1/cronjobs/:id/reschedule", "application/json", cronJobController.RescheduleCronJob, middlewaresMap, dbService, secureProperties, nil)
	routesMap["ValidateCronExpressionRoute"] = proto.NewRoute("POST", "/api/v1/cronjobs/validate", "application/json", cronJobController.ValidateCronExpression, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
