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

* Prefer small functions. Less than 10 lines of code when possible.
* Small, non-technical utility functions go in `./utils`
* Use external library with parcimony
* Document each functions, method and structure. State the behavior of the function and its invariants. Document the parameters and the returns arguments.
* Use hard assertions a lot.
* Error out only if failures are acceptable or expectable. If not, panic.
* Comment explains implementations key-considerations. Use with parcimony. Don't paraphrase the code.
* Less package is better to avoid dependency cycles.
* Use docs.go for general package-level user documentations. Keep it less than 1000 characters.
* Use Readme.<name>.md files for package-level maintainer documentation. Keep it less than 2000 character per file.
* Use new structures and interfaces with parcimony.

* Small utility functions should be generic when possible (with regard for complexity) and they should be moved in a large `./utils` package when they are non-functional.
* When a function exists in the standard library of Go, we should not reimplement it.
* We use external (and non-standard) dependencies with parcimony but we may rely on gnark/gnark-crypto as much as possible.
* The code should be documented. Each function, each structure and each method must be documented. The documentation states what the function does. What the inputs are for. What the output will be. And if there is any pre-post condition and if there are side-effects. It also explains the error cases. The documentation should be succinct and informative.
* The code should be commented but with parcimony. Comments are not here to paraphrase the code. They are here to explain key design choices: "why this part is implemented like this and not like that", and non-trivial edge-case handling. It is a good-practice to comment bug-patch especially if they are non-trivials/
* Differentiating between panic-failures and error-failures is important. Whenever an error originate from a developper it is OK and encouraged to fail with panic. For instance, if a sanity-check fails or if an assertion is broken. If the error is to be expected by the application (like an external server not responding, or an external server providing invalid data), then this should be handled and good errors should be used
* Use small functions as much as possible. Many small and reusable functions are better than big, monolithic and ad-hoc functions. Thus, when a reusable function can be distilled from 2 bigger functions it should be done. Bonus point if the distilled function is non-functional and can be moved to the utils package. However, functions should be as context-free as possible. E.G. it should be possible to explain what the function does without having to explain WHY we need and the context in which it is needed.
* We prefer using less package to avoid dependency cycles.
* New structures and interfaces should be used with parcimony. Keeping the number of mental indirection under control is important to keep the code readable. It should be generic however and relying on type safety when possible is also good. So it's all about balancing. If we can drastically improve the type safety with almost no increase in complexity, then it's a good change. If we can reduce complexity a lot without sacrificing too much type safety then it is also good.
* Documentation-wise, we prefer having a `docs.go` file per package as a way to document the code. The package documention should explain the utility of the package, so it should focus on the "what" and not so much on the "how". So it can be understood as package user documentation. For maintainer documentation, we prefer to use `Readme.md` files. The main goal of the Readme.md is to provide context for the maintainer. If possible, we want to avoid as much as possible duplication in the documentation: it is better to refer to other package's documentation whenever that is possible. If the package is context heavy, we shall break the Readme.md into several files, each focusing on one aspect of the package. When we do that, the md document should be formatted as `Readme.<name>.md`. Each document should be reasonably small and condensed. Feel free to update the documentation, whenever you think some key information is missing but remember that it has to stay reasonably small and structured so do it with parcimony. Sometime, the key information is better to add on the function documentation.
* When doing a task, focus only on this task as much as you can.
* Testing is a fundamental aspect of the project. They have to be thorough but they also have to be maintainable. In Go, a good approach is table-driven tests as they are easy to extend and maintain and are also a good way to extend the coverage.
* When creating new function, method or structure, document the invariants (pre-conditions, post-conditions, invariant) of the structure. It's also good to assert them.