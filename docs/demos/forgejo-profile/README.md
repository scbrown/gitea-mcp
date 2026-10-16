# Forgejo Actions profile demonstration

Actual local MCP stdio CLI discovery, using https://example.invalid and no credential:
upstream v1.7.0 (`bfe0d4c9b06e71a60d22b7ce81a55f01ee0157b0`) advertises 13 read
and 3 write run methods. Candidate `db26a00eca0def6aa7846a1c42c1375b13bde41f`
with `--actions-profile forgejo-15` advertises 2 read and 1 write methods.
Stale job/log/rerun requests return an explicit unsupported result before HTTP.
Tests separately count HTTP calls, including a supported positive control.

Reproduce with `python3 demo.py /path/to/original /path/to/candidate`.
Replay the actual util-linux recording with
`scriptreplay --log-out output.log --log-timing timing.log`.
The small sanitized recording is retained here with the source for review.
This demonstrates candidate behavior, not an installed service or error-rate recovery.

This fork starts from the public upstream tag above and retains its MIT license.
GitHub CI builds and tests source; it does not publish releases or install services.

CI retains a static Linux/amd64 tarball, SHA-256 checksums, build information and
exact source commit. Release publication selects a successful main-commit artifact;
PR artifacts are candidate evidence. Publishing a release and installing it are
separate steps. Neither a source push nor CI artifact upload deploys a service.
