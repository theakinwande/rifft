# api-diff

A fast CLI tool that compares two OpenAPI 3.x specs and tells you exactly what changed — and whether it will break your consumers.

```
❌ BREAKING     DELETE /users/{id}            Endpoint removed
❌ BREAKING     POST /orders                  Required field 'currency' added to request body
❌ BREAKING     GET /products/{id}            Field 'price' type changed: number → string
⚠️  WARNING     GET /orders                   New enum value 'REFUNDED' added to 'status'
✅ NON-BREAKING POST /products                New optional field 'tags' added
✅ NON-BREAKING GET /health                   New endpoint added

Summary: 3 breaking, 1 warning, 2 non-breaking
```

---

## Install

```bash
go install github.com/theakinwande/api-diff/cmd@latest
```

Or build from source:

```bash
git clone https://github.com/theakinwande/api-diff
cd api-diff
go build -o api-diff ./cmd/main.go
```

---

## Usage

```bash
# Compare two local spec files
api-diff v1.yaml v2.yaml

# Compare from URLs
api-diff https://api.example.com/v1/openapi.yaml https://api.example.com/v2/openapi.yaml

# Output as JSON
api-diff v1.yaml v2.yaml --format json

# Exit with code 1 if any breaking changes are found (useful for CI)
api-diff v1.yaml v2.yaml --fail-on-breaking
```

---

## Change Detection

| Change | Classification |
|---|---|
| Endpoint removed | ❌ BREAKING |
| HTTP method changed | ❌ BREAKING |
| Required request body field added | ❌ BREAKING |
| Field type changed | ❌ BREAKING |
| Required parameter added | ❌ BREAKING |
| Parameter removed | ❌ BREAKING |
| New enum value added | ⚠️ WARNING |
| New endpoint added | ✅ NON-BREAKING |
| Optional field added | ✅ NON-BREAKING |
| Description or example changed | ✅ NON-BREAKING |

---

## JSON Output

```bash
api-diff v1.yaml v2.yaml --format json
```

```json
{
  "summary": {
    "breaking": 3,
    "warning": 1,
    "non_breaking": 2
  },
  "changes": [
    {
      "type": "BREAKING",
      "method": "DELETE",
      "path": "/users/{id}",
      "field": "",
      "description": "Endpoint removed"
    }
  ]
}
```

---

## CI / GitHub Actions

Add this to your workflow to block PRs that introduce breaking API changes:

```yaml
# .github/workflows/api-diff.yml
name: API Contract Check

on:
  pull_request:
    paths:
      - 'specs/**'

jobs:
  api-diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Install api-diff
        run: go install github.com/theakinwande/api-diff/cmd@latest

      - name: Check for breaking changes
        run: api-diff specs/openapi-base.yaml specs/openapi.yaml --fail-on-breaking
```

---

## Running Tests

```bash
go test ./...
```

19 tests across `parser`, `differ`, and `reporter` packages.

---

## Project Structure

```
api-diff/
├── cmd/main.go              # CLI entrypoint (Cobra)
├── parser/openapi.go        # Loads specs from files or URLs
├── differ/diff.go           # Core change detection engine
├── reporter/text.go         # Colored terminal output
├── reporter/json.go         # JSON output
├── testdata/
│   ├── v1.yaml              # Sample baseline spec
│   └── v2.yaml              # Sample updated spec (exercises all change types)
└── go.mod
```

---

## Built With

- [kin-openapi](https://github.com/getkin/kin-openapi) — OpenAPI 3.x parsing
- [cobra](https://github.com/spf13/cobra) — CLI framework
- [color](https://github.com/fatih/color) — Terminal colors

---
