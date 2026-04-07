---
name: manage-local-db
description: Access and manage the local PostgreSQL database for this datastar.chat project. Use this skill whenever the user asks about database content, thread data, messages, or wants to query/modify the local Postgres database. This includes showing conversation threads, message history, thread metadata, or any database inspection tasks.
---

# Manage Local Database

This skill helps you interact with the local PostgreSQL database for the datastar.chat application.

## Database Connection

Connection string (from .env):
```
postgresql://postgres:postgres@localhost:5432/datastar_chat?sslmode=disable
```

Quick psql command:
```bash
psql "postgresql://postgres:postgres@localhost:5432/datastar_chat" -c "YOUR_QUERY"
```

## Schema

### threads table
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PRIMARY KEY | Thread identifier |
| title | TEXT NOT NULL | Thread title |
| archived_at | TIMESTAMPTZ | Null if not archived |
| is_streaming | BOOLEAN | True if currently streaming |
| created_at | TIMESTAMPTZ | Creation timestamp |
| updated_at | TIMESTAMPTZ | Last update timestamp |

### messages table
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PRIMARY KEY | Message identifier |
| thread_id | UUID FK → threads | Parent thread |
| role | TEXT | 'user', 'assistant', or 'system' |
| content | TEXT | Message content |
| created_at | TIMESTAMPTZ | Creation timestamp |

Indexes: `idx_messages_thread_id`, `idx_messages_thread_id_created_at`

## Common Queries

### List all threads (non-archived)
```sql
SELECT id, title, created_at FROM threads WHERE archived_at IS NULL ORDER BY created_at DESC;
```

### Get thread details
```sql
SELECT * FROM threads WHERE id = 'UUID';
```

### Get all messages in a thread
```sql
SELECT role, content, created_at FROM messages WHERE thread_id = 'UUID' ORDER BY created_at ASC;
```

### Count messages per thread
```sql
SELECT thread_id, COUNT(*) as msg_count FROM messages GROUP BY thread_id;
```

### Search messages
```sql
SELECT m.*, t.title FROM messages m JOIN threads t ON m.thread_id = t.id WHERE m.content ILIKE '%search_term%';
```

## When User Asks About Data

1. **If asking about a specific thread** (with UUID or title):
   - Query the thread row
   - Query messages for that thread_id
   - Present in readable format

2. **If asking about all threads/messages**:
   - List threads with counts
   - Show recent activity

3. **If asking to modify data**:
   - Confirm the change
   - Execute UPDATE/DELETE with proper WHERE clause
   - Verify the change

## Output Format

Present thread data as:
```
**Thread:** [title] (UUID)
Created: [timestamp]
Status: [active/archived/streaming]

**Messages:**
[role] ([time]): [content preview or full content]
```

Use `psql` for quick queries or Go database code if modifying via the application layer.
