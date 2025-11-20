package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chimort/avito_test_task/iternal/repository"
)

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
