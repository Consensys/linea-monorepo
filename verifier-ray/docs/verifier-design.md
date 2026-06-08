# Verifier Design

## verifier.verify Architecture

```
┌──────────────────────────────────────────────┐
│  verifier.verify                             │  ← entry point
│  1. replay  — drive transcript, derive coins │
│  2. route   — wrap coins+rounds in Context   │
│  3. dispatch — call each sub-verifier        │
└────────────────┬─────────────────────────────┘
                 │ protocol.Context
                 │  all_coins  []Coin
                 │  rounds     []RoundMessage
                 ▼
        ┌────────────────┬──────────────────┐
        ▼                ▼                  ▼
  vanishing.verify  logderiv.verify   ...future...
  (math only)       (math only)
```

## Package Structure

```
src/
├── protocol/       RoundMessage, ColumnMessage, Coin, Scalar, Sampler, Spec, Context, replay()
├── query/          sub-verifiers: vanishing, (logderiv, rangecheck, ...)
└── verifier.zig    entry point: Systems, ProofData, verify()
```

`protocol/` is internally split into three files:

- `types.zig` — wire types: `RoundMessage`, `ColumnMessage`, `Coin`, `Scalar`, `Visibility`
- `sampler.zig` — `Sampler`: compile-time parametric Fiat-Shamir driver; only called by `replay`
- `root.zig` — `Spec`, `Context`, `replay()`; re-exports the public surface

## Coin Generation

Coins are KoalaBear extension-field elements derived via the Fiat-Shamir
transform: the prover's round messages are absorbed into a Poseidon2
Merkle-Damgård hasher; after each round the hasher state is squeezed to
produce verifier challenges.

### Compile-time: counts come from the protocol spec

Each compiler in prover-ray registers coins by calling `round.NewCoinField`
during system construction. The Go codegen reads those registrations and emits
them as compile-time constants:

```
round_coin_counts  = [0, 2, 0, M, M]   // coins squeezed after each round
round_coin_offsets = [0, 0, 2, 2, 2+M] // start index in all_coins[]
total_round_coins  = 2 + 2*M
```

`round_coin_counts[0]` is always 0 — no coins are derived before the first
round message is absorbed. `protocol.Sampler` is parametric on `advance_counts`
at compile time, so coin counts are never a runtime decision.

### Codegen pipeline

The static Zig data file is produced by a Go pipeline that walks the compiled
`wiop.System` from prover-ray:

```
prover-ray (Go)
    │
    └── wiop.System          compiled IOP: rounds, coins, verifier actions
            │
            ├──────────────────────────────────────┐
            ▼                                      ▼
    BuildVanishingSystem()               BuildLogDerivSystem()   (future)
    → VanishingSystem{                   → LogDerivSystem{ ... }
        Modules,
        RoundCoinCounts, ...}
            │                                      │
            └──────────────┬───────────────────────┘
                           ▼
                  generator.System{
                      Rounds:    [...],
                      Vanishing: &vanishingSystem,
                      LogDeriv:  &logDerivSystem,  // future
                  }
                           │
                           ▼
                  generator.Generate()    emit Zig source
                      ├── emitVanishingSystem()  → module/expression/bucket constants
                      ├── emitLogDerivSystem()   → (future)
                      └── spec + systems literals (shared across all sub-verifiers)
                           │
                           ▼
                  src/generated/stub.zig   static Zig data, committed or generated at build time
                           │
                           ▼  (comptime parameters)
                  verifier.verify()        runtime proof checker
                      ├── protocol.replay()    absorb round messages, squeeze all coins
                      ├── vanishing.verify()   evaluate expression DAGs, check quotient identity
                      └── logderiv.verify()    (future)
```

To add a sub-verifier:

1. **Go side** — add `Build*System()` + a field to `generator.System` + `emit*System()` in `emitters.go`.
2. **Zig side** — follow the four-step how-to comment at the top of `verifier.zig`
   (`Systems` field → `ProofData` field → dispatch call; import is step 0).

`protocol.Spec` and `protocol.replay` are **shared and unchanged** — the new
sub-verifier reads its coins from the same pre-derived `ctx.all_coins`.

The generated file exports exactly two public constants:

`spec` and `systems` are passed as `comptime` parameters to `verifier.verify`;
they carry no runtime cost beyond what the Zig compiler inlines.

### Runtime: protocol.replay absorbs and squeezes

```
for each round i:
    absorb columns  → oracle commitments or public column values
    absorb cells    → opened scalar values
    squeeze round_coin_counts[i+1] coins → all_coins[coin_offsets[i+1]..]
```

`protocol.replay` is the only function that touches the Fiat-Shamir
transcript. It runs once, before any sub-verifier is called.

## What the Vanishing Sub-verifier Does

`vanishing.verify` receives pre-derived coins via `ctx.all_coins` and cell
openings via `ctx.rounds[i].cells`. It never touches the transcript.

For each module it:

1. Reads `merge_coin` (α) and `eval_coin` (r) from `ctx.all_coins` at the
   offsets fixed by the compiled system.
2. Computes the domain annihilator `Z_H(r) = r^n − 1`.
3. Evaluates the expression DAG, resolving `coin_value`, `cell_value`,
   `column_claim`, and `constant` leaves.
4. Aggregates vanishing numerators: `P_agg(r) = Σ αⁱ · Pᵢ(r) · Cᵢ(r)`.
5. Reconstructs the quotient: `Q(r) = Σ (r^n)^k · Qₖ(r)`.
6. Checks the PLONK identity: `P_agg(r) = Z_H(r) · Q(r)`.

## What Changed from the Previous Design

| Concern | Before | After |
|---|---|---|
| Transcript ownership | inside `vanishing.verify` | `protocol.replay`, called once by `verifier.verify` |
| Coin count per round | `next_round_coin_count` field in `RoundMessage` | compile-time constant in `protocol.Spec` |
| Sub-verifier input | `rounds + claims` (transcript data mixed with math) | `ctx + claims` (coins pre-derived, math only) |

The sub-verifier contract is now: **given pre-derived coins and cell openings,
check the mathematical identity. Nothing else.**

To add a new sub-verifier, only `verifier.zig` changes — see the how-to
comment at the top of that file.
