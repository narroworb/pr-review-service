package models

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
	ID        string
	Name      string
	AuthorID  int64
	Status    PRStatus
	Reviewers []User
}
