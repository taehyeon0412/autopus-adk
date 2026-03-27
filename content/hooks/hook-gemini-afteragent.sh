#!/bin/sh
# hook-gemini-afteragent.sh — Gemini CLI AfterAgent hook for autopus result collection.
# Reads hook JSON from stdin, extracts prompt_response,
# writes result.json and done signal to the session directory.
# POSIX shell compatible. No jq dependency — uses python3 for JSON.
set -e

SESSION_ID="${AUTOPUS_SESSION_ID:-}"
if [ -z "$SESSION_ID" ]; then
  exit 0
fi

# Validate session ID to prevent path traversal (alphanumeric, hyphen, underscore only)
case "$SESSION_ID" in
  *[!a-zA-Z0-9_-]*) exit 0 ;;
esac

SESSION_DIR="/tmp/autopus/${SESSION_ID}"
if [ ! -d "$SESSION_DIR" ]; then
  exit 0
fi

# Determine round-scoped file names when AUTOPUS_ROUND is set (integer-only).
case "${AUTOPUS_ROUND:-}" in *[!0-9]*) AUTOPUS_ROUND="" ;; esac
if [ -n "$AUTOPUS_ROUND" ]; then
  RESULT_FILE="${SESSION_DIR}/gemini-round${AUTOPUS_ROUND}-result.json"
  DONE_FILE="${SESSION_DIR}/gemini-round${AUTOPUS_ROUND}-done"
else
  RESULT_FILE="${SESSION_DIR}/gemini-result.json"
  DONE_FILE="${SESSION_DIR}/gemini-done"
fi

# Read hook JSON from stdin and extract prompt_response via python3.
# Input is passed via stdin (not argv) to avoid shell injection.
python3 -c "
import json, sys
data = json.load(sys.stdin)
msg = data.get('prompt_response', '')
if not msg:
    sys.exit(0)
result = {'output': msg, 'exit_code': 0}
with open(sys.argv[1], 'w') as f:
    json.dump(result, f)
" "${RESULT_FILE}"

chmod 600 "${RESULT_FILE}"

# Write done signal (empty file).
: > "${DONE_FILE}"
