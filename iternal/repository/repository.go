package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/chimort/avito_test_task/iternal/api"
	"github.com/lib/pq"
)

type UserRepo interface {
	UpdateActive(ctx context.Context, userID string, isActive bool) (*api.User, error)
	TeamAdd(ctx context.Context, teamName string, teamMembers []api.TeamMember) (*api.Team, error)
	GetTeam(ctx context.Context, teamName string) (*api.Team, error)
	PullRequestCreate(ctx context.Context, pullRequestId string, pullRequestName string, authorId string) (*api.PullRequest, error)
	PullRequestMerge(ctx context.Context, pullRequestId string) (*api.PullRequest, error)
	PullRequestReassign(ctx context.Context, pullRequestId string, oldUserId string) (*api.PullRequest, string, error)
	GetPRsByReviewer(ctx context.Context, reviewerId string) ([]*api.PullRequestShort, error)
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
	defer func() {
		_ = tx.Rollback()
	}()

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
	defer func() { _ = rows.Close() }()

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

func (r *UserRepository) PullRequestCreate(ctx context.Context, pullRequestId string, pullRequestName string, authorId string) (*api.PullRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	res, err := tx.ExecContext(ctx,
		`insert into pull_requests (id, title, author_id)
		select $1, $2, id from users where id = $3`,
		pullRequestId, pullRequestName, authorId)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrPRExists
		}
		return nil, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrUserNotFound
	}

	rows, err := tx.QueryContext(ctx,
		`select u.id
		from users u
		join user_teams ut on u.id = ut.user_id
		where ut.team_name = (
		select team_name from user_teams where user_id = $1 limit 1
		) and u.id <> $1 and u.is_active = true
		order by random()
		limit 2`,
		authorId)

	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var reviewerIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewerIDs = append(reviewerIDs, id)
	}

	for _, reviewerID := range reviewerIDs {
		if _, err := tx.ExecContext(ctx, `
			insert into pr_reviewers(pr_id, reviewer_id)
			values ($1, $2)
			on conflict do nothing	
		`, pullRequestId, reviewerID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	pr := &api.PullRequest{
		PullRequestId:     pullRequestId,
		PullRequestName:   pullRequestName,
		AuthorId:          authorId,
		AssignedReviewers: reviewerIDs,
		CreatedAt:         func() *time.Time { t := time.Now(); return &t }(),
		MergedAt:          nil,
		Status:            "OPEN",
	}
	return pr, nil
}

func (r *UserRepository) PullRequestMerge(ctx context.Context, pullRequestId string) (*api.PullRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var currentStatus string
	err = tx.QueryRowContext(ctx,
		`select status from pull_requests where id = $1 for update`, pullRequestId).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPRNotFound
		}
		return nil, err
	}

	if currentStatus == "MERGED" {
		var pr api.PullRequest
		err = tx.QueryRowContext(ctx,
			`select id, title, author_id, status, created_at, merged_at from pull_requests where id = $1`,
			pullRequestId,
		).Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
		if err != nil {
			return nil, err
		}

		rows, err := tx.QueryContext(ctx,
			`select reviewer_id from pr_reviewers where pr_id = $1`, pullRequestId)
		if err != nil {
			return nil, err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			pr.AssignedReviewers = append(pr.AssignedReviewers, id)
		}

		return &pr, nil
	}

	mergedAt := time.Now()
	_, err = tx.ExecContext(ctx,
		`update pull_requests set status = 'MERGED', merged_at = $2 where id = $1`,
		pullRequestId, mergedAt)
	if err != nil {
		return nil, err
	}

	var pr api.PullRequest
	err = tx.QueryRowContext(ctx,
		`select id, title, author_id, status, created_at, merged_at from pull_requests where id = $1`,
		pullRequestId,
	).Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx,
		`select reviewer_id from pr_reviewers where pr_id = $1`, pullRequestId)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *UserRepository) PullRequestReassign(ctx context.Context, pullRequestId string, oldUserId string) (*api.PullRequest, string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var authorId, status, pullRequestName string
	var createdAt time.Time

	err = tx.QueryRowContext(ctx,
		`select author_id, status, title, created_at
		from pull_requests
		where id = $1`,
		pullRequestId,
	).Scan(&authorId, &status, &pullRequestName, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrPRNotFound
		}
		return nil, "", err
	}

	if status == "MERGED" {
		return nil, "", ErrPRMerged
	}

	var reviewerExists int
	err = tx.QueryRowContext(ctx,
		`select 1 
		 from pr_reviewers 
		 where pr_id = $1 and reviewer_id = $2`,
		pullRequestId, oldUserId,
	).Scan(&reviewerExists)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrReviewerNotAssign
		}
		return nil, "", err
	}

	var newReviewer string
	err = tx.QueryRowContext(ctx,
		`select u.id
		from users u
		join user_teams ut on ut.user_id = u.id
		where ut.team_name = (
				select team_name
				from user_teams
				where user_id = $2
				limit 1
		)
		and u.is_active = true
		and u.id <> $2
		and u.id <> $3
		and u.id not in (
				select reviewer_id
				from pr_reviewers
				where pr_id = $1
		)
		order by random()
		limit 1`,
		pullRequestId, oldUserId, authorId,
	).Scan(&newReviewer)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", ErrNoCandidates
		}
		return nil, "", err
	}

	_, err = tx.ExecContext(ctx,
		`delete from pr_reviewers 
		 where pr_id = $1 and reviewer_id = $2`,
		pullRequestId, oldUserId,
	)
	if err != nil {
		return nil, "", err
	}

	_, err = tx.ExecContext(ctx,
		`insert into pr_reviewers(pr_id, reviewer_id) 
		 values ($1, $2)`,
		pullRequestId, newReviewer,
	)
	if err != nil {
		return nil, "", err
	}

	rows, err := tx.QueryContext(ctx,
		`select reviewer_id 
		 from pr_reviewers 
		 where pr_id = $1`,
		pullRequestId,
	)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = rows.Close() }()

	var reviewers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, "", err
		}
		reviewers = append(reviewers, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, "", err
	}

	pr := &api.PullRequest{
		PullRequestId:     pullRequestId,
		PullRequestName:   pullRequestName,
		AuthorId:          authorId,
		AssignedReviewers: reviewers,
		Status:            api.PullRequestStatus(status),
		CreatedAt:         &createdAt,
		MergedAt:          nil,
	}

	return pr, newReviewer, nil
}

func (r *UserRepository) GetPRsByReviewer(ctx context.Context, reviewerId string) ([]*api.PullRequestShort, error) {
	rows, err := r.db.QueryContext(ctx,
		`select pr.id, pr.title, pr.author_id, pr.status
		 from pull_requests pr
		 join pr_reviewers prr on prr.pr_id = pr.id
		 where prr.reviewer_id = $1`, reviewerId)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var prs []*api.PullRequestShort
	for rows.Next() {
		var pr api.PullRequestShort
		var status string
		if err := rows.Scan(&pr.PullRequestId, &pr.PullRequestName, &pr.AuthorId, &status); err != nil {
			return nil, err
		}
		pr.Status = api.PullRequestShortStatus(status)
		prs = append(prs, &pr)
	}
	return prs, nil
}
