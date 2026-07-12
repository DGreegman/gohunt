-- name: CreateApplication :one
INSERT INTO applications (job_id, status, notes)
VALUES($1, 'new', $2)
RETURNING id, job_id, status, applied_at, notes, next_action, next_action_date, created_at;

-- name: ListApplications :many
SELECT a.id, a.job_id, a.status, a.applied_at, a.notes,
        a.next_action, a.next_action_date, a.created_at, 
        j.title, j.company, j.url
FROM applications a
JOIN jobs j ON j.id = a.job_id
ORDER BY a.created_at DESC;

-- name: GetApplication :one
SELECT id, job_id, status, applied_at, notes, next_action, next_action, next_action_date, created_at
FROM applications
WHERE id = $1;


-- name: UpdateApplication :one
UPDATE applications
SET status = COALESCE(sqlc.narg('status'), status),
        notes = COALESCE(sqlc.narg('notes'), notes),
        next_action = COALESCE(sqlc.narg('next+_action'), next_action),
        applied_at = CASE
                WHEN sqlc.narg('status')::text = 'applied' AND applied_at IS NULL THEN now()
                ELSE applied_at
        END
WHERE id = $1
RETURNING id, job_id, status, applied_at, notes, next_action, next_action_date, created_at;