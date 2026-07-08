-- name: CreateJob :one 
INSERT INTO jobs (
    title, company, source, url, description, description_hash, location, remote, posted_at, hash, raw_json
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 
)
ON CONFLICT (hash) DO NOTHING
RETURNING id;

-- name: CountJobs :one
SELECT count(*) FROM jobs; 


-- name: ListJobs :many
SELECT j.id, j.title, j.company, j.source, j.url, j.location, j.remote,
       j.posted_at, j.link_status, j.created_at,
       s.fit_score, s.rationale
FROM jobs j
LEFT JOIN job_scores s ON s.job_id = j.id
ORDER BY 
    CASE WHEN @sort_by_score::bool THEN s.fit_score END DESC NULLS LAST,
    j.created_at DESC
LIMIT $1 OFFSET $2;


-- name: ListUnscoredJobs :many
SELECT j.id, j.title, j.company, j.source, j.url, j.description, j.description_hash, j.location, j.posted_at
FROM jobs j
LEFT JOIN job_scores s ON s.job_id = j.id
WHERE s.id IS NULL
ORDER BY j.created_at DESC
LIMIT $1;

-- name: CreateJobScore :one
INSERT INTO job_scores (
    job_id, fit_score, dimension_scores, matched_skills, missing_skills, rationale
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id;