package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Health check
	r.Get("/health", h.HealthCheck)

	// Teams
	r.Post("/team/add", h.CreateTeam)
	r.Get("/team/get", h.GetTeam)
	r.Post("/team/deactivate-all", h.DeactivateTeam) // Bonus task

	// Users
	r.Post("/users/setIsActive", h.SetIsActive)
	r.Get("/users/getReview", h.GetUserReviews)

	// Pull Requests
	r.Post("/pullRequest/create", h.CreatePR)
	r.Post("/pullRequest/merge", h.MergePR)
	r.Post("/pullRequest/reassign", h.ReassignReviewer)

	// Stats (Bonus task)
	r.Get("/stats", h.GetStats)

	return r
}
