.PHONY: build test lint format typecheck check clean install deps-check

install:
	npm install

build:
	npx tsc

test:
	npx vitest run

lint:
	npx oxlint .
	npx biome check .

format:
	npx biome format --write .

typecheck:
	npx tsc --noEmit

check: format lint typecheck test build
	@echo "All checks passed."

deps-check:
	npx knip || echo "WARN: unused dependencies detected"

clean:
	rm -rf dist/ coverage/ node_modules/.cache/
