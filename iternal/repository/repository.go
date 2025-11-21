package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/lib/pq"
)

var ErrTeamExists = errors.New("team already exists")
var ErrTeamNotFound = errors.New("team not found")

type UserRepo interface {
	UpdateActive(ctx context.Context, userID string, isActive bool) (*api.User, error)
	TeamAdd(ctx context.Context, teamName string, teamMembers []api.TeamMember) (*api.Team, error)
	GetTeam(ctx context.Context, teamName string) (*api.Team, error)
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) TeamAdd(ctx context.Context, teamName string, teamMembers []api.TeamMember) (*api.Team, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `insert into team (name) values ($1)`, teamName)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrTeamExists
		}
		return nil, err
	}

	for _, member := range teamMembers {
		_, err := tx.ExecContext(ctx,
			`insert into users (id, name, is_active) values ($1, $2, $3)
		on conflict (id) do update set is_active = EXCLUDED.is_active`,
		 member.UserId, member.Username, member.IsActive)
		if err != nil {
			return nil, err
		}
	}

	for _, member := range teamMembers {
		_, err := tx.ExecContext(ctx,
			`insert into user_teams (user_id, team_name) values ($1, $2)
		on conflict do nothing`, member.UserId, teamName)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &api.Team{
		TeamName: teamName,
		Members:  teamMembers,
	}, nil
}

func (r *UserRepository) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	query := `
	select u.id, u.name, u.is_active
	from user_teams ut
	join users u on ut.user_id = u.id
	where ut.team_name = $1
	`

	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []api.TeamMember
	for rows.Next() {
		var m api.TeamMember
		if err := rows.Scan(&m.UserId, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if len(members) == 0 {
		return nil, ErrTeamNotFound
	}

	team := &api.Team{
    	TeamName: teamName,
    	Members:  members,
	}
	return team, nil
}

func (r *UserRepository) UpdateActive(ctx context.Context, userID string, isActive bool) (*api.User, error) {

	var user api.User

	query := `
		SELECT 
			u.id,
			u.name,
			u.is_active,
			COALESCE(ut.team_name, '')
		FROM users u
		LEFT JOIN user_teams ut ON ut.user_id = u.id
		WHERE u.id = $1
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserId,
		&user.Username,
		&user.IsActive,
		&user.TeamName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	_, err = r.db.ExecContext(ctx, `UPDATE users SET is_active = $1 WHERE id = $2`, isActive, userID)
	if err != nil {
		return nil, err
	}

	user.IsActive = isActive

	return &user, nil
}
