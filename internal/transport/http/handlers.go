package http

import (
	"encoding/json"
	"net/http"
	"pr-review-service/internal/domain"
	"pr-review-service/internal/service"
)

type Handler struct {
	teamService *service.TeamService
	userService *service.UserService
	prService   *service.PRService
}

func NewHandler(
	teamService *service.TeamService,
	userService *service.UserService,
	prService *service.PRService,
) *Handler {
	return &Handler{
		teamService: teamService,
		userService: userService,
		prService:   prService,
	}
}

// CreateTeam POST /team/add
func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	createdTeam, err := h.teamService.CreateTeam(r.Context(), &team)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"team": createdTeam,
	})
}

// GetTeam GET /team/get
func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "team_name is required")
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, team)
}

// SetIsActive POST /users/setIsActive
func (h *Handler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

// CreatePR POST /pullRequest/create
func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	pr, err := h.prService.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"pr": pr,
	})
}

// MergePR POST /pullRequest/merge
func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	pr, err := h.prService.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pr": pr,
	})
}

// ReassignReviewer POST /pullRequest/reassign
func (h *Handler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "invalid request body")
		return
	}

	pr, replacedBy, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": replacedBy,
	})
}

// GetUserReviews GET /users/getReview
func (h *Handler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "user_id is required")
		return
	}

	prs, err := h.userService.GetUserReviews(r.Context(), userID)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

// GetStats  GET /stats (Бонус)
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.userService.GetStats(r.Context(), 10)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats,
	})
}

// DeactivateTeam POST /team/{team_name}/deactivate-all (Бонус)
func (h *Handler) DeactivateTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondError(w, http.StatusBadRequest, "INVALID_INPUT", "team_name is required")
		return
	}

	if err := h.prService.DeactivateTeamAndReassign(r.Context(), teamName); err != nil {
		handleDomainError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "team deactivated and reviewers reassigned successfully",
	})
}

// HealthCheck GET /health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}
