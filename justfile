migrate:
	atlas schema apply \
	  --to "file://db/schema.sql" \
	  --url "sqlite://tmp/app.db" \
	  --dev-url "sqlite://dev?mode=memory"

build-linux:
	mkdir -p dist
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/afastchat ./cmd/server

rsync-linux: build-linux
	rsync -avz dist/afastchat hetzner-box:/srv/afastchat/afastchat
	ssh hetzner-box 'chown afastchat:afastchat /srv/afastchat/afastchat && chmod +x /srv/afastchat/afastchat'

format-templates:
    bunx prettier -w templates
