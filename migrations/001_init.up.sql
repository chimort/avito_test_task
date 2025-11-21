CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true
);

create table if not exists team (
    name text not null unique
);

create table if not exists user_teams(
    user_id text not null references users(id) on delete cascade,
    team_name text not null references team(name) on delete cascade,
    PRIMARY KEY(user_id, team_name)
);

create table if not exists pull_requests (
    id text PRIMARY KEY,
    title varchar(100) not null,
    author_id text not null references users(id) on delete cascade,
    status varchar(50) not null DEFAULT 'OPEN' check(status in ('OPEN', 'MERGED'))
);

create table if not exists pr_reviewers (
    pr_id text NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    reviewer_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY(pr_id, reviewer_id)
);