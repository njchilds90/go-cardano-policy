# Changelog

All notable changes to this project will be documented in this file.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Versioning follows [Semantic Versioning](https://semver.org/).

## [1.0.0] - 2026-02-25

### Added
- `Script` type with all Cardano native script variants: `sig`, `all`, `any`, `atLeast`, `before`, `after`
- `NewSigScript`, `NewAllScript`, `NewAnyScript`, `NewAtLeastScript`, `NewBeforeScript`, `NewAfterScript` constructors
- `Validate` — recursive structural validation
- `Compute` — Policy ID derivation (Blake2b-224 of CBOR-encoded script with type prefix)
- `ComputeWithContext` — context-aware variant for agents and servers
- `MustCompute`, `MustSigScript` — panic variants for constants and tests
- `ComputeMany` — batch policy ID computation
- `ToJSON` / `FromJSON` — cardano-cli compatible JSON serialization with round-trip fidelity
- `IsTimeLocked` — recursive time-lock detection
- `PolicyID` type with `String()` and `Bytes()` methods
- `ErrInvalidScript` structured error type
- Pure-Go Blake2b-224 implementation (zero external dependencies)
- Minimal inline CBOR encoder for Cardano native script format
- Table-driven tests with race detector
- GitHub Actions CI across Go 1.21, 1.22, 1.23
- Dependency check: CI fails if any external module is introduced
