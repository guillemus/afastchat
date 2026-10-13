-- +goose Up
-- create "openrouter_connections" table
CREATE TABLE `openrouter_connections` (
  `user_id` text NOT NULL,
  `sealed_key` blob NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`user_id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- +goose Down
-- reverse: create "openrouter_connections" table
DROP TABLE `openrouter_connections`;
