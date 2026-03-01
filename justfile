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

format-templates:
    bunx prettier -w templates
