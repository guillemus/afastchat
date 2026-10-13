-- +goose Up
-- create "users" table
CREATE TABLE `users` (
  `id` text NULL,
  `email` text NOT NULL,
  `display_name` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`id`)
);
-- create index "users_email_lower_idx" to table: "users"
CREATE UNIQUE INDEX `users_email_lower_idx` ON `users` ((lower(email)));
-- create "auth_identities" table
CREATE TABLE `auth_identities` (
  `id` text NULL,
  `user_id` text NOT NULL,
  `provider` text NOT NULL,
  `subject` text NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "auth_identities_provider_subject" to table: "auth_identities"
CREATE UNIQUE INDEX `auth_identities_provider_subject` ON `auth_identities` (`provider`, `subject`);
-- create index "auth_identities_user_id_provider" to table: "auth_identities"
CREATE UNIQUE INDEX `auth_identities_user_id_provider` ON `auth_identities` (`user_id`, `provider`);
-- create "sessions" table
CREATE TABLE `sessions` (
  `token` text NULL,
  `data` blob NOT NULL,
  `expiry` real NOT NULL,
  PRIMARY KEY (`token`)
);
-- create index "sessions_expiry_idx" to table: "sessions"
CREATE INDEX `sessions_expiry_idx` ON `sessions` (`expiry`);
-- create "threads" table
CREATE TABLE `threads` (
  `id` text NULL,
  `user_id` text NOT NULL,
  `title` text NOT NULL DEFAULT '',
  `archived_at` datetime NULL,
  `is_streaming` integer NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CHECK (is_streaming IN (0, 1))
);
-- create index "idx_threads_user_id" to table: "threads"
CREATE INDEX `idx_threads_user_id` ON `threads` (`user_id`);
-- create "messages" table
CREATE TABLE `messages` (
  `id` text NULL,
  `thread_id` text NOT NULL,
  `type` text NOT NULL,
  `seq` integer NOT NULL,
  `content` text NOT NULL,
  `tool_call_id` text NOT NULL DEFAULT '',
  `tool_name` text NOT NULL DEFAULT '',
  `provider_response_id` text NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  `updated_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`thread_id`) REFERENCES `threads` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_messages_thread_id" to table: "messages"
CREATE INDEX `idx_messages_thread_id` ON `messages` (`thread_id`);
-- create index "idx_messages_thread_id_seq" to table: "messages"
CREATE UNIQUE INDEX `idx_messages_thread_id_seq` ON `messages` (`thread_id`, `seq`);
-- create index "idx_messages_created_at" to table: "messages"
CREATE INDEX `idx_messages_created_at` ON `messages` (`thread_id`, `created_at`, `seq`);
-- create index "idx_messages_provider_response_id" to table: "messages"
CREATE INDEX `idx_messages_provider_response_id` ON `messages` (`provider_response_id`);
-- create index "idx_messages_tool_call_id" to table: "messages"
CREATE INDEX `idx_messages_tool_call_id` ON `messages` (`tool_call_id`);
-- create "usage" table
CREATE TABLE `usage` (
  `id` text NULL,
  `user_id` text NOT NULL,
  `thread_id` text NOT NULL,
  `message_id` text NULL,
  `provider` text NOT NULL,
  `model` text NOT NULL,
  `input_tokens` integer NOT NULL DEFAULT 0,
  `cached_input_tokens` integer NOT NULL DEFAULT 0,
  `output_tokens` integer NOT NULL DEFAULT 0,
  `reasoning_tokens` integer NOT NULL DEFAULT 0,
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`thread_id`) REFERENCES `threads` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `2` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_usage_user_id_created_at" to table: "usage"
CREATE INDEX `idx_usage_user_id_created_at` ON `usage` (`user_id`, `created_at`);
-- create index "idx_usage_thread_id_created_at" to table: "usage"
CREATE INDEX `idx_usage_thread_id_created_at` ON `usage` (`thread_id`, `created_at`);
-- create index "idx_usage_message_id" to table: "usage"
CREATE INDEX `idx_usage_message_id` ON `usage` (`message_id`);

-- +goose Down
-- reverse: create index "idx_usage_message_id" to table: "usage"
DROP INDEX `idx_usage_message_id`;
-- reverse: create index "idx_usage_thread_id_created_at" to table: "usage"
DROP INDEX `idx_usage_thread_id_created_at`;
-- reverse: create index "idx_usage_user_id_created_at" to table: "usage"
DROP INDEX `idx_usage_user_id_created_at`;
-- reverse: create "usage" table
DROP TABLE `usage`;
-- reverse: create index "idx_messages_tool_call_id" to table: "messages"
DROP INDEX `idx_messages_tool_call_id`;
-- reverse: create index "idx_messages_provider_response_id" to table: "messages"
DROP INDEX `idx_messages_provider_response_id`;
-- reverse: create index "idx_messages_created_at" to table: "messages"
DROP INDEX `idx_messages_created_at`;
-- reverse: create index "idx_messages_thread_id_seq" to table: "messages"
DROP INDEX `idx_messages_thread_id_seq`;
-- reverse: create index "idx_messages_thread_id" to table: "messages"
DROP INDEX `idx_messages_thread_id`;
-- reverse: create "messages" table
DROP TABLE `messages`;
-- reverse: create index "idx_threads_user_id" to table: "threads"
DROP INDEX `idx_threads_user_id`;
-- reverse: create "threads" table
DROP TABLE `threads`;
-- reverse: create index "sessions_expiry_idx" to table: "sessions"
DROP INDEX `sessions_expiry_idx`;
-- reverse: create "sessions" table
DROP TABLE `sessions`;
-- reverse: create index "auth_identities_user_id_provider" to table: "auth_identities"
DROP INDEX `auth_identities_user_id_provider`;
-- reverse: create index "auth_identities_provider_subject" to table: "auth_identities"
DROP INDEX `auth_identities_provider_subject`;
-- reverse: create "auth_identities" table
DROP TABLE `auth_identities`;
-- reverse: create index "users_email_lower_idx" to table: "users"
DROP INDEX `users_email_lower_idx`;
-- reverse: create "users" table
DROP TABLE `users`;
