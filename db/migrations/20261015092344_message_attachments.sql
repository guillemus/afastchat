-- +goose Up
-- create "message_attachments" table
CREATE TABLE `message_attachments` (
  `id` text NULL,
  `message_id` text NOT NULL,
  `user_id` text NOT NULL,
  `object_key` text NOT NULL,
  `mime_type` text NOT NULL,
  `filename` text NOT NULL,
  `position` integer NOT NULL,
  `created_at` datetime NOT NULL DEFAULT (CURRENT_TIMESTAMP),
  PRIMARY KEY (`id`),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`message_id`) REFERENCES `messages` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "message_attachments_object_key" to table: "message_attachments"
CREATE UNIQUE INDEX `message_attachments_object_key` ON `message_attachments` (`object_key`);
-- create index "idx_message_attachments_message_id" to table: "message_attachments"
CREATE INDEX `idx_message_attachments_message_id` ON `message_attachments` (`message_id`, `position`);
-- create index "idx_message_attachments_user_id" to table: "message_attachments"
CREATE INDEX `idx_message_attachments_user_id` ON `message_attachments` (`user_id`);

-- +goose Down
-- reverse: create index "idx_message_attachments_user_id" to table: "message_attachments"
DROP INDEX `idx_message_attachments_user_id`;
-- reverse: create index "idx_message_attachments_message_id" to table: "message_attachments"
DROP INDEX `idx_message_attachments_message_id`;
-- reverse: create index "message_attachments_object_key" to table: "message_attachments"
DROP INDEX `message_attachments_object_key`;
-- reverse: create "message_attachments" table
DROP TABLE `message_attachments`;
