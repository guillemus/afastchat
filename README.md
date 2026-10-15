# afastchat

<a href="https://designobo.com/s/5e2b5700d57265d3cdd625bb9e29541e6eec67a7f7725d9d0092709b76e3baf1"><img src=".github/prototyped-with-designobo.svg" alt="Prototyped with Designobo" height="20"></a>

A [T3 Chat](https://t3.chat) clone, live at [afastchat.com](https://afastchat.com) on a single AWS t3.micro.

Bring your own [OpenRouter](https://openrouter.ai) key and chat with any model, with web search, image attachments and streaming that survives reloads.

## How it works

- **The server renders HTML.** Pages are [templ](https://templ.guide) components. [Datastar](https://data-star.dev) sends user actions to the server and applies the HTML it sends back, morphing it into the page.
- **Updates arrive over SSE.** Each page keeps a Server-Sent Events stream open, and the server pushes HTML patches down it: streamed tokens, new chats, title changes and toasts.
- **Chat turns are durable.** Each reply runs as a [DBOS](https://www.dbos.dev) workflow, and tokens go out over an embedded [NATS](https://nats.io) server. The reply survives a closed tab. Any tab or device showing that chat picks up the stream, and a reconnect resumes where it left off.
- **The sidebar is virtualized on the server.** The browser sends its scroll position, and the server renders only the visible rows.
- **It ships as one binary.** Go with embedded assets, no Docker.

## Running it yourself

The easiest way to use it is [afastchat.com](https://afastchat.com). To run your own copy you need:

- a Google OAuth client, for login,
- a Cloudflare R2 bucket, for image uploads. R2 is the only storage backend right now; PRs adding other image backends are welcome,
- an OpenRouter API key, which you connect inside the app after logging in.

Install [mise](https://mise.jdx.dev) and [bun](https://bun.sh), then put these in `~/.config/afastchat.com/.env`, which mise loads:

```sh
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
OPENROUTER_TOKEN_KEY=   # encrypts stored OpenRouter keys: openssl rand -base64 32
R2_ENDPOINT=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET=
```

```sh
bun install
mise run
```

## License

[MIT](LICENSE)

Prototyped with [Designobo](https://designobo.com/s/5e2b5700d57265d3cdd625bb9e29541e6eec67a7f7725d9d0092709b76e3baf1).
