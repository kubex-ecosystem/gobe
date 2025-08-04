package routes

import (
	"github.com/gin-gonic/gin"
	contacts "github.com/rafa-mori/gobe/internal/controllers/contacts"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	l "github.com/rafa-mori/logz"

	"net/http"
)

type ContactRoutes struct {
	ar.IRouter
}

func NewContactRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		l.ErrorCtx("Router is nil for ContactRoute", nil)
		return nil
	}
	rtl := *rtr

	handler := contacts.ContactController{}

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)

	dbService := rtl.GetDatabaseService()

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["PostContactRoute"] = NewRoute(http.MethodPost, "/api/v1/contact", "application/json", handler.PostContact, middlewaresMap, dbService, secureProperties)
	routesMap["GetContactRoute"] = NewRoute(http.MethodGet, "/api/v1/contact", "application/json", handler.GetContact, middlewaresMap, dbService, secureProperties)
	routesMap["HandleContactRoute"] = NewRoute(http.MethodPost, "/api/v1/contact/handle", "application/json", handler.HandleContact, middlewaresMap, dbService, secureProperties)

	return routesMap
}
