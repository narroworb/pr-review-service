CREATE TABLE IF NOT EXISTS teams (
    team_id SERIAL PRIMARY KEY,
    name VARCHAR(100)
);

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(100),
    is_active BOOLEAN,
    team_id INT REFERENCES teams(team_id)
);

CREATE TABLE IF NOT EXISTS pull_requests (
    pr_id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(100),
    author_id VARCHAR(100) REFERENCES users(user_id),
    pr_status VARCHAR(6) DEFAULT 'OPEN'
);

CREATE TABLE IF NOT EXISTS pull_requests_reviewers (
    pr_id VARCHAR(100) REFERENCES pull_requests(pr_id),
    reviewer_id VARCHAR(100) REFERENCES users(user_id)
);