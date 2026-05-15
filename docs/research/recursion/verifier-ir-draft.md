#  Unify Verify Across Compression and Aggregation



## Motivation

**Goal.** Refactor the verify logic that appears inside both *compression* and *aggregation*, and disentangle it from the surrounding commit / recommit / final-wrap steps that today's imperative chain conflates with it. The output is four declarative primitives — `InitialCommit`, `Verify`, `Recommit`, `FinalWrap` — composing into clean pipelines and making the load-bearing invariant structural: **every `Verify` must be followed by `Recommit`**.

`Verify` is the most visible win because it is the *common* operation across the two flows that look superficially different today:

- **Compression** uses `Verify` with the `WizardRewrite` backend (today's `selfrecursion.SelfRecurse` — rewrites a prior Vortex's verifier as wizard constraints).
- **Aggregation** uses `Verify` with the `PlonkRecurse` backend (today's `recursion.DefineRecursionOf` — builds a PLONK-in-wizard circuit that verifies k child proofs).

The compression chain (which embeds the `selfrecursion.SelfRecurse` step as its `Verify` half) is confusing for three reasons:

1. **The same skeleton is repeated 9 times across 5 files**, hand-tuned with parameters like `WithTargetColSize`, `ForceNumOpenedColumns`, and `VortexBlowup` that vary only slightly between call sites.
2. **Verification and commitment are mixed inside what looks like one operation**, even though they are distinct: `selfrecursion.SelfRecurse` produces wizard constraints, not a proof; the follow-up `Arcane → [MPTS] → Vortex` is what actually commits.
3. **Aggregation and compression are independent operations** that the imperative code conflates. A reader cannot tell from a call sequence which lines do aggregation, which do compression, and which do both.


---

## 1. The observation — where the compression chain is used

The "compression chain" skeleton is uniform:

```
SelfRecurse → CleanUp → (MiMC | Poseidon2) → Arcane(...) → [MPTS] → Vortex(...)
```

It appears in nine places across five files, falling into four structural roles:

| # | Site | Hash | Role |
|---|---|---|---|
| 1 | [zkevm/full.go:78-146](zkevm/full.go#L78-L146) `fullInitialCompilationSuite` | Poseidon2 | initial-compress (4 rounds, blowup 2→8→16→16) |
| 2 | [zkevm/full.go:151-219](zkevm/full.go#L151-L219) `fullInitialCompilationSuiteLarge` | Poseidon2 | scaled-up variant |
| 3 | [zkevm/full.go:223-266](zkevm/full.go#L223-L266) `fullSecondCompilationSuite` | Poseidon2 | post-recursion-compress |
| 4 | [protocol/distributed/segment_compilation.go:227-272](protocol/distributed/segment_compilation.go#L227-L272) | Poseidon2 | distributed-segment-compress |
| 5 | [protocol/distributed/segment_compilation.go:320-362](protocol/distributed/segment_compilation.go#L320-L362) | Poseidon2 | segment-recursion-compress |
| 6 | [circuits/pi-interconnection/keccak/.../segment_compilation.go:202-259](circuits/pi-interconnection/keccak/prover/protocol/distributed/segment_compilation.go#L202-L259) | **MiMC** | mirror of #4 |
| 7 | [circuits/pi-interconnection/keccak/.../segment_compilation.go:305-338](circuits/pi-interconnection/keccak/prover/protocol/distributed/segment_compilation.go#L305-L338) | **MiMC** | mirror of #5 |
| 8 | [circuits/pi-interconnection/keccak/customize_keccak.go:60-107](circuits/pi-interconnection/keccak/customize_keccak.go#L60-L107) | **MiMC** | custom-keccak-compress |
| 9 | [cmd/dev-tools/corset-checker/main.go:53-95](cmd/dev-tools/corset-checker/main.go#L53-L95) | Poseidon2 | dev/test suite |

The axes of variation are bounded:

- **Hash backend** — MiMC (pi-interconnection family) vs Poseidon2 (main protocol). The two trees are parallel copies of the same logic.
- **Round count** — 2 to 4.
- **Blowup progression** — typical values 2, 8, 16.
- **Preceding `plonkinwizard.Compile`** — present iff this is the first compress after building the original IOP.
- **MPTS injection** — present iff the next vortex anchors a verifying-key root.
- **Terminal markers** — `PremarkAsSelfRecursed()`, `AddPrecomputedMerkleRootToPublicInputs(...)` on the last vortex of any chain.

Sites `#4` and `#6` are *the same chain in different hash backends* — they will drift unless unified.

---

## 2. Aggregation and compression are different — aggregation always needs compression after

The compression chain from §1 is one of two **pipeline-level operations**. The other is **aggregation**. They are independent at the pipeline level — each is itself built from finer-grained primitives introduced in §3.

### Aggregation

`recursion.DefineRecursionOf(MaxNumProof=k)` — one PLONK recursion call that verifies k child proofs in a single statement.

- Arity-2 example: [conglomeration_hierarchical.go:319-324](circuits/pi-interconnection/keccak/prover/protocol/distributed/conglomeration_hierarchical.go#L319-L324) (`MaxNumProof: 2`).
- Arity-1 example: final segment wrap at [segment_compilation.go:287-302](circuits/pi-interconnection/keccak/prover/protocol/distributed/segment_compilation.go#L287-L302).

The gnark verifier circuit is **itself arity-1**. Arity-k aggregation is `MaxNumProof` independent gnark instances batched into a single PLONK witness via `PlonkCtx.Run` ([recursion.go:302-363](protocol/compiler/recursion/recursion.go#L302-L363)).

### Compression

The chain from §1 — a compile-time wizard rewrite that shrinks an IOP. Selfrecursion (`selfrecursion.SelfRecurse`) is one step inside it — specifically the *Verify* half. The *Recommit* half (`Arcane → [MPTS] → Vortex`) is what actually shrinks the proof.

### The invariant: aggregation must be followed by compression

> Every aggregation step must be followed by a compression chain.

Reason: an aggregation step produces a wizard containing **k proof-verifications worth of constraints**. Without compression, the next level would have to verify a wizard ~k× larger; iterating gives exponential blowup over depth. Compression re-folds those constraints into a compact committed wizard so the next layer sees something the same shape as before.

**Conglomeration shows this clearly.** [conglomeration_hierarchical.go:295-313](circuits/pi-interconnection/keccak/prover/protocol/distributed/conglomeration_hierarchical.go#L295-L313):

```go
conglo.Compile(comp, ...)                                    // ← bare aggregation
d.CompiledConglomeration = CompileSegment(conglo, params)    // ← compression chain
```

`ModuleConglo.Compile` is just `recursion.DefineRecursionOf(MaxNumProof=2)` — pure aggregation, no compression. The compression that makes the resulting wizard usable is in the subsequent `CompileSegment` call. 

---

## 3. Four primitives in pipeline order

A correct unification splits the pipeline into **four primitive node types** that compose in dataflow order:

```
InitialCommit  →  (Verify → Recommit) × N rounds  →  [FinalWrap]
```

- **`InitialCommit`** — every pipeline starts here.
- **`(Verify → Recommit)`** — the repeating unit; one full compression or aggregation step. Verify never appears without a Recommit after it.
- **`FinalWrap`** — optional terminal step; present only at the root of the proof tree (publishes the proof to L1).


### `InitialCommit`

```
InitialCommit{params}   :   original_IOP → (claim, proof)
```


### `Verify`

```
Verify{backend, arity}   :   (claims, proofs) → wizard_with_verification_constraints
```

Consumes k child proofs and rewrites their verification obligations into wizard constraints. **Produces an uncommitted wizard, not a proof** — that's why it must always be followed by `Recommit`.

Two backends:

| Backend | Arity | Mechanism | Today's code |
|---|---|---|---|
| `WizardRewrite` | 1 | Compile-time wizard rewrite — verifier checks become new wizard columns/constraints. | `selfrecursion.SelfRecurse` |
| `PlonkRecurse` | k | Build a PLONK-in-wizard circuit that verifies one child per gnark instance; arity-k batches k instances into a single PLONK statement. | `recursion.DefineRecursionOf{MaxNumProof:k}` |

`PlonkRecurse` *contains* `WizardRewrite` internally ([recursion.go:259](protocol/compiler/recursion/recursion.go#L259) calls `selfrecursion.RecurseOverCustomCtx()` per vortex context). So the IR must model nesting.

Claim representation, proof representation, and Fiat-Shamir placement differ by backend:

| | `WizardRewrite` | `PlonkRecurse` |
|---|---|---|
| Claim representation | wizard columns (Merkle roots, row counts, PI) | flat field elements (X, Ys, Commitments, Pubs) |
| Proof representation | wizard runtime state (`MerkleProofs`, `OpenedColumns`, …) | `wizard.Proof` + matrices as gnark witness |
| FS location | external — coins pre-sampled before the rewrite runs | internal — `KoalaFS` runs inside the gnark circuit |
| File refs | [selfrecursion/context.go:15-150](protocol/compiler/selfrecursion/context.go#L15-L150); [vortex/verifier.go:127-153](protocol/compiler/vortex/verifier.go#L127-L153) | [recursion/circuit.go:26-35](protocol/compiler/recursion/circuit.go#L26-L35); [recursion/recursion.go:65-91](protocol/compiler/recursion/recursion.go#L65-L91) |

### `Recommit`

```
Recommit{params, role}   :   wizard → (claim, proof)
```

Takes the uncommitted wizard produced by `Verify` (or by the initial Arcane pass) and produces a fresh `(claim, proof)` via a new vortex commitment.

Today: `compiler.Arcane(...) → [mpts.Compile(...)] → vortex.Compile(...)`.

**Mandatory after every `Verify`** — this is the invariant from §2 made structural.

Role flag:
- `Inner` — a middle compression step.
- `TerminalChild` — last compression in this branch; sets `PremarkAsSelfRecursed()`, marking the result as a leaf for the parent aggregator ("treat me as input, don't compress me further").

`AddPrecomputedMerkleRootToPublicInputs(...)` (verifying-key root exposure) and MPTS injection live on `Recommit` as `expose: *PIName` and `mpts: *MPTSProfile` fields.

### `FinalWrap`

```
FinalWrap{curve, params}   :   wizard → published_proof
```

Terminates the root of the proof tree by compiling the final wizard into a monolithic gnark PLONK proof in the target curve (today: BN254). The only primitive whose output is a public artifact rather than a wizard. Bounded by BN254's `nbConstraints + nbPublic ≤ 2^26` ceiling — the reason `ForceNumOpenedColumns(48)` is necessary in `TreeAggregationFinalCompilationSuite`.

---

## 4. How to use the primitives — pipeline shapes

The four primitives compose into three pipeline shapes that cover every flow in the codebase today.

### 4.1 Compression pipeline (single segment / leaf)

A GL/LPP segment proof. Output is consumed by a parent aggregator, never directly published.

```
       original_IOP
            │
            ▼
       InitialCommit{blowup: 2, opened: 256, expose: lppMerkleRoot?}
            │
            ▼
    ┌─ Verify{WizardRewrite, arity: 1}            ┐
    │       │                                     │
    │       ▼                                     │  N-1 inner rounds
    │  Recommit{Inner, target: 1<<15, ...}        │
    │       │                                     │
    └──────...                                    ┘
            │
            ▼
       Verify{WizardRewrite, arity: 1}
            │
            ▼
       Recommit{TerminalChild, expose: VerifyingKey}      ← PremarkAsSelfRecursed
            │
            ▼
       (claim, proof)            ← consumed by parent aggregator or by FinalWrap
```

Same shape covers [protocol/distributed/segment_compilation.go:202-258](protocol/distributed/segment_compilation.go#L202-L258) (Poseidon2) and [circuits/pi-interconnection/.../segment_compilation.go:202-258](circuits/pi-interconnection/keccak/prover/protocol/distributed/segment_compilation.go#L202-L258) (MiMC). They differ only by `Recommit{Hash: …}`.

### 4.2 Aggregation step (one tree node)

A k-to-1 aggregator consuming child proofs from k pipelines of shape 4.1. Notice that the post-aggregation compression chain is *the same* shape 4.1 body — the invariant in §2 made concrete.

```
   (claim_1, proof_1)   ...   (claim_k, proof_k)
            \                       /
             \_____________________/
                       │
                       ▼
            Verify{PlonkRecurse, arity: k}              ← aggregation
                       │
                       ▼
                Recommit{Inner}                          ← committed aggregated wizard
                       │
                       ▼
        (Verify{WizardRewrite, arity: 1} ; Recommit{Inner})*   ← MUST compress after
                       │
                       ▼
               Recommit{TerminalChild}
                       │
                       ▼
                  (claim, proof)
```

Conglomeration ([conglomeration_hierarchical.go:295-313](circuits/pi-interconnection/keccak/prover/protocol/distributed/conglomeration_hierarchical.go#L295-L313)) is exactly this with k=2.

### 4.3 Full pipeline — tree aggregation terminating at L1

The full proof pipeline composes the two shapes above and terminates at `FinalWrap`:

```
        [2^d leaf pipelines, shape 4.1]
                    │
                    ▼
        [level 0: 2^(d-1) aggregation steps, shape 4.2, k=2]
                    │
                    ▼
        [level 1: 2^(d-2) aggregation steps, shape 4.2, k=2]
                    │
                    ▼
                  ...
                    │
                    ▼
        [level d-1: 1 aggregation step (root)]
                    │
                    ▼
        FinalWrap{curve: BN254}                          ← TERMINATES tree
                    │
                    ▼
             published proof                             ← what L1 verifies
```

Two termination points coexist: every `Recommit{TerminalChild}` is *locally* terminal for its branch (and serves as input to the next level); `FinalWrap` is *globally* terminal for the whole tree.

### Record types — the pipeline IR as data

A `Pipeline` is a flat slice of nodes; each node is one of the four primitives. The interpreter walks this slice in order and emits the corresponding `wizard.ContinueCompilation(...)` calls. (`oneof` below is sum-type notation — concretely an interface in Go.)

```go
type Pipeline []Stage
type Stage   = oneof { InitialCommit, Verify, Recommit, FinalWrap }

type InitialCommit struct {
    Hash          HashBackend       // MiMC | Poseidon2
    Blowup        int               // typically 2
    OpenedColumns int               // typically 256
    SIS           SISParams
    SISThreshold  int
    ExposeRoot    *PublicInputName  // app-level PI, e.g., lppMerkleRoot
}

type Verify struct {
    Backend                VerifyBackend        // WizardRewrite | PlonkRecurse
    Arity                  int                  // 1 for WizardRewrite; k for PlonkRecurse
    Params                 recursion.Parameters // PlonkRecurse only
    PublicInputRestriction []string             // currently special-cased
}

type Recommit struct {
    Hash          HashBackend
    TargetColSize int                  // 1<<14, 1<<15, ...
    StitcherMin   int
    Blowup        int                  // typically 8
    OpenedColumns int                  // typically 40-48
    SIS           SISParams
    SISThreshold  int
    MPTS          *MPTSProfile         // optional; mpts.Compile inserted before vortex
    Role          RecommitRole         // Inner | TerminalChild
    ExposeRoot    *PublicInputName     // optional VK root exposure
}

type FinalWrap struct {
    Curves       []ECCurve             // typically [BN254]; old pipeline was [BLS12_377, BW6_761, BN254]
    PlonkParams  PlonkParams
    PublicInputs []PublicInputName
}
```

---

## 5. What's new and how to proceed

### What's just a representation change (no behavior change)

- Folding the 9 chain sites into one interpreter over `[]Stage`.
- Unifying the MiMC and Poseidon2 trees — sites `#4` and `#6` collapse to one record with `Hash` set differently.
- Exposing `TargetColSize`, blowup, opened-cols as data instead of literals at five call sites.


### Open questions to lock down before coding

1. **Public-input plumbing.** Three special-cased PI groups exist: `lppMerkleRootPublicInput`, `VerifyingKey*PublicInput`, `conglomeration*`. The IR needs first-class PI annotations (`passthrough | fold-into-digest | expose-as-root`) so the special-casing in [segment_compilation.go:266-282](circuits/pi-interconnection/keccak/prover/protocol/distributed/segment_compilation.go#L266-L282) goes away. **Get this wrong and the IR won't capture all pipelines.**
2. **FS binding for arity ≥ 2.** Each aggregation must consume child claims into its transcript seed (or admit sibling malleability). Today this is implicit in `KoalaFS`'s init order — the IR should make it a property of the `Verify` node.
3. **Compile-time vs prove-time split.** `Verify{WizardRewrite}` is compile-time; `Verify{PlonkRecurse}` is gnark-circuit-build (still compile-time, before prove); `Recommit` is always compile-time wizard rewrite; `FinalWrap` is gnark-circuit-build + prove-time PLONK. The IR is a *plan*; realization is a separate concern that must not leak into the IR shape.
4. **Verifier-side replay.** Letting the verifier (native and gnark) read the same pipeline IR to replay is a strictly larger lift than the prover-side refactor — needs explicit design for what gets serialized, versioning, and FS transcript binding. Out of scope for the next steps below; flagged here for completeness.

### Concrete next steps, in dependency order

1. **Lock down PI annotation vocabulary** (open question #1). One page.
2. **Define the `Stage` sum-type and primitive records** to exactly cover sites #1, #4, #6 (one Poseidon2, two MiMC).
3. **Build the interpreter** that consumes `[]Stage` and emits today's `wizard.ContinueCompilation(...)` calls. No new capability yet.
4. **Migrate site #1** behind a feature flag. Compare wizard hashes / proof sizes / prove-time against the imperative version.
5. **Migrate sites #4 and #6 together**, eliminating the MiMC/Poseidon2 duplicated tree.
6. **Add `FinalWrap`**, migrate finalwrap and conglomeration through the IR.
7. **Then** consider the planner, search, and serializable-IR-for-verifier work.

Steps 1-5 are the "small data refactor" portion. Steps 6-7 are where the IR pays off.
