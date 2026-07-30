# Migration authoring

A migration rule implements `migrate.Rule`. Its ID and version become part of reports, so choose them as public API.

`Plan` may inspect files and return `pawnkit-core/textedit` edits. It must honor cancellation and must not write to disk.

Mark a rule `safe` only when the parser and semantic model prove the edit belongs to the intended code. Macro-derived, unresolved, and heuristic matches are `review-required`.

Test the case you want to change, nearby code that must stay untouched, malformed input, and a second run over the migrated result. That last test catches rules which keep rewriting their own output.

## Library migrations

Treat each direction as a separate migration. Do not assume that reversing a
rule restores the original program.

A library migration can coordinate:

- dependency and tool configuration changes;
- include replacements;
- symbol-aware call, callback, constant, and tag edits;
- a report of unsupported or ambiguous behavior.

Keep these changes in one preview when they must succeed together. Refuse the
migration if its required source library, target profile, or semantic facts are
missing. Use compiler and runtime fixtures for behavior that static analysis
cannot prove.

Built-in migration groups do not need a new tool. A public recipe or plugin
format shared with other repositories does require an RFC before implementation.
