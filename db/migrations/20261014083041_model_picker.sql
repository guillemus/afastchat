-- +goose Up
-- add column "selected_model" to table: "users"
ALTER TABLE `users` ADD COLUMN `selected_model` text NOT NULL DEFAULT '';
-- create "user_favorite_models" table
CREATE TABLE `user_favorite_models` (
  `user_id` text NOT NULL,
  `model_id` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`user_id`, `model_id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- +goose Down
-- reverse: create "user_favorite_models" table
DROP TABLE `user_favorite_models`;
-- reverse: add column "selected_model" to table: "users"
ALTER TABLE `users` DROP COLUMN `selected_model`;
