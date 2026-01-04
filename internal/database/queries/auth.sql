-- name: CreateAccount :one
INSERT INTO accounts (
    id, phone, email
) VALUES (
             $1, $2, $3
         )
RETURNING *;

-- name: GetAccountByPhone :one
SELECT * FROM accounts
WHERE phone = $1 LIMIT 1;

-- name: GetAccountByEmail :one
SELECT * FROM accounts
WHERE email = $1 LIMIT 1;