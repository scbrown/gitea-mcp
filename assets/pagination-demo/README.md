# Invalid pages: reject before HTTP

The same Actions receiver controls run against two isolated source trees.
Before: published `f28c56b06c7f2461aa3901a51761617290245a05` plus the current
receiver test. Its clamp sends page 1 for explicit zero, which fails the zero
outbound request assertion. After: this introducing repair commit rejects
explicit zero, negative, malformed, Boolean, fractional, null and overflowing
pages before HTTP. Omitted page still sends 1; both numeric and string page 2
send 2. Generic integer and ID parsing keeps its historical behavior.

The recording prints actual received URLs and explicit no-request assertions.
It requires a behavior failure on the baseline and refuses to count a build
failure as evidence. Both trees use the module's declared Go 1.27 toolchain.
This is an isolated HTTP fixture, not a production deployment or proof of the
old seven-day reliability criterion. The earlier clamp behavior and its review
finding remain in Git history; they are superseded by this rejection contract.

Download `output.log` and `timing.log`, then use util-linux:

```sh
scriptreplay --log-out output.log --log-timing timing.log
```

Replay was checked. `output.log` is also a readable transcript; no inline player
is claimed. These small files stay in the repository's assets directory.
To reproduce, create an isolated checkout of the exact baseline above, copy the
candidate's `operation/actions/profile_test.go` into that baseline, then run
from the candidate root:

```sh
BASELINE_DIR=/path/to/isolated-baseline bash assets/pagination-demo/run.sh
```

The script does not modify either source tree.
