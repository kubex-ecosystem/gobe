package routes

import (
	l "github.com/faelmori/logz"
	contacts "github.com/rafa-mori/gobe/internal/controllers/contacts"
	ar "github.com/rafa-mori/gobe/internal/interfaces"

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
	middlewaresMap := make(map[string]any)

	dbService := rtl.GetDatabaseService()

	routesMap["PostContactRoute"] = NewRoute(http.MethodPost, "/contact", "application/json", handler.PostContact, middlewaresMap, dbService)
	routesMap["GetContactRoute"] = NewRoute(http.MethodGet, "/contact", "application/json", handler.GetContact, middlewaresMap, dbService)
	routesMap["HandleContactRoute"] = NewRoute(http.MethodPost, "/contact/handle", "application/json", handler.HandleContact, middlewaresMap, dbService)

	return routesMap
}
