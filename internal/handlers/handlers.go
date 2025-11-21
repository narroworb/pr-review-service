package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/narroworb/pr-review-service/internal/models"
)

type DatabaseInterface interface {
	GetTeamByName(string) (models.Team, error)
	CreateTeam(string) (int64, error)
	GetUserByID(string) (models.User, error)
	CreateUser(models.User) error
	GetUsersInTeam(int64) ([]models.User, error)
	InsertTeamInTransaction(string, []models.User) error
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
	w.Header().Set("Content-Type", "application/json")
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
		w.Header().Set("Content-Type", "application/json")
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
			w.Header().Set("Content-Type", "application/json")
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
