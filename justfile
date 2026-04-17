migrate-dev:
	source .env && atlas schema apply \
	  --to "file://db/schema.sql" \
	  --url "${DATABASE_URL}&search_path=public" \
	  --dev-url "docker://postgres/16/dev?search_path=public"

migrate-prod:
	source .env && atlas schema apply \
	  --to "file://db/schema.sql" \
	  --url "${PRODUCTION_DATABASE_URL}&search_path=public" \
	  --dev-url "docker://postgres/16/dev?search_path=public"

build-linux:
	mkdir -p dist
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/afastchat ./cmd/server

rsync-linux: build-linux
	rsync -avz dist/afastchat hetzner-box:/srv/afastchat/afastchat
	ssh hetzner-box 'chown afastchat:afastchat /srv/afastchat/afastchat && chmod +x /srv/afastchat/afastchat'

format-templates:
    bunx prettier -w templates
