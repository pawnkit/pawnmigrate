# Security policy

Report vulnerabilities through GitHub's private
[security advisory form](https://github.com/pawnkit/pawnmigrate/security/advisories/new).
Do not open a public issue before a fix is available.

Pawn source, manifests, paths, and migration plans are untrusted. Path escapes,
unsafe edits, crashes, hangs, and failed rollback are in scope.

Apply requires a clean Git worktree by default. It checks source snapshots
again before writing, uses atomic file replacement, and rolls back earlier
writes when a later write fails.
