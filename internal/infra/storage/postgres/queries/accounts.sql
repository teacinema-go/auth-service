-- name: CreateAccount :exec
INSERT INTO accounts (id, phone, email) VALUES ($1, $2, $3);

-- name: UpdateAccountIsPhoneVerified :exec
UPDATE accounts
set is_phone_verified = $2
WHERE id = $1;

-- name: UpdateAccountIsEmailVerified :exec
UPDATE accounts
set is_email_verified = $2
WHERE id = $1;

-- name: GetAccountByPhone :one
SELECT * FROM accounts
WHERE phone = $1 LIMIT 1;

-- name: GetAccountByEmail :one
SELECT * FROM accounts
WHERE email = $1 LIMIT 1;