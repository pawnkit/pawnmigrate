# Compatibility

The current release knows about these migrations:

| Migration | Safety | What it changes |
|---|---|---|
| `project.manifest-schema-v1` | Safe | JSON and YAML project manifests |
| `source.openmp-include` | Review required | Parsed `a_samp` include directives |
| `api.deprecated-calls` | Review required | Calls matched against API metadata and semantic analysis |

Constants, tags, callbacks, and hooks are left alone because their meaning often
depends on project context.

The open.mp include migration needs review because pawnmigrate does not select a
target profile yet. It is never part of a default apply.

Migration IDs and versions appear in reports. If a transformation changes, its version should change too, so an old report still says what actually ran.
