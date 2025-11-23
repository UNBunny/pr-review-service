CREATE TABLE teams (
    name VARCHAR(255) PRIMARY KEY
);

CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL REFERENCES teams(name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE pull_requests (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP
);

CREATE TABLE pull_request_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (pull_request_id, user_id)
);

CREATE INDEX idx_users_team ON users(team_name);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_pr_author ON pull_requests(author_id);
CREATE INDEX idx_pr_status ON pull_requests(status);
CREATE INDEX idx_pr_reviewers_user ON pull_request_reviewers(user_id);



-- Constraint: не более 2 ревьюеров на PR
CREATE OR REPLACE FUNCTION check_reviewers_count()
RETURNS TRIGGER AS $$
BEGIN
    IF (SELECT COUNT(*) FROM pull_request_reviewers WHERE pull_request_id = NEW.pull_request_id) >= 2 THEN
        RAISE EXCEPTION 'PR cannot have more than 2 reviewers';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_max_reviewers
    BEFORE INSERT ON pull_request_reviewers
    FOR EACH ROW
    EXECUTE FUNCTION check_reviewers_count();