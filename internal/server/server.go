package server

import (
	"hitalent-test/internal/handlers"
	"hitalent-test/internal/ui"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(deptHandler *handlers.DepartmentHandler, empHandler *handlers.EmployeeHandler) http.Handler {
	r := chi.NewRouter()

	r.Get("/live", handlers.LivenessProbe)
	r.Get("/ready", handlers.ReadinessProbe)
	r.Get("/startup", handlers.StartupProbe)

	r.Get("/", ui.ServeIndex)

	r.Route("/departments", func(r chi.Router) {
		r.Post("/", deptHandler.Create)
		r.Get("/{id}", deptHandler.GetByID)
		r.Patch("/{id}", deptHandler.Update)
		r.Delete("/{id}", deptHandler.Delete)

		r.Post("/{id}/employees", empHandler.Create)
	})

	return r
}
