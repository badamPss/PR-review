CREATE SCHEMA IF NOT EXISTS pr_review;

CREATE TABLE IF NOT EXISTS pr_review.teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pr_review.users (
    id   VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    team_id BIGINT REFERENCES pr_review.teams(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pr_review.pull_requests (
    id BIGSERIAL PRIMARY KEY,
    pull_request_id VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES pr_review.users(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    reviewers TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_users_team_id ON pr_review.users(team_id);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON pr_review.users(is_active);
CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pr_review.pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pr_review.pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_pull_requests_pull_request_id ON pr_review.pull_requests(pull_request_id);

