-- name: CreateRefreshToken :one
INSERT INTO
	refresh_tokens (
		token,
		created_at,
		updated_at,
		user_id,
		expires_at,
		revoked_at
	)
VALUES
	($1, now(), now(), $2, $3, $4)
RETURNING
	*;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET
	updated_at = now(),
	revoked_at = now()
WHERE
	token = $1
RETURNING
	*;

-- name: GetUserFromRefreshToken :one
SELECT
	users.*
FROM
	users
	JOIN refresh_tokens ON refresh_tokens.user_id = users.id
WHERE
	refresh_tokens.token = $1
	AND revoked_at IS NULL
	AND expires_at > NOW();
