CREATE SCHEMA IF NOT EXISTS pr_review;

CREATE TABLE IF NOT EXISTS pr_review.team (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pr_review.user (
    id   VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    team_id BIGINT REFERENCES pr_review.team(id) ON DELETE SET NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pr_review.pull_request (
    id BIGSERIAL PRIMARY KEY,
    pull_request_id VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES pr_review.user(id) ON DELETE RESTRICT,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    reviewers TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_user_team_id ON pr_review.user(team_id);
CREATE INDEX IF NOT EXISTS idx_user_is_active ON pr_review.user(is_active);
CREATE INDEX IF NOT EXISTS idx_pull_request_pull_request_id ON pr_review.pull_request(pull_request_id);

