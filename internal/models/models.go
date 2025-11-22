package models

import "time"

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type User struct {
	ID       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
	GroupID  int64  `json:"-"`
}

type Team struct {
	ID   int64
	Name string
}

type PullRequest struct {
	ID        string     `json:"pull_request_id"`
	Name      string     `json:"pull_request_name"`
	AuthorID  string     `json:"author_id"`
	Status    PRStatus   `json:"status"`
	Reviewers []User     `json:"-"`
	MergedAt  *time.Time `json:"-"`
}

type UserStats struct {
	UserID          string `json:"user_id"`
	PRReviewerCount int64  `json:"count_pr_reviewer"`
	PRAuthorCount   int64  `json:"count_pr_author"`
}

type TeamStats struct {
	TeamName      string `json:"team_name"`
	UsersCount    int64  `json:"users_count"`
	AllPRCount    int64  `json:"all_pr_count"`
	MergedPRCount int64  `json:"merged_pr_count"`
	OpenPRCount   int64  `json:"open_pr_count"`
}
