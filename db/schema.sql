CREATE TABLE threads (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    archived_at DATETIME,
    is_streaming INTEGER NOT NULL DEFAULT 0 CHECK (is_streaming IN (0, 1)),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN (
        'user_message',
        'assistant_message',
        'system_message',
        'assistant_reasoning',
        'assistant_tool_call',
        'tool_output'
    )),
    -- Per thread message order. Used for stable conversation rendering.
    seq INTEGER NOT NULL,
    content BLOB NOT NULL,
    tool_call_id TEXT NOT NULL DEFAULT '',
    tool_name TEXT NOT NULL DEFAULT '',
    provider_response_id TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_thread_id ON messages(thread_id);
CREATE UNIQUE INDEX idx_messages_thread_id_seq ON messages(thread_id, seq);
CREATE INDEX idx_messages_created_at ON messages(thread_id, created_at, seq);
CREATE INDEX idx_messages_provider_response_id ON messages(provider_response_id);
CREATE INDEX idx_messages_tool_call_id ON messages(tool_call_id);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX sessions_token_hash_idx ON sessions(token_hash);
CREATE INDEX sessions_user_id_idx ON sessions(user_id);
CREATE INDEX sessions_expires_at_idx ON sessions(expires_at);
