# Migration authoring

A migration rule implements `migrate.Rule`. Its ID and version become part of reports, so choose them as public API.

`Plan` may inspect files and return `pawnkit-core/textedit` edits. It must honor cancellation and must not write to disk.

Mark a rule `safe` only when the parser and semantic model prove the edit belongs to the intended code. Macro-derived, unresolved, and heuristic matches are `review-required`.

Test the case you want to change, nearby code that must stay untouched, malformed input, and a second run over the migrated result. That last test catches rules which keep rewriting their own output.
