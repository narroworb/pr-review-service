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
		Members []User
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
