- `/Users/guillem/Documents/projects/designobo.com` is the reference project for all design decisions: tooling, mise tasks, templ/Datastar view patterns, UI components, toasts, and coding standards. When unsure how to do something, check how designobo does it.
- Deploy target is VPS, no Docker. Prefer single Go binary with embedded public assets.
- Local env vars live in `~/.config/afastchat.com/.env`, loaded by mise.

# commands

- sqlite db at `tmp/app.db`. Use `mise run db:sql "SQL"` to execute raw SQL against it.
- `db/schema.sql` is the source of truth. After changing it, run `mise run db:migrate <name>` to generate a goose migration in `db/migrations/` from the diff. goose applies migrations on server start; never hand-write migration files.
- to check TypeScript, run `mise run tscheck`.
- to lint (golangci, sumtype, design tokens), run `mise run lint`.
- to format everything (templ, Go, TypeScript, CSS), run `mise run format`.

# views

- Views are templ components in `views/`, shared UI in `views/ui/`. Frontend TypeScript lives in `views/scripts/` and is bundled into `public/index.js`.
- Prefix client-only Datastar signals with `_`. Unprefixed signals are sent with every request.
- Style with Tailwind utilities using the tokens in `views/styles.css`; `mise run lint:styles` rejects raw palette colors and default text sizes.
- Show user feedback through the shared toast: `x.patchToast(w, r, ui.Toast{...})`.

# production

You won't run `ssh hetzner-box` by yourself without asking for permission. But you are allowed to run `mise prod:status` and `mise prod:logs` only if it makes sense for the task.
You also won't deploy anything on prod without asking for permission.
