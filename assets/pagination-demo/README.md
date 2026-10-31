# Actions pagination: actual requests before and after

The focused reproducer calls the Actions read handler against a local HTTP
fixture that requires a positive page. With the parser from base
`8ad2fb3ccbf59c623098e073ea43202beceee4b4`, omitted page sends 1 but explicit
zero sends 0 and the receiver refuses it. With this change, omitted, zero and
negative pages send 1; explicit page 2 still sends 2. The recording prints the
actual received request URLs, including both the expected baseline failure and
successful candidate calls. It does not show a production MCP deployment or
claim that omitted-page production calls were failing.

Source: base above plus the parser and wire controls in this directory's
introducing commit. Target: isolated HTTP fixture, not a live service. Small
recording files remain in the repository's assets directory.

Download `output.log` and `timing.log`, then replay with util-linux:

```sh
scriptreplay --log-out output.log --log-timing timing.log
```

Replay was checked. `output.log` is the readable transcript; no inline playback
is claimed. To repeat from the candidate repository root:

```sh
git show 8ad2fb3ccbf59c623098e073ea43202beceee4b4:pkg/params/params.go > ../before-params.go
bash assets/pagination-demo/run.sh
```

The script temporarily substitutes the baseline parser, restores the candidate
on exit, and requires the baseline refusal before running the candidate. Run in
an isolated checkout with no concurrent writer.
