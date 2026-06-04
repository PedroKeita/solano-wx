package handlers

import (
	"net/http"

	_ "solano-wx/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

func NewDocsHandler() http.Handler {
	return httpSwagger.WrapHandler
}
