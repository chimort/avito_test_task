INSERT INTO team (name)
VALUES ('backend')
ON CONFLICT (name) DO NOTHING;

INSERT INTO users (id, name, is_active)
VALUES (123, 'test_user', false)
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_teams (user_id, team_name)
VALUES (123, 'backend')
ON CONFLICT (user_id, team_name) DO NOTHING;
