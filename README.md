# go-cardano-policy

[![CI](https://github.com/njchilds90/go-cardano-policy/actions/workflows/ci.yml/badge.svg)](https://github.com/njchilds90/go-cardano-policy/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/njchilds90/go-cardano-policy.svg)](https://pkg.go.dev/github.com/njchilds90/go-cardano-policy)
[![Go Report Card](https://goreportcard.com/badge/github.com/njchilds90/go-cardano-policy)](https://goreportcard.com/report/github.com/njchilds90/go-cardano-policy)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
![Zero Dependencies](https://img.shields.io/badge/dependencies-zero-brightgreen)

**Zero-dependency Go library for Cardano native script construction and policy ID computation.**

Build and hash Cardano minting policies in pure Go — no CBOR libraries, no Cardano SDK, no network required. Deterministic, stateless, AI-agent-friendly.

---

## Why?

Every Cardano NFT or token project needs a **Policy ID** — the unique fingerprint of the minting policy. Computing one requires:

1. Constructing a **native script** (single-sig, multisig, time-locked, or combinations)
2. CBOR-encoding it using Cardano's ledger format
3. Hashing with **Blake2b-224**

Existing Go Cardano libraries bury this deep in large, dependency-heavy SDKs. `go-cardano-policy` exposes it as a clean, focused, zero-dependency API.

---

## Install
```bash
go get github.com/njchilds90/go-cardano-policy
```

Requires Go 1.21+. Zero external dependencies.

---

## Usage

### Single-Sig Policy (most common for NFTs)
```go
import policy "github.com/njchilds90/go-cardano-policy"

sig, err := policy.NewSigScript("e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd")
if err != nil {
    log.Fatal(err)
}

id, err := policy.Compute(sig)
if err != nil {
    log.Fatal(err)
}

fmt.Println(id) // 56-char hex policy ID
```

### Time-Locked NFT Policy (lock minting before a slot)
```go
sig, _ := policy.NewSigScript("e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd")
lock, _ := policy.NewBeforeScript(60000000) // only mintable before slot 60000000
all, _ := policy.NewAllScript(sig, lock)

id, err := policy.Compute(all)
fmt.Println(id)
fmt.Println(policy.IsTimeLocked(all)) // true
```

### 2-of-3 Multisig Policy
```go
sig1, _ := policy.NewSigScript("e09d36c79dec9bd1b3d9e152247701cd0bb0f68e8cd748d2a058bcd")
sig2, _ := policy.NewSigScript("a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d")
sig3, _ := policy.NewSigScript("deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

multisig, _ := policy.NewAtLeastScript(2, sig1, sig2, sig3)
id, _ := policy.Compute(multisig)
fmt.Println(id)
```

### Export to cardano-cli JSON
```go
jsonBytes, err := policy.ToJSON(all)
// Write to policy.json — feed directly to cardano-cli transaction policyid
os.WriteFile("policy.json", jsonBytes, 0644)
```

### Batch Compute
```go
ids, err := policy.ComputeMany(sig1, sig2, sig3)
```

### Context Support (for AI agents and servers)
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

id, err := policy.ComputeWithContext(ctx, script)
```

---

## Supported Script Types

| Type | Constructor | Description |
|------|------------|-------------|
| `sig` | `NewSigScript(keyHash)` | Requires one key's signature |
| `all` | `NewAllScript(scripts...)` | All sub-scripts must pass |
| `any` | `NewAnyScript(scripts...)` | At least one sub-script must pass |
| `atLeast` | `NewAtLeastScript(n, scripts...)` | At least n sub-scripts must pass |
| `before` | `NewBeforeScript(slot)` | Valid only before the given slot |
| `after` | `NewAfterScript(slot)` | Valid only after the given slot |

---

## For AI Agents

`go-cardano-policy` is designed for automated pipelines:

- **Pure functions** — no global state, no side effects
- **Deterministic** — same input always produces same Policy ID
- **Context-aware** — cancel long-running batch operations
- **Structured errors** — `ErrInvalidScript` with typed fields
- **JSON I/O** — machine-readable, round-trippable

---

## Design

- **Zero external dependencies** — only Go standard library
- **Embedded Blake2b-224** — pure Go implementation, no `golang.org/x/crypto`
- **CBOR encoding** — minimal, inline encoder for Cardano's native script format
- **Validates before hashing** — errors surface before any computation

---

## Related Libraries

| Library | Purpose |
|---------|---------|
| [go-cardano-metadata](https://github.com/njchilds90/go-cardano-metadata) | CIP-25 / CIP-68 NFT metadata |
| [go-cardano-fees](https://github.com/njchilds90/go-cardano-fees) | Transaction fee estimation |
| [go-cardano-asset](https://github.com/njchilds90/go-cardano-asset) | Asset name / fingerprint utilities |

---

## License

MIT — see [LICENSE](LICENSE)
