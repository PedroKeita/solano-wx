package handlers

import (
	"net/http"
)

func NewDashboardHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/dashboard/index.html")
	}
}
