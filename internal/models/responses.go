package models

import (
	"time"
)

type AddTeamResponse struct {
	Team AddTeamRequest `json:"team"`
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type GetTeamResponse struct {
	Team struct {
		Name    string `json:"team_name"`
		Members []User `json:"members"`
	} `json:"team"`
}

type SetUserIsActiveResponse struct {
	User struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		TeamName string `json:"team_name"`
		IsActive bool   `json:"is_active"`
	} `json:"user"`
}

type CreatePRResponse struct {
	PR struct {
		PRID      string   `json:"pull_request_id"`
		PRName    string   `json:"pull_request_name"`
		AuthorID  string   `json:"author_id"`
		Status    PRStatus `json:"status"`
		Reviewers []string `json:"assigned_reviewers"`
	} `json:"pr"`
}

type MergePRResponse struct {
	PR struct {
		PRID      string    `json:"pull_request_id"`
		PRName    string    `json:"pull_request_name"`
		AuthorID  string    `json:"author_id"`
		Status    PRStatus  `json:"status"`
		Reviewers []string  `json:"assigned_reviewers"`
		MergedAt  time.Time `json:"mergedAt"`
	} `json:"pr"`
}

type ReassignPRResponse struct {
	PR struct {
		PRID      string   `json:"pull_request_id"`
		PRName    string   `json:"pull_request_name"`
		AuthorID  string   `json:"author_id"`
		Status    PRStatus `json:"status"`
		Reviewers []string `json:"assigned_reviewers"`
	} `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}

type GetReviewResponse struct {
	UserID string        `json:"user_id"`
	PR     []PullRequest `json:"pull_requests"`
}

type GetStatsUsersResponse struct {
	PRStats []UserStats `json:"statistic"`
}

type GetStatsTeamsResponse struct {
	PRStats []TeamStats `json:"statistic"`
}

type GetStatsPRsResponse struct {
	PRStats map[string]int64 `json:"statistic_count_reviewers"`
}

type DeactivateAllUsersInTeamResponse struct {
	TeamName string `json:"team_name"`
	Users    []User `json:"users"`
}

type DeactivateUsersByIDResponse struct {
	Users         []User   `json:"users"`
	NotFoundUsers []string `json:"not_found_users"`
}
