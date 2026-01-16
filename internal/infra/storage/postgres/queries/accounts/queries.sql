-- name: Create :one
INSERT INTO accounts (
    id, phone, email
) VALUES (
             $1, $2, $3
         )
RETURNING *;

-- name: UpdateIsPhoneVerified :exec
UPDATE accounts
set is_phone_verified = $2
WHERE id = $1;

-- name: UpdateIsEmailVerified :exec
UPDATE accounts
set is_email_verified = $2
WHERE id = $1;

-- name: GetByPhone :one
SELECT * FROM accounts
WHERE phone = $1 LIMIT 1;

-- name: GetByEmail :one
SELECT * FROM accounts
WHERE email = $1 LIMIT 1;