-- name: InsertAdminAccessEvent :one
INSERT INTO admin_access_events (
    actor_user_id,
    target_user_id,
    resource_type,
    resource_id,
    action,
    outcome,
    request_id,
    safe_metadata
)
VALUES (
    sqlc.arg(actor_user_id),
    sqlc.narg(target_user_id),
    sqlc.arg(resource_type),
    sqlc.narg(resource_id),
    sqlc.arg(action),
    sqlc.arg(outcome),
    sqlc.arg(request_id),
    sqlc.arg(safe_metadata)
)
RETURNING *;

-- name: ListAdminAccessEventsByActor :many
SELECT *
FROM admin_access_events
WHERE actor_user_id = sqlc.arg(actor_user_id)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size);

-- name: ListAdminAccessEventsByTarget :many
SELECT *
FROM admin_access_events
WHERE target_user_id = sqlc.arg(target_user_id)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_size);
