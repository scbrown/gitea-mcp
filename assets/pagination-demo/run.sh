#!/bin/bash
set -euo pipefail
: "${BASELINE_DIR:?Set BASELINE_DIR to an isolated f28c56b source tree with the current profile_test.go copied in}"
baseline_output=$(mktemp)
trap 'rm -f "$baseline_output"' EXIT
printf 'BEFORE: published clamp implementation f28c56b with current receiver controls\n'
if (cd "$BASELINE_DIR" && go test -count=1 -v -run '^TestForgejo15DiscoveryAndStaleCalls$' ./operation/actions) > "$baseline_output" 2>&1; then
  cat "$baseline_output"
  printf 'ERROR: baseline unexpectedly satisfied rejection controls\n'
  exit 1
fi
cat "$baseline_output"
grep -F 'sent HTTP' "$baseline_output" >/dev/null
if grep -F '[build failed]' "$baseline_output" >/dev/null; then
  printf 'ERROR: a build failure is not a baseline behavior demonstration\n'
  exit 1
fi
printf 'AFTER: invalid explicit pages reject before HTTP; omitted1 and valid2 still send requests\n'
go test -count=1 -v -run '^TestForgejo15DiscoveryAndStaleCalls$' ./operation/actions
