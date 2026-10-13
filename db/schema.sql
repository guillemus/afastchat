-- Allowed values are enforced in app code with Go string enum types, not DB CHECK constraints.
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    display_name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

CREATE TABLE auth_identities (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    -- Allowed values: google.
    provider TEXT NOT NULL,
    -- Stable provider account identifier. Google calls this "sub".
    subject TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (provider, subject),
    UNIQUE (user_id, provider)
);

CREATE TABLE openrouter_connections (
    user_id TEXT NOT NULL PRIMARY KEY REFERENCES users(id),
    -- AES-GCM sealed OpenRouter API key.
    sealed_key BLOB NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

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
    -- Allowed values:
    -- user_message: user input.
    -- assistant_message: assistant response.
    -- system_message: system instruction.
    -- assistant_reasoning: assistant reasoning trace.
    -- assistant_tool_call: assistant tool request.
    -- tool_output: tool result.
    type TEXT NOT NULL,
    -- Per thread message order. Used for stable conversation rendering.
    seq INTEGER NOT NULL,
    content TEXT NOT NULL,
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

CREATE TABLE usage (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    thread_id TEXT NOT NULL REFERENCES threads(id),
    message_id TEXT REFERENCES messages(id),
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    input_tokens INTEGER NOT NULL DEFAULT 0,
    cached_input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    reasoning_tokens INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_usage_user_id_created_at ON usage(user_id, created_at);
CREATE INDEX idx_usage_thread_id_created_at ON usage(thread_id, created_at);
CREATE INDEX idx_usage_message_id ON usage(message_id);
