package app

import (
	"github.com/gin-gonic/gin"
	cts "github.com/rafa-mori/gobe/internal/app/controllers/app/contacts"
	proto "github.com/rafa-mori/gobe/internal/app/router/types"
	ar "github.com/rafa-mori/gobe/internal/contracts/interfaces"
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

	handler := cts.ContactController{}

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)

	dbService := rtl.GetDatabaseService()

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["PostContactRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/contact", "application/json", handler.PostContact, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetContactRoute"] = proto.NewRoute(http.MethodGet, "/api/v1/contact", "application/json", handler.GetContact, middlewaresMap, dbService, secureProperties, nil)
	routesMap["HandleContactRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/contact/handle", "application/json", handler.HandleContact, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
