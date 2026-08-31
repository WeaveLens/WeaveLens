# Phase 11 — Graph Export

## Objective

Allow users to export a generated infrastructure graph without rescanning AWS.

## Formats

Implement in this order:

1. JSON
2. Draw.io
3. SVG
4. PNG

If implementation complexity becomes excessive, prioritize JSON and Draw.io.

## Architecture

Export operates on the canonical graph:

```text
Graph
 ↓
Export Service/Module
 ↓
Output
```

Export logic must not depend on AWS SDK.

## JSON

Export the canonical graph representation.

## Draw.io

Generate a Draw.io-compatible diagram containing:

* resources;
* relationships;
* labels;
* categories;
* visual metadata.

The resulting file should be editable.

## SVG / PNG

Reuse graph/layout information where practical.

Do not duplicate graph logic.

## Testing

Test:

* empty graph;
* simple graph;
* complex graph;
* duplicate resources;
* relationship rendering;
* valid JSON;
* valid Draw.io output.

## Acceptance Criteria

A graph displayed by WeaveLens can be exported without performing another AWS scan.

## Git

Commit:

```text
feat(export): add infrastructure graph export
```

Do not proceed automatically.
