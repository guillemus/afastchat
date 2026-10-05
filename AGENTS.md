sqlite db at `tmp/app.db`.
After `db/schema.sql` changes, run `mise run migrate` to apply Atlas migration with auto-approve.
For frontend JS type checks, run `mise run typecheck-scripts`.
Deploy target is VPS, no Docker. Prefer single Go binary with embedded public assets.
