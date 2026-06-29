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