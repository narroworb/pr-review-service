package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

func (p *PostgresDB) GetTeamByName(teamName string) (models.Team, error) {
	r := p.db.QueryRow("SELECT team_id, name FROM teams WHERE name=$1", teamName)

	var team models.Team

	if err := r.Scan(&team.ID, &team.Name); err != nil {
		return models.Team{}, err
	}
	return team, nil
}

func (p *PostgresDB) CreateTeam(teamName string) (int64, error) {
	r := p.db.QueryRow("INSERT INTO teams (name) VALUES ($1) RETURNING team_id", teamName)

	var teamID int64

	if err := r.Scan(&teamID); err != nil {
		return -1, err
	}
	return teamID, nil
}

func (p *PostgresDB) GetUserByID(userID string) (models.User, error) {
	r := p.db.QueryRow("SELECT user_id, name, is_active, team_id FROM users WHERE user_id=$1", userID)

	var user models.User

	if err := r.Scan(&user.ID, &user.Name, &user.IsActive, &user.GroupID); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (p *PostgresDB) CreateUser(user models.User) error {
	_, err := p.db.Exec("INSERT INTO users (user_id, name, is_active, team_id) VALUES ($1, $2, $3, $4)", user.ID, user.Name, user.IsActive, user.GroupID)
	return err
}

func (p *PostgresDB) GetUsersInTeam(teamID int64) ([]models.User, error) {
	r, err := p.db.Query("SELECT user_id, name, is_active, team_id FROM users WHERE team_id=$1", teamID)
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

func (p *PostgresDB) InsertTeamInTransaction(teamName string, users []models.User) error {
	t, err := p.db.Begin()
	if err != nil {
		return err
	}

	r := t.QueryRow("INSERT INTO teams (name) VALUES ($1) RETURNING team_id", teamName)
	var teamID int64
	if err := r.Scan(&teamID); err != nil {
		t.Rollback()
		return err
	}

	for _, user := range users {
		_, err := t.Exec("INSERT INTO users (user_id, name, is_active, team_id) VALUES ($1, $2, $3, $4)", user.ID, user.Name, user.IsActive, teamID)
		if err != nil {
			t.Rollback()
			return err
		}
	}

	return t.Commit()
}
