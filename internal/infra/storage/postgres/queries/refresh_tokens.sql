-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, account_id, token_hash, expires_at)
VALUES ($1,$2,$3,$4);

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1 AND expires_at > NOW();

-- name: DeleteRefreshTokenByHash :execrows
DELETE FROM refresh_tokens
WHERE token_hash = $1;

-- name: DeleteRefreshTokensByAccountID :exec
DELETE FROM refresh_tokens
WHERE account_id = $1;