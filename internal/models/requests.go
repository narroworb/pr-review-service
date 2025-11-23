package models

type AddTeamRequest struct {
	TeamName string `json:"team_name"`
	Members  []struct {
		UserID   string `json:"user_id"`
		Name     string `json:"username"`
		IsActive bool   `json:"is_active"`
	} `json:"members"`
}

type SetUserIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type CreatePRRequest struct {
	PRID     string `json:"pull_request_id"`
	PRName   string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

type MergePRRequest struct {
	PRID string `json:"pull_request_id"`
}

type ReassignPRRequest struct {
	PRID          string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
}

type DeactivateAllUsersInTeamRequest struct {
	TeamName string `json:"team_name"`
}

type DeactivateUsersByIDRequest struct {
	UserNames []string `json:"user_names"`
}
