#!/bin/bash
set -euo pipefail
cp pkg/params/params.go candidate-params.tmp
trap 'cp candidate-params.tmp pkg/params/params.go; rm -f candidate-params.tmp' EXIT
cp ../before-params.go pkg/params/params.go
printf 'BEFORE: Actions read with omitted page and then page=0 (expected refusal)\n'
if go test -count=1 -v -run '^TestForgejo15DiscoveryAndStaleCalls$' ./operation/actions; then
  printf 'ERROR: baseline unexpectedly accepted zero page\n'; exit 1
fi
cp candidate-params.tmp pkg/params/params.go
printf 'AFTER: omitted/zero/negative page becomes1; explicit page2 remains2\n'
go test -count=1 -v -run '^TestForgejo15DiscoveryAndStaleCalls$' ./operation/actions
