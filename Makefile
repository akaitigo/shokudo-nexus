.PHONY: build test lint format check quality clean proto

build:
	cd backend && $(MAKE) build
	cd frontend && $(MAKE) build

test:
	cd backend && $(MAKE) test
	cd frontend && $(MAKE) test

lint:
	cd backend && $(MAKE) lint
	cd frontend && $(MAKE) lint
	buf lint

format:
	cd backend && $(MAKE) format
	cd frontend && $(MAKE) format
	buf format -w

proto:
	buf generate

check: format lint test build
	@echo "All checks passed."

quality:
	@echo "=== Quality Gate ==="
	@test -f LICENSE || { echo "ERROR: LICENSE missing. Fix: add MIT LICENSE file"; exit 1; }
	@! grep -rn "TODO\|FIXME\|HACK\|console\.log\|println\|print(" backend/internal/ frontend/src/ 2>/dev/null | grep -v "node_modules" || { echo "ERROR: debug output or TODO found. Fix: remove before ship"; exit 1; }
	@! grep -rn "password=\|secret=\|api_key=\|sk-\|ghp_" backend/ frontend/src/ 2>/dev/null | grep -v '\$$\$${' | grep -v "node_modules" || { echo "ERROR: hardcoded secrets. Fix: use env vars with no default"; exit 1; }
	@test ! -f PRD.md || ! grep -q "\[ \]" PRD.md || { echo "ERROR: unchecked acceptance criteria in PRD.md"; exit 1; }
	@echo "OK: automated quality checks passed"
	@echo "Manual checks required: README quickstart, demo GIF, input validation, ADR >=1"

clean:
	cd backend && $(MAKE) clean
	cd frontend && $(MAKE) clean
