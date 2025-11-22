package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/narroworb/pr-review-service/internal/models"
)

type DatabaseInterface interface {
	GetTeamByName(string) (models.Team, error)
	CreateTeam(string) (int64, error)
	GetUserByID(string) (models.User, error)
	CreateUser(models.User) error
	GetUsersInTeam(int64) ([]models.User, error)
	InsertTeamInTransaction(string, []models.User) error
	GetUserWithTeamByID(string) (models.User, string, error)
	UpdateUserActivity(string, bool) error
	GetPRByID(string) (models.PullRequest, error)
	GetActiveUsersInTeamExcAuthor(int64, string) ([]models.User, error)
	InsertPRInTransaction(models.PullRequest) error
	GetReviewersByPRID(string) ([]string, error)
	SetMergedStatusPR(string) (time.Time, error)
	FoundAvailableReviewerPR(string, []string, string) (string, error)
	SwapReviewerInPR(string, string, string) error
	GetPRByReviewerID(string) ([]models.PullRequest, error)
}

type HandlersRepo struct {
	db DatabaseInterface
}

func NewHandlersRepo(db DatabaseInterface) *HandlersRepo {
	return &HandlersRepo{
		db: db,
	}
}

type addTeamRequest struct {
	TeamName string `json:"team_name"`
	Members  []struct {
		UserID   string `json:"user_id"`
		Name     string `json:"username"`
		IsActive bool   `json:"is_active"`
	} `json:"members"`
}

type addTeamResponse struct {
	Team addTeamRequest `json:"team"`
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func serverError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	var e errorResponse
	e.Error.Code = "SERVER_ERROR"
	e.Error.Message = "try again later"
	json.NewEncoder(w).Encode(e)
}

func (h *HandlersRepo) AddTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req addTeamRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	team, err := h.db.GetTeamByName(req.TeamName)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get team in handler /team/add: %v", err)
		serverError(w)
		return
	}
	if team.ID != 0 {
		w.WriteHeader(http.StatusBadRequest)
		var e errorResponse
		e.Error.Code = "TEAM_EXISTS"
		e.Error.Message = fmt.Sprintf("team_name %s already exists", req.TeamName)
		json.NewEncoder(w).Encode(e)
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
		user, err := h.db.GetUserByID(m.UserID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("error in get user in handler /team/add: %v", err)
			serverError(w)
			return
		}
		if user.ID != "" {
			w.WriteHeader(http.StatusBadRequest)
			var e errorResponse
			e.Error.Code = "USER_EXISTS"
			e.Error.Message = fmt.Sprintf("user with id=%s already belongs to another team", m.UserID)
			json.NewEncoder(w).Encode(e)
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

	if err := h.db.InsertTeamInTransaction(req.TeamName, users); err != nil {
		log.Printf("error in apply transaction to create team in handler /team/add: %v", err)
		serverError(w)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(addTeamResponse{Team: req})
}

type getTeamResponse struct {
	Team struct {
		Name    string `json:"team_name"`
		Members []models.User
	} `json:"team"`
}

func (h *HandlersRepo) GetTeam(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	team, err := h.db.GetTeamByName(teamName)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get team in handler /team/get/%s: %v", teamName, err)
		serverError(w)
		return
	}
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var resp getTeamResponse
	resp.Team.Name = team.Name
	members, err := h.db.GetUsersInTeam(team.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get users in handler /team/get/%s: %v", teamName, err)
		serverError(w)
		return
	}
	resp.Team.Members = members

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type setUserIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type setUserIsActiveResponse struct {
	User struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		TeamName string `json:"team_name"`
		IsActive bool   `json:"is_active"`
	} `json:"user"`
}

func (h *HandlersRepo) SetUserIsActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req setUserIsActiveRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, teamName, err := h.db.GetUserWithTeamByID(req.UserID)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get user in handler /users/setIsActive: %v", err)
		serverError(w)
		return
	}

	var resp setUserIsActiveResponse
	resp.User.UserID, resp.User.Username, resp.User.TeamName, resp.User.IsActive = user.ID, user.Name, teamName, req.IsActive

	if user.IsActive == req.IsActive {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	err = h.db.UpdateUserActivity(user.ID, req.IsActive)
	if err != nil {
		log.Printf("error in update user in handler /users/setIsActive: %v", err)
		serverError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type createPRRequest struct {
	PRID     string `json:"pull_request_id"`
	PRName   string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type createPRResponse struct {
	PR struct {
		PRID      string          `json:"pull_request_id"`
		PRName    string          `json:"pull_request_name"`
		AuthorID  string          `json:"author_id"`
		Status    models.PRStatus `json:"status"`
		Reviewers []string        `json:"assigned_reviewers"`
	} `json:"pr"`
}

func (h *HandlersRepo) CreatePR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req createPRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := h.db.GetPRByID(req.PRID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in create pr in handler /pullRequest/create: %v", err)
		serverError(w)
		return
	}
	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
		var e errorResponse
		e.Error.Code = "PR_EXISTS"
		e.Error.Message = fmt.Sprintf("PR with id=%s already exists", req.PRID)
		json.NewEncoder(w).Encode(e)
		return
	}

	user, _, err := h.db.GetUserWithTeamByID(req.AuthorID)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get user in handler /pullRequest/create: %v", err)
		serverError(w)
		return
	}

	reviewers, err := h.db.GetActiveUsersInTeamExcAuthor(user.GroupID, user.ID)
	if err != nil {
		log.Printf("error in get reviewers in handler /pullRequest/create: %v", err)
		serverError(w)
		return
	}

	pr := models.PullRequest{
		ID:        req.PRID,
		Name:      req.PRName,
		AuthorID:  req.AuthorID,
		Status:    models.PRStatusOpen,
		Reviewers: reviewers,
	}

	err = h.db.InsertPRInTransaction(pr)
	if err != nil {
		log.Printf("error in insert pr in handler /pullRequest/create: %v", err)
		serverError(w)
		return
	}

	var resp createPRResponse

	resp.PR.PRID, resp.PR.PRName, resp.PR.AuthorID, resp.PR.Status, resp.PR.Reviewers = pr.ID, pr.Name, pr.AuthorID, pr.Status, make([]string, 0, 2)
	for _, u := range pr.Reviewers {
		resp.PR.Reviewers = append(resp.PR.Reviewers, u.ID)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type mergePRRequest struct {
	PRID string `json:"pull_request_id"`
}

type mergePRResponse struct {
	PR struct {
		PRID      string          `json:"pull_request_id"`
		PRName    string          `json:"pull_request_name"`
		AuthorID  string          `json:"author_id"`
		Status    models.PRStatus `json:"status"`
		Reviewers []string        `json:"assigned_reviewers"`
		MergedAt  time.Time       `json:"mergedAt"`
	} `json:"pr"`
}

func (h *HandlersRepo) MergePR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req mergePRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pr, err := h.db.GetPRByID(req.PRID)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get pr in handler /pullRequest/merge: %v", err)
		serverError(w)
		return
	}

	var resp mergePRResponse
	resp.PR.PRID, resp.PR.PRName, resp.PR.AuthorID, resp.PR.Status = pr.ID, pr.Name, pr.AuthorID, pr.Status
	resp.PR.Reviewers, err = h.db.GetReviewersByPRID(pr.ID)
	if err != nil {
		log.Printf("error in get reviewers in handler /pullRequest/merge: %v", err)
		serverError(w)
		return
	}

	if pr.Status == models.PRStatusMerged {
		resp.PR.MergedAt = *pr.MergedAt
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.PR.MergedAt, err = h.db.SetMergedStatusPR(pr.ID)
	if err != nil {
		log.Printf("error in update status in handler /pullRequest/merge: %v", err)
		serverError(w)
		return
	}
	resp.PR.Status = models.PRStatusMerged

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type reassignPRRequest struct {
	PRID          string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
}

type reassignPRResponse struct {
	PR struct {
		PRID      string          `json:"pull_request_id"`
		PRName    string          `json:"pull_request_name"`
		AuthorID  string          `json:"author_id"`
		Status    models.PRStatus `json:"status"`
		Reviewers []string        `json:"assigned_reviewers"`
	} `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}

func (h *HandlersRepo) ReassignPR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req reassignPRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pr, err := h.db.GetPRByID(req.PRID)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get pr in handler /pullRequest/reassign: %v", err)
		serverError(w)
		return
	}

	if pr.Status == models.PRStatusMerged {
		w.WriteHeader(http.StatusConflict)
		var e errorResponse
		e.Error.Code = "PR_MERGED"
		e.Error.Message = "cannot reassign on merged PR"
		json.NewEncoder(w).Encode(e)
		return
	}

	_, err = h.db.GetUserByID(req.OldReviewerID)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get user in handler /pullRequest/reassign: %v", err)
		serverError(w)
		return
	}

	var resp reassignPRResponse
	resp.PR.PRID, resp.PR.PRName, resp.PR.AuthorID, resp.PR.Status = pr.ID, pr.Name, pr.AuthorID, pr.Status
	resp.PR.Reviewers, err = h.db.GetReviewersByPRID(pr.ID)

	if err != nil {
		log.Printf("error in get reviewers in handler /pullRequest/reassign: %v", err)
		serverError(w)
		return
	}
	if (len(resp.PR.Reviewers) == 0) || (len(resp.PR.Reviewers) == 1 && resp.PR.Reviewers[0] != req.OldReviewerID) ||
		(len(resp.PR.Reviewers) == 2 && resp.PR.Reviewers[0] != req.OldReviewerID && resp.PR.Reviewers[1] != req.OldReviewerID) {
		w.WriteHeader(http.StatusConflict)
		var e errorResponse
		e.Error.Code = "NOT_ASSIGNED"
		e.Error.Message = fmt.Sprintf("reviewer with id=%s is not assigned to PR with id=%s", req.OldReviewerID, req.PRID)
		json.NewEncoder(w).Encode(e)
		return
	}

	availableReviewerID, err := h.db.FoundAvailableReviewerPR(req.PRID, resp.PR.Reviewers, resp.PR.AuthorID)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusConflict)
		var e errorResponse
		e.Error.Code = "NO_CANDIDATE"
		e.Error.Message = "no active replacement candidate in team"
		json.NewEncoder(w).Encode(e)
		return
	}
	if err != nil {
		log.Printf("error in found available reviewer in handler /pullRequest/reassign: %v", err)
		serverError(w)
		return
	}

	err = h.db.SwapReviewerInPR(req.PRID, req.OldReviewerID, availableReviewerID)
	if err != nil {
		log.Printf("error in update reviewer in handler /pullRequest/reassign: %v", err)
		serverError(w)
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

type getReviewResponse struct {
	UserID string               `json:"user_id"`
	PR     []models.PullRequest `json:"pull_requests"`
}

func (h *HandlersRepo) GetReview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err := h.db.GetUserByID(userID)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("error in get user in handler /users/setIsActive: %v", err)
		serverError(w)
		return
	}

	prs, err := h.db.GetPRByReviewerID(userID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error in get user in handler /users/setIsActive: %v", err)
		serverError(w)
		return
	}

	var resp getReviewResponse
	resp.UserID = userID
	resp.PR = prs

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
