# Contributing

PawnKit is maintained by volunteers, so reviews may take a little time.

Migration ideas are welcome, but a useful rule needs a narrow purpose and a
before-and-after fixture. Mark edits unsafe when a maintainer must make a
judgment that the tool cannot verify.

Run these checks before opening a pull request:

```sh
go test ./...
go vet ./...
CGO_ENABLED=1 go test -race ./...
```

Plans must remain reviewable. Preserve the clean-worktree guard, snapshot
checks, atomic writes, and rollback behavior when changing apply logic.
