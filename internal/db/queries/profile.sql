-- name: GetProfile :one
SELECT id, name, current_role_title, skills, experience_years, target_roles, preferred_stack, salary_min, salary_max, location_pref, remote_only, timezone, updated_at
FROM user_profile
WHERE id = 1;

-- name: UpsertProfile :one
INSERT INTO user_profile (
    id, name, current_role_title, skills, experience_years,
    target_roles, preferred_stack, salary_min, salary_max,
    location_pref, remote_only, timezone, updated_at
) VALUES (
    1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now()
)
ON CONFLICT (id) DO UPDATE SET 
    name                 = EXCLUDED.name,
    current_role_title   = EXCLUDED.current_role_title,
    skills               = EXCLUDED.skills,
    experience_years     = EXCLUDED.experience_years,
    target_roles         = EXCLUDED.target_roles,
    preferred_stack      = EXCLUDED.preferred_stack,
    salary_min           = EXCLUDED.salary_min,
    salary_max           = EXCLUDED.salary_max,
    location_pref        = EXCLUDED.location_pref,
    remote_only          = EXCLUDED.remote_only,
    timezone             = EXCLUDED.timezone,
    updated_at           = now()
RETURNING id, name, current_role_title, skills, experience_years,
          target_roles, preferred_stack, salary_min, salary_max,
          location_pref, remote_only, timezone, updated_at;