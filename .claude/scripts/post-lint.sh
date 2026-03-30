#!/usr/bin/env bash
set -euo pipefail

input="$(cat)"
file="$(jq -r '.tool_input.file_path // .tool_input.path // empty' <<< "$input")"

# Go files
case "$file" in
  *.go)
    cd "$(git rev-parse --show-toplevel 2>/dev/null)/backend" || exit 0
    if ! command -v golangci-lint &>/dev/null; then
      go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 2>/dev/null || { echo "WARN: golangci-lint install failed" >&2; exit 0; }
    fi
    if ! command -v gofumpt &>/dev/null; then
      go install mvdan.cc/gofumpt@latest 2>/dev/null || true
    fi
    golangci-lint run --fix "$file" >/dev/null 2>&1 || true
    diag="$(golangci-lint run "$file" 2>&1 | head -20)"
    if [ -n "$diag" ]; then
      jq -Rn --arg msg "$diag" \
        '{ hookSpecificOutput: { hookEventName: "PostToolUse", additionalContext: $msg } }'
    fi
    exit 0
    ;;
esac

# TypeScript files
case "$file" in
  *.ts|*.tsx|*.js|*.jsx)
    npx biome format --write "$file" >/dev/null 2>&1 || true
    npx oxlint --fix "$file" >/dev/null 2>&1 || true
    diag="$(npx oxlint "$file" 2>&1 | head -20)"
    if [ -n "$diag" ]; then
      jq -Rn --arg msg "$diag" \
        '{ hookSpecificOutput: { hookEventName: "PostToolUse", additionalContext: $msg } }'
    fi
    exit 0
    ;;
esac

# Proto files
case "$file" in
  *.proto)
    ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)
    cd "$ROOT"
    buf format -w 2>&1 | head -5
    buf lint 2>&1 | head -20
    exit 0
    ;;
esac
