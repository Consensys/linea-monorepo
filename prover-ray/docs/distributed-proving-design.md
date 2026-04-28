# Distributed Proving — Design Document

## 1. Context: the current Limitless prover

The current implementation lives in `prover/backend/execution/limitless/` and
`prover/protocol/distributed/`. It proves one EVM block by running a single
monolithic pipeline on one machine:

```
Bootstrapper (one giant wizard.ProverRuntime)
    │
    │  Arithmetization.Assign() writes ALL columns into the runtime
    │
    ▼
SegmentRuntime()
    │  Scans the completed runtime, slices columns into ModuleWitnessGL[]
    │  and ModuleWitnessLPP[] for every (module, segment) pair
    ▼
GL segment provers  ──────────────────────────────────────────┐
    │  Run in parallel, each produces a SegmentProof            │
    │  with LppCommitment = Poseidon2(LPP columns)              │
    ▼                                                           │
    Wait for ALL GL proofs to finish                           │
    │                                                           │
    ▼                                                           │
GetSharedRandomnessFromSegmentProofs(allGLProofs)              │
    │  Multiset-hash all (moduleIdx, segIdx, LppCommitment)     │
    │  → field.Octuplet (the shared Fiat-Shamir seed)           │
    ▼                                                           │
LPP segment provers                                            │
    │  Each uses sharedRandomness as InitialFiatShamirState     │
    │  Run in parallel                                          │
    ▼                                                           │
HierarchicalConglomeration ◄───────────────────────────────────┘
    │  Receives all GL + LPP proofs via a channel
    │  Binary-tree reduction: 2 proofs → 1, until 1 remains
    ▼
Outer BLS12-377 proof
```

Key code references:
- `prove.go:RunDistributedPipeline` — the pipeline orchestration
- `distribute.go:GetSharedRandomnessFromSegmentProofs` — shared randomness
- `distribute.go:SegmentRuntime` — column extraction from the runtime
- `module_witness.go` — `ModuleWitnessGL`, `ModuleWitnessLPP` types


## 2. Problems with the current architecture

### 2.1 Serial GL → shared randomness → LPP dependency

LPP provers cannot start until **every** GL proof has finished. In a 100-segment
block with, say, 50 GL and 50 LPP segments, the LPP provers sit idle while the
slowest GL prover finishes. Shared randomness is derived from the GL proofs'
`LppCommitment` outputs, but that commitment is just a Poseidon2 hash of the raw
LPP column data — it does not actually require running the GL circuit.

### 2.2 One process, one machine

`RunDistributedPipeline` is a single Go function on a single machine. All
witnesses live in RAM simultaneously. The concurrency model is goroutines within
one OS process, not tasks distributed across a cluster.

### 2.3 No fault tolerance

If any segment prover panics, the `errgroup` cancels all others and the whole
block fails. There is no retry, no checkpointing, no partial-result recovery.

### 2.4 Straggler problem

With N concurrent goroutines, the wall-clock time of the slowest segment sets
the time for the whole batch. In a homogeneous goroutine pool a 10 % performance
outlier extends total time by close to 10 %; in a heterogeneous cluster the
effect is much worse.

### 2.5 Monolithic segmentation

`SegmentRuntime` scans the entire wizard runtime after a single monolithic
`Arithmetization.Assign` call. Arithmetization and the prover are tightly
coupled through `wizard.ProverRuntime` — a heavy internal type. There is no
way to pipeline tracing with proving.


## 3. Key design decisions in the new architecture

### Decision 1: N static `.bin` files replace the monolithic `zkevm.bin`

The arithmetization team ships one `.bin` file per **segment kind** (e.g.
`keccak.bin`, `alu.bin`, `ec_precompiles.bin`). Each `.bin` file defines:
- Which columns the segment kind has
- The constraint system (all constraints except vanishing)
- The **Fiat-Shamir schedule**: a list of column-ID groups, one per round

The prover is compiled from these `.bin` files. Neither side tells the other
what a segment kind proves at runtime; both sides derive the same column
structure independently by compiling the same `.bin` file.

### Decision 2: The Fiat-Shamir schedule makes preflight explicit

Compiling a `.bin` file yields `FSSchedule [][]ColumnID`:

```
FSSchedule[0]   = columns committed before any verifier challenge
                = the "looked-up" / LPP columns
                = what arithmetization must emit early (preflight)
                = what the prover commits to derive shared randomness

FSSchedule[1..] = segment-local rounds; not shared across kinds
```

Shared randomness = commit each segment's `FSSchedule[0]` columns →
multiset-hash all commitments. This requires no GL circuit execution.

### Decision 3: Arithmetization owns segmentation and tracing

In the old design the prover calls `SegmentRuntime(runtime)` to pull columns
out of its own internal `wizard.ProverRuntime`. In the new design arithmetization
produces typed, pre-segmented witnesses (`ModuleWitnessGL`, `ModuleWitnessLPP`)
directly. The prover never sees a `wizard.ProverRuntime` from arithmetization.

### Decision 4: Task-queue model replaces the synchronized pipeline

Each unit of work is a stateless, idempotent **task** identified by
`(blockID, kindIndex, segmentIndex)`. Tasks are placed in a shared queue and
consumed by any available worker on any machine. There is no global barrier.

The **only coordination event** is:
> all `PreflightCommitTask`s for a block complete
> → compute shared randomness (single hash, microseconds)
> → enqueue all `LPPProveTask`s

GL tasks are never gated on this event.


## 4. Arithmetization ↔ Prover interface

### 4.1 Setup time (once per deployment)

```
Both sides compile the same .bin files:

    distributed.CompileSegmentKind("keccak.bin", vkRoot) → *SegmentKind
    distributed.CompileSegmentKind("alu.bin",    vkRoot) → *SegmentKind
    ...

    // Prover derives its config from the same SegmentKinds:
    arithmetization.Configure(distributed.ArithmetizationConfig(kinds))

    // Prover compiles its circuits from the same SegmentKinds:
    distributed.CompileProver(kinds[i]) → *compiledProver
```

`SegmentKind.ArithmetizationBlueprint()` maps circuit structure to the value
type `arithmetization.ModuleBlueprint` — a plain struct with no pointers to
circuit internals, safe to serialise or send across a process boundary.

The mapping that crosses the boundary:
```
SegmentKind.FSSchedule[0]  →  ModuleBlueprint.LPPColumnIDs
SegmentKind.AllColumns     →  ModuleBlueprint.AllColumnIDs
SegmentKind.ReceivedValues →  ModuleBlueprint.ReceivedValuesAcc*
SegmentKind.N0SelectorIDs  →  ModuleBlueprint.N0SelectorIDs
```

Arithmetization does not need to know about Fiat-Shamir, LPP, or shared
randomness. From its perspective `LPPColumnIDs` is simply "the columns I must
emit first".

### 4.2 Per-block runtime

```
arithmetization.Run(tracePath) → (preflightCh, traceCh)
```

| Channel | Content | When sent |
|---------|---------|-----------|
| `preflightCh` | `[]PreflightSegment` — one entry per (kind, segment), containing only `LPPColumnIDs` columns | As soon as `FSSchedule[0]` columns are assigned; **before** the full trace is done |
| `traceCh` | `TracingResult` — `ModuleWitnessGL[]` + `ModuleWitnessLPP[]` | When the full trace is expanded |

`ModuleWitnessLPP.InitialFiatShamirState` is **always zero** when arithmetization
sends it. Injecting shared randomness is the prover's exclusive responsibility.

### 4.3 What arithmetization never does

- Does not compute commitments
- Does not run any hash involving randomness
- Does not wait for any prover result
- Does not know the word "Fiat-Shamir"


## 5. Data flow in the new architecture

```
.bin files (static, shipped by arithmetization team)
    │
    ├── CompileSegmentKind() ──────────────────────────────────────────────┐
    │       → SegmentKind (FSSchedule, AllColumns, ...)                    │
    │       → arithmetization.Config (via ArithmetizationConfig())         │
    │       → compiledProver (prover circuit, per kind)                    │
    │                                                                       │
    │  SETUP: both sides run this once; no runtime hand-off needed         │
    └──────────────────────────────────────────────────────────────────────┘

Per-block proving:

arithmetization.Run(tracePath)
    │
    ├─ [fast] preflight pass: extract FSSchedule[0] columns
    │         ──► preflightCh ──► PreflightCommitTask (per segment) ──► queue
    │                                     │
    │                             Worker: CommitLPPColumns()
    │                                     │ (Poseidon2 Merkle root)
    │                                     ▼
    │                             Coordinator.OnPreflightResult()
    │                                     │
    │                         [all N commits received]
    │                                     │
    │                             GetSharedRandomness(commitments)
    │                                     │ (single multiset hash)
    │                                     ▼
    │                             for each LPP witness:
    │                               w.InitialFiatShamirState = sharedRandomness
    │                               queue.EnqueueLPP(LPPProveTask)
    │
    └─ [slow] full segmentation: extract all columns
              ──► traceCh ──► for each GL witness:
                                  queue.EnqueueGL(GLProveTask)
                              (LPP tasks already enqueued by coordinator above)

Workers (stateless, any machine, any number):
    │
    ├── GLProveTask  → prove GL circuit → write proof → OnProofResult
    ├── LPPProveTask → prove LPP circuit → write proof → OnProofResult
    └── MergeTask   → merge 2 proofs → write merged proof → OnMergeResult

Coordinator.OnProofResult (greedy):
    │
    ├── [< 2 proofs available] → wait
    └── [≥ 2 proofs available] → EnqueueMerge immediately
                                      │
                               [root merge] → final proof
```


## 6. How each identified issue is addressed

| Issue | Root cause in current design | Solution |
|-------|------------------------------|----------|
| **LPP waits for all GL proofs** | Shared randomness derived from GL proof outputs (`LppCommitment`) | Shared randomness derived from raw column commitments in preflight — no GL circuit needed |
| **One crash fails the block** | `errgroup` cancels all goroutines | Task-queue: failed tasks are re-delivered to any live worker; completed work is not lost |
| **One machine, one process** | `RunDistributedPipeline` is a single function call | Workers are stateless processes that pull from a shared queue; run on any number of machines |
| **Memory pressure** | All witnesses held in RAM until all proofs done | Each task carries only its own witness; proved witnesses can be freed immediately |
| **Straggler slows the batch** | Fixed goroutine-per-segment assignment | Work-stealing: fast workers pull the next task immediately; no worker idles waiting for a slow peer |
| **Sync before conglomeration** | Conglomeration started after all segment proofs landed | Greedy merge pairing: any two available proofs → `MergeTask` immediately; merge tree grows as leaves arrive |
| **Tight coupling through `wizard.ProverRuntime`** | `SegmentRuntime` scans prover's internal runtime | Arithmetization produces typed pre-segmented witnesses (`ModuleWitnessGL`, `ModuleWitnessLPP`); prover never sees the runtime |


## 7. Package structure

```
prover-ray/
├── arithmetization/
│   ├── types.go           — PreflightSegment, ModuleWitnessGL, ModuleWitnessLPP,
│   │                         ModuleBlueprint, TracingResult
│   │                         (the typed values that cross the boundary)
│   └── arithmetization.go — Arithmetization.Configure(cfg) + Run(tracePath)
│                             (two-channel output: preflight early, full trace later)
│
├── distributed/
│   ├── segment_kind.go    — SegmentKind: compiled from one .bin file;
│   │                         holds FSSchedule, AllColumns, ReceivedValues*, N0Selectors
│   │                         CompileSegmentKind(binPath) and CompileProver(kind)
│   ├── blueprint.go       — SegmentKind.ArithmetizationBlueprint()
│   │                         ArithmetizationConfig(kinds) — projects circuit structure
│   │                         into the plain ModuleBlueprint value type
│   └── preflight.go       — CommitLPPColumns(seg) → LPPCommitment
│                             CommitAll(segs) → []LPPCommitment
│                             GetSharedRandomness([]LPPCommitment) → field.Octuplet
│
└── pipeline/
    ├── task.go            — PreflightCommitTask, GLProveTask, LPPProveTask, MergeTask
    │                         TaskQueue interface, WorkerResult
    ├── coordinator.go     — Coordinator: tracks preflight commits per block,
    │                         fires shared-randomness event, pairs proofs for merging
    └── pipeline.go        — Prover.Prove() — enqueues tasks, returns immediately
                             Worker.Run()   — stateless task executor, any machine
```


## 8. RISC-V zkVM versus zkEVM — impact on the design

The new design targets a RISC-V zkVM (Risc 5), not the EVM. This section
clarifies how sharding works in the RISC-V setting and what it means for the
design.

### 8.1 How sharding works: the shard is the primary unit

Execution is a **single continuous stream of CPU cycles**. Sharding cuts this
stream into time-bounded slices:

```
cycle:  0   1   2   3   4   5   6   7   8   9  10  ...
        ├─────────────────── shard 0 ──────────────┤ ├────── shard 1 ──
                              ▲
                        shard boundary: some chip hit its row threshold
```

**Each CPU cycle contributes rows to zero or more chips:**

```
cycle N executes a memory-load ALU instruction:
    CPU chip:     +1 row  (always)
    Memory chip:  +1 row  (memory access)
    ALU chip:     +1 row  (arithmetic)
    Hash chip:    +0 rows (no hash this cycle)
    Bitwise chip: +0 rows
```

**The shard boundary fires when any chip's row counter reaches its threshold.**
Thresholds are measured in rows, not in CPU cycles:

```
cpu.bin    threshold: 2^20 rows  → shard ends after at most 2^20 cycles
memory.bin threshold: 2^18 rows  → may end earlier if memory-heavy
alu.bin    threshold: 2^20 rows
hash.bin   threshold: 2^16 rows  → lower because hash circuits are expensive
```

Thresholds may be uniform or per-chip. Different thresholds let the system
target roughly equal prover time per chip regardless of circuit complexity.

**Each shard contains exactly one CPU segment plus one chip segment per chip
that was active in that cycle range.** A chip that received zero rows in a
shard produces no segment and no proof for that shard.

### 8.2 The CPU trace is the clock; chips accumulate passively

The CPU trace is the backbone of execution:
- It is the only trace that is inherently sequential (cycle N+1 depends on
  cycle N's register file and PC).
- It is the source for detecting shard boundaries — you count chip rows as
  you scan the CPU trace, and cut when any counter hits its threshold.
- It is on the **query side** of the LogUp bus: the CPU trace says "I performed
  this operation", and the chip traces say "here is the record proving it".

The chip/module traces are derived from the CPU trace:
- **Memory chip**: one row per memory read/write instruction in the CPU trace.
- **ALU chip**: one row per arithmetic instruction.
- **Hash chip**: one row per hash invocation.
- etc.

Chips do not drive the shard boundary; they just accumulate rows as the CPU
scans forward. The CPU trace is the single source of truth for when to cut.

### 8.3 Within a shard: CPU is query side, chips are table side

Each shard's proof set covers:

```
shard k:
    CPU segment k      — proves "I executed cycles [C_k, C_{k+1}) correctly"
                         query side: looks up that each chip claim was satisfied
    Memory segment k   — proves "here are all memory ops in cycles [C_k, C_{k+1})"
                         table side: looked up by the CPU trace
    ALU segment k      — same for ALU ops
    Hash segment k     — same for hash ops
    …
```

The CPU segment is the query side; chip segments are the table side. This maps
directly to the GL/LPP split already in the design:

```
CPU segment  → proves as a GL (Global Lookup) witness
               its LogUp queries are checked against the chip table commitments
Chip segments → provide the table columns
                FSSchedule[0] of each chip = the preflight columns
                (committed early to derive shared randomness)
```

### 8.4 The carry chain spans shards

Between shards, the CPU state (register file, PC, memory root hash) must be
forwarded from shard k to shard k+1. This is the `ReceivedValuesGlobal` carry:

```
arithmetization (sequential — cannot be parallelised):
    scan shard 0 → emit carry_0 (registers, PC, mem root)
    scan shard 1 with carry_0 as input → emit carry_1
    scan shard 2 with carry_1 as input → …

proving (fully parallel — carry is already baked into the witness):
    ProveTask(shard 0, carry_0)  ─┐
    ProveTask(shard 1, carry_1)  ─┼─ all enqueued simultaneously
    ProveTask(shard 2, carry_2)  ─┘
```

The sequential dependency lives entirely inside arithmetization's tracing pass.
Once each shard's witness includes its incoming carry, all shard provers are
independent. No change to the task-queue or coordinator is needed.

### 8.5 FSSchedule[0] = chip table columns (not CPU columns)

Shared randomness is derived from the chip table columns (the table side of the
LogUp bus), not from the CPU trace (the query side). This is why preflight is
scoped to chip/functional segments:

```
Preflight per shard k:
    memory_chip.FSSchedule[0] for shard k  ← committed early
    alu_chip.FSSchedule[0] for shard k     ← committed early
    hash_chip.FSSchedule[0] for shard k    ← committed early
    (NOT the CPU columns — those are on the query side)

After all preflight commits:
    GetSharedRandomness(all shard × chip commitments) → single field.Octuplet
    → injected into every CPU segment witness as InitialFiatShamirState
```

A practical benefit: chip table columns for a given shard can be extracted as
soon as that shard's cycles have been scanned, independent of the full
sequential CPU carry chain. Preflight can therefore start as shards are emitted,
not only after all shards are done.

### 8.6 Dynamic segment counts replace fixed TracesLimits

In zkEVM each module had a compile-time maximum size (`TracesLimits`). Segment
counts were bounded and known statically.

In RISC-V the trace length depends on the program being proved:
- A shard with no hash invocations produces zero hash-chip rows → no hash proof.
- A memory-heavy program produces many memory rows → more shards or more
  memory-chip proofs per shard.
- Total shard count is unknown until tracing completes.

Consequences for the design:

1. **`TotalSegmentCount` is fully dynamic.** Arithmetization reports it at
   runtime; it cannot be known at setup time.

2. **The coordinator handles 0-row chips gracefully.** A chip with zero rows
   in a shard produces no task, no preflight commitment, and no proof. It
   contributes nothing to shared randomness for that shard.

3. **The merge tree depth is dynamic.** The coordinator's greedy pairing
   strategy in `OnProofResult` handles this naturally.

### 8.7 .bin file taxonomy for RISC-V

```
One .bin file per chip type (static, decided by the arithmetization team):

    cpu.bin       — CPU execution trace (temporal, ordered carry)
                    role:   query side of all chip lookups
                    carry:  register file, PC, memory root hash

    memory.bin    — memory read/write records (functional, unordered)
    alu.bin       — arithmetic and logic operations (functional)
    hash.bin      — Keccak / Poseidon2 invocations (functional)
    bitwise.bin   — bitwise operations (functional)
    …

Each shard produces: 1 CPU proof + 1 proof per active chip in that shard.
Empty chips (0 rows in a shard) produce no proof.
```

### 8.8 Summary: what changes, what does not

| Aspect | zkEVM | RISC-V zkVM | Design impact |
|--------|-------|-------------|---------------|
| Primary unit | Module (grouped by operation type) | Shard (grouped by cycle range) | Shard = 1 CPU segment + N chip segments |
| Trace structure | Per-module traces only | CPU temporal trace + per-chip traces derived from it | CPU trace is the clock that decides shard boundaries |
| Shard boundary | Static TracesLimits per module | Any chip hits its row threshold | Threshold per chip; detected during CPU trace scan |
| Segment ordering | All segments independent | CPU segments ordered (carry chain); chip segments unordered | Arithmetization is sequential per shard; proving is fully parallel |
| Preflight source | Any looked-up column | Chip table columns only (not CPU columns) | FSSchedule[0] scoped to chip/functional kinds |
| Segment count | Bounded by static TracesLimits | Dynamic, program-dependent | TotalSegmentCount at runtime; empty chips produce no task |
| Carry richness | LogUp accumulator | LogUp accumulator + register file + PC + memory root | Larger ReceivedValuesGlobal; same mechanism |
| .bin file count | ~1 (monolithic zkevm.bin) | N (one per chip type) | Already assumed in the new design |

The task-queue model, the two-channel arithmetization interface, the FS schedule
concept, the coordinator, and the fault-tolerance properties are all unchanged.
RISC-V refines **what arithmetization produces** (shard-structured instead of
module-structured) but not **how the prover consumes it**.

## 9. Remaining open questions

1. **`CommitLPPColumns` hash scheme**: must use the exact same Poseidon2 Merkle
   tree the GL circuit uses for its `lppMerkleRootPublicInput`. A native
   (non-circuit) version of this hash needs to be extracted from the `poseidon2`
   compiler package and made available in `distributed/preflight.go`.

2. **Consistency check**: the GL proof outputs `LppCommitment` as a public input.
   The pipeline should assert this matches the value computed in preflight as a
   sanity check. Currently the pseudocode notes this in `SegmentProof.LppCommitment`
   but does not enforce it.

3. **Coordinator persistence**: for crash recovery the coordinator's per-block
   state (received commitments, available proof paths) must be checkpointed to a
   durable store before results are acknowledged from the queue.

4. **Task queue backend**: the `TaskQueue` interface is abstract. The simplest
   implementation is a Redis stream or a persistent Go channel backed by a
   write-ahead log. Choice affects at-least-once vs. exactly-once semantics.
