# Contributing

Thank you for your interest in `go-cardano-policy`!

## Philosophy

This library is intentionally minimal and zero-dependency. Please respect that:

- **No external imports.** The `go.mod` must never list a `require` block.
- **No global state.** All exported functions must be pure and stateless.
- **Deterministic outputs.** Same input = same output, always.
- **Cardano spec fidelity.** CBOR encoding and policy ID computation must match `cardano-cli` exactly.

## How to Contribute

1. Fork the repository
2. Create a branch: `git checkout -b feature/your-feature`
3. Write code and tests (table-driven preferred)
4. Ensure `go test -race ./...` passes
5. Ensure `go vet ./...` passes
6. Open a pull request with a clear description

## Adding a New Script Type

If Cardano introduces a new native script type in a future era:
1. Add the `ScriptType` constant in `script.go`
2. Add a constructor `NewXxxScript(...)`
3. Add a case in `Validate`
4. Add a case in `writeScript` in `policy.go`
5. Add table-driven tests covering valid and invalid inputs

## Reporting Bugs

Please include:
- Go version (`go version`)
- The script JSON that produced the wrong Policy ID
- The Policy ID you computed and the one `cardano-cli` produced

## License

By contributing, you agree your contributions are licensed under MIT.
```

---

## Release & Verification Instructions
```
RELEASE STEPS (GitHub Web UI Only)
====================================

1. Verify CI is green:
   → Go to: https://github.com/njchilds90/go-cardano-policy/actions
   → Confirm the latest CI run shows all green checkmarks

2. Create the tag:
   → Go to: https://github.com/njchilds90/go-cardano-policy/releases
   → Click "Draft a new release"
   → Click "Choose a tag" → type: v1.0.0 → click "Create new tag: v1.0.0 on publish"
   → Target branch: main

3. Fill in the release form:
   Title:   v1.0.0 — Initial Release
   Body:
   ───────────────────────────────────────
   ## go-cardano-policy v1.0.0

   First stable release.

   ### What's included
   - Native script construction: `sig`, `all`, `any`, `atLeast`, `before`, `after`
   - Policy ID computation (Blake2b-224, CBOR-encoded, Cardano ledger-compatible)
   - cardano-cli JSON serialization / deserialization
   - Zero external dependencies
   - Context support for AI agent pipelines
   - Table-driven tests with race detector

   ### Install
```
   go get github.com/njchilds90/go-cardano-policy@v1.0.0
```
   ───────────────────────────────────────

4. Click "Publish release"

5. Verify on pkg.go.dev (wait ~5-15 minutes):
   → https://pkg.go.dev/github.com/njchilds90/go-cardano-policy@v1.0.0

   If it hasn't appeared, force-fetch it:
   → https://pkg.go.dev/github.com/njchilds90/go-cardano-policy@v1.0.0?tab=doc
   (visiting the URL triggers the indexer)

SEMANTIC VERSIONING GUIDANCE
==============================
v1.0.x  — Bug fixes, no API changes
v1.x.0  — New script types, new helpers (backward compatible)
v2.0.0  — Only if Script struct shape must change (breaking)
