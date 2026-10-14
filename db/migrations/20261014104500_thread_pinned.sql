-- +goose Up
-- add column "pinned_at" to table: "threads"
ALTER TABLE `threads` ADD COLUMN `pinned_at` datetime NULL;

-- +goose Down
-- reverse: add column "pinned_at" to table: "threads"
ALTER TABLE `threads` DROP COLUMN `pinned_at`;
