package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"
	"github.com/narroworb/pr-review-service/internal/models"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error in open connection: %v", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("error in postgres ping: %v", err)
	}
	return &PostgresDB{
		db: conn,
	}, nil
}

func (p *PostgresDB) Close() {
	p.db.Close()
}

func (p *PostgresDB) RunMigrations() error {
	driver, err := postgres.WithInstance(p.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("error postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("error in init migrations: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error in apply migrations: %v", err)
	}

	return nil
}

func (p *PostgresDB) GetTeamByName(ctx context.Context, teamName string) (models.Team, error) {
	r := p.db.QueryRowContext(ctx, "SELECT team_id, name FROM teams WHERE name=$1", teamName)

	var team models.Team

	if err := r.Scan(&team.ID, &team.Name); err != nil {
		return models.Team{}, err
	}
	return team, nil
}

func (p *PostgresDB) CreateTeam(ctx context.Context, teamName string) (int64, error) {
	r := p.db.QueryRowContext(ctx, "INSERT INTO teams (name) VALUES ($1) RETURNING team_id", teamName)

	var teamID int64

	if err := r.Scan(&teamID); err != nil {
		return -1, err
	}
	return teamID, nil
}

func (p *PostgresDB) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	r := p.db.QueryRowContext(ctx, "SELECT user_id, name, is_active, team_id FROM users WHERE user_id=$1", userID)

	var user models.User

	if err := r.Scan(&user.ID, &user.Name, &user.IsActive, &user.GroupID); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (p *PostgresDB) CreateUser(ctx context.Context, user models.User) error {
	_, err := p.db.ExecContext(ctx, "INSERT INTO users (user_id, name, is_active, team_id) VALUES ($1, $2, $3, $4)", user.ID, user.Name, user.IsActive, user.GroupID)
	return err
}

func (p *PostgresDB) GetUsersInTeam(ctx context.Context, teamID int64) ([]models.User, error) {
	r, err := p.db.QueryContext(ctx, "SELECT user_id, name, is_active, team_id FROM users WHERE team_id=$1", teamID)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0, 2)

	for r.Next() {
		var user models.User

		if err := r.Scan(&user.ID, &user.Name, &user.IsActive, &user.GroupID); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (p *PostgresDB) InsertTeamInTransaction(ctx context.Context, teamName string, users []models.User) error {
	t, err := p.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	r := t.QueryRowContext(ctx, "INSERT INTO teams (name) VALUES ($1) RETURNING team_id", teamName)
	var teamID int64
	if err := r.Scan(&teamID); err != nil {
		t.Rollback()
		return err
	}

	for _, user := range users {
		_, err := t.ExecContext(ctx, "INSERT INTO users (user_id, name, is_active, team_id) VALUES ($1, $2, $3, $4)", user.ID, user.Name, user.IsActive, teamID)
		if err != nil {
			t.Rollback()
			return err
		}
	}

	return t.Commit()
}

func (p *PostgresDB) GetUserWithTeamByID(ctx context.Context, userID string) (models.User, string, error) {
	r := p.db.QueryRowContext(ctx, "SELECT user_id, users.name, is_active, users.team_id, teams.name FROM users INNER JOIN teams ON users.team_id=teams.team_id WHERE user_id=$1", userID)

	var user models.User
	var teamName string

	if err := r.Scan(&user.ID, &user.Name, &user.IsActive, &user.GroupID, &teamName); err != nil {
		return models.User{}, "", err
	}
	return user, teamName, nil
}

func (p *PostgresDB) UpdateUserActivity(ctx context.Context, userID string, isActive bool) error {
	_, err := p.db.ExecContext(ctx, "UPDATE users SET is_active=$1 WHERE user_id=$2", isActive, userID)
	return err
}

func (p *PostgresDB) GetPRByID(ctx context.Context, pRID string) (models.PullRequest, error) {
	r := p.db.QueryRowContext(ctx, "SELECT pr_id, name, author_id, pr_status, merged_at FROM pull_requests WHERE pr_id=$1", pRID)

	var pr models.PullRequest

	if err := r.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.MergedAt); err != nil {
		return models.PullRequest{}, err
	}
	return pr, nil
}

func (p *PostgresDB) GetActiveUsersInTeamExcAuthor(ctx context.Context, teamID int64, userID string) ([]models.User, error) {
	r, err := p.db.QueryContext(ctx, `WITH pr_count AS (SELECT reviewer_id, COUNT(*) AS cnt FROM pull_requests_reviewers GROUP BY reviewer_id)
 		SELECT user_id, name, is_active, team_id FROM users LEFT JOIN pr_count ON users.user_id=pr_count.reviewer_id
 		WHERE user_id!=$1 AND is_active AND team_id=$2 ORDER BY COALESCE(cnt, 0) LIMIT 2;`,
		userID, teamID)

	if err != nil {
		return nil, err
	}

	reviewers := make([]models.User, 0, 2)

	for r.Next() {
		var user models.User
		if err := r.Scan(&user.ID, &user.Name, &user.IsActive, &user.GroupID); err != nil {
			return nil, err
		}

		reviewers = append(reviewers, user)
	}

	return reviewers, nil
}

func (p *PostgresDB) InsertPRInTransaction(ctx context.Context, pr models.PullRequest) error {
	t, err := p.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	_, err = t.ExecContext(ctx, "INSERT INTO pull_requests (pr_id, name, author_id, pr_status) VALUES ($1, $2, $3, $4)", pr.ID, pr.Name, pr.AuthorID, pr.Status)
	if err != nil {
		t.Rollback()
		return err
	}

	for _, reviewer := range pr.Reviewers {
		_, err := t.ExecContext(ctx, "INSERT INTO pull_requests_reviewers (pr_id, reviewer_id) VALUES ($1, $2)", pr.ID, reviewer.ID)
		if err != nil {
			t.Rollback()
			return err
		}
	}

	return t.Commit()
}

func (p *PostgresDB) GetReviewersByPRID(ctx context.Context, pRID string) ([]string, error) {
	r, err := p.db.QueryContext(ctx, "SELECT reviewer_id FROM pull_requests_reviewers WHERE pr_id=$1", pRID)
	if err != nil {
		return nil, err
	}
	reviewersID := make([]string, 0, 2)

	for r.Next() {
		var userID string
		if err := r.Scan(&userID); err != nil {
			return nil, err
		}
		reviewersID = append(reviewersID, userID)
	}

	return reviewersID, nil
}

func (p *PostgresDB) SetMergedStatusPR(ctx context.Context, pRID string) (time.Time, error) {
	r := p.db.QueryRowContext(ctx, `UPDATE pull_requests SET pr_status=$1, merged_at=NOW() WHERE pr_id=$2 RETURNING merged_at`, models.PRStatusMerged, pRID)

	var mergedAt time.Time
	if err := r.Scan(&mergedAt); err != nil {
		return time.Time{}, err
	}

	return mergedAt, nil
}

func (p *PostgresDB) FoundAvailableReviewerPR(ctx context.Context, pRID string, reviewersID []string, authorID string) (string, error) {
	r := p.db.QueryRowContext(ctx, `WITH team AS (SELECT team_id FROM users u INNER JOIN pull_requests pr ON pr.author_id=u.user_id WHERE pr.pr_id=$1),
	pr_count AS (SELECT reviewer_id, COUNT(*) AS cnt FROM pull_requests_reviewers GROUP BY reviewer_id)
	SELECT user_id FROM users u INNER JOIN team t ON u.team_id=t.team_id 
	LEFT JOIN pr_count prc ON prc.reviewer_id=u.user_id 
	WHERE is_active AND user_id != ALL($2) AND user_id != $3 ORDER BY cnt LIMIT 1`, pRID, pq.Array(reviewersID), authorID)

	var reviewerID string
	if err := r.Scan(&reviewerID); err != nil {
		return "", err
	}
	return reviewerID, nil
}

func (p *PostgresDB) SwapReviewerInPR(ctx context.Context, pRID, oldReviewerID, newReviewerID string) error {
	_, err := p.db.ExecContext(ctx, `UPDATE pull_requests_reviewers SET reviewer_id=$1 WHERE pr_id=$2 AND reviewer_id=$3`, newReviewerID, pRID, oldReviewerID)
	return err
}

func (p *PostgresDB) GetPRByReviewerID(ctx context.Context, reviewerID string) ([]models.PullRequest, error) {
	r, err := p.db.QueryContext(ctx, `SELECT pr.pr_id, name, author_id, pr_status FROM pull_requests_reviewers prr 
	INNER JOIN pull_requests pr 
	ON prr.pr_id=pr.pr_id
	WHERE reviewer_id=$1`, reviewerID)
	if err != nil {
		return nil, err
	}
	pullrequests := make([]models.PullRequest, 0, 1)

	for r.Next() {
		var pr models.PullRequest
		if err := r.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		pullrequests = append(pullrequests, pr)
	}

	return pullrequests, nil
}

func (p *PostgresDB) GetCountPRStatsByUser(ctx context.Context) ([]models.UserStats, error) {
	r, err := p.db.QueryContext(ctx,
		`SELECT u.user_id, COALESCE(a.cnt_author, 0) AS cnt_author,
			COALESCE(r.cnt_reviewer, 0) AS cnt_reviewer
		FROM users u
		LEFT JOIN (
			SELECT author_id, COUNT(*) AS cnt_author
			FROM pull_requests
			GROUP BY author_id
		) a ON u.user_id = a.author_id
		LEFT JOIN (
			SELECT reviewer_id, COUNT(*) AS cnt_reviewer
			FROM pull_requests_reviewers
			GROUP BY reviewer_id
		) r ON u.user_id = r.reviewer_id
		ORDER BY cnt_author DESC;
	`)
	if err != nil {
		return nil, err
	}
	stats := make([]models.UserStats, 0, 10)

	for r.Next() {
		var s models.UserStats
		if err := r.Scan(&s.UserID, &s.PRAuthorCount, &s.PRReviewerCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (p *PostgresDB) GetCountPRStatsByTeam(ctx context.Context) ([]models.TeamStats, error) {
	r, err := p.db.QueryContext(ctx,
		`SELECT
			t.name,
			COUNT(DISTINCT u.user_id) AS users_count,
			COUNT(pr.pr_id) AS total_pr,
			COUNT(pr.pr_id) FILTER (WHERE pr.pr_status = 'OPEN')  AS open_pr,
			COUNT(pr.pr_id) FILTER (WHERE pr.pr_status = 'MERGED') AS merged_pr
		FROM teams t
		LEFT JOIN users u ON u.team_id = t.team_id
		LEFT JOIN pull_requests pr ON pr.author_id = u.user_id
		GROUP BY t.name
		ORDER BY t.name;
		`)
	if err != nil {
		return nil, err
	}
	stats := make([]models.TeamStats, 0, 10)

	for r.Next() {
		var s models.TeamStats
		if err := r.Scan(&s.TeamName, &s.UsersCount, &s.AllPRCount, &s.OpenPRCount, &s.MergedPRCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

func (p *PostgresDB) GetCountReviewerStatsByPR(ctx context.Context) (map[string]int64, error) {
	r, err := p.db.QueryContext(ctx,
		`SELECT
			pr_id,
			COUNT(*) AS reviewers_count
		FROM pull_requests_reviewers
		GROUP BY pr_id
		ORDER BY pr_id;
		`)
	if err != nil {
		return nil, err
	}
	stats := make(map[string]int64)

	for r.Next() {
		var pr_id string
		var count int64
		if err := r.Scan(&pr_id, &count); err != nil {
			return nil, err
		}
		stats[pr_id] = count
	}

	return stats, nil
}
