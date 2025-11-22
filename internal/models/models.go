package models

import "time"

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type User struct {
	ID       string `json:"user_id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	GroupID  int64  `json:"-"`
}

type Team struct {
	ID   int64
	Name string
}

type PullRequest struct {
	ID        string    `json:"pull_request_id"`
	Name      string    `json:"pull_request_name"`
	AuthorID  string    `json:"author_id"`
	Status    PRStatus  `json:"status"`
	Reviewers []User    `json:"-"`
	MergedAt  *time.Time `json:"-"`
}
