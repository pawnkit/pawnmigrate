# pawnmigrate

`pawnmigrate` updates Pawn projects without hiding what it plans to change.

## Install

Download a binary from the
[release page](https://github.com/pawnkit/pawnmigrate/releases), or install it
with Go:

```sh
go install github.com/pawnkit/pawnmigrate/cmd/pawnmigrate@latest
```

Check the installed version with `pawnmigrate --version`.

Start with a diff:

```sh
pawnmigrate --project . --output diff
```

If the diff looks right, apply it from a clean Git worktree:

```sh
pawnmigrate --project . --apply
```

Files are checked again before they are written. If a write fails, `pawnmigrate` tries to restore earlier files. Pawn source is formatted after migration.

To see which rules apply without building a plan:

```sh
pawnmigrate --project . --status
```

The default run includes only migrations marked safe. Pass `--allow-unsafe` to
include changes that need review, then inspect the diff before applying them.

Use `--only` to run specific migrations:

```sh
pawnmigrate --project . --only project.manifest-schema-v1,source.openmp-include --output diff
```

JSON output is available for tools and CI with `--output json`. `--allow-dirty` bypasses the clean-worktree check, but it also makes recovery harder.

Exit code `0` means the command completed, `2` reports invalid input or a
migration failure, and `3` reports an output or internal failure.

See [migration compatibility](docs/compatibility.md) for the rules currently shipped.

## Contributing

Migration ideas work best as a small before-and-after example. See
[CONTRIBUTING.md](CONTRIBUTING.md) for safety and review expectations.
