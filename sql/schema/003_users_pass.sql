-- +goose Up
ALTER TABLE users 
ADD hashed_pass TEXT UNIQUE NOT NULL DEFAULT 'unset';
-- +goose Down
ALTER TABLE users DROP hashed_pass;