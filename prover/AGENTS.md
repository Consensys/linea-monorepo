# AGENTS.md — prover

> Inherits all rules from [root AGENTS.md](../AGENTS.md). Only overrides and additions below.

## Package Overview

Go-based ZK proof generation service for Linea. Uses gnark for circuit compilation, gnark-crypto for cryptographic primitives, and go-kzg-4844 for KZG polynomial commitments.

## How to Run

```bash
# Build
cd prover && make build

# Run tests (excludes corset, includes fuzz-light)
cd prover && go test ./... -tags nocorset,fuzzlight -timeout 30m

# Static checks
cd prover && gofmt -l .
cd prover && golangci-lint run

# Build Docker image
docker build -f prover/Dockerfile -t consensys/linea-prover:local .
```

## Go-Specific Conventions

- **Go version:** 1.24.6 (see `go.mod`)
- **Formatting:** `gofmt` (standard Go formatting, tabs)
- **Linting:** `golangci-lint`
- **Indentation:** tabs (per `.editorconfig` Go section)
- **Build tags:** `nocorset` and `fuzzlight` for standard test runs
- **Cross-compilation:** Targets Linux AMD64 (musl) and Darwin ARM64

### Key Dependencies

| Library | Purpose |
|---------|---------|
| `github.com/consensys/gnark` | ZK circuit compiler |
| `github.com/consensys/gnark-crypto` | Cryptographic primitives |
| `github.com/consensys/go-corset` | Constraint system |
| `github.com/crate-crypto/go-kzg-4844` | KZG polynomial commitments |
| `github.com/sirupsen/logrus` | Logging |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/prometheus/client_golang` | Metrics |

### Directory Structure

```
backend/     Backend logic
circuits/    Arithmetic circuits
zkevm/       ZK-EVM implementation
symbolic/    Symbolic execution
crypto/      Cryptographic utilities
protocol/    Protocol definitions
public-input/ Public input handling
config/      Configuration
```

## Go-Specific Safety Rules

- Never commit proving keys or large binary assets (checked via `.gitignore`)
- Cryptographic code changes require careful review — affects proof validity
- Test timeouts are 30 minutes due to proof generation complexity
- Memory-intensive operations: ensure adequate heap allocation

## Testing

- CI workflow: `.github/workflows/prover-testing.yml`
- Static checks run first (gofmt, golangci-lint)
- Compressor tests run separately from main test suite
- 30-minute timeout for test suite

## Agent Rules (Overrides)

- Always run `gofmt` and `golangci-lint` before proposing Go changes
- Do not modify circuit definitions without understanding the ZK proof implications
- Binary assets in `prover-assets/` are version-controlled selectively — check `.gitignore` exceptions

## Code style

* Less package is better to avoid dependency cycles.
* Use external library with parcimony
* Document each functions, method and structure. State the behavior of the function and its invariants. Document the parameters and the returns arguments.
* Only create new structures they are worth the added maintainance cost or if they decrease the maintainance cost.
* Prefer small functions. Less than 10 lines of code when possible.
* Small, non-technical utility functions go in `./utils`
* Use hard assertions a lot.
* Files should have 80 characters or less per lines, must not have more than 120.
* Files should not have more than 1000 lines of code unless there is a very good reason to.
* Error out only if failures are acceptable or expectable. If not, panic.
* Comment explains implementations key considerations. Use with parcimony. Don't paraphrase the code.
* Use docs.go for general package-level user documentations. Keep it short. Less than 1000 characters.
* Use Readme.<name>.md files for package-level maintainer documentation. Keep it less than 5000 character per file.
* If you want to document more stuffs, factor out the documentation in more Readme files but keep them shorts.
* Use new structures and interfaces with parcimony. Extend existing one when possible. 
