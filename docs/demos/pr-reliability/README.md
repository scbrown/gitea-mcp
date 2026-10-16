# Pull request reliability demo

Baseline source: `5ab502ab33109cdd1bafa91f8156b5cc997e381d`.
Candidate source: `81b85b0b520795b33b7519e13ad5d6d34047a7f6`.
Both binaries were built with Go 1.27.0; the candidate reports a clean VCS build.

Replay the terminal recording with:

```sh
scriptreplay --log-timing timing.log --log-out output.log
```

Or run `python3 demo.py /path/to/baseline /path/to/candidate`.
The script uses the actual MCP stdio protocol with a private HTTP fixture and
no credentials or live API writes.

The exact up-to-date 500 changes from error to `changed: false`; an empty PR
changes from empty success to explicit error. One transient read 503 changes
from failure to success after one retry. A Git stderr 500 and a write 503 remain
errors, with exactly one POST each. A normal read is the positive control.
The recording proves local client behavior, not deployment or repair of server
storage stalls, Git locks, or a historical panic.
