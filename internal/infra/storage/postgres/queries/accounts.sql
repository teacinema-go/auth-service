-- name: CreateAccount :one
INSERT INTO accounts (id, phone, email, role)
VALUES ($1, $2, $3, $4)
ON CONFLICT (phone, email) DO NOTHING
RETURNING id;

-- name: AccountExistsByPhone :one
SELECT EXISTS (
    SELECT 1 FROM accounts WHERE phone = $1
) AS exists;

-- name: AccountExistsByEmail :one
SELECT EXISTS (
    SELECT 1 FROM accounts WHERE email = $1
) AS exists;

-- name: GetAccountByPhone :one
SELECT * FROM accounts
WHERE phone = $1 LIMIT 1;

-- name: GetAccountByEmail :one
SELECT * FROM accounts
WHERE email = $1 LIMIT 1;