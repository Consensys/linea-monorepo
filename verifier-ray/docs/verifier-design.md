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
  (constraint identity)       (constraint identity)
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
- `sampler.zig` — `Sampler`: comptime-only namespace parametrized by per-round squeeze counts; the runtime transcript is owned by `replay` and passed by pointer, so ordering is enforced by `inline for` with no runtime check
- `root.zig` — `Spec`, `Context`, `replay()`; re-exports the public surface

## Coin Generation

Coins are KoalaBear extension-field elements derived via the Fiat-Shamir
transform: the prover's round messages are absorbed into a Poseidon2
Merkle-Damgård hasher; after each round the hasher state is squeezed to
produce verifier challenges.

### Compile-time: counts come from the protocol spec

Coin counts are fixed at compile time in the protocol spec:

```
round_coin_counts  = [0, 2, 0, M, M]   // coins squeezed after each round
round_coin_offsets = [0, 0, 2, 2, 2+M] // start index in all_coins[]
total_round_coins  = 2 + 2*M
```

`round_coin_counts[0]` is always 0 — no coins are derived before the first
round message is absorbed. `protocol.Sampler` is parametric on `advance_counts`
at compile time, so coin counts are never a runtime decision.

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
| Sampler state | `Sampler` owned transcript + `current_round` (runtime ordering check) | comptime-only namespace; transcript owned by `replay`, ordering enforced by `inline for` |
| Coin array allocation | heap-allocated via allocator, freed after verify | stack-allocated `[total_coins]Coin` in `replay`; no allocator needed |

The sub-verifier contract is now: **given pre-derived coins and cell openings,
check the mathematical identity. Nothing else.**

To add a new sub-verifier, only `verifier.zig` changes — see the how-to
comment at the top of that file.
