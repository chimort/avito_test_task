package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/chimort/avito_test_task/iternal/repository"
	"github.com/lib/pq"
)

var globalTime = time.Now()

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *repository.UserRepository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	repo := repository.NewUserRepository(db)
	return db, mock, repo
}

func TestUserRepository_TeamAdd(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
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
	})

	t.Run("exists", func(t *testing.T) {
		teamName := "backend"
		members := []api.TeamMember{{UserId: "u1", Username: "Alice", IsActive: true}}
		mock.ExpectBegin()
		mock.ExpectExec("(?i)INSERT INTO team").
			WithArgs(teamName).
			WillReturnError(&pq.Error{Code: "23505"})
		mock.ExpectRollback()

		_, err := repo.TeamAdd(ctx, teamName, members)
		if !errors.Is(err, repository.ErrTeamExists) {
			t.Fatalf("expected ErrTeamExists, got %v", err)
		}
	})
}

func TestUserRepository_UpdateActive(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := "123"
		isActive := true

		rows := sqlmock.NewRows([]string{"id", "name", "is_active", "team_name"}).
			AddRow(userID, "testuser", false, "devteam")
		mock.ExpectQuery("SELECT .* FROM users u").WithArgs(userID).WillReturnRows(rows)
		mock.ExpectExec("UPDATE users SET is_active").WithArgs(isActive, userID).WillReturnResult(sqlmock.NewResult(1, 1))

		user, err := repo.UpdateActive(ctx, userID, isActive)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !user.IsActive {
			t.Errorf("expected active true, got false")
		}
	})

	t.Run("not found", func(t *testing.T) {
		userID := "notfound"
		mock.ExpectQuery("SELECT .* FROM users u").WithArgs(userID).WillReturnError(sql.ErrNoRows)
		_, err := repo.UpdateActive(ctx, userID, true)
		if err != sql.ErrNoRows {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
	})
}

func TestUserRepository_GetTeam(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		teamName := "backend"
		rows := sqlmock.NewRows([]string{"id", "name", "is_active"}).
			AddRow("u1", "Alice", true).
			AddRow("u2", "Bob", true)
		mock.ExpectQuery("(?i)SELECT .* FROM user_teams").WithArgs(teamName).WillReturnRows(rows)

		team, err := repo.GetTeam(ctx, teamName)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(team.Members) != 2 {
			t.Errorf("expected 2 members, got %d", len(team.Members))
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("(?i)SELECT .* FROM user_teams").WithArgs("missing").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "is_active"}))
		_, err := repo.GetTeam(ctx, "missing")
		if !errors.Is(err, repository.ErrTeamNotFound) {
			t.Fatalf("expected ErrTeamNotFound, got %v", err)
		}
	})
}

func TestUserRepository_PullRequestCreate(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("insert into pull_requests").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("select u.id from users u join user_teams").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("u2"))
		mock.ExpectExec("insert into pr_reviewers").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		pr, err := repo.PullRequestCreate(ctx, "pr1", "Test PR", "u1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if pr.PullRequestId != "pr1" {
			t.Errorf("unexpected PR %+v", pr)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("insert into pull_requests").WillReturnResult(sqlmock.NewResult(1, 0))
		mock.ExpectRollback()

		_, err := repo.PullRequestCreate(ctx, "pr2", "Test PR", "missing")
		if err != repository.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func TestUserRepository_PullRequestMerge(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("merge success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("select status from pull_requests").WithArgs("pr1").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("OPEN"))
		mock.ExpectExec("update pull_requests set status = 'MERGED'").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("select id, title, author_id, status, created_at, merged_at from pull_requests").
			WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author_id", "status", "created_at", "merged_at"}).
				AddRow("pr1", "Test PR", "u1", "MERGED", globalTime, globalTime))
		mock.ExpectQuery("select reviewer_id from pr_reviewers").WillReturnRows(sqlmock.NewRows([]string{"reviewer_id"}).AddRow("u2"))
		mock.ExpectCommit()

		pr, err := repo.PullRequestMerge(ctx, "pr1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if pr.Status != "MERGED" {
			t.Errorf("expected MERGED, got %v", pr.Status)
		}
	})
}

func TestUserRepository_PullRequestReassign(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("select author_id, status, title, created_at from pull_requests").
			WithArgs("pr1").
			WillReturnRows(sqlmock.NewRows([]string{"author_id", "status", "title", "created_at"}).
				AddRow("u1", "OPEN", "Test PR", globalTime))
		mock.ExpectQuery("select 1 from pr_reviewers").
			WithArgs("pr1", "u2").
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
		mock.ExpectQuery("select u.id from users u join user_teams").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("u3"))
		mock.ExpectExec("delete from pr_reviewers").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("insert into pr_reviewers").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("select reviewer_id from pr_reviewers").WillReturnRows(sqlmock.NewRows([]string{"reviewer_id"}).AddRow("u3"))
		mock.ExpectCommit()

		pr, newReviewer, err := repo.PullRequestReassign(ctx, "pr1", "u2")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if newReviewer != "u3" {
			t.Errorf("unexpected new reviewer: %v", newReviewer)
		}
		if len(pr.AssignedReviewers) != 1 {
			t.Errorf("unexpected assigned reviewers: %+v", pr.AssignedReviewers)
		}
	})
}

func TestUserRepository_GetPRsByReviewer(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()
	ctx := context.Background()

	mock.ExpectQuery("select pr.id, pr.title, pr.author_id, pr.status from pull_requests pr join pr_reviewers prr").
		WithArgs("u2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "author_id", "status"}).AddRow("pr1", "Test PR", "u1", "OPEN"))

	prs, err := repo.GetPRsByReviewer(ctx, "u2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(prs) != 1 || prs[0].PullRequestId != "pr1" {
		t.Errorf("unexpected PRs: %+v", prs)
	}
}
