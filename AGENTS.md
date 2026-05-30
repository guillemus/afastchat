sqlite db at `tmp/app.db`.
After `db/schema.sql` changes, run `just migrate` to apply Atlas migration with auto-approve.
For frontend JS type checks, run `just typecheck-scripts`.
Deploy target is VPS, no Docker. Prefer single Go binary with embedded public assets.
