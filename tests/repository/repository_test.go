package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/repository"
	"github.com/lib/pq"
)

func TestUserRepository_TeamAdd_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	teamName := "backend"
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
		{UserId: "u2", Username: "Bob", IsActive: true},
	}

	mock.ExpectBegin()

	mock.ExpectExec("(?i)INSERT INTO team").
		WithArgs(teamName).
		WillReturnResult(sqlmock.NewResult(1, 1))

	for _, m := range members {
		mock.ExpectExec("(?i)INSERT INTO users").
			WithArgs(m.UserId, m.Username, m.IsActive).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	for _, m := range members {
		mock.ExpectExec("(?i)INSERT INTO user_teams").
			WithArgs(m.UserId, teamName).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectCommit()

	team, err := repo.TeamAdd(ctx, teamName, members)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if team.TeamName != teamName || len(team.Members) != len(members) {
		t.Errorf("unexpected team result: %+v", team)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_TeamAdd_Exists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	teamName := "backend"
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
	}

	mock.ExpectBegin()
	mock.ExpectExec("(?i)INSERT INTO team").
		WithArgs(teamName).
		WillReturnError(&pq.Error{Code: "23505"})
	mock.ExpectRollback()

	_, err = repo.TeamAdd(ctx, teamName, members)
	if !errors.Is(err, repository.ErrTeamExists) {
		t.Fatalf("expected ErrTeamExists, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_TeamAdd_UserInsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	teamName := "backend"
	members := []api.TeamMember{
		{UserId: "u1", Username: "Alice", IsActive: true},
	}

	mock.ExpectBegin()

	mock.ExpectExec("(?i)INSERT INTO team").
		WithArgs(teamName).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("(?i)INSERT INTO users").
		WithArgs(members[0].UserId, members[0].Username, members[0].IsActive).
		WillReturnError(errors.New("db error"))

	mock.ExpectRollback()

	_, err = repo.TeamAdd(ctx, teamName, members)
	if err == nil || err.Error() != "db error" {
		t.Fatalf("expected db error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_UpdateActive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	userID := "123"
	isActive := true

	rows := sqlmock.NewRows([]string{"id", "name", "is_active", "team_name"}).
		AddRow(userID, "testuser", false, "devteam")

	mock.ExpectQuery("SELECT .* FROM users u").
		WithArgs(userID).
		WillReturnRows(rows)

	mock.ExpectExec("UPDATE users SET is_active").
		WithArgs(isActive, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user, err := repo.UpdateActive(ctx, userID, isActive)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.UserId != userID || !user.IsActive {
		t.Errorf("unexpected user result: %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_UpdateActive_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	userID := "notfound"

	mock.ExpectQuery("SELECT .* FROM users u").
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.UpdateActive(ctx, userID, true)
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_GetTeam(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	defer db.Close()

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	teamName := "backend"

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name", "is_active"}).
			AddRow("u1", "Alice", true).
			AddRow("u2", "Bob", true)

		mock.ExpectQuery("(?i)SELECT .* FROM user_teams").
			WithArgs(teamName).
			WillReturnRows(rows)

		team, err := repo.GetTeam(ctx, teamName)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if team.TeamName != teamName {
			t.Errorf("expected team_name %s, got %s", teamName, team.TeamName)
		}
		if len(team.Members) != 2 {
			t.Errorf("expected 2 members, got %d", len(team.Members))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("(?i)SELECT .* FROM user_teams").
			WithArgs("missing").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_active"}))

		_, err := repo.GetTeam(ctx, "missing")
		if !errors.Is(err, repository.ErrTeamNotFound) {
			t.Fatalf("expected ErrTeamNotFound, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("(?i)SELECT .* FROM user_teams").
			WithArgs("backend").
			WillReturnError(errors.New("db error"))

		_, err := repo.GetTeam(ctx, "backend")
		if err == nil || err.Error() != "db error" {
			t.Fatalf("expected db error, got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
}

