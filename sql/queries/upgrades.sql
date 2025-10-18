-- name: UpgradeUserById :execrows
UPDATE users
SET is_chirpy_red = '1'
WHERE id = $1;