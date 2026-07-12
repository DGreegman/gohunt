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
