# Call-shape compatibility demonstration

Real local MCP stdio calls use the same private loopback HTTP fixture and no
credential. Baseline is upstream v1.7.0 (`bfe0d4c9b06e71a60d22b7ce81a55f01ee0157b0`);
candidate is clean `5ab502ab33109cdd1bafa91f8156b5cc997e381d` with
`--actions-profile forgejo-15`.

Watch bare errors become method-specific guidance and examples. Actions `list`
with `in_progress` becomes the supported `running` request; PR `list` actually
lists, and `edit` makes the intended PATCH to the fixture. Missing parameters
and unavailable Forgejo jobs send zero HTTP requests. Per-call HTTP counts and
successful fixture reads provide positive controls. The write example contacts
only this disposable fixture; it proves no production permission or adoption.

Reproduce: `python3 demo.py /path/to/baseline /path/to/candidate`.
Replay the retained util-linux recording:
`scriptreplay --log-out output.log --log-timing timing.log`.
The small sanitized recording stays with the source for review; replay verified.
Portable DP rules were separately exported/imported twice in private databases
(7 imported, then 7 skipped), without changing live aliases. Native hook firing
and seven-day error-rate acceptance remain separate deployment checks.
