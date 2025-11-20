package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chimort/avito_test_task/iternal/api"
)

type UserRepo interface {
	UpdateActive(ctx context.Context, userID string, isActive bool) (*api.User, error)
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
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
