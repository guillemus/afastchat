CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    paid_at DATETIME,
    payment_customer_id TEXT NOT NULL DEFAULT '',
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

CREATE TABLE payments (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    -- Allowed values: polar, stripe.
    provider TEXT NOT NULL DEFAULT 'polar',
    provider_checkout_id TEXT NOT NULL,
    -- Allowed values:
    -- lifetime: lifetime purchase.
    -- topup: credit top-up.
    type TEXT NOT NULL,
    -- Allowed values:
    -- created: checkout created.
    -- succeeded: payment succeeded.
    -- failed: payment failed.
    -- expired: checkout expired.
    status TEXT NOT NULL,
    amount_cents INTEGER NOT NULL,
    credit_cents INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_payment_checkouts_provider_checkout_id ON payments(provider, provider_checkout_id);
CREATE INDEX idx_payment_checkouts_user_id_created_at ON payments(user_id, created_at);

CREATE TABLE payment_events (
    id TEXT PRIMARY KEY,
    -- Allowed values: polar, stripe.
    provider TEXT NOT NULL DEFAULT 'polar',
    provider_event_id TEXT NOT NULL,
    type TEXT NOT NULL,
    payload TEXT NOT NULL,
    processed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_payment_events_provider_event_id ON payment_events(provider, provider_event_id);
CREATE INDEX idx_payment_events_type_created_at ON payment_events(provider, type, created_at);

CREATE TABLE billing_ledger (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    payment_checkout_id TEXT REFERENCES payments(id),
    usage_id TEXT REFERENCES usage(id),
    -- Allowed values:
    -- - initial_credit: signup credit.
    -- - topup_credit: purchased credit.
    -- - usage_debit: usage charge.
    -- - adjustment: manual adjustment.
    type TEXT NOT NULL,
    amount_cents INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_billing_ledger_user_id_created_at ON billing_ledger(user_id, created_at);
CREATE UNIQUE INDEX idx_billing_ledger_payment_checkout_id_type ON billing_ledger(payment_checkout_id, type) WHERE payment_checkout_id IS NOT NULL;
CREATE UNIQUE INDEX idx_billing_ledger_usage_id ON billing_ledger(usage_id) WHERE usage_id IS NOT NULL;
