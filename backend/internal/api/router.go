package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter builds the chi router with CORS, standard middlewares, public routes, and protected admin routes.
func NewRouter(h *Handler, allowedOrigins []string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", h.Health)

	r.Route("/api/v1", func(apiRouter chi.Router) {
		apiRouter.Get("/health", h.Health)
		apiRouter.Get("/nodes", h.ListPublicNodes)

		apiRouter.Get("/map/heatmap", h.GetHeatmap)
		apiRouter.Get("/map/point", h.GetPoint)
		apiRouter.Get("/map/playback/timestamps", h.GetPlaybackTimestamps)

		apiRouter.Post("/access/verify", h.VerifyAccess)
		apiRouter.Get("/access/session", h.CheckSession)
		apiRouter.Post("/access/logout", h.Logout)

		apiRouter.Route("/admin", func(adminRouter chi.Router) {
			adminRouter.Use(h.auth.RequireAdminMiddleware)

			adminRouter.Get("/nodes", h.ListAdminNodes)
			adminRouter.Post("/nodes", h.CreateAdminNode)
			adminRouter.Patch("/nodes/{id}", h.UpdateAdminNode)
			adminRouter.Delete("/nodes/{id}", h.DeleteAdminNode)
			adminRouter.Get("/nodes/{id}/readings", h.GetAdminNodeReadings)
			adminRouter.Post("/nodes/{id}/readings/batch", h.BatchInsertNodeReadings)
		})
	})

	return r
}
