package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/narroworb/pr-review-service/internal/models"
)

type DatabaseInterface interface {
	GetTeamByName(context.Context, string) (models.Team, error)
	CreateTeam(context.Context, string) (int64, error)
	GetUserByID(context.Context, string) (models.User, error)
	CreateUser(context.Context, models.User) error
	GetUsersInTeam(context.Context, int64) ([]models.User, error)
	InsertTeamInTransaction(context.Context, string, []models.User) error
	GetUserWithTeamByID(context.Context, string) (models.User, string, error)
	UpdateUserActivity(context.Context, string, bool) error
	GetPRByID(context.Context, string) (models.PullRequest, error)
	GetActiveUsersInTeamExcAuthor(context.Context, int64, string) ([]models.User, error)
	InsertPRInTransaction(context.Context, models.PullRequest) error
	GetReviewersByPRID(context.Context, string) ([]string, error)
	SetMergedStatusPR(context.Context, string) (time.Time, error)
	FoundAvailableReviewerPR(context.Context, string, []string, string) (string, error)
	SwapReviewerInPR(context.Context, string, string, string) error
	GetPRByReviewerID(context.Context, string) ([]models.PullRequest, error)
	GetCountPRStatsByUser(context.Context) ([]models.UserStats, error)
	GetCountPRStatsByTeam(context.Context) ([]models.TeamStats, error)
	GetCountReviewerStatsByPR(context.Context) (map[string]int64, error)
}

type HandlersRepo struct {
	db DatabaseInterface
}

func NewHandlersRepo(db DatabaseInterface) *HandlersRepo {
	return &HandlersRepo{
		db: db,
	}
}

func writeError(w http.ResponseWriter, code, message string, statusCode int) {
	w.WriteHeader(statusCode)
	var e models.ErrorResponse
	e.Error.Code = code
	e.Error.Message = message
	json.NewEncoder(w).Encode(e)
}

func (h *HandlersRepo) AddTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req models.AddTeamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BAD_REQUEST", "invalid json body of request", http.StatusBadRequest)
		return
	}

	team, err := h.db.GetTeamByName(ctx, req.TeamName)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get team in handler /team/add: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}
	if team.ID != 0 {
		writeError(w, "TEAM_EXISTS", fmt.Sprintf("team_name %s already exists", req.TeamName), http.StatusBadRequest)
		return
	}

	// teamID, err := h.db.CreateTeam(req.TeamName)
	// if err != nil {
	// 	log.Printf("error in create team in handler /team/add: %v", err)
	// 	serverError(w)
	// 	return
	// }

	users := make([]models.User, 0, len(req.Members))

	for _, m := range req.Members {
		user, err := h.db.GetUserByID(ctx, m.UserID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("error in get user in handler /team/add: %v", err)
			writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
			return
		}
		if user.ID != "" {
			writeError(w, "USER_EXISTS", fmt.Sprintf("user with id=%s already belongs to another team", m.UserID), http.StatusBadRequest)
			return
		}

		users = append(users, models.User{
			ID:       m.UserID,
			Name:     m.Name,
			IsActive: m.IsActive,
		})

		// err = h.db.CreateUser(models.User{
		// 	ID:       m.UserID,
		// 	Name:     m.Name,
		// 	IsActive: m.IsActive,
		// 	GroupID:  teamID,
		// })

		// if err != nil {
		// 	log.Printf("error in create user in handler /team/add: %v", err)
		// 	serverError(w)
		// 	return
		// }
	}

	if err := h.db.InsertTeamInTransaction(ctx, req.TeamName, users); err != nil {
		log.Printf("error in apply transaction to create team in handler /team/add: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.AddTeamResponse{Team: req})
}

func (h *HandlersRepo) GetTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, "BAD_REQUEST", "empty query parameter team_name", http.StatusBadRequest)
		return
	}
	team, err := h.db.GetTeamByName(ctx, teamName)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get team in handler /team/get?team_name=%s: %v", teamName, err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}
	if err == sql.ErrNoRows {
		writeError(w, "TEAM_NOT_FOUND", fmt.Sprintf("there is no team with name: %s", teamName), http.StatusNotFound)
		return
	}

	var resp models.GetTeamResponse
	resp.Team.Name = team.Name
	members, err := h.db.GetUsersInTeam(ctx, team.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get users in handler /team/get?team_name=%s: %v", teamName, err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}
	resp.Team.Members = members

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlersRepo) SetUserIsActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req models.SetUserIsActiveRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BAD_REQUEST", "invalid json body of request", http.StatusBadRequest)
		return
	}

	user, teamName, err := h.db.GetUserWithTeamByID(ctx, req.UserID)

	if err == sql.ErrNoRows {
		writeError(w, "USER_NOT_FOUND", fmt.Sprintf("there is not user with id=%s", req.UserID), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get user in handler /users/setIsActive: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	var resp models.SetUserIsActiveResponse
	resp.User.UserID, resp.User.Username, resp.User.TeamName, resp.User.IsActive = user.ID, user.Name, teamName, req.IsActive

	if user.IsActive == req.IsActive {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	err = h.db.UpdateUserActivity(ctx, user.ID, req.IsActive)
	if err != nil {
		log.Printf("error in update user in handler /users/setIsActive: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlersRepo) CreatePR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req models.CreatePRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BAD_REQUEST", "invalid json body of request", http.StatusBadRequest)
		return
	}

	_, err := h.db.GetPRByID(ctx, req.PRID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in create pr in handler /pullRequest/create: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}
	if err == nil {
		writeError(w, "PR_EXISTS", fmt.Sprintf("PR with id=%s already exists", req.PRID), http.StatusBadRequest)
		return
	}

	user, _, err := h.db.GetUserWithTeamByID(ctx, req.AuthorID)
	if err == sql.ErrNoRows {
		writeError(w, "USER_NOT_FOUND", fmt.Sprintf("there is no user with id=%s", req.AuthorID), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get user in handler /pullRequest/create: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	reviewers, err := h.db.GetActiveUsersInTeamExcAuthor(ctx, user.GroupID, user.ID)
	if err != nil {
		log.Printf("error in get reviewers in handler /pullRequest/create: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	pr := models.PullRequest{
		ID:        req.PRID,
		Name:      req.PRName,
		AuthorID:  req.AuthorID,
		Status:    models.PRStatusOpen,
		Reviewers: reviewers,
	}

	err = h.db.InsertPRInTransaction(ctx, pr)
	if err != nil {
		log.Printf("error in insert pr in handler /pullRequest/create: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	var resp models.CreatePRResponse

	resp.PR.PRID, resp.PR.PRName, resp.PR.AuthorID, resp.PR.Status, resp.PR.Reviewers = pr.ID, pr.Name, pr.AuthorID, pr.Status, make([]string, 0, 2)
	for _, u := range pr.Reviewers {
		resp.PR.Reviewers = append(resp.PR.Reviewers, u.ID)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlersRepo) MergePR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req models.MergePRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BAD_REQUEST", "invalid json body of request", http.StatusBadRequest)
		return
	}

	pr, err := h.db.GetPRByID(ctx, req.PRID)
	if err == sql.ErrNoRows {
		writeError(w, "PR_NOT_FOUND", fmt.Sprintf("there is no pull request with id=%s", req.PRID), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get pr in handler /pullRequest/merge: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	var resp models.MergePRResponse
	resp.PR.PRID, resp.PR.PRName, resp.PR.AuthorID, resp.PR.Status = pr.ID, pr.Name, pr.AuthorID, pr.Status
	resp.PR.Reviewers, err = h.db.GetReviewersByPRID(ctx, pr.ID)
	if err != nil {
		log.Printf("error in get reviewers in handler /pullRequest/merge: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	if pr.Status == models.PRStatusMerged {
		resp.PR.MergedAt = *pr.MergedAt
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.PR.MergedAt, err = h.db.SetMergedStatusPR(ctx, pr.ID)
	if err != nil {
		log.Printf("error in update status in handler /pullRequest/merge: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}
	resp.PR.Status = models.PRStatusMerged

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlersRepo) ReassignPR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req models.ReassignPRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "BAD_REQUEST", "invalid json body of request", http.StatusBadRequest)
		return
	}

	pr, err := h.db.GetPRByID(ctx, req.PRID)
	if err == sql.ErrNoRows {
		writeError(w, "PR_NOT_FOUND", fmt.Sprintf("there is no pull request with id=%s", req.PRID), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get pr in handler /pullRequest/reassign: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	if pr.Status == models.PRStatusMerged {
		writeError(w, "PR_MERGED", "cannot reassign on merged PR", http.StatusConflict)
		return
	}

	_, err = h.db.GetUserByID(ctx, req.OldReviewerID)
	if err == sql.ErrNoRows {
		writeError(w, "USER_NOT_FOUND", fmt.Sprintf("there is no user with id=%s", req.OldReviewerID), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get user in handler /pullRequest/reassign: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	var resp models.ReassignPRResponse
	resp.PR.PRID, resp.PR.PRName, resp.PR.AuthorID, resp.PR.Status = pr.ID, pr.Name, pr.AuthorID, pr.Status
	resp.PR.Reviewers, err = h.db.GetReviewersByPRID(ctx, pr.ID)

	if err != nil {
		log.Printf("error in get reviewers in handler /pullRequest/reassign: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}
	if !slices.Contains(resp.PR.Reviewers, req.OldReviewerID) {
		writeError(w, "NOT_ASSIGNED", fmt.Sprintf("reviewer with id=%s is not assigned to PR with id=%s", req.OldReviewerID, req.PRID), http.StatusConflict)
		return
	}

	availableReviewerID, err := h.db.FoundAvailableReviewerPR(ctx, req.PRID, resp.PR.Reviewers, resp.PR.AuthorID)
	if err == sql.ErrNoRows {
		writeError(w, "NO_CANDIDATE", "no active replacement candidate in team", http.StatusConflict)
		return
	}
	if err != nil {
		log.Printf("error in found available reviewer in handler /pullRequest/reassign: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	err = h.db.SwapReviewerInPR(ctx, req.PRID, req.OldReviewerID, availableReviewerID)
	if err != nil {
		log.Printf("error in update reviewer in handler /pullRequest/reassign: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	if resp.PR.Reviewers[0] == req.OldReviewerID {
		resp.PR.Reviewers[0] = availableReviewerID
	} else {
		resp.PR.Reviewers[1] = availableReviewerID
	}
	resp.ReplacedBy = availableReviewerID

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlersRepo) GetReview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, "BAD_REQUEST", "empty query parameter user_id", http.StatusBadRequest)
		return
	}

	_, err := h.db.GetUserByID(ctx, userID)
	if err == sql.ErrNoRows {
		writeError(w, "USER_NOT_FOUND", fmt.Sprintf("there is no user with id=%s", userID), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get user in handler /users/setIsActive: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	prs, err := h.db.GetPRByReviewerID(ctx, userID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get pr in handler /users/setIsActive: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	var resp models.GetReviewResponse
	resp.UserID = userID
	resp.PR = prs

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlersRepo) GetStatsByUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	stats, err := h.db.GetCountPRStatsByUser(ctx)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get stats in handler /stats/users: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	var resp models.GetStatsUsersResponse
	resp.PRStats = stats
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlersRepo) GetStatsByTeams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	stats, err := h.db.GetCountPRStatsByTeam(ctx)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get stats in handler /stats/users: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	var resp models.GetStatsTeamsResponse
	resp.PRStats = stats
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *HandlersRepo) GetStatsByPRs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	stats, err := h.db.GetCountReviewerStatsByPR(ctx)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get stats in handler /stats/users: %v", err)
		writeError(w, "SERVER_ERROR", "try again later", http.StatusInternalServerError)
		return
	}

	resp := models.GetStatsPRsResponse{
		PRStats: stats,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
