-- jobs: every listing we fetch, deduplicated by hash
CREATE TABLE jobs (
    id              BIGSERIAL PRIMARY KEY,
    title           TEXT NOT NULL,
    company         TEXT NOT NULL,
    source          TEXT NOT NULL,
    url             TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    description_hash TEXT NOT NULL DEFAULT '',
    location        TEXT NOT NULL DEFAULT '',
    remote          BOOLEAN NOT NULL DEFAULT FALSE,
    posted_at       TIMESTAMPTZ,
    hash            TEXT NOT NULL UNIQUE,
    link_status     TEXT NOT NULL DEFAULT 'unknown',
    raw_json        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- job_scores: AI fit score for a job, one row per job
CREATE TABLE job_scores (
    id              BIGSERIAL PRIMARY KEY,
    job_id          BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    fit_score       INTEGER NOT NULL,
    dimension_scores JSONB,
    matched_skills  TEXT[],
    missing_skills  TEXT[],
    rationale       TEXT,
    scored_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (job_id)
);

-- applications: a job the user is actively pursuing
CREATE TABLE applications (
    id               BIGSERIAL PRIMARY KEY,
    job_id           BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    status           TEXT NOT NULL DEFAULT 'new'
                     CHECK (status IN ('new','applied','interview','offer','rejected')),
    applied_at       TIMESTAMPTZ,
    notes            TEXT,
    next_action      TEXT,
    next_action_date DATE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- drafts: AI-generated application materials, versioned
CREATE TABLE drafts (
    id             BIGSERIAL PRIMARY KEY,
    application_id BIGINT NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    cover_letter   TEXT,
    custom_answers JSONB,
    version        INTEGER NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- user_profile: single-user profile feeding the scorer and drafter
CREATE TABLE user_profile (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    current_role_title TEXT,
    skills          TEXT[],
    experience_years INTEGER,
    target_roles    TEXT[],
    preferred_stack TEXT[],
    salary_min      INTEGER,
    salary_max      INTEGER,
    location_pref   TEXT,
    remote_only     BOOLEAN NOT NULL DEFAULT FALSE,
    timezone        TEXT NOT NULL DEFAULT 'UTC',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- source_logs: one row per fetch run per source, for observability
CREATE TABLE source_logs (
    id          BIGSERIAL PRIMARY KEY,
    source_name TEXT NOT NULL,
    fetched_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    jobs_found  INTEGER NOT NULL DEFAULT 0,
    jobs_new    INTEGER NOT NULL DEFAULT 0,
    errors      TEXT,
    duration_ms BIGINT NOT NULL DEFAULT 0
);

-- blacklist: companies or keywords to always filter out
CREATE TABLE blacklist (
    id    BIGSERIAL PRIMARY KEY,
    type  TEXT NOT NULL CHECK (type IN ('company','keyword')),
    value TEXT NOT NULL
);