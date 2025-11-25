CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(500) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_pr_author_id ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user_id ON pr_reviewers(user_id);
