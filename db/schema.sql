CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BLOB NOT NULL,
    expiry REAL NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions(expiry);

CREATE TABLE threads (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    title TEXT NOT NULL DEFAULT '',
    archived_at DATETIME,
    is_streaming INTEGER NOT NULL DEFAULT 0 CHECK (is_streaming IN (0, 1)),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_threads_user_id ON threads(user_id);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    thread_id TEXT NOT NULL REFERENCES threads(id),
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
