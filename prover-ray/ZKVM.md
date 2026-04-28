# zkVM Comparison: RISC Zero, SP1, Linea Limitless, Nexus, ZKsync Airbender, and ZisK

How six ZK proving systems handle large computations: splitting execution into segments, proving each segment, and stitching the proofs back together.

**Evidence basis:** RISC Zero and SP1 claims are grounded in source code from their public repositories (file paths and code snippets verified on `main`). SP1 proof system details (§5.2.1–5.2.3, §11.1) verified against the `sp1` codebase on `main`: field is KoalaBear (not BabyBear), PCS is Jagged → Stacked Basefold, hash is Poseidon2, security target 100 bits. Nexus claims are grounded in the `nexus-zkvm` codebase (`main` branch) plus official blog posts. Linea claims include source-level references to the `linea-monorepo` codebase (see Sections 5.3, 5.4, and 6.3). **ZKsync Airbender** claims below are grounded in this repository (`zksync-airbender`): `docs/overview.md`, `docs/circuit_overview.md`, `docs/philosophy_and_logic.md`, `docs/delegation_circuits.md`, `circuit_defs/trace_and_split/src/lib.rs`, `prover/src/tracers/oracles/mod.rs`, `execution_utils/src/proofs.rs`, and representative verifier FRI code under `circuit_defs/risc_v_cycles/verifier/`. **Product-facing details** (six-stage proving pipeline, **FFLONK** on-chain wrapper, Blake2s/Blake3, GPU target) are taken from the official [ZKsync Airbender Overview](https://docs.zksync.io/zk-stack/components/zksync-airbender) (ZKsync Docs). **ZisK** (Polygon) claims are grounded in the `zisk` repository (`main` branch): `core/src/lib.rs`, `core/src/mem.rs`, `core/src/zisk_definitions.rs`, `core/src/zisk_inst.rs`, `pil/zisk.pil`, `sdk/src/prover/backend.rs`, `sdk/src/prover/mod.rs`, `ziskos/entrypoint/src/io.rs`, `ziskos/entrypoint/src/syscalls/`, `zisk-contracts/ZiskVerifier.sol`, `zisk-contracts/PlonkVerifier.sol`, `book/getting_started/precompiles.md`, `book/getting_started/distributed_execution.md`, and the workspace `Cargo.toml`.

---

## Table of Contents

1. [Background: How a zkVM Works](#1-background-how-a-zkvm-works)
2. [Key Terminology](#2-key-terminology)
3. [The Core Problem](#3-the-core-problem)
4. [Architecture at a Glance](#4-architecture-at-a-glance)
5. [Trace Structure Deep Dive](#5-trace-structure-deep-dive) ← **Diagrams showing concrete trace layouts** — **§5.2.1–5.2.3** SP1 proof system + Jagged PCS + SepticDigest — **§5.5–5.6** ZKsync Airbender — **§5.7** ZisK
6. [Segmentation Strategies](#6-segmentation-strategies) — **§6.5** ZKsync Airbender — **§6.6** ZisK
7. [Precompile / Accelerator Design](#7-precompile--accelerator-design) — **§7.5** ZKsync Airbender — **§7.6** ZisK
8. [Cross-Segment Consistency](#8-cross-segment-consistency)
9. [Proof Aggregation Pipelines](#9-proof-aggregation-pipelines) — **§9.5** ZKsync Airbender — **§9.6** ZisK
10. [Parallelism Summary](#10-parallelism-summary)
11. [Prover Orchestration & Segment Communication](#11-prover-orchestration--segment-communication) — **§11.1** SP1 — **§11.4** Linea Limitless — **§11.5** ZisK
12. [Trade-offs](#12-trade-offs)
13. [Information Gaps](#13-information-gaps)

---

## 1. Background: How a zkVM Works

This section explains the foundational concepts for readers unfamiliar with virtual machines or RISC-V.

### 1.1 What is a Virtual Machine?

A **virtual machine (VM)** is a software simulation of a computer. Instead of running on real hardware, programs run on a "fake" CPU implemented in software. The VM executes your program instruction by instruction, maintaining the illusion of a real computer.

**Why use a VM?** Portability (run the same program anywhere) and observability (record everything the program does).

### 1.2 The Core Components

A VM simulates three main components:

| Component | What it is | Analogy |
|-----------|------------|---------|
| **CPU** | The "brain" that fetches, decodes, and executes instructions one at a time | A worker following a recipe step by step |
| **Registers** | Small, fast storage slots inside the CPU (RISC-V has 32, named x0–x31) | Scratch paper on the worker's desk |
| **Memory** | Larger, slower storage where the program and data live | A filing cabinet the worker can access |

**How they interact:**
1. The CPU fetches the next instruction from memory (using the **program counter** to know where to look)
2. The CPU decodes the instruction to understand what to do
3. The CPU executes the instruction — this might involve reading/writing registers, reading/writing memory, or computing a result
4. The program counter advances to the next instruction (or jumps somewhere else)

### 1.3 What is RISC-V?

**RISC-V** is a specification for how a CPU should work — what instructions it understands, what registers it has, how memory is addressed. It's an open standard (no licensing fees), which is why zkVMs use it.

**Instructions** are the tiny operations a CPU can do:
- `ADD x3, x1, x2` — add registers x1 and x2, store result in x3
- `LOAD x5, [0x1000]` — read memory at address 0x1000 into register x5
- `STORE x5, [0x1000]` — write register x5 to memory at address 0x1000
- `BEQ x1, x2, label` — if x1 equals x2, jump to label
- `ECALL` — ask the environment to do something special (see below)

A compiled Rust or C program is just a long sequence of these simple instructions.

### 1.4 What is an ECALL (System Call)?

An **ecall** (environment call) is how a program asks for something it can't do itself. When the CPU hits an `ecall` instruction, it pauses normal execution and transfers control to a special handler.

In a zkVM context, ecalls are used to invoke **precompiles** — optimized circuits for expensive operations like hashing or elliptic curve math. Instead of proving thousands of basic instructions for a SHA-256 hash, the guest says "please hash this" via ecall, and the zkVM uses a specialized circuit that's much cheaper to prove.

### 1.5 Guest and Host

A zkVM has two sides:

| | Guest | Host |
|---|---|---|
| **What it is** | The program being proved | The machine running the prover |
| **What it sees** | Only what the host explicitly provides | Everything (files, network, etc.) |
| **What it can do** | Compute; request precompiles via ecall | Execute the guest; provide inputs; generate proof |

**Analogy:** The guest is a person in a sealed room doing math on a whiteboard. The host is outside, passing inputs under the door and reading the final answer. The proof says "the person did the math correctly" without revealing the intermediate steps.

### 1.6 From Execution to Proof

Here's how a zkVM turns program execution into a proof:

```
┌─────────────────────────────────────────────────────────────────┐
│                         YOUR PROGRAM                            │
│                    (compiled to RISC-V)                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          EXECUTOR                               │
│  Runs the program instruction by instruction.                   │
│  Records every step into an EXECUTION TRACE (a big table).      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      EXECUTION TRACE                            │
│  ┌───────┬──────┬───────┬─────┬─────┬─────┬─────────┐          │
│  │ Cycle │  PC  │ Instr │ x1  │ x2  │ x3  │ Memory  │          │
│  ├───────┼──────┼───────┼─────┼─────┼─────┼─────────┤          │
│  │   0   │0x100 │ LOAD  │  ?  │  0  │  0  │ read    │          │
│  │   1   │0x104 │ LOAD  │ 42  │  ?  │  0  │ read    │          │
│  │   2   │0x108 │ ADD   │ 42  │ 10  │  ?  │  —      │          │
│  │   3   │0x10C │ STORE │ 42  │ 10  │ 52  │ write   │          │
│  │  ...  │ ...  │  ...  │ ... │ ... │ ... │  ...    │          │
│  └───────┴──────┴───────┴─────┴─────┴─────┴─────────┘          │
│  Each row = one clock cycle. Millions of rows for real programs.│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          PROVER                                 │
│  Checks that the trace satisfies CONSTRAINTS:                   │
│    • "If instruction is ADD, then x3 = x1 + x2"                │
│    • "PC increments by 4 unless it's a jump"                   │
│    • "Memory reads return the last written value"              │
│  Produces a cryptographic PROOF.                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                          VERIFIER                               │
│  Checks the proof in milliseconds.                              │
│  Never sees the full trace — only the proof.                   │
│  Outputs: ACCEPT or REJECT.                                     │
└─────────────────────────────────────────────────────────────────┘
```

### 1.7 The Memory Checking Problem

Here's a tricky issue: the CPU can read and write memory in any order. At cycle 100, you write 42 to address 0x500. At cycle 5000, you read from 0x500 — you should get 42. But the trace is organized by time, not by address. How do you prove memory is consistent?

**Three approaches:**

| Approach | How it works | Used by |
|----------|--------------|---------|
| **Merkle trees** | Keep a hash tree of all memory; every access updates/checks the root | RISC Zero |
| **Permutation argument** | Sort memory accesses by address; prove the sorted list is a permutation of the chronological list; check consistency in sorted order | SP1, Nexus |
| **Timestamp-based** | Each memory cell tracks its last access time; check that accesses happen in order | Nexus |
| **Shuffle RAM + timestamps + chunk linkage** | Registers and RAM share a unified shuffle-style memory argument with per-access timestamps; execution is split into chunks with **lazy init** / **teardown** rows so the first access in a chunk matches the last prior access’s value and timestamp | ZKsync Airbender (`docs/circuit_overview.md`; `trace_and_split` / `chunk_lazy_init_and_teardown` in `prover/src/tracers/oracles/mod.rs`; `FinalRegisterValue` / `split_timestamp` in `execution_utils/src/proofs.rs`) |
| **Dedicated memory state machines + bus** | Memory accesses are delegated from the main state machine to specialized sub-state-machines (aligned, unaligned, input) via a memory proxy; proved with bus-style interactions between main SM and memory SMs; chunks linked via segment-boundary memory state | ZisK (`core/src/lib.rs`: Memory Proxy → Memory Aligned / Unaligned / Input; `state-machines/mem/`, `pil/zisk.pil`) |

### 1.8 Why Segmentation?

A real program might run for billions of cycles. The trace would have billions of rows — too big to prove at once (exhausts memory and time).

**Solution:** Split the trace into **segments** (e.g., 1 million rows each), prove each segment separately, then aggregate the proofs.

**The challenge:** Some relationships span segments. For temporal segmentation (slicing by time), the main cross-segment relationship is **memory** — a write in segment 1 might be read in segment 5. Different zkVMs handle this differently (see Section 7).

**Key insight:** Temporal segmentation keeps most lookups local. If you dispatch a precompile at cycle 1000, the precompile executes at cycle 1000 — both sides are in the same segment. Only memory (where writes and reads can be far apart in time) spans segments.

---

## 2. Key Terminology

| Term | Meaning |
|---|---|
| **Execution trace** | A table of all intermediate values produced while running a program. The prover must show this table satisfies certain constraints. |
| **AIR** | Algebraic Intermediate Representation — the constraint system that defines what a valid trace looks like. |
| **Segment / Shard / Chunk** | A slice of the full trace that is proved independently. RISC Zero calls them "segments," SP1 calls them "shards," Linea calls them "segments," ZisK calls them "chunks." |
| **po2** | "Power of 2" — the log2 of a trace's row count. A trace with 2^20 rows has po2 = 20. Traces are always padded to a power-of-two length. |
| **Chip (SP1) / Module (Linea)** | A self-contained sub-circuit responsible for one job (e.g., CPU logic, Keccak hashing, memory). Each has its own AIR. |
| **Lookup argument (LogUp)** | A cryptographic mechanism to prove that values in one table appear in another (e.g., "this byte is in the range 0–255"). Uses log-derivative sums: if the total sum across all tables is zero, the multisets match. |
| **LogUp-GKR** | A more efficient variant of LogUp that uses the GKR sumcheck protocol. Used by SP1 V6. |
| **Cross-table lookup** | A lookup where the "sender" row is in one chip/module and the "receiver" row is in a different chip/module. |
| **Cross-segment lookup** | A lookup where the two sides land in different segments. This is the hard problem — each segment is proved independently, so matching them requires extra machinery. |
| **Fiat-Shamir transcript** | Converts an interactive proof into a non-interactive one by deriving "random" challenges from a hash of the proof data so far. Each proof (or segment proof) can have its own independent transcript. |
| **Multiset hash (EC)** | An algebraic accumulator on an elliptic curve: hash each interaction onto a curve point, sum them. If the total is the point at infinity, all sends were matched by receives. Needs no shared randomness between proofs. |
| **PCS** | Polynomial Commitment Scheme — the mechanism for committing to polynomials and opening them at chosen points. Different PCS choices (FRI, KZG, etc.) give different trade-offs in proof size, prover time, and trust assumptions. |
| **STARK / FRI** | A proof system based on polynomial commitments via the Fast Reed-Solomon IOP. Produces large proofs but needs no trusted setup. |
| **Circle STARK** | A STARK variant (used by Stwo) that operates over a circle group in a Mersenne prime field (M31), enabling fast native 32-bit arithmetic. |
| **Groth16** | A SNARK that produces tiny (~260 byte) proofs but requires a trusted setup. Typically used as the final "wrapper" proof for on-chain verification. |
| **Recursion** | Verifying a proof inside another proof. Used to aggregate many segment proofs into one. |
| **Conflation (Linea)** | The process of batching many EVM blocks into a single proving job. The conflated trace across all modules is what gets segmented by Limitless. |
| **Horizontal segmentation (Limitless)** | Partitioning the proof along **different modules** (different `ModuleName` values: HUB-A, HUB-B, KECCAK, ARITH-OPS, …). Each module is its own circuit family; cross-module consistency uses lookups and the GL→LPP pipeline. |
| **Vertical segmentation (Limitless)** | Partitioning **one module’s trace** into multiple **row bands** (multiple segment proofs for the same module). `SegmentModuleIndex` identifies which band; `IsFirst` / `IsLast` mark ends of the vertical chain. Needed when a module’s column height exceeds what a single proof can cover. |
| **Chunk (Airbender)** | A **temporal** slice of execution with trace length a power of two; executable cycles per chunk are **`trace_size - 1`** (`circuit_defs/trace_and_split/src/lib.rs`). Docs cite **~2²²** cycles per batch (`docs/philosophy_and_logic.md`). |
| **Delegation (Airbender)** | A **separate circuit** (precompile) triggered via **CSR `0x7c0`**, with witness `DelegationWitness` and proofs collected per `delegation_type` in `ProgramProof` (`docs/delegation_circuits.md`, `execution_utils/src/proofs.rs`). |
| **Chunk (ZisK)** | A temporal slice of execution with **2¹⁸ = 262,144 steps** (`CHUNK_SIZE_BITS = 18` in `core/src/zisk_definitions.rs`). Execution is re-run per chunk in parallel during witness generation. Maximum total steps: **2³⁶ − 1** (`DEFAULT_MAX_STEPS`). |
| **PIL / airgroup (ZisK)** | The constraint system is defined in **PIL** (Polynomial Identity Language), compiled into a single **`airgroup Zisk`** that includes the main state machine plus all secondary SMs (ROM, memory, binary, arith, keccakf, sha256f, poseidon2, blake2, arith_eq, big_int, DMA) (`pil/zisk.pil`). |
| **Secondary state machine (ZisK)** | A specialized sub-circuit that handles a class of operations delegated from the main state machine. Analogous to SP1's "chip" or Airbender's "delegation circuit." Examples: Binary Basic, Binary Extension, Arith, Memory Aligned, Keccak-f, SHA-256-f, Poseidon2, BLAKE2 (`core/src/lib.rs`). |
| **Vadcop (ZisK)** | The intermediate proof format produced by the Proofman proving stack. Vadcop Final proofs can be **compressed** and optionally wrapped into a **Plonk** or **FFLONK** SNARK for on-chain verification (`sdk/src/prover/mod.rs`). |

---

## 3. The Core Problem

A single monolithic execution trace is too large to prove in one shot — it exhausts memory, time, or both. The standard solution is:

1. **Segment** the execution into smaller pieces
2. **Prove** each segment independently (ideally in parallel)
3. **Aggregate** the segment proofs into a single final proof

The systems differ in *how* they segment, *what relationships span segments*, and *how they handle those cross-segment relationships*.

---

## 4. Architecture at a Glance

| | RISC Zero | SP1 (V6) | Linea Limitless | Nexus (main) | ZKsync Airbender | ZisK (Polygon) |
|---|---|---|---|---|---|---|
| **VM type** | RISC-V (rv32im) | RISC-V (rv32im) | EVM | RISC-V (RV32I + optional M) | RISC-V (rv32im), machine mode, no traps | RISC-V (rv64ima) → transpiled to custom Zisk ISA (64-bit operands) |
| **Proof system** | STARK (univariate FRI + DEEP-ALI) over BabyBear | STARK (multilinear Basefold + sumcheck) over KoalaBear | Wizard-IOP → Plonk + Groth16 | Stwo (circle STARK) | STARK (FRI) over Mersenne31 → FFLONK on-chain | PIL-defined AIRs + Proofman (Vadcop) over Goldilocks → Plonk/FFLONK on-chain |
| **Segmentation** | Temporal (time slices) | Temporal (~2M cycle chunks) | Structural (by module + row bands) | None (single proof) | Temporal (~2²² cycle chunks) | Temporal (~2¹⁸ step chunks) |
| **Precompile design** | All in one circuit (muxed rows) | Separate chip per precompile | Separate module per precompile | Custom opcodes + extension components | Separate delegation circuit per precompile type | Separate state machine per precompile, dispatched via syscalls |
| **Cross-segment mechanism** | Merkle root chaining | EC multiset hash (memory only) | Shared randomness + partial log-derivative sums | N/A (all lookups intra-proof) | Shuffle RAM with lazy init/teardown between segments | Dedicated memory SMs + bus interactions; cross-chunk linkage to be investigated |

---

## 5. Trace Structure Deep Dive

This section shows what the execution trace actually looks like in each system — the concrete table layouts that get segmented and proved.

### 5.1 RISC Zero: Single Unified Table

RISC Zero uses a **single trace table** where every row has columns for all possible operations. Selector bits determine which constraints are active for each row.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    RISC ZERO: SINGLE UNIFIED TRACE TABLE                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Every row has columns for ALL block types.                                      │
│ Only the active block type's columns contain meaningful data.                   │
│ Selector bits tell the constraint system which constraints to check.            │
└─────────────────────────────────────────────────────────────────────────────────┘

┌───────┬───────────────────┬──────────────┬──────────────────┬──────────────────┐
│ Cycle │  Selector Bits    │  Block Type  │  CPU Columns     │ Precompile Cols  │
│       │ [two-hot encoding]│              │  (when active)   │ (when active)    │
├───────┼───────────────────┼──────────────┼──────────────────┼──────────────────┤
│   0   │ [1,0,0,0,...]     │ InstReg      │ PC, x1, x2, x3   │ [zeros]          │
│   1   │ [1,0,0,0,...]     │ InstReg      │ PC, x1, x2, x3   │ [zeros]          │
│   2   │ [0,1,0,0,...]     │ InstEcall    │ ecall dispatch   │ [zeros]          │
│   3   │ [0,0,1,0,...]     │ EcallBigInt  │ args from regs   │ [zeros]          │
│   4   │ [0,0,0,1,...]     │ BigInt       │ [unused]         │ data[0:15]       │
│   5   │ [0,0,0,1,...]     │ BigInt       │ [unused]         │ data[16:31]      │
│   6   │ [0,0,0,1,...]     │ BigInt       │ [unused]         │ polynomial check │
│   7   │ [1,0,0,0,...]     │ InstResume   │ return from ecall│ [zeros]          │
│  ...  │       ...         │     ...      │       ...        │       ...        │
└───────┴───────────────────┴──────────────┴──────────────────┴──────────────────┘

SEGMENT STRUCTURE:
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Segment 0                    │ Segment 1                    │ Final segment    │
│ rows 0 to ~1M                │ rows ~1M to ~2M              │ (smaller)        │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Each segment contains:                                                          │
│   • REPLICATED lookup tables (U8: 0-255, U16: 0-65535) = 4,112+ fixed rows     │
│   • Merkle root of memory state at boundary                                     │
│   • Self-contained LogUp grand product (starts at 1, ends at 1)                │
│   • Independent Fiat-Shamir transcript                                          │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**Key properties:**
- All operations share the same column layout
- Segment shape depends only on row count (po2), not which operations ran
- Only 13 possible shapes → 13 pre-compiled normalizer programs

#### 5.1.1 RISC Zero Proof System: Univariate STARK with DEEP-ALI

RISC Zero uses a **univariate polynomial** STARK, not multilinear + sumcheck. The key components:

| Component | RISC Zero's Choice | Source |
|-----------|-------------------|--------|
| **Field** | BabyBear (p = 15 × 2²⁷ + 1 = 2,013,265,921) | `risc0/zkp/src/core/field/baby_bear.rs` |
| **Extension field** | Degree-4 extension of BabyBear | For 128-bit security target |
| **Polynomial type** | Univariate | Trace columns encoded as univariate polynomials over roots of unity |
| **PCS** | FRI (Fast Reed-Solomon IOP) | `risc0/zkp/src/prove/fri.rs`, `risc0/zkp/src/verify/fri.rs` |
| **Constraint system** | AIR (Algebraic Intermediate Representation) | Randomized AIR with preprocessing (RAP) |
| **Protocol** | DEEP-ALI + batched FRI | `risc0/zkp/src/prove/soundness.rs` |
| **Hash function** | Poseidon2 | Used for Merkle trees and Fiat-Shamir |
| **Memory/permutation** | PLONK-style grand product | For memory consistency |
| **Range checks** | PLOOKUP-style lookup | For byte/u16 range checks |

**DEEP-ALI explained:** Instead of checking constraints at many points inside the trace domain, DEEP-ALI checks at a **single random point z outside the domain**. This works because:

1. Prover commits to trace polynomials (Merkle tree of evaluations)
2. Verifier picks random z outside the evaluation domain (via Fiat-Shamir)
3. Prover reveals polynomial values at z
4. Prover constructs the **DEEP quotient**: `Q(x) = [F(x) - F(z)] / (x - z)`
5. FRI proves Q(x) is low-degree — this single FRI proof covers all trace polynomials

**Batching multiple polynomials:** All trace polynomials are combined via random linear combination into one batched polynomial before the DEEP quotient. Only one FRI proof is needed regardless of how many trace columns exist.

**FRI parameters (from `risc0/zkp/src/lib.rs`):**

| Parameter | Value | Meaning |
|-----------|-------|---------|
| `INV_RATE` | 4 | Reed-Solomon expansion rate = 1/4 (ρ = 0.25) |
| `QUERIES` | 50 | Number of FRI query rounds |
| `FRI_FOLD` | 16 | Folding factor per round (2⁴) |
| `FRI_MIN_DEGREE` | 256 | Stop folding when polynomial degree reaches this |

**Security level (from `risc0/zkp/src/prove/soundness.rs`):**
- Toy model conjecture (ethSTARK): ~97 bits
- Conjectured (DEEP-FRI Conjecture 2.3): ~75 bits  
- Proven (list-decoding regime): ~42 bits

**Contrast with SP1 V6:** SP1 switched to **multilinear polynomials + sumcheck** over KoalaBear (via the `slop` stack, successor to Plonky3). The key differences:

| Aspect | RISC Zero (Univariate + FRI) | SP1 V6 (Multilinear + Basefold) |
|--------|------------------------------|--------------------------------|
| Polynomial representation | Single variable, degree = trace length | Multiple variables, degree ≤ 1 per variable |
| Trace indexing | Rows indexed by roots of unity | Rows indexed by binary hypercube |
| Main protocol | DEEP-ALI | Zerocheck sumcheck + Basefold |
| Field | BabyBear (p = 15 × 2²⁷ + 1) | KoalaBear (p3 KoalaBear, degree-4 extension) |
| PCS | FRI | Jagged → Stacked Basefold (FRI-config) |
| Hash | Poseidon2 over BabyBear | Poseidon2 over KoalaBear (16-wide state) |

#### 5.1.2 Trace Padding and Prover Cost

**Why padding exists.** RISC Zero's STARK requires NTT-based polynomial operations (interpolation, evaluation, FRI folding), which only work on domains whose size is a power of 2. Real execution traces almost never land on an exact power of 2, so the trace is padded up to the next one. If a segment uses 3,000 rows of real computation, the trace becomes 2¹² = 4,096 rows — the remaining ~1,096 rows are filled with dummy data (zeros or copies that trivially satisfy constraints). The range of allowed trace sizes is `MIN_PO2 = 11` (2,048 rows) to `MAX_CYCLES_PO2 = 24` (16M rows).

**Column width is fixed at compile time; row count is decided at runtime.** The circuit layout — how many columns exist, which block types they support, the constraint definitions — is determined entirely by the zirgen compiler and is the same for every segment. The trace height (`2^po2`) is chosen at runtime based on how many execution cycles the segment actually consumed. This means the system supports **dynamic data volume**: the prover adapts to the workload without recompiling the circuit. Source: `set_po2()` in `risc0/zkp/src/prove/prover.rs`, `segment_limit_po2` in `risc0/zkvm/src/host/server/exec/executor.rs`.

**The prover pays for the full padded trace, not just the real data.** Every prover-side operation runs over the entire `2^po2` domain, including padding rows:

| Operation | Domain size | Notes |
|-----------|-------------|-------|
| NTT / interpolation | `2^po2 × num_columns` | `batch_interpolate_ntt` in `make_coeffs()` operates on the full witness buffer |
| Low-degree extension (LDE) | `4 × 2^po2 × num_columns` | `batch_expand_into_evaluate_ntt` in `PolyGroup::new()` expands every column to `INV_RATE × 2^po2` evaluations |
| Merkle tree | `4 × 2^po2` leaves | Every padded row contributes a leaf; all leaves are hashed |
| Constraint evaluation | `4 × 2^po2` rows | `eval_check` evaluates the constraint polynomial over the full LDE domain |
| FRI | Full LDE domain | Folding and queries span all `4 × 2^po2` points |

The padding rows are "free" only in the sense that they trivially satisfy constraints (no risk of a soundness issue). They are **not free** in prover time, memory, or proof size — the NTT, Merkle hashing, and FRI query costs are proportional to `2^po2`, not to the number of real rows.

**Worst-case overhead.** Since the trace rounds up to a power of 2, the worst case is just over half padding: a segment needing 2^k + 1 rows pays for 2^(k+1), nearly doubling the work. The `padding_row_count` field in `SegmentInfo` (`risc0/zkvm/src/host/server/session.rs`) records this gap explicitly.

**Mitigation via segmentation.** The primary lever is choosing the right segment size. The executor splits execution into segments whose `po2` is the smallest power of 2 that fits (configurable via `segment_limit_po2`). This keeps the padding overhead bounded but does not eliminate it.

**Contrast with multi-table systems.** In systems like SP1 (per-chip tables) or ZisK (per-SM tables), each table is independently sized to its actual usage — and SP1 goes further than simple per-chip power-of-2 rounding (see §5.2.2a for details). SP1 pads each chip's trace to the next multiple of 32 (not the next power of 2), and the Jagged PCS commits only to the dense stored rows across all chips, so a chip with 100 rows pays for ~128 rows of prover work, not for the shard's maximum table height. RISC Zero's single-table design forces all block types into one domain size, so a segment with 1M ALU rows and 100 SHA rows commits the SHA columns at 1M rows too. The mux/selector design (§5.1) mitigates this by overlaying block types onto shared columns — there are no dedicated "SHA columns" sitting empty — but the fundamental cost is still `2^po2 × total_column_width` for the entire table.

### 5.2 SP1: Multiple Chips (Separate Tables)

SP1 uses **separate STARK tables (chips)** for different operations. Each chip has its own AIR and columns. Chips communicate via cross-table lookups.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                      SP1: MULTI-CHIP ARCHITECTURE                               │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Each chip is a separate STARK table with its own AIR.                           │
│ Chips communicate via cross-table lookups (LogUp-GKR).                          │
│ Sharding is temporal: CPU→precompile lookups always land in the same shard.     │
└─────────────────────────────────────────────────────────────────────────────────┘

SHARD N (time slice ~2M cycles):
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  ┌───────────────────────────┐      ┌───────────────────────────┐              │
│  │       CPU CHIP            │      │     MEMORY CHIP           │              │
│  ├───────────────────────────┤      ├───────────────────────────┤              │
│  │ Cycle │ PC   │ Op  │ regs │      │ Addr │ Value │ Time │ Op  │              │
│  ├───────┼──────┼─────┼──────┤      ├──────┼───────┼──────┼─────┤              │
│  │   0   │0x100 │LOAD │ ...  │◄────►│0x200 │  42   │  0   │READ │              │
│  │   1   │0x104 │ADD  │ ...  │      │0x204 │  10   │  1   │READ │              │
│  │   2   │0x108 │ECALL│ ...  │──┐   │0x200 │  52   │  3   │WRITE│              │
│  │  ...  │ ...  │ ... │ ...  │  │   │ ...  │ ...   │ ...  │ ... │              │
│  └───────────────────────────┘  │   └───────────────────────────┘              │
│                                 │                                               │
│  ┌───────────────────────────┐  │   ┌───────────────────────────┐              │
│  │     KECCAK CHIP           │  │   │     SHA256 CHIP           │              │
│  ├───────────────────────────┤  │   ├───────────────────────────┤              │
│  │ Row │ State[25] │ Round   │◄─┘   │ Row │ W[16] │ H[8] │Round │              │
│  ├─────┼───────────┼─────────┤      ├─────┼───────┼──────┼──────┤              │
│  │  0  │ [input]   │    0    │      │  0  │ [msg] │[init]│  0   │              │
│  │  1  │ [after θ] │    1    │      │  1  │[expand│[rnd] │  1   │              │
│  │ ... │    ...    │   ...   │      │ ... │  ...  │ ...  │ ...  │              │
│  └───────────────────────────┘      └───────────────────────────┘              │
│                                                                                 │
│  ┌───────────────────────────┐      ┌───────────────────────────┐              │
│  │   BYTE TABLE (fixed)      │      │     GLOBAL CHIP           │              │
│  ├───────────────────────────┤      ├───────────────────────────┤              │
│  │ Value │ (range 0-255)     │      │ Cross-shard memory        │              │
│  │   0   │                   │      │ accumulated onto EC point │              │
│  │   1   │                   │      │ (SepticDigest)            │              │
│  │  ...  │                   │      └───────────────────────────┘              │
│  │  255  │                   │                                                  │
│  └───────────────────────────┘      Cross-table lookups (LogUp-GKR)            │
│                                     resolved within this shard                  │
└─────────────────────────────────────────────────────────────────────────────────┘

CROSS-SHARD MEMORY CONSISTENCY:
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Shard 0          │ Shard 1          │ ... │ Machine Verifier                    │
│ SepticDigest_0   │ SepticDigest_1   │     │ Σ(all digests) = point at infinity  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

**Key properties:**
- Each chip has specialized columns for its operation
- Shard shape varies based on which chips are active
- Dynamic compilation with caching for shape normalization
- EC multiset hash for cross-shard memory (no shared randomness)

#### 5.2.1 SP1 Proof System: Multilinear Basefold + Sumcheck over KoalaBear

SP1 V6 ("Hypercube") uses a **multilinear polynomial** STARK, not univariate + FRI like RISC Zero. The key components:

| Component | SP1's Choice | Source |
|-----------|-------------|--------|
| **Field** | KoalaBear (p3 KoalaBear prime) | `crates/primitives/src/lib.rs` — `pub type SP1Field = KoalaBear` |
| **Extension field** | Degree-4 binomial extension (x⁴ − 3) | `BinomialExtensionField<SP1Field, 4>` |
| **Polynomial type** | Multilinear (MLE over boolean hypercube) | Trace columns as `PaddedMle` / `Mle` from `slop_multilinear` |
| **PCS** | Jagged PCS → Stacked Basefold (FRI-config) | `crates/hypercube/src/verifier/shard.rs` |
| **Constraint system** | AIR per chip (zerocheck sumcheck) | `crates/hypercube/src/prover/zerocheck/` |
| **Protocol** | Zerocheck via sumcheck + Basefold + LogUp-GKR | `crates/hypercube/src/logup_gkr/` |
| **Hash function** | Poseidon2 over KoalaBear (16-wide state) | `slop/crates/koala-bear/src/koala_bear_poseidon2.rs` |
| **Memory/permutation** | SepticDigest EC multiset hash | `crates/hypercube/src/septic_digest.rs` |
| **Lookup argument** | LogUp-GKR (sumcheck per GKR round) | `crates/hypercube/src/logup_gkr/prover.rs` |

**IOP context (`SP1GlobalContext`):** The `KoalaBearDegree4Duplex` struct ties everything together — Poseidon2 duplex challenger (16 elements absorbed per round, 8 squeezed), sponge hasher, truncated permutation compressor, and Merkle tree configuration (`slop/crates/koala-bear/src/koala_bear_poseidon2.rs`).

**Basefold configuration:** SP1's shard PCS is `StackedPcsVerifier` (stacked Basefold), wrapped by `JaggedPcsVerifier` (see §5.2.2 for Jagged PCS). The stack combines multiple multilinear polynomials into a single batched Basefold commitment. Basefold is parameterized by `FriConfig` (blowup, query count, proof-of-work), making it FRI-like but operating on multilinear polynomials.

**FRI / Basefold parameters (from `crates/primitives/src/fri_params.rs`):**

| Parameter | Core | Shrink/Wrap | Meaning |
|-----------|------|-------------|---------|
| `LOG_BLOWUP` | 2 | 3 | Reed-Solomon expansion rate = 1/2^blowup |
| `PROOF_OF_WORK_BITS` | 16 | 22 | Grinding: required zero bits in PoW |
| `num_queries` | derived | derived | Computed from security target |
| `GKR_GRINDING_BITS` | 12 | — | Grinding for LogUp-GKR challenges |

**Security level (from `crates/primitives/src/fri_params.rs`):**
- Target: **100 bits** (`SP1_TARGET_BITS_OF_SECURITY = 100`)
- Query count formula accounts for recent Gruen–Diamond analysis; increased from 84 to 94 queries for `log_blowup=1` to "be on the safe side" (`slop/crates/primitives/src/lib.rs`)

**LogUp-GKR explained:** SP1's lookup argument uses the GKR protocol (Grand Product / Sum over a layered circuit), with a **sumcheck** at each GKR round:

1. `LogupGkrCpuTraceGenerator::generate_gkr_circuit` builds layers of padded MLEs over row × interaction variables, plus an interaction layer.
2. `GkrProverImpl::prove_gkr_circuit` walks layers GKR-style; each layer's `prove_gkr_round` produces a `LogupGkrRoundProof` containing numerator/denominator evaluations and a `PartialSumcheckProof`.
3. **Grinding barrier:** Before challenges are drawn, a proof-of-work requiring `GKR_GRINDING_BITS` (12) zero bits is verified — this hardens the Fiat-Shamir transcript against adversarial challenge selection.
4. After all rounds, `logup_evaluations` (per-chip polynomial openings) are verified against the accumulated sumcheck claims.

**Zerocheck:** Constraint checking uses a **zerocheck** protocol — equivalent to proving that a multivariate polynomial vanishes on the boolean hypercube. SP1's `ZeroCheckPoly` sums all variables except the last over the hypercube, leaving a univariate polynomial that the verifier checks via sumcheck (`crates/hypercube/src/prover/zerocheck/sum_as_poly.rs`).

**Outer recursion:** For the final wrap/SNARK stage, SP1 switches to `SP1OuterGlobalContext = BNGC<...>` — a BN254-side algebraic context, analogous to RISC Zero's `identity_p254` step that bridges BabyBear to BN254 for Groth16.

#### 5.2.2 Jagged PCS: Variable-Sized Tables Without Waste

**The problem:** In a multi-chip architecture, each chip's trace table has a different number of rows. A shard with 100,000 CPU rows, 50 Keccak rows, and 0 SHA-256 rows would waste enormous space if every table were padded to the same power-of-two height. The PCS needs to commit to all tables, but a naive approach "pays" for the largest table's size on every table.

**What Jagged PCS does:** It allows the prover to commit to a collection of **columns with different heights** (a "jagged" layout), while still using a single dense multilinear PCS (Stacked Basefold) under the hood. The verifier is convinced that each column's committed values are correct — including that the "padding" beyond a column's real height is treated as zero — without the prover paying for full-size tables everywhere.

**How it works (three-phase protocol):**

1. **Commit:** The prover passes batches of `PaddedMle` (padded multilinear extensions). Real data goes through the stacked Basefold; the **outer digest** binds the row and column counts (including padding/dummy tables) via Poseidon2 hash + compress with the inner Basefold commitment. Columns with zero rows are filtered out and not committed at all.

2. **Main sumcheck (Hadamard product):** To prove that evaluations at a random point correspond to the jagged layout, the prover constructs a **Hadamard product** of two multilinears:
   - A **"long" interleaved MLE** — the dense committed data, stacked across all columns
   - A **jagged indicator MLE** — a multilinear extension of a function that maps (row, column) indices in the 2D jagged array to indices in the flattened 1D vector. This indicator is built from **prefix sums of column areas** and evaluated via a **branching program** (following [HR18](https://eccc.weizmann.ac.il/report/2018/161/)).

   A sumcheck reduces the N-variate evaluation claim to a single point for the dense PCS.

3. **Jagged evaluation sub-proof:** A second sumcheck (`JaggedSumcheckEvalProof`) certifies the jagged indicator polynomial's evaluation at the challenge point, using the branching-program structure for efficiency.

4. **Dense PCS:** Finally, the stacked Basefold proves the opening at the reduced point.

**Key types (from `slop/crates/jagged/`):**

| Type | Role |
|------|------|
| `JaggedPcsVerifier<GC, C>` | Wraps a stacked PCS verifier; adds jagged protocol |
| `JaggedProver<GC, Proof, C>` | Wraps a stacked PCS prover; handles commit/prove |
| `JaggedPcsProof` | Proof object: dense PCS proof + sumcheck proof + jagged eval proof + row/column counts |
| `JaggedLittlePolynomialProverParams` | Parameters for the jagged indicator polynomial (prefix sums, branching program) |

**Integration in SP1:** `ShardVerifier` owns a `JaggedPcsVerifier`, and every `ShardProof` carries a `JaggedPcsProof` as its evaluation proof. The recursion circuit reimplements verification via `RecursiveJaggedPcsVerifier`.

**What it brings to the table:**
- **Pay-per-use:** A chip with 50 rows in a shard that has 100,000 CPU rows only "pays" for 50 rows worth of prover work for that chip, not 100,000.
- **Sparse chips are essentially free:** A chip with zero rows in a shard is filtered out entirely — no commitment, no proof work.
- **Unified batching:** Despite the variable sizes, all chips in a shard share a single Basefold commitment and a single opening proof, keeping the proof compact.
- **No table replication:** Unlike RISC Zero (which pays 4,112+ rows per segment for lookup tables), SP1's Jagged PCS means unused chips contribute no overhead.

**Contrast with other systems:**

| System | How variable-sized tables are handled |
|--------|--------------------------------------|
| **RISC Zero** | N/A — single unified table, all rows have same column layout regardless of operation |
| **SP1** | Jagged PCS — each chip has own height, padded columns treated as zero, single batched commitment |
| **Linea** | Each module is a separate Plonk circuit with its own domain size; different modules can have different sizes |
| **ZKsync Airbender** | Main table is fixed size; delegation circuits have own sizes; handling to be investigated |
| **ZisK** | Each secondary SM has its own trace size; Proofman dynamically plans instance counts and sizes per execution; virtual tables with configurable max rows (`set_max_num_rows_virtual(1 << 21)` in PIL) |

#### 5.2.2a SP1 Trace Padding and Prover Cost (Parallel to §5.1.2)

**Why padding exists.** Like RISC Zero, SP1's polynomial commitment scheme requires trace tables to be aligned to certain boundaries. However, the alignment granularity and the way the prover "pays" for padding are fundamentally different.

**Chip-level padding: multiple of 32, not power of 2.** Each chip pads its trace to the **next multiple of 32** (minimum 16 rows), not the next power of 2. The function `next_multiple_of_32` in `crates/hypercube/src/util.rs` controls this:

```rust
pub fn next_multiple_of_32(n: usize, fixed_height: Option<usize>) -> usize {
    if let Some(height) = fixed_height {
        if n > height {
            panic!("fixed height is too small: ...");
        }
        height
    } else {
        n.next_multiple_of(32).max(16)
    }
}
```

A chip with 100 real rows gets 128 rows; one with 3,000 rows gets 3,008. The worst-case overhead is 31 rows, compared to RISC Zero's nearly-2× from power-of-2 rounding. When a **proof shape** is active (for fixed-shape recursion circuits), the height can be pinned to a specific value via `Program::fixed_log2_rows` (`crates/core/executor/src/program.rs`), which looks up a per-chip log₂ height from the `preprocessed_shape`.

**The prover pays for real data, not the padded domain.** This is the critical architectural difference from RISC Zero. SP1 uses a Jagged PCS (§5.2.2) backed by Stacked Basefold, which together ensure prover work is proportional to the **sum of real chip areas**, not to the shard-wide maximum:

| Layer | What it pays for | Source |
|-------|------------------|--------|
| **Trace generation** | `num_rows` per chip (multiple-of-32 padded) — stored as a dense `RowMajorMatrix` | `MachineAir::generate_trace` in `crates/hypercube/src/air/machine.rs` |
| **`PaddedMle` wrapping** | Virtual domain is `2^max_log_row_count` (shard-wide), but extra rows are **implicit zeros** — never materialized | `PaddedMle::padded_with_zeros` in `slop/crates/multilinear/src/padded.rs` |
| **Jagged PCS commit** | Filters out zero-row chips entirely; commits only dense `Mle` inner data; padding values **ignored and treated as zero** | `JaggedProver::commit_multilinears` in `slop/crates/jagged/src/prover.rs` |
| **Stacked Basefold commit** | Total area = `Σ(stored_rows × width)` across all chips, then pads only to the next multiple of `2^log_stacking_height` | `StackedPcsProver::commit_multilinears` in `slop/crates/stacked/src/prover.rs` |
| **RS encoding (DFT)** | Runs on the dense stacked blob, not per-chip `2^L` arrays | `BasefoldEncoder::encode` in `slop/crates/basefold-prover/src/encoder.rs` |
| **Zerocheck / sumcheck** | Uses `num_real_entries` to set a **threshold point** per chip — the sumcheck only needs to consider real rows | `ShardProver` in `crates/hypercube/src/prover/shard.rs` |

**Concrete comparison (same workload as §5.1.2):**

| Scenario | RISC Zero prover cost | SP1 prover cost |
|----------|----------------------|-----------------|
| Chip/segment with 3,000 real rows | 4,096 rows (2¹²) — 27% wasted | 3,008 rows (next multiple of 32) — 0.3% wasted |
| Chip with 2,049 real rows | 4,096 rows (2¹²) — 50% wasted | 2,080 rows — 1.5% wasted |
| Chip with 100 real rows in a shard where max chip has 1M rows | 1M rows (shares domain) | 128 rows (independent sizing via Jagged PCS) |
| Chip with 0 rows | Still allocated at `2^po2` domain | Filtered out entirely — zero cost |

**The `max_log_row_count` is virtual, not materialized.** A common source of confusion: every `PaddedMle` in a shard is tagged with the same `max_log_row_count` (the shard's maximum multilinear variable count), and zerocheck assertions verify `num_variables == max_log_row_count`. But this is the **protocol-level domain**, not the prover's actual work. The dense data in the `Mle` inner tensor has dimensions `stored_rows × width`, and all expensive operations (DFT, Merkle, sumcheck) run on the dense representation. The virtual zeros only appear in the algebraic protocol as implicit padding, handled by the `PaddedMle` abstraction without materialization.

**Summary:** SP1 effectively pays only for the real data plus trivial alignment overhead (≤31 rows per chip + stacking alignment across the shard). This contrasts sharply with RISC Zero, where the prover pays for the full `2^po2 × total_columns` domain including all padding rows.

#### 5.2.3 SepticDigest: The Cross-Shard Memory Bridge

SP1's cross-shard memory consistency uses an **elliptic curve multiset hash** on a purpose-built septic curve. Here is the full mechanism:

**The curve:** `y² = x³ + 45x + 41z³` over the septic extension field `F_{p^7} = F_p[z]/(z^7 − 3z − 5)`, where p is the KoalaBear prime. This gives a curve with a large enough group order for 100-bit security, while keeping coordinates in the same field as the proof system.

**Hashing interactions onto the curve:**

1. **Message encoding:** Each memory interaction is encoded as an 8-limb `u32` message. The interaction `kind` (Memory, Global, etc.) is OR'd into the high bits of the first limb.

2. **Trial-and-increment lift:** For offsets 0..255, the prover builds a 16-element Poseidon2 input `m_trial` (8 message limbs + offset encoding in limb 7 + 8 zero padding), hashes through Poseidon2, and takes the first 7 elements of the hash output as an x-coordinate in `F_{p^7}`. If `curve_formula(x)` yields a quadratic residue, the square root gives `y` and `(x, y)` is a valid curve point.

3. **Send/receive sign convention:** **Sends negate** the curve point; **receives keep** the original sign. This ensures that a matched send-receive pair cancels to the identity.

4. **Accumulation:** The `GlobalChip` in each shard maintains a running EC sum (`SepticCurveComplete`) of all curve points from that shard's memory interactions. The final affine point is the shard's `global_cumulative_sum`.

**Verification:** The machine verifier starts from `vk.initial_global_cumulative_sum` (which encodes the program's initial memory image as negated curve points), adds each shard's `global_cumulative_sum`, and checks the result equals `SepticDigest::zero()` (a fixed anchor point). If all sends were matched by receives across the full execution, the multiset is balanced and the sum is the identity.

**Why a septic extension?** The extension degree 7 gives coordinates with `7 × 31 = 217` bits per coordinate (assuming 31-bit KoalaBear limbs). The Hasse bound guarantees the curve group order is close to `p^7 ≈ 2^217`, providing ample security margin for the multiset hash.

**Source files:** `crates/hypercube/src/septic_curve.rs`, `crates/hypercube/src/septic_extension.rs`, `crates/hypercube/src/septic_digest.rs`, `crates/core/machine/src/global/mod.rs`, `crates/core/machine/src/operations/global_interaction.rs`.

### 5.3 Linea Limitless: Multi-Module Structural Segmentation

Linea uses **separate modules** for each EVM operation, with segmentation by **module and row-band** (structural), not by time.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    LINEA LIMITLESS: MULTI-MODULE ARCHITECTURE                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│ Each EVM operation has its own MODULE with specialized constraints.             │
│ Segmentation is STRUCTURAL (by module + row bands), NOT temporal.               │
│ Cross-module lookups span segments → requires shared randomness.                │
└─────────────────────────────────────────────────────────────────────────────────┘

FULL CONFLATED TRACE (multiple EVM blocks):
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  MODULE: HUB-A (central dispatch)       MODULE: HUB-B (memory/gas)             │
│  ┌───────────────────────────┐          ┌───────────────────────────┐          │
│  │ hub.STAMP │ hub.OPCODE    │          │ mxp.* │ mmio.* │ gas.*    │          │
│  ├───────────┼───────────────┤          ├───────┼────────┼──────────┤          │
│  │     1     │   PUSH1       │─────────►│ [memory expansion]       │          │
│  │     2     │   ADD         │          │ [gas calculations]       │          │
│  │     3     │   CALL        │──┐       │        ...               │          │
│  │    ...    │    ...        │  │       └───────────────────────────┘          │
│  └───────────────────────────┘  │                                               │
│                                 │                                               │
│  MODULE: KECCAK                 │       MODULE: ARITH-OPS                       │
│  ┌───────────────────────────┐  │       ┌───────────────────────────┐          │
│  │ keccak.* │ rom.* │ rlp.*  │◄─┘       │ add.* │ mul.* │ exp.*    │          │
│  ├──────────┼───────┼────────┤          ├───────┼───────┼──────────┤          │
│  │ [hash computation]        │          │ [arithmetic operations]  │          │
│  │ [24 rounds keccak-f]      │          │ [polynomial checks]      │          │
│  └───────────────────────────┘          └───────────────────────────┘          │
│                                                                                 │
│  MODULE: ECDSA                          MODULE: BLS-G1                          │
│  ┌───────────────────────────┐          ┌───────────────────────────┐          │
│  │ ecrecover.* │ ext.*       │          │ bls_g1_add.* │ msm.*     │          │
│  └───────────────────────────┘          └───────────────────────────┘          │
│                                                                                 │
│  ... (15+ modules: SHA2, P256, BLS-G2, BLS-PAIRING, MODEXP, BN-EC-OPS, etc.)   │
└─────────────────────────────────────────────────────────────────────────────────┘

STRUCTURAL SEGMENTATION:
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Each module is segmented into row-bands, producing TWO proof types:             │
│                                                                                 │
│  ┌─────────────────────────────────────────────────────────────────────────┐   │
│  │ GL (Global + Local)              │ LPP (Log-deriv, Permutation, Poly)   │   │
│  │ • Column commitments             │ • Cross-module lookup arguments      │   │
│  │ • Local constraints              │ • Grand product (permutations)       │   │
│  │ • Global constraints             │ • Horner evaluation (polynomials)    │   │
│  │ • Produces: LppCommitment        │ • Uses: shared randomness from GL    │   │
│  └─────────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────────┘
```

#### Horizontal vs vertical segmentation (Limitless)

Limitless uses two orthogonal axes; the codebase names them **horizontal** and **vertical**:

| Axis | What it splits | In code / witness | Why it exists |
|------|------------------|-------------------|---------------|
| **Horizontal** | The conflated trace into **separate modules** (different EVM operation clusters / AIRs). | `ModuleName` is documented as identifying an *horizontal* module (`prover/protocol/distributed/module_discovery.go`). Each module (HUB-A, KECCAK, …) is proved as its own segment family. | Specialization: each module has its own columns and constraints; cross-module links are lookups, not one giant table. |
| **Vertical** | **One module’s** assignment into **several row-band proofs** when the trace is too tall for a single circuit. | `SegmentModuleIndex` is the vertical slice index; `IsFirst` / `IsLast` mark whether a segment is the first or last vertical instance of that module (`ModuleWitnessGL`, `ModuleGL` in `prover/protocol/distributed/module_witness.go`, `module_gl.go`). | “Limitless” scale: a single module (e.g. HUB-A) can still exceed one Plonk domain, so it is **stacked vertically** into multiple GL/LPP segment pairs that chain via `SentValuesGlobal` / `ReceivedValuesGlobal`. |

**Intuition:** Think of the full conflated trace as a grid. **Horizontal** cuts separate **columns of responsibility** (which submodule owns which constraints). **Vertical** cuts **along the row axis** within one submodule when height must be chunked.

**Relation to GL/LPP:** GL and LPP are not a third segmentation axis; they are **two proof phases per (horizontal, vertical) cell**—first commitments and local/global checks (GL), then lookup/permutation/polynomial arguments (LPP) using challenges derived from all GL outputs.

**Soundness note:** Some local constraints may only be safe if they do not span vertical boundaries incorrectly. The compiler guards against patterns that would only hold if the module had a **single** vertical segment; forcing one vertical segment would remove chunking and hurt scalability (`ModuleGL.InsertLocal` in `prover/protocol/distributed/module_gl.go`).

**Linea Module Clusters (from codebase):**

| Module | Key Tables | Typical Size |
|--------|------------|--------------|
| **HUB-A** | `hub.*`, `hub×4.*` | 262K - 1M rows |
| **HUB-B** | `mxp.*`, `mmio.*`, `mmu.*`, `gas.*` | 16K - 1M rows |
| **ARITH-OPS** | `add.*`, `mul.*`, `mod.*`, `exp.*`, `u32.*`, `bin.*` | 32K - 2M rows |
| **KECCAK** | `keccak.*`, `rom.*`, `shakiradata.*` | 16K - 8M rows |
| **ECDSA** | `ecrecover.*`, `ext.*` | 32K - 65K rows |
| **SHA2** | sha2 packing, blocks, hash | 512 - 16K rows |
| **BLS-G1/G2** | BLS curve operations | varies |
| **MODEXP** | modular exponentiation | 256 - 65K rows |
| **TINY-STUFFS** | `romlex.*`, `txndata.*`, public input | 1 - 262K rows |

### 5.4 Two-Phase Proving Pipeline (Linea)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         LINEA TWO-PHASE PIPELINE                                │
└─────────────────────────────────────────────────────────────────────────────────┘

PHASE 1: GL Proofs (all parallel)
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│ HUB-A-GL │ │ HUB-B-GL │ │KECCAK-GL │ │ARITH-GL  │  ...
└────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
     │            │            │            │
     └────────────┴────────────┴────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ SHARED RANDOMNESS = Poseidon2(MSetHash(moduleIndex, segmentIndex, LppCommit)...) │
└─────────────────────────────────────────────────────────────────────────────────┘
                       │
                       ▼
PHASE 2: LPP Proofs (all parallel, using shared randomness)
┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐
│ HUB-A-LPP │ │ HUB-B-LPP │ │KECCAK-LPP │ │ARITH-LPP  │  ...
│ LogDeriv  │ │ LogDeriv  │ │ LogDeriv  │ │ LogDeriv  │
│ = partial │ │ = partial │ │ = partial │ │ = partial │
└─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘
      │             │             │             │
      └─────────────┴─────────────┴─────────────┘
                       │
                       ▼
HIERARCHICAL CONGLOMERATION (2-by-2 binary tree):
┌─────────────────────────────────────────────────────────────────────────────────┐
│ At each merge:                                                                  │
│   sumLogDerivative = child1.LogDerivativeSum + child2.LogDerivativeSum         │
│   prodGrandProduct = child1.GrandProduct × child2.GrandProduct                 │
│   sumHorner = child1.HornerSum + child2.HornerSum                              │
└─────────────────────────────────────────────────────────────────────────────────┘
                       │
                       ▼
OUTER CIRCUIT (execution-limitless):
┌─────────────────────────────────────────────────────────────────────────────────┐
│ FINAL ASSERTIONS:                                                               │
│   • LogDerivativeSum == 0  (all cross-module lookups satisfied)                │
│   • GrandProduct == 1      (all permutations valid)                            │
│   • HornerSum == 0         (all polynomial checks pass)                        │
│   • VK Merkle root matches (all segments from valid circuits)                  │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 5.5 Trace Structure Comparison Summary

> **Naming note:** Airbender calls its segments **"chunks"** and its precompiles **"delegation circuits."** ZisK calls its precompiles **"secondary state machines."** In the tables below we use the document's standard terminology (segment, precompile) for easy comparison.

| Aspect | RISC Zero | SP1 | Linea Limitless | ZKsync Airbender | ZisK |
|--------|-----------|-----|-----------------|------------------|------|
| **Field / PCS** | BabyBear / FRI (DEEP-ALI) | KoalaBear / Jagged Basefold (sumcheck) | — / Plonk + Groth16 | Mersenne31 / FRI | Goldilocks / Proofman (Vadcop) |
| **Trace layout** | Single unified table | Multiple chips/tables | Multiple modules | Single main table + separate delegation tables | Main SM + multiple secondary SM tables (one PIL airgroup) |
| **Column sharing** | All columns in every row | Each chip has own columns | Each module has own columns | Shared columns in main table; each delegation type has own columns | Main SM has its own columns; each secondary SM has specialized columns |
| **Segmentation** | Temporal (by cycle) | Temporal (~2M cycles) | Structural (module + row bands) | Temporal (~2²² cycles per segment) | Temporal (~2¹⁸ steps per chunk) |
| **Precompile rows** | Same table, different selector | Separate chip | Separate module | Separate delegation circuit (not muxed into main table) | Separate secondary state machine (delegated via bus) |
| **Cross-segment lookups** | None (all local) | None (chip lookups local) | Yes (cross-module) | None documented; continuity via shuffle RAM init/teardown | To be investigated |
| **Memory consistency** | Merkle root chaining | EC multiset hash | Permutation within modules | Timestamped shuffle RAM + lazy init/teardown at boundaries | Dedicated memory SMs (aligned/unaligned/input) via memory proxy; cross-chunk mechanism to be investigated |
| **Shared randomness** | No | No | Yes (GL → LPP barrier) | No (per-segment Fiat-Shamir) | To be investigated |
| **Table replication** | Yes (4,112+ rows/segment) | No | No | To be investigated | Virtual lookup tables grouped by PIL config (`set_group_virtual_tables`); replication strategy to be investigated |
| **Shape normalization** | 13 fixed programs | Dynamic cache (LRU `SP1NormalizeCache`) | Fixed per module | Multiple machine configs; exact count to be investigated | Proofman plans SM instance counts and sizes per execution; normalization details to be investigated |
| **Variable-size handling** | N/A (uniform rows) | Jagged PCS (pay only for rows used) | Per-module domain sizes | To be investigated | Proofman dynamically plans number and size of SM instances; exact mechanism to be investigated |

### 5.6 ZKsync Airbender: Trace Structure

The main trace is a **single table**: each **row = one CPU cycle**, each **column = a witness variable** (PC, opcode flags, ALU intermediates, memory-argument fields). The trace length is a power of two; a segment contains `trace_size - 1` executable cycles. Instruction dispatch uses heavy **opcode muxing** ("execute all, choose one" — `docs/circuit_overview.md`) rather than separate tables per opcode family. Bytecode is fetched via a **ROM lookup** keyed by PC.

**Precompiles** (called "delegation circuits") are **separate tables** with their own AIR and column layout — similar to SP1's chip model, not RISC Zero's single-table muxing. Dispatch is via **CSR `0x7c0`**. Implemented types: **BLAKE2** (round + extended control) and **BigInt/u256** (`docs/delegation_circuits.md`). Register and memory accesses from delegations participate in the **same shuffle RAM argument** as the main table.

```
SEGMENT (= "chunk" in Airbender naming):

┌───────┬──────────┬────────────────────┬────────────────────┬────────────┐
│ Cycle │ PC       │ Opcode flags       │ ALU / decoder /    │ Shuffle RAM│
│       │ (limbs)  │ (orthogonal select)│ lookups, ranges    │ + timestamp│
├───────┼──────────┼────────────────────┼────────────────────┼────────────┤
│   0   │ 0x1000…  │ ADD                │ …                  │ mem events │
│   1   │ 0x1004…  │ LOAD               │ …                  │ mem events │
│   2   │ 0x1008…  │ CSRRW → delegation │ …                  │ mem events │
│  …    │ …        │ …                  │ …                  │ …          │
│ pad   │ (zeros)  │ —                  │ —                  │ —          │
└───────┴──────────┴────────────────────┴────────────────────┴────────────┘

TEMPORAL SEGMENTATION:

┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
│ Segment 0            │  │ Segment 1            │  │ Segment 2 …          │
│ ~2²² cycles          │  │ ~2²² cycles          │  │                      │
│ Main table → STARK   │  │ Main table → STARK   │  │                      │
├──────────────────────┤  ├──────────────────────┤  ├──────────────────────┤
│ Teardown: final      │──│ Lazy init: first     │  │ Same pattern         │
│ (timestamp, value)   │  │ accesses match prior │  │                      │
│ per touched word     │  │ segment's teardown   │  │                      │
└──────────────────────┘  └──────────────────────┘  └──────────────────────┘

DELEGATION (separate proofs, not rows in the main table):

┌────────────────────────┐   ┌────────────────────────┐
│ BLAKE2 circuit         │   │ BigInt/u256 circuit     │
│ Own columns + STARK    │   │ Own columns + STARK     │
└────────┬───────────────┘   └────────┬───────────────┘
         └───────────┬───────────────┘
                     ▼
         ProgramProof { base_layer_proofs[], delegation_proofs[type] }
```

**Cross-segment memory linkage:** Each segment records a sorted **lazy init / teardown** list of `(address, timestamp, value)` for every memory word touched. The next segment's lazy-init entries must match the prior segment's teardown values — analogous to RISC Zero's Merkle-root chaining but using a shuffle argument, and to SP1's SepticDigest but using explicit init/teardown rows.

**Source files:** `circuit_defs/trace_and_split/src/lib.rs`, `prover/src/tracers/oracles/mod.rs`, `execution_utils/src/proofs.rs`, `docs/circuit_overview.md`, `docs/delegation_circuits.md`.

### 5.7 ZisK: Main + Secondary State Machines (PIL Airgroup)

ZisK uses a **main state machine** plus multiple **secondary state machines**, all defined in a single PIL airgroup. The main SM executes Zisk instructions (transpiled from RISC-V rv64ima ELF binaries); complex operations are delegated to specialized secondary SMs via bus-style interactions.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                   ZisK: MAIN + SECONDARY STATE MACHINES                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│ One PIL airgroup ("airgroup Zisk") containing all state machines.              │
│ Main SM handles instruction dispatch; secondary SMs prove specific operations. │
│ Communication via bus-style lookups within the airgroup.                       │
└─────────────────────────────────────────────────────────────────────────────────┘

CHUNK N (temporal slice, ~2¹⁸ steps):
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                                                                 │
│  ┌───────────────────────────┐      ┌───────────────────────────┐              │
│  │     MAIN STATE MACHINE    │      │     ROM STATE MACHINE     │              │
│  ├───────────────────────────┤      ├───────────────────────────┤              │
│  │ Step│ PC   │ Op  │ a,b,c │      │ Fixed instruction table   │              │
│  ├─────┼──────┼─────┼───────┤      │ (transpiled from ELF)     │              │
│  │  0  │0x800 │ ADD │ ...   │◄────►│ Keyed by PC               │              │
│  │  1  │0x808 │LOAD │ ...   │      └───────────────────────────┘              │
│  │  2  │0x810 │ECALL│ ...   │──┐                                               │
│  │ ... │ ...  │ ... │ ...   │  │   ┌───────────────────────────┐              │
│  └───────────────────────────┘  │   │   MEMORY PROXY           │              │
│                                 │   ├───────────────────────────┤              │
│                                 │   │ → Memory Aligned SM      │              │
│                                 │   │ → Memory Unaligned SM    │              │
│                                 │   │ → Memory Input SM        │              │
│                                 │   └───────────────────────────┘              │
│                                 │                                               │
│  ┌───────────────────────────┐  │   ┌───────────────────────────┐              │
│  │     KECCAK-F SM           │  │   │     SHA-256-F SM          │              │
│  ├───────────────────────────┤  │   ├───────────────────────────┤              │
│  │ Keccak-f 1600 permutation │◄─┤   │ SHA-256 extend+compress  │              │
│  └───────────────────────────┘  │   └───────────────────────────┘              │
│                                 │                                               │
│  ┌───────────────────────────┐  │   ┌───────────────────────────┐              │
│  │     ARITH SM              │  │   │  BINARY PROXY             │              │
│  ├───────────────────────────┤  │   ├───────────────────────────┤              │
│  │ Arithmetic operations     │◄─┘   │ → Binary Basic SM → Table │              │
│  │ arith_eq, arith_eq_384    │      │ → Binary Extension → Table│              │
│  │ big_int_add               │      └───────────────────────────┘              │
│  └───────────────────────────┘                                                 │
│                                                                                 │
│  ┌───────────────────────────┐      ┌───────────────────────────┐              │
│  │    POSEIDON2 SM           │      │    BLAKE2 SM              │              │
│  ├───────────────────────────┤      ├───────────────────────────┤              │
│  │ Poseidon2 compression     │      │ BLAKE2 round function     │              │
│  └───────────────────────────┘      └───────────────────────────┘              │
│                                                                                 │
│  ┌───────────────────────────┐      ┌───────────────────────────┐              │
│  │    DMA SMs                │      │  FREQUENT-OPS SM          │              │
│  ├───────────────────────────┤      ├───────────────────────────┤              │
│  │ dma, dma_rom, dma_pre_post│      │ Common small operations   │              │
│  │ dma_64_aligned, unaligned │      └───────────────────────────┘              │
│  └───────────────────────────┘                                                 │
│                                                                                 │
│  Virtual Tables: ARITH_TABLE, BINARY_TABLE, BINARY_EXTENSION_TABLE,            │
│  KECCAKF_TABLE, MEMORY_ALIGN_ROM, DMA_ROM, etc.                               │
└─────────────────────────────────────────────────────────────────────────────────┘

TEMPORAL SEGMENTATION:

┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
│ Chunk 0              │  │ Chunk 1              │  │ Chunk 2 …            │
│ ~2¹⁸ steps           │  │ ~2¹⁸ steps           │  │                      │
│ Main + all secondary │  │ Main + all secondary │  │                      │
│ SMs → PIL airgroup   │  │ SMs → PIL airgroup   │  │                      │
├──────────────────────┤  ├──────────────────────┤  ├──────────────────────┤
│ Cross-chunk memory   │──│ linkage via memory   │──│ SM boundary state    │
│ boundary state       │  │ proxy                │  │                      │
└──────────────────────┘  └──────────────────────┘  └──────────────────────┘
```

**Key architectural properties:**

- **Transpilation**: Guest programs compile to RISC-V rv64ima ELF binaries, then are transpiled to a **Zisk ROM** — a sequence of Zisk instructions, each with opcode and 64-bit operands `a`, `b`, producing `(c, flag)` (`core/src/zisk_inst.rs`, `core/src/riscv2zisk.rs`).
- **Single PIL airgroup**: All state machines (main + secondaries) are defined in one `airgroup Zisk` block (`pil/zisk.pil`). This is architecturally between RISC Zero's single-table muxing and SP1's fully independent chips — the SMs are separate tables but share one constraint compilation unit.
- **Proof delegation**: The main SM delegates complex operations (binary, arithmetic, memory, hashing, ECC) to secondary SMs. Each secondary SM can have multiple instances, planned dynamically per execution.
- **Virtual tables**: Lookup tables (arith, binary, keccak, etc.) are grouped into virtual tables with configurable max sizes (`set_max_std_tables_bits(20)`, `set_max_num_rows_virtual(1 << 21)` in `pil/zisk.pil`).
- **Public inputs**: 32 × 64-bit values (2048 bits total), first 32 bytes reserved for the program verification key.

**Proof system details:**

| Component | ZisK's Choice | Source |
|-----------|--------------|--------|
| **Field** | Goldilocks (p = 2⁶⁴ − 2³² + 1) | `sdk/src/prover/backend.rs` — `ProofMan<Goldilocks>` |
| **Constraint system** | PIL (Polynomial Identity Language) | `pil/zisk.pil` |
| **Proving stack** | Proofman (pil2-proofman v0.16.1) | Workspace `Cargo.toml` — git dep on `pil2-proofman` |
| **Intermediate proof** | Vadcop Final (optionally compressed) | `sdk/src/prover/mod.rs` — `ZiskProof::VadcopFinal` / `VadcopFinalCompressed` |
| **SNARK wrapper** | Plonk or FFLONK | `sdk/src/prover/mod.rs` — `ZiskProof::Plonk` / `ZiskProof::Fflonk`, `SnarkProtocol::Plonk` / `SnarkProtocol::Fflonk` |
| **Hash function** | To be investigated (Proofman-internal) | |
| **PCS** | To be investigated (Proofman-internal) | |
| **Security level** | To be investigated | |

**Source files:** `pil/zisk.pil`, `core/src/lib.rs`, `core/src/zisk_inst.rs`, `core/src/riscv2zisk.rs`, `core/src/zisk_definitions.rs`, `sdk/src/prover/backend.rs`, `sdk/src/prover/mod.rs`.

---

## 6. Segmentation Strategies

This is the most important architectural difference between the five systems. It determines how cross-segment relationships are handled, which in turn determines parallelism properties.

**Key insight — temporal segmentation prevents cross-segment lookups:**

When you segment by time (RISC Zero, SP1), both sides of any lookup happen at the same moment:
- CPU dispatches Keccak at cycle 1000 → Keccak executes at cycle 1000 → both in the same segment

The only exception is **memory**, where writes and reads can be far apart in time:
- Write to address 0x500 at cycle 1000 (segment 1) → Read from 0x500 at cycle 50000 (segment 5)

This is why RISC Zero and SP1 only need special handling for memory across segments — everything else stays local.

When you segment by structure/module (Linea), lookups between modules inherently span segments:
- Hub module dispatches to Keccak module → they're in different segments by construction → lookup spans segments

### 6.1 RISC Zero — Time Slices, Everything Local

The executor splits execution into time-ordered segments. Each segment is a self-contained STARK proof.

**Lookup table replication:** Lookup tables (U8: 0–255, U16: 0–65535) are materialized in every segment's trace. This costs a fixed `LOOKUP_TABLE_CYCLES = ((1 << 8) + (1 << 16)) / 16 = 4,112` rows per segment regardless of actual computation (`risc0/circuit/rv32im/src/execute/platform.rs`).

**Why replication is necessary:** Each segment runs its own LogUp grand product (starts at 1, must end at 1). A lookup "pull" in segment *i* can only be balanced by a "push" in the same segment. There is no cross-segment accumulator.

**Memory across segments:** A Merkle tree of the full memory state is committed at each segment boundary. The aggregator checks `segment[i].end_root == segment[i+1].start_root`.

#### 6.1.1 The Merkle Tree Commits to the Entire Address Space

The Merkle tree covers the **full 32-bit address space** (4 GB), not just the memory touched by a particular segment. The tree structure:

| Parameter | Value | Explanation |
|-----------|-------|-------------|
| Address space | 4 GB | Full 32-bit (0x00000000 to 0xFFFFFFFF) |
| Page size | 1 kB | Each leaf = 1024 bytes of memory |
| Number of pages | 2^22 | 4 GB ÷ 1 kB = 4,194,304 pages |
| **Tree depth** | **22 levels** | log2(4,194,304) = 22 |

```
                    [Root]                     ← committed at segment boundary
                   /      \
                 ...      ...
                /            \
           [node]            [node]            ← 22 levels of internal nodes
          /      \          /      \
       [leaf]  [leaf]    [leaf]  [leaf]        ← each leaf = hash of 1 kB page
```

**Why commit to the entire address space?**

1. **Deterministic Image ID:** The initial memory state (your compiled program + initial data) hashes to a single 256-bit `ImageID`. This uniquely identifies the program regardless of which segments run.

2. **Simple cross-segment verification:** The aggregator only needs to check that adjacent segments agree on the root: `segment[i].end_root == segment[i+1].start_root`. No complex state reconciliation.

3. **Sparse representation:** Most of the 4M pages are never touched. Untouched pages have a known "empty" hash, so the prover doesn't materialize all leaves—only the Merkle paths to accessed pages.

**Paging cost:** When a segment first accesses a page, it must "page in" by verifying the Merkle path from leaf to root (22 hashes). Modified pages must be "paged out" at segment end. This costs **1,094 to 5,130 cycles per page** depending on caching. Programs with scattered memory access patterns pay more paging overhead than those with good locality.

All constraints are local. No shared randomness between segments. Fully parallel proving.

**How the executor chooses segment boundaries:**

The po2 is not a fixed configuration — it is computed per-segment from actual row usage.

1. **Row-cost accounting:** A `BlockTracker` tracks cumulative row cost as weighted "row points." Each instruction and precompile adds points corresponding to its block types (e.g., `Add` costs `InstReg + UnitAddSub` points; a BigInt ecall costs `EcallBigInt + N × BigInt` points). Paging (Merkle operations) also adds points. Points are converted to rows by dividing by `POINTS_PER_ROW` (5,040).

2. **Segment boundary trigger:** Before each instruction, the executor checks whether estimated rows exceed the threshold: `2^segment_limit_po2 - max_insn_rows`. The headroom (`max_insn_rows` = 25,000 for po2 >= 15; 2,000 for smaller segments) ensures one worst-case instruction always fits.

3. **Per-segment po2:** After a segment ends, its proof po2 is `ceil(log2(used_rows))`. The trace is padded to exactly `2^po2` rows.

4. **Result in practice:** A long computation produces many "full" segments at the ceiling po2 (default `DEFAULT_SEGMENT_LIMIT_PO2 = 20` = ~1M rows, from `risc0/circuit/rv32im/src/execute/mod.rs`) plus a smaller final segment (po2 as low as 12). Users can override the ceiling via `ExecutorEnv::segment_limit_po2`.

5. **Trade-off:** Smaller ceiling = more segments = more overhead (table replication + paging + more lift/join proofs), but each segment proves faster and uses less memory. Larger ceiling = fewer segments but each is more expensive.

### 6.2 SP1 — Time Slices with Separate Chips

SP1 has many specialized chips (CPU, memory, byte table, Keccak, SHA-256, etc.), each with its own AIR. Chips communicate via cross-table lookups (LogUp sends/receives).

**Sharding is temporal:** The executor checks after every instruction whether the trace exceeds a size threshold (`element_threshold` on total trace area, or `height_threshold` on max table height). When either is exceeded, the current shard is emitted and a new one starts. There is no pre-built full trace — shards stream out during execution.

**The key property — chip lookups never span shards:** Because sharding is time-based, both sides of any CPU-to-precompile dispatch happen at the same execution cycle and land in the same shard. Only memory accesses (e.g., a write in shard 1, a read in shard 5) span shards.

**Within a shard:** All chips share one Fiat-Shamir transcript. LogUp-GKR verifies that all sends/receives balance locally.

**Across shards (memory only):** Each shard accumulates its cross-shard memory interactions onto an elliptic curve point (`SepticDigest`) via the `GlobalChip`. The machine verifier sums all shards' digests and checks the total equals the point at infinity. This is purely algebraic — no shared randomness needed.

Each shard gets a fresh Fiat-Shamir challenger seeded only with the (fixed) verification key. All shards are fully parallel with no synchronization barrier.

**Version history:**
- V1–V3: All cross-table interactions used LogUp with partial sums — required shared randomness across shards (synchronization barrier).
- V4 (Turbo): Replaced LogUp for memory with EC multiset hash — eliminated the barrier for memory.
- V6 (Hypercube): Switched to multilinear proofs with LogUp-GKR. Per-shard Fiat-Shamir eliminates all cross-shard barriers. Jagged PCS for variable-sized tables ("pay only for rows you use").

### 6.3 Linea Limitless — Structural Segmentation

Linea proves EVM execution across many blocks in a single "conflation." Each EVM operation (arithmetic, Keccak, ECDSA, etc.) has its own module with a specialized constraint system.

Unlike SP1 and RISC Zero, segmentation is **structural** (by module and row-band), not temporal:

**Horizontal vs vertical (summary):** **Horizontal** segmentation chooses *which module* (which slice of the arithmetization). **Vertical** segmentation splits *that module’s rows* into multiple segment proofs when needed. See [Horizontal vs vertical segmentation (Limitless)](#horizontal-vs-vertical-segmentation-limitless) in Section 5.3.

1. **Build the full conflated trace** across all modules
2. **Segment** each module's trace into row-bands, producing two types of segments: GL segments and LPP segments
3. **Prove all GL segments** in parallel — each produces a commitment
4. **Derive shared randomness** from GL commitments (via Poseidon2 hash)
5. **Prove all LPP segments** in parallel — using the shared randomness as the lookup challenges (gamma/alpha for the log-derivative argument)
6. **Aggregate** via hierarchical 2-by-2 conglomeration: partial log-derivative sums are added, grand products are multiplied

**GL vs LPP explained:**
- **GL (Global + Local):** Contains column commitments and constraint checks that don't require cross-module randomness. Produces the `LppCommitment` used to derive shared randomness.
- **LPP (Log-derivative, Permutation, Polynomial):** Contains the cross-module lookup arguments, grand products for permutations, and Horner evaluations. Requires the shared randomness derived from all GL proofs.

**Why shared randomness is unavoidable here:** Structural segmentation splits cross-module lookups across segments. For example, the Hub module dispatching to the Keccak module means the "send" row is in one segment and the "receive" row is in another. The log-derivative lookup argument requires both sides to use the same random challenge. Each segment computes a partial sum; the final aggregator checks that all partial sums total to zero.

**The two-phase barrier:** All GL segments must be proved before shared randomness can be derived. Only then can LPP segments be proved. This creates a single synchronization barrier.

**Source-level evidence (Linea codebase):**
- Distributed pipeline orchestration: `prover/backend/execution/limitless/prove.go` → `RunDistributedPipeline`, `Prove`
- Shared randomness derivation: `prover/protocol/distributed/distribute.go` → `GetSharedRandomnessFromSegmentProofs()`
- GL/LPP module definitions: `prover/protocol/distributed/module_gl.go`, `module_lpp.go`
- Hierarchical conglomeration: `prover/protocol/distributed/conglomeration_hierarchical.go`
- Module discovery advices: `prover/zkevm/limitless.go` → `DiscoveryAdvices()`

### 6.4 Nexus — No Segmentation (Single Proof)

The Stwo prover takes the entire execution trace and produces one proof. `Machine::prove` commits all traces (preprocessed + main + interaction) and calls `stwo::prover::prove` once. There is no loop over sub-traces and no recursive aggregation.

Utility functions like `UniformTrace::split_by` exist in the codebase for partitioning block vectors, but nothing in the proving pipeline consumes split traces as separate proofs.

**Consequence:** The full trace must fit in memory. For programs that exceed this limit, no fallback exists in the current codebase.

**This is a known limitation, not a design choice.** The Nexus team's "zkVM 3.0 and Beyond" blog post (May 2025) explicitly lists "monolithic trace" and "no modular proofs" as current limitations:

> *"We plan to decompose the single trace matrix into per-component traces, where each trace captures the behavior of an isolated subsystem. Each of these traces will come with its own set of constraints and its own proof. To maintain global consistency, we'll link shared variables across components using LogUp-based digest constraints."*

This planned approach resembles SP1's chip model (separate AIR per component, cross-component LogUp) rather than RISC Zero's single-table design or Linea's structural segmentation.

The Nexus Roadmap (updated Jan 2026) targets **zkVM 4.0 for mid-2026** with recursive composition, batching for parallel execution, and instruction sorting.

Sources: [zkVM 3.0 and Beyond](https://blog.nexus.xyz/zkvm-3-0-and-beyond-toward-modular-distributed-zero-knowledge-proofs/), [The Nexus Roadmap](https://blog.nexus.xyz/the-nexus-roadmap/).

### 6.5 ZKsync Airbender — Temporal Segmentation

Airbender segments by time: each segment is ~2²² cycles; a full execution can span ~2³⁶ cycles (`docs/philosophy_and_logic.md`). Each segment yields an independent STARK/FRI proof. RAM and delegation arguments connect segments globally.

Like RISC Zero and SP1, temporal segmentation prevents cross-segment precompile lookups — the CSR dispatch and the delegation work both occur within the same segment's time window. Only **memory** can span segments, handled by **shuffle RAM + lazy init/teardown** (not Merkle roots like RISC Zero, not an EC accumulator like SP1).

All segment proofs are **parallel** — no Linea-style GL/LPP synchronization barrier is documented.

### 6.6 ZisK — Temporal Segmentation with Chunk-Parallel Witness Generation

ZisK segments by time: each chunk contains **2¹⁸ = 262,144 steps** (`CHUNK_SIZE_BITS = 18` in `core/src/zisk_definitions.rs`). Maximum execution is **2³⁶ − 1 steps** (`DEFAULT_MAX_STEPS`), yielding up to ~262K chunks for the longest programs.

**Witness generation parallelism:** A first execution collects the minimum trace required to split execution into chunks. Each chunk is then re-executed in parallel, generating witness data for the main SM and all secondary SMs (`core/src/lib.rs`).

**Instance planning:** After execution, the Proofman plans the **number and size of secondary SM instances** required to contain all delegated operations. Different chunks may produce different instance counts for each secondary SM (e.g., more Keccak instances in chunks with heavy hashing).

Like RISC Zero, SP1, and Airbender, temporal segmentation prevents cross-segment precompile lookups — the dispatch and precompile work both happen within the same chunk's time window. Only **memory** can span chunks, handled by the memory proxy state machines.

**Distributed proving:** The distributed system splits work into **partial contributions** assigned to workers based on compute capacity, then a global challenge phase, then proof generation, then **aggregation** where one worker collects all partial proofs and produces the final proof (`book/getting_started/distributed_execution.md`).

---

## 7. Precompile / Accelerator Design

How each system handles accelerated operations (hash functions, elliptic curve arithmetic, modular math) affects whether cross-table lookups arise and how complex the proving architecture becomes.

### 7.1 RISC Zero — Everything in One Table

All operations — regular RISC-V instructions and precompiles — share a **single trace table**. Each row activates exactly one "block type" via a two-hot selector (`major × minor`). A `BigInt` row and an `InstReg` row occupy the same column layout; only the selector bits differ.

**The single-table structure:**

```
Row │ Selector bits │ Block Type   │ Shared columns...
────┼───────────────┼──────────────┼──────────────────────────────────
  0 │ [1,0,0,0,...] │ InstReg      │ [reg data used] [bigint cols = 0]
  1 │ [1,0,0,0,...] │ InstReg      │ [reg data used] [bigint cols = 0]
  2 │ [0,1,0,0,...] │ InstEcall    │ [dispatch data] [bigint cols = 0]
  3 │ [0,0,1,0,...] │ EcallBigInt  │ [args from regs][bigint cols = 0]
  4 │ [0,0,0,1,...] │ BigInt       │ [unused]        [bigint cols used]
  5 │ [0,0,0,1,...] │ BigInt       │ [unused]        [bigint cols used]
  6 │ [0,0,0,1,...] │ BigInt       │ [unused]        [bigint cols used]
  7 │ [1,0,0,0,...] │ InstResume   │ [reg data used] [bigint cols = 0]
```

Every row has columns for every block type, but only the active block type's columns contain meaningful data. The selector bits tell the constraint system which constraints to check for that row.

**Dispatch flow:** Guest calls `sys_bigint(...)` → RISC-V `ecall` instruction → circuit's `InstEcall` block → machine-mode dispatch on register A7 → precompile block types activate (e.g., `EcallBigInt` → `BigInt` rows).

**BigInt accelerator — prover/verifier interaction:**

The BigInt precompile uses **polynomial identity testing** (Schwartz-Zippel lemma) to prove modular arithmetic efficiently. Here's how the prover and verifier interact:

1. **Prover executes and provides the answer:** When the guest calls `sys_bigint(a, b, modulus)`, the prover (host) actually computes `c = a * b mod m` using normal math. This answer `c` is written into the trace as "advice" (witness data).

2. **Prover commits to the trace:** The prover converts the trace to polynomials and commits (hashes) them. At this point, the prover is locked in — can't change the trace.

3. **Verifier provides random challenge z:** Via Fiat-Shamir, a random field element `z` is derived from the commitment. The prover couldn't have predicted this when building the trace.

4. **Prover evaluates polynomials at z:** Across multiple `BigInt` rows, the circuit evaluates the operands as polynomials at z:
   - Row 4: Accumulate `a(z)` from bytes 0–15
   - Row 5: Continue accumulating `a(z)`, start `b(z)`
   - Row 6: Continue, compute `c(z)`, `m(z)`, quotient `k(z)`
   - Row 7: Final check — assert `a(z) * b(z) = c(z) + k(z) * m(z)`

5. **Why cheating fails:** If the prover gave a wrong answer `c'`, the polynomial identity `a(x) * b(x) = c'(x) + k(x) * m(x)` would be false. A false polynomial identity of degree N has at most N roots. The probability that random `z` hits a root is ~N/field_size ≈ 0. So checking at one random point is enough.

**Why this is efficient:** Instead of proving thousands of constraint rows for grade-school multiplication, you prove ~N/16 rows (where N is byte length) that just evaluate polynomials. Much cheaper.

**Poseidon2:** Multi-step hashing across dedicated block types (`EcallP2` → `P2Step` → `P2ExtRound` / `P2IntRounds`). Memoized — repeated hashes of the same data reuse the first computation.

**SHA-256:** No dedicated block type. Implemented as synthetic microcode using generic building blocks (memory reads/writes, standard arithmetic units). Slower than dedicated circuits but avoids adding more block types.

**Supported:** SHA-256, Keccak, secp256k1, P-256, Ed25519/Curve25519, RSA, BLS12-381, BigInt modular arithmetic, KZG.

**Consequence:** No cross-table lookups at all. The "dispatch" from CPU to precompile is just selector bits changing from one row to the next within the same table. This simplicity is why RISC Zero's normalization only needs 13 pre-compiled programs (one per po2) — the table shape is always the same regardless of which precompiles ran.

#### 7.1.1 Three Accelerator Paths (Not All BigInt)

RISC Zero has three distinct acceleration mechanisms, each with dedicated block types:

| Accelerator | Block Types | What it handles |
|-------------|-------------|-----------------|
| **BigInt VM** | `EcallBigInt` → `BigInt` | All elliptic curve operations (secp256k1, P-256, Ed25519, BLS12-381), RSA, modular arithmetic |
| **Poseidon2** | `EcallP2` → `P2Step` → `P2ExtRound` / `P2IntRounds` / `P2Block` | Poseidon2 hashing (zkVM-native hash) |
| **None (microcode)** | Generic instruction blocks (`InstReg`, `InstLoad`, etc.) | SHA-256 runs as synthetic RISC-V microcode — no dedicated circuit |

The BigInt accelerator is the most general-purpose: it runs a programmable "bibc" bytecode that can express arbitrary modular arithmetic. Patched cryptography crates (e.g., `k256` for secp256k1) compile down to bibc programs. But Poseidon2 has its own dedicated circuit for maximum efficiency, and SHA-256 has no acceleration at all.

#### 7.1.2 How a Precompile Call Spans Multiple Rows

A precompile invocation (e.g., an ECDSA signature verification via the BigInt accelerator) **cannot fit in a single cycle**. It generates multiple consecutive rows in the trace:

**Example: BigInt ecall for elliptic curve operation**

```
Cycle │ Block Type    │ What happens
──────┼───────────────┼─────────────────────────────────────────────────
  N   │ InstEcall     │ Guest executes `ecall`; trap to machine mode
 N+1  │ EcallBigInt   │ Dispatch header: read blob_ptr, program size, etc.
 N+2  │ BigInt        │ First bibc instruction: load 16 bytes from memory
 N+3  │ BigInt        │ Second bibc instruction: load another 16 bytes
 ...  │ BigInt        │ (one row per bibc VM instruction)
 N+k  │ BigInt        │ Final bibc instruction: polynomial identity check
N+k+1 │ InstResume    │ Return to guest code
```

The number of `BigInt` rows depends on the operation's complexity. For a 256-bit curve operation, this could be dozens of rows. The block tracker accounts for this:

```rust
// From risc0/circuit/rv32im/src/execute/block_tracker.rs
pub fn track_ecall_bigint(&mut self, verify_program_size: u64) {
    self.blocks.add_block(BlockType::EcallBigInt);           // 1 row
    self.blocks.add_blocks(BlockType::BigInt, verify_program_size);  // N rows
    self.blocks.add_blocks(BlockType::MakeTable, (verify_program_size + 1).div_ceil(8));
}
```

#### 7.1.3 Witness Structure for BigInt Rows

Each `BigInt` row stores a fixed-size slice of the operation's data:

```c
// From risc0/circuit/rv32im-sys/cxx/rv32im/witness/bigint.h
struct BigIntWitness {
  uint32_t cycle;           // Which cycle this row represents
  uint32_t mm;              // Machine mode flag
  uint32_t first;           // Is this the first BigInt row in the sequence?
  PhysMemReadWitness inst;  // The bibc instruction word
  PhysMemReadWitness baseReg; // Base register for memory addressing
  uint32_t data[4];         // 16 bytes of bigint data (4 × u32)
  uint32_t prevCycle[4];    // Memory consistency: previous access cycle per word
  uint32_t prevValue[4];    // Memory consistency: previous value per word
};
```

**Key observation:** Each row handles **16 bytes** (4 × 32-bit words). A 256-bit (32-byte) operand requires at least 2 rows just to load. An operation like `a × b mod m` with 256-bit operands needs to load `a`, `b`, `m`, compute the result `c`, and verify the polynomial identity — hence many rows.

There is no special-cased structure for specific operations like `ecrecover(hash, v, r, s)`. The inputs live in guest memory; the bibc program reads them 16 bytes at a time into generic `BigInt` rows. The proof sees only the generic row structure, not the semantic meaning of the data.

#### 7.1.4 Why Precompile Rows Stay in the Same Segment

Because RISC Zero uses **temporal segmentation**, all rows for a single precompile invocation are consecutive in time:

- `InstEcall` at cycle N
- `EcallBigInt` at cycle N+1  
- `BigInt` rows at cycles N+2, N+3, ..., N+k
- `InstResume` at cycle N+k+1

These rows land in the **same segment** (unless the segment boundary happens to fall in the middle, which would be rare and handled by the segment-splitting logic). The dispatch and all computation rows share the same Fiat-Shamir transcript. No cross-segment lookups are needed for precompile operations.

This is the key advantage of temporal segmentation: the "send" (CPU dispatches precompile) and "receive" (precompile executes) happen at the same moment in time, so they're always co-located. Only **memory** — where a write at cycle 1000 might be read at cycle 50000 — can span segments.

### 7.2 SP1 — Separate Chip per Precompile

Each precompile is a dedicated STARK table (chip) with its own AIR. The CPU chip dispatches via `ecall`; the precompile chip records events and generates its own trace. Chips communicate via cross-table lookups (send/receive interactions compiled into LogUp).

**Supported:** SHA-256, Keccak-256, secp256k1, secp256r1, Ed25519, BLS12-381, RSA, BigInt, uint256.

**Consequence:** Cross-table lookups between CPU and precompile chips are needed, but temporal sharding ensures both sides land in the same shard. No cross-shard lookup problem for precompiles.

### 7.3 Linea — Separate Module per Precompile

Each EVM precompile is a completely separate arithmetic module with its own constraint system, maximally specialized for each operation. The Hub module dispatches to precompile modules via cross-module lookups.

**Supported:** Keccak, ECDSA/ecrecover, ModExp, EC Add/Mul/Pair, SHA-256, BLS (G1/G2 Add, MSM, Map, Pairing), EIP-4844 Point Evaluation, P256.

**Consequence:** Cross-module lookups are intrinsic. After structural segmentation, these lookups span segments — this is what drives the need for shared randomness.

### 7.4 Nexus — Custom Opcodes + Extension Components

**Execution side:** Precompiles use reserved custom RISC-V opcodes (`custom-0/1/2`), not Linux-style `ECALL` syscalls. Dynamic precompiles are described in the ELF via `.note.nexus-precompiles`. An `InstructionExecutorRegistry` maps opcode → executor at runtime.

**Proving side (Keccak — the only implemented precompile):** Split across two layers:
- **KeccakChip** (in the main chip tuple): A bridge that records inputs, runs `keccakf` in the prover to derive witness data, and sets a flag. It does *not* constrain the cryptographic result — the `IsCustomKeccak` constraints are minimal, with a TODO for fuller decoding/register-access constraints (see gap #9).
- **Keccak extension components** (separate Stwo components): `PermutationMemoryCheck`, two `KeccakRound` instances, `XorTable`, `BitNotAndTable`, `BitRotateTable`. These contain the actual Keccak arithmetic. Their LogUp contributions must balance with the main trace so the total multiset sum is zero.

**Default vs Keccak-enabled:** The public `prove()` calls `prove_with_extensions(&[], ...)` — Keccak extensions are off by default. Using `keccakf` in a guest requires explicitly passing `ExtensionComponent::keccak_extensions()`.

**Not yet supported:** Out-of-crate or dynamically loaded precompile circuits.

**Consequence:** Architecturally closer to SP1/Linea (dedicated sub-circuits + cross-component LogUp) than to RISC Zero's single muxed table. But everything is one Stwo proof — no cross-segment issues because there are no segments.

### 7.5 ZKsync Airbender — Delegation Circuits via CSR

**Dispatch:** Programs invoke precompiles via `CSRRW` on CSR `0x7c0` (not `ECALL`). Each delegation type is a **separate compiled circuit** with its own trace and lookup tables (`docs/delegation_circuits.md`).

**Implemented types:** BLAKE2 (single round + extended control) and BigInt/u256 (ADD, SUB, MUL, EQ, carry, memcopy). Full production coverage — to be investigated.

**Integration:** All register and memory touches from delegations participate in the **same unified shuffle RAM argument** as ordinary instructions. This is similar to SP1 (separate AIR per accelerator + cross-table argument) but different from RISC Zero (everything muxed into one table).

**Supported operations (from docs):** BLAKE2s/Blake3 hashing, U256 field arithmetic, elliptic curve operations (secp256k1, secp256r1, BLS12), modular exponentiation.

### 7.6 ZisK — Separate Secondary State Machines via Syscalls

**Dispatch:** Guest programs invoke precompiles via RISC-V `ecall` instructions, which route to syscall handlers in the ZisK OS layer (`ziskos/entrypoint/src/syscalls/`). Each precompile is a **separate secondary state machine** with its own PIL constraints and trace.

**Architecture:** The main SM identifies delegated operations and collects required data; a first execution pass gathers the minimum trace, then chunks are re-executed in parallel, each producing witness data for all secondary SMs. The executor plans the number and size of SM instances dynamically per chunk (`core/src/lib.rs`).

**Supported precompiles (from `book/getting_started/precompiles.md` and `ziskos/entrypoint/src/syscalls/`):**

| Category | Operations |
|----------|-----------|
| **Integer arithmetic** | `add256`, `arith256` (mul+add), `arith256_mod` (modular mul+add), `arith384_mod` |
| **Hashing** | Keccak-f 1600, SHA-256-f (extend+compress), Poseidon2 (compression), BLAKE2 |
| **Elliptic curves** | secp256k1 add/dbl, secp256r1 (P-256) add/dbl, BN254 curve add/dbl, BLS12-381 curve add/dbl |
| **Field extensions** | BN254 Fp2 complex add/sub/mul, BLS12-381 Fp2 complex add/sub/mul |

**PIL-level state machines (from `pil/zisk.pil`):** main, rom, mem, mem_align, mem_align_byte, frequent_ops, binary (basic + extension + add), arith, big_int_add, arith_eq, arith_eq_384, keccakf, sha256f, poseidon2, blake2br, DMA (dma, dma_rom, dma_pre_post, dma_64_aligned, dma_unaligned), dual_range.

**Integration:** Secondary SMs communicate with the main SM via bus-style lookups defined in PIL. All SMs are part of a single `airgroup Zisk`, so intra-chunk lookups are always local (no cross-chunk lookup problem for precompiles, similar to temporal segmentation in RISC Zero/SP1/Airbender).

**Contrast with other systems:**

| System | Precompile integration | Dispatch mechanism |
|--------|----------------------|-------------------|
| **RISC Zero** | Same table, muxed rows | `ecall` → block types in unified trace |
| **SP1** | Separate chip (STARK table) per precompile | `ecall` → chip-specific trace |
| **Linea** | Separate module per EVM precompile | Hub dispatch → cross-module lookup |
| **Airbender** | Separate delegation circuit | `CSRRW` on CSR `0x7c0` |
| **ZisK** | Separate secondary state machine | `ecall` → syscall → secondary SM |

## 8. Cross-Segment Consistency

This is the core technical differentiator. When you split a trace into segments, some relationships inevitably span segments. How you handle them determines parallelism.

| | What spans segments | Mechanism | Shared randomness needed? |
|---|---|---|---|
| **RISC Zero** | Memory state only | Merkle root chaining | No |
| **SP1** | Memory only (chip lookups are local) | EC multiset hash (SepticDigest) | No |
| **Linea** | Cross-module lookups (set membership) | Partial log-derivative sums + shared randomness | Yes (one barrier) |
| **Nexus** | N/A (no segments) | All lookups resolved in single proof | N/A |
| **ZKsync Airbender** | RAM + registers across segments | Lazy init/teardown + timestamped shuffle RAM | No (per-segment Fiat-Shamir) |
| **ZisK** | Memory across chunks (SM lookups are local) | Dedicated memory SMs via memory proxy; cross-chunk linkage mechanism to be investigated | To be investigated |

### Why SP1 avoids shared randomness but Linea cannot

The root cause is the segmentation strategy:

- **SP1 (temporal):** Both sides of a CPU→precompile lookup happen at the same execution cycle, so they always land in the same shard. Only memory spans shards, and memory uses a randomness-free EC hash.
- **Linea (structural):** The Hub module and the Keccak module are in different segments by construction. Their lookup relationship spans segments, requiring shared random challenges for the log-derivative argument.
- **ZKsync Airbender (temporal):** Same logic as RISC Zero/SP1 — delegation dispatch and execution land in the same segment. Cross-segment issues are dominated by RAM/register continuity, handled by shuffle RAM + lazy init/teardown.
- **ZisK (temporal):** Same logic — secondary SM dispatch and execution land in the same chunk. Cross-chunk issues are dominated by memory continuity, handled by the memory proxy SMs. Whether shared randomness is needed for the Proofman aggregation pipeline — to be investigated.

## 9. Proof Aggregation Pipelines

After segment proofs are generated, they must be aggregated into one final proof. Segment proofs are heterogeneous (different sizes, different chips active), but the aggregator circuit needs homogeneous inputs. This is the **normalization problem**.

### Why normalization matters

The aggregator circuit must be compiled ahead of time. It needs to know the exact "shape" of the proof it's verifying — how many columns, which constraints, etc. If segment proofs have different shapes, you'd need a different aggregator for each shape.

**RISC Zero (single table):** All segments have the same column layout regardless of what operations ran. The only thing that varies is the size (po2). Result: only 13 possible shapes (po2 = 12 to 24), so 13 pre-compiled normalizer programs suffice.

**SP1 (multiple tables/chips):** Each shard activates a different subset of chips. A shard with Keccak has different columns than one without. With ~20 chips, there are potentially millions of shape combinations. Result: dynamic compilation with caching — compile each unique shape once, reuse from cache.

### 9.1 RISC Zero: lift → join → resolve → identity_p254 → shrink_wrap

```
SegmentReceipt (per-segment rv32im STARK)
    ↓ lift (one per segment, all parallel)
SuccinctReceipt (recursion-circuit STARK, uniform shape)
    ↓ join (pairwise binary tree)
single SuccinctReceipt
    ↓ resolve (discharges assumptions from proof composition — no SP1 equivalent)
unconditional SuccinctReceipt
    ↓ identity_p254 (re-proves with Poseidon over BN254 field — bridges BabyBear to Groth16)
SuccinctReceipt (BN254-native)
    ↓ shrink_wrap (Groth16)
Groth16Receipt (~260 bytes)
```

**Shape normalization:** 13 pre-compiled lift programs, one per allowed po2 (`LIFT_PO2_RANGE = 12..=24` in `risc0/circuit/recursion/src/lib.rs`). No dynamic compilation needed — the single-table architecture means all segments of the same po2 have identical shape regardless of which precompiles ran. Default max is `DEFAULT_MAX_PO2 = 22` for 97-bit security target.

**Receipt types:** CompositeReceipt (collection of raw segment STARKs) → SuccinctReceipt (single recursion STARK, constant-time verification) → Groth16Receipt (constant-size SNARK).

### 9.2 SP1: normalize → compress → shrink → wrap

```
ShardProof (per-shard, heterogeneous — different chips active per shard)
    ↓ normalize (one per shard, all parallel)
RecursionProof (fixed uniform shape; carries forward SepticDigest accumulator)
    ↓ compress (aggregates N normalized proofs into one)
single RecursionProof
    ↓ shrink (reduces proof size)
    ↓ wrap (converts to Groth16)
Groth16 proof
```

**Shape normalization:** Dynamic LRU cache (`SP1NormalizeCache`) keyed on shard shape. Each unique combination of active chips is compiled once; subsequent shards of the same shape reuse the cached circuit.

**Key differences from RISC Zero:**
- Compress is N-ary (takes multiple proofs at once); RISC Zero's join is strictly pairwise.
- SP1 handles deferred proofs through the `global_cumulative_sum` mechanism rather than an explicit resolve step.

### 9.3 Linea: Hierarchical Conglomeration

```
GL segment proofs (all parallel)
    ↓ derive shared randomness (Poseidon2 over multiset of module/segment/LppCommitment)
LPP segment proofs (all parallel, using shared randomness)
    ↓ hierarchical 2-by-2 conglomeration
root proof
    ↓ outer circuit checks: LogDerivativeSum == 0, GrandProduct == 1, HornerSum == 0
gnark proof (Groth16)
```

> **Note:** Source-level evidence for this pipeline is now included in Section 5.4 and Section 6.3. Shared randomness is `poseidon2_koalabear.HashVec` over a multiset built from each GL proof’s `(moduleIndex, segmentIndex, lppCommitment)` (`prover/protocol/distributed/distribute.go` → `GetSharedRandomnessFromSegmentProofs`).

### 9.4 Nexus: None

No aggregation pipeline exists in the current codebase. The output is a single Stwo `StarkProof` + metadata.

Recursive composition is planned for zkVM 4.0 (mid-2026). See Section 6.4.

### 9.5 ZKsync Airbender: Segments → Recursion → FFLONK

**Per-segment proving pipeline** ([ZKsync Airbender Overview](https://docs.zksync.io/zk-stack/components/zksync-airbender)):

1. **Stage 1** — Witness LDEs + trace commitments
2. **Stage 2** — Lookup and memory argument setup
3. **Stage 3** — STARK quotient polynomial
4. **Stage 4** — DEEP polynomial (FRI batching)
5. **Stage 5** — FRI IOPP proof

**Aggregation:** Multiple segment proofs + delegation proofs are bundled into `ProgramProof` (`execution_utils/src/proofs.rs`). The verifier itself is compiled to RISC-V and proven recursively through several layers (`recursion_layer.bin`, `recursion_log_23_layer.bin`, `final_recursion_layer.bin` in `tools/verifier/`).

**Stage 6 — SNARK wrapper:** The final recursive proof is wrapped into **FFLONK** for on-chain verification ([ZKsync Airbender Overview](https://docs.zksync.io/zk-stack/components/zksync-airbender)). Proof size, curve, and parameters — to be investigated.

**Shape normalization:** Multiple machine configurations (full ISA, reduced ISA, recursion-only) imply several verifier keys; exact count to be investigated.

### 9.6 ZisK: Witness → Vadcop → Aggregation → Plonk/FFLONK

```
Zisk program (RISC-V ELF)
    ↓ transpile (Riscv2zisk)
ZiskRom
    ↓ execute + plan (Executor)
Witness computation (per-chunk, parallel)
    ↓ prove per state machine instance
Individual SM proofs
    ↓ aggregate (Proofman, recursively)
Vadcop Final proof (optionally compressed)
    ↓ wrap (Plonk or FFLONK)
SNARK proof (on-chain verifiable)
```

**Witness computation:** The executor runs the program once to collect minimal trace, then splits into chunks of 2¹⁸ steps. Each chunk is re-executed in parallel, generating witness data for all state machines. The Proofman plans the number and size of SM instances (`core/src/lib.rs`).

**Per-SM proving:** Individual proofs are generated for every state machine instance (main, ROM, memory, binary, arith, keccak, etc.) using PIL constraints compiled into the `airgroup Zisk`.

**Recursive aggregation:** Individual SM proofs are aggregated into aggregated proofs, recursively, until a single **Vadcop Final** proof remains (`sdk/src/prover/backend.rs` — `aggregate_proofs`). The proof can optionally be **compressed** (`ZiskProof::VadcopFinalCompressed`).

**SNARK wrapper:** The Vadcop Final proof is wrapped into a **Plonk** or **FFLONK** SNARK for on-chain verification (`sdk/src/prover/mod.rs` — `SnarkProtocol::Plonk` / `SnarkProtocol::Fflonk`). The Plonk proving key is installed via `ziskup setup_snark`.

**On-chain verifier:** The Solidity `ZiskVerifier` contract (`zisk-contracts/ZiskVerifier.sol`) extends a snarkJS-generated `PlonkVerifier`. Verification hashes `SHA-256(programVK || publicValues || rootCVadcopFinal) mod BN254_scalar_field` and passes the digest + proof to `verifyProof`. Deployable on any EVM chain.

**Shape normalization:** The Proofman dynamically plans SM instance sizes per execution; how many distinct verifier keys are needed and how shapes are normalized — to be investigated.

**Source files:** `sdk/src/prover/backend.rs`, `sdk/src/prover/mod.rs`, `zisk-contracts/ZiskVerifier.sol`, `zisk-contracts/PlonkVerifier.sol`, `book/getting_started/quickstart.md`.

---

## 10. Parallelism Summary

| | Parallelism model | Barriers | Root cause |
|---|---|---|---|
| **RISC Zero** | All segments in parallel | None | All constraints local; per-segment randomness |
| **SP1** | All shards in parallel | None | Per-shard Fiat-Shamir; EC hash for memory needs no randomness |
| **Linea** | Two phases: GL then LPP | One (GL → derive randomness → LPP) | Structural segmentation splits lookups; log-derivative argument needs shared challenges |
| **Nexus** | Single proof | None (but trace must fit in memory) | No segmentation |
| **ZKsync Airbender** | All segments in parallel per phase | One (memory commitments → shared challenges → proofs) | Temporal segmentation; shared `ExternalChallenges` derived from all commitments before proving |
| **ZisK** | Chunk witness generation in parallel; distributed workers in parallel | At least one (partial contributions → global challenge → prove); exact barrier count to be investigated | Temporal segmentation; Proofman orchestrates multi-phase proving |

---

## 11. Prover Orchestration & Segment Communication

How do segment provers coordinate at runtime? Do they exchange messages, share memory, or operate in isolation? This section covers the concrete inter-segment communication mechanisms used during proof generation.

| | Coordination model | Inter-segment communication | Synchronization barriers |
|---|---|---|---|
| **RISC Zero** | Actor-based task queue (`r0vm`); Factory dispatches to Worker pool | None between segments — tasks flow through central Factory actor | **None** — segments proved independently; joins happen as segment receipts arrive |
| **SP1** | Controller/Worker with async task dispatch (`SP1Controller` → `WorkerClient` → per-task-type mpsc queues) | None between shards — all coordination flows through controller task queues | **None** — shards proved independently; compress tree reduces as proofs arrive |
| **Linea** | Single-process pipeline (`RunDistributedPipeline`); goroutine pools + channels | No direct segment-to-segment messages; GL/LPP proofs stream to a hierarchical conglomeration consumer | **Two barriers:** (1) all GL proofs finished → `GetSharedRandomnessFromSegmentProofs` → LPP; (2) conglomeration consumes ordered stream until closed |
| **Nexus** | N/A (single proof, no segments) | N/A | N/A |
| **ZKsync Airbender** | Central orchestrator + GPU worker pool | None between segments — channels to orchestrator only | One barrier: memory commitments → shared challenges → proofs |
| **ZisK** | Coordinator/Worker (gRPC); MPI for multi-process; Proofman orchestrates SM witness + proving | Workers compute partial contributions → coordinator derives global challenge → workers prove → aggregator produces final proof | At least one barrier: partial contributions → global challenge → proving (three-phase distributed pipeline) |

### 11.1 SP1: Controller/Worker with Async Task Dispatch

Shards never communicate with each other. SP1 uses an **async controller/worker architecture** where the controller submits typed tasks to workers via channels, and a `CompressTree` aggregates results as they stream in.

**Architecture overview:**

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                       SP1 PROVER ORCHESTRATION                               │
└──────────────────────────────────────────────────────────────────────────────┘

Phase 1: Execution + Splicing (streaming)

  ┌────────────────────┐
  │  Executor           │
  │  (runs RISC-V       │
  │   program)          │
  └──────┬─────────────┘
         │ ShardBoundary per instruction
         ▼
  ┌────────────────────┐     ┌────────────────────┐
  │ SplicingEngine     │     │ SendSpliceEngine   │
  │ (num_splicing_     │────►│ (parallel upload +  │
  │  workers)          │     │  submit ProveShard) │
  └────────────────────┘     └──────┬─────────────┘
                                    │ submit_task(ProveShard)
                                    ▼
Phase 2: Core + Normalize Proving (parallel, independent)

         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
  │ CoreWorker 0 │  │ CoreWorker 1 │  │ CoreWorker N │
  │ RISC-V STARK │  │ RISC-V STARK │  │ RISC-V STARK │
  │ + normalize  │  │ + normalize  │  │ + normalize  │
  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           │ normalized proofs stream to
                           ▼
Phase 3: Compress Tree (pipelined, as proofs arrive)

  ┌──────────────────────────────────────────────┐
  │ CompressTree                                  │
  │ Ordering: precompile | deferred | core | mem  │
  │ Batches proofs → submit RecursionReduce       │
  │ 2-ary compose programs merge batches          │
  └──────────────────┬───────────────────────────┘
                     │ single compressed proof
                     ▼
Phase 4: Shrink + Wrap (sequential)

  ┌────────────────────┐
  │ Shrink             │
  │ (reduce proof size)│
  └──────┬─────────────┘
         ▼
  ┌────────────────────┐
  │ Wrap               │
  │ (→ Groth16/Plonk)  │
  └────────────────────┘
```

**Controller (`SP1Controller`):** Drives the end-to-end pipeline: submits `CoreExecute`, subscribes to streaming `ProofData` over the task message channel, runs `CompressTree::reduce_proofs` for aggregation, then submits `ShrinkWrap` / `Groth16Wrap` after compress completes.

**Worker (`LocalWorkerClient`):** Maintains a `HashMap<TaskType, mpsc::Sender>` — one bounded channel per task kind (capacity 1). Task types include `ProveShard`, `RecursionReduce`, `ShrinkWrap`, etc. The full-node worker (`init.rs`) spawns a handler loop per task type that dispatches to `SP1ProverEngine`.

**Task dispatch mechanism:**

| Mechanism | Role |
|-----------|------|
| `WorkerClient::submit_task` | Dispatches work by `TaskType` (abstract; local impl uses per-type mpsc) |
| `AsyncEngine` + `WorkerQueue` | Worker pool with semaphore for backpressure; tasks acquire a permit, pop any idle worker, run, return worker |
| `ProverSemaphore` | Caps concurrent heavy proof work across all components |
| `CompressTree` | Internal unbounded channel "proof queue" plus event streams; batches normalized proofs and submits `RecursionReduce` tasks |

**Parallelism configuration:**

| Pool | Config key | Role |
|------|-----------|------|
| Splicing | `num_splicing_workers` | Parallel shard serialization + upload |
| Core proving | `num_core_workers` + `core_buffer_size` | Parallel RISC-V STARK proving |
| Recursion pipeline | Separate counts for prepare/execute/prove | Chained via `Chain::new(prepare, Chain::new(executor, prove))` |
| Shrink/Wrap | Sequential (one at a time) | Final step runs only once |

**Key properties:**
- **No synchronization barriers:** Each shard gets its own Fiat-Shamir challenger seeded from the (fixed) verification key. No shard depends on any other shard's output.
- **Pipelined compression:** The `CompressTree` starts reducing as soon as proofs arrive — no need to wait for all shards. Shard ordering within the tree is: precompile shards → deferred shards → core shards → memory shards.
- **No GPU in orchestration crate:** The `crates/prover` package is CPU-only orchestration; GPU proving (if any) lives behind the `SP1ProverComponents` abstraction in separate crates.

**Source files:** `crates/prover/src/worker/controller/mod.rs`, `crates/prover/src/worker/controller/core.rs`, `crates/prover/src/worker/controller/compress.rs`, `crates/prover/src/worker/controller/splicing.rs`, `crates/prover/src/worker/controller/global.rs`, `crates/prover/src/worker/prover/core.rs`, `crates/prover/src/worker/prover/recursion.rs`, `crates/prover/src/worker/client/local.rs`, `crates/prover/src/worker/node/full/init.rs`, `slop/crates/futures/src/pipeline.rs`.

### 11.2 ZKsync Airbender: Orchestrator + Channel-Based Dispatch

Segments never communicate with each other. The architecture is a centralized **orchestrator + worker pool** pattern using `crossbeam_channel` for all messaging.

**Phase 1 — CPU tracing (parallel, no inter-worker state):**
Each CPU trace worker independently re-simulates the full RISC-V program from the start but only materializes the chunk assigned to it (by `chunk_index % split_count`). Workers do not exchange state — they fast-forward to their assigned segment. Results are sent as `WorkerResult` messages (containing `SetupAndTeardownChunk` or `CyclesChunk`) on an unbounded channel to the orchestrator.

**Orchestrator pairing:**
The `ExecutionProver` collects CPU results and pairs `SetupAndTeardownChunk` + `CyclesChunk` by matching `index` in two `HashMap`s. Only when both halves arrive does it dispatch a `GpuWorkRequest`. Segments never see each other's data.

**Phase 2 — GPU memory commitments (parallel, independent):**
Each segment's memory commitment is an independent `MemoryCommitmentRequest` routed through a `GpuManager` job queue to GPU threads via bounded(0) channels (one pair per GPU device). GPU workers process one request at a time in isolation.

**Barrier — shared challenges:**
After **all** memory commitments are collected, the orchestrator derives a single `ExternalChallenges` via Fiat-Shamir over setup caps, final register values, and all memory commitments (`fs_transform_for_memory_and_delegation_arguments` → `ExternalChallenges::draw_from_transcript_seed` in `gpu_prover/src/execution/prover.rs`). This is the **only synchronization barrier** in the pipeline.

**Phase 3 — GPU proving (parallel, independent):**
Every segment receives the **same** `ExternalChallenges` in its `ProofRequest` and is proved independently. GPU workers remain unaware of other segments.

**Intra-segment parallelism:**
Within a single segment proof, a `worker::Worker` (wrapping a rayon `ThreadPool`) parallelizes LDE computation, Merkle tree construction, and quotient evaluation across CPU cores.

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                     AIRBENDER PROVER ORCHESTRATION                           │
└──────────────────────────────────────────────────────────────────────────────┘

Phase 1: CPU Tracing (parallel, independent)

  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
  │ CPU worker 0 │  │ CPU worker 1 │  │ CPU worker 2 │  …
  │ simulates    │  │ simulates    │  │ simulates    │
  │ full program │  │ full program │  │ full program │
  │ emits seg 0  │  │ emits seg 1  │  │ emits seg 2  │
  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           ▼
              ┌─────────────────────────┐
              │  Orchestrator           │
              │  (ExecutionProver)      │
              │  pairs chunks by index  │
              └────────────┬────────────┘
                           │
Phase 2: Memory Commitments (parallel, independent)
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
  │ GPU: commit  │  │ GPU: commit  │  │ GPU: commit  │
  │ seg 0 memory │  │ seg 1 memory │  │ seg 2 memory │
  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           ▼
              ┌─────────────────────────┐
              │  BARRIER                │
              │  Derive shared          │
              │  ExternalChallenges     │
              │  from all commitments   │
              └────────────┬────────────┘
                           │
Phase 3: Proving (parallel, independent, same challenges)
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
  │ GPU: prove   │  │ GPU: prove   │  │ GPU: prove   │
  │ segment 0    │  │ segment 1    │  │ segment 2    │
  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
         │                 │                 │
         └─────────────────┼─────────────────┘
                           ▼
              ProgramProof { base_layer_proofs[], delegation_proofs[] }
```

**Source files:** `gpu_prover/src/execution/prover.rs`, `gpu_prover/src/execution/gpu_manager.rs`, `gpu_prover/src/execution/gpu_worker.rs`, `gpu_prover/src/execution/messages.rs`, `worker/src/lib.rs`.

### 11.3 RISC Zero: Actor-Based Task Queue (r0vm)

RISC Zero's production prover (`r0vm`) uses an **actor-based architecture** with a central Factory actor that dispatches tasks to a pool of Worker actors. Segments never communicate directly — all coordination flows through the Factory.

**Architecture overview:**

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        RISC ZERO PROVER ORCHESTRATION                       │
└─────────────────────────────────────────────────────────────────────────────┘

                        ┌─────────────────────┐
                        │   ProofJob Actor    │
                        │   (per proof job)   │
                        │   - tracks segments │
                        │   - manages joins   │
                        └──────────┬──────────┘
                                   │ submit_task()
                                   ▼
                        ┌─────────────────────┐
                        │   Factory Actor     │
                        │   - task queue      │
                        │   - resource mgmt   │
                        │   - worker dispatch │
                        └──────────┬──────────┘
                                   │
            ┌──────────────────────┼──────────────────────┐
            ▼                      ▼                      ▼
   ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
   │  Worker Actor 0  │   │  Worker Actor 1  │   │  Worker Actor N  │
   │  ┌───────────┐  │   │  ┌───────────┐  │   │  ┌───────────┐  │
   │  │ CPU Queue │  │   │  │ CPU Queue │  │   │  │ CPU Queue │  │
   │  │ (execute, │  │   │  │ (execute, │  │   │  │ (execute, │  │
   │  │ preflight)│  │   │  │ preflight)│  │   │  │ preflight)│  │
   │  └───────────┘  │   │  └───────────┘  │   │  └───────────┘  │
   │  ┌───────────┐  │   │  ┌───────────┐  │   │  ┌───────────┐  │
   │  │ GPU Queue │  │   │  │ GPU Queue │  │   │  │ GPU Queue │  │
   │  │ (prove,   │  │   │  │ (prove,   │  │   │  │ (prove,   │  │
   │  │ lift,join)│  │   │  │ lift,join)│  │   │  │ lift,join)│  │
   │  └───────────┘  │   │  └───────────┘  │   │  └───────────┘  │
   └─────────────────┘   └─────────────────┘   └─────────────────┘
```

**Task flow for a proof job:**

1. **Execute** → Executor runs the program, emits segments as they complete
2. **ProveSegment** → For each segment:
   - CPU: `Preflight` (witness generation)
   - GPU: `ProveSegmentCore` (STARK proof)
3. **Lift** → Convert each SegmentReceipt to SuccinctReceipt (one per segment, parallel)
4. **Join** → Pairwise binary tree aggregation (as receipts arrive)
5. **Resolve** → Discharge assumptions (if any)
6. **ShrinkWrap** → Final Groth16 wrapper

**Key properties:**

- **No synchronization barriers:** Segments are proved independently. Each segment gets its own Fiat-Shamir transcript seeded from the (fixed) verification key.
- **Pipelined joins:** The `ProofJob` actor maintains a `joins: HashMap<Range, SuccinctReceipt>` and calls `maybe_join()` after each lift/join completes. Joins happen as soon as adjacent pairs are ready — no need to wait for all segments.
- **Resource-aware scheduling:** The Factory tracks CPU cores and GPU tokens per task type. A po2=20 segment proof requires more GPU tokens than po2=12.

**Contrast with simple sequential prover:**

The library's `ProverImpl::prove_session()` iterates segments sequentially:

```rust
for segment in session.segments() {
    segments.push(self.prove_segment(ctx, &segment)?);
}
```

This is simpler but doesn't parallelize. The `r0vm` actor system is the production-grade parallel implementation.

**Source files:** `risc0/r0vm/src/actors/factory.rs`, `risc0/r0vm/src/actors/worker.rs`, `risc0/r0vm/src/actors/job/proof.rs`, `risc0/r0vm/src/actors/protocol.rs`.

### 11.4 Linea Limitless: single prover, streaming conglomeration

The Limitless path does **not** use a distributed cluster of segment provers talking to each other. One Go process runs `RunDistributedPipeline` (`prover/backend/execution/limitless/prove.go`): bootstrapper → parallel GL jobs → shared randomness → parallel LPP jobs → hierarchical conglomeration, then the outer gnark circuit.

**Phase 1 — Bootstrapper (sequential):** Builds/loads the witness layout and segment blueprints; loads the verification-key Merkle tree root used across segments.

**Phase 2 — GL proofs (parallel, bounded concurrency):** `errgroup` with `numConcurrentSubProverJobs` (default 4). Each job runs `RunGL` for one GL witness index, stores the `SegmentProof` for randomness, and **sends the same proof** on a buffered `proofStream` channel to the background conglomeration goroutine (so GL proofs are merged into the tree as they complete). Job order is **longest-first** (reverse index) to reduce tail idle.

**Barrier 1 — Shared randomness:** After **all** GL jobs succeed, `GetSharedRandomnessFromSegmentProofs` runs **in the main goroutine** (see §9.3). Only then do LPP jobs start, each receiving the same `field.Octuplet` challenges.

**Phase 3 — LPP proofs (parallel, same pool size):** Same pattern as GL: `RunLPP(cfg, i, sharedRandomness, …)`; results go to `proofStream`. Optional `runtime.GC` / `FreeOSMemory` between phases to cap heap use.

**Phase 4 — Conglomeration (background consumer):** A goroutine loads the compiled conglomeration mmap, runs `RunConglomerationHierarchical` reading from `proofStream` until it receives `numGL + numLPP` proofs, then the main flow **closes** the channel and waits for the final root `SegmentProof`.

**Inter-segment “communication” in-circuit:** **Vertical** adjacency within a module uses `SentValuesGlobal` / `ReceivedValuesGlobal` and `IsFirst` / `IsLast` (`prover/protocol/distributed/module_gl.go`). **Cross-module** consistency is closed in the outer execution-limitless circuit (log-derivative sum, grand product, Horner sum), not by prover messages.

**Source files:** `prover/backend/execution/limitless/prove.go`, `prover/protocol/distributed/distribute.go`, `prover/zkevm/limitless.go`.

### 11.5 ZisK: Coordinator/Worker with Three-Phase Distributed Pipeline

ZisK supports both **single-process** and **distributed** proving. The single-process path uses Proofman directly; the distributed path uses a Coordinator/Worker architecture over gRPC, with optional MPI for multi-process proving within a single worker.

**Architecture overview:**

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                        ZisK PROVER ORCHESTRATION                             │
└──────────────────────────────────────────────────────────────────────────────┘

Single-process path:

  ┌────────────────────┐
  │  Executor           │
  │  (runs RISC-V       │
  │   program, plans    │
  │   SM instances)     │
  └──────┬─────────────┘
         │ chunks (2¹⁸ steps each)
         ▼
  ┌────────────────────┐
  │ Proofman            │
  │ (witness gen per    │
  │  chunk, parallel)   │
  │ (prove per SM       │
  │  instance)          │
  │ (aggregate          │
  │  recursively)       │
  └──────┬─────────────┘
         │ Vadcop Final proof
         ▼
  ┌────────────────────┐
  │ SNARK wrapper       │
  │ (Plonk / FFLONK)   │
  └────────────────────┘

Distributed path:

  ┌─────────────────────┐
  │   Coordinator        │     Client submits prove request
  │   (gRPC server)      │     with inputs + compute capacity
  │   - assigns workers  │
  │   - orchestrates     │
  │     3 phases         │
  └──────────┬──────────┘
             │
  ┌──────────┼──────────────────────┐
  ▼          ▼                      ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ Worker 0     │  │ Worker 1     │  │ Worker N     │
│ ELF + ROM    │  │ ELF + ROM    │  │ ELF + ROM    │
│ Proving key  │  │ Proving key  │  │ Proving key  │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
Phase 1: Partial Contributions (parallel, independent)
       │                 │                 │
       └─────────────────┼─────────────────┘
                         ▼
              ┌─────────────────────────┐
              │  BARRIER                │
              │  Derive global          │
              │  challenge              │
              └────────────┬────────────┘
                           │
Phase 2: Proving (parallel, same challenge)
       ┌─────────────────┼─────────────────┐
       ▼                 ▼                 ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ Worker 0     │  │ Worker 1     │  │ Worker 2     │
│ partial proof│  │ partial proof│  │ partial proof│
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       └─────────────────┼─────────────────┘
                         ▼
Phase 3: Aggregation (one worker selected)
              ┌─────────────────────────┐
              │  Aggregator worker      │
              │  collects all partial   │
              │  proofs → final proof   │
              └────────────┬────────────┘
                           │
                           ▼
              Vadcop Final / SNARK proof
```

**Key properties:**

- **Workers report compute capacity** when they register. The coordinator assigns workers sequentially until the requested capacity is met.
- **Workers do not communicate with each other** — all coordination flows through the coordinator via gRPC.
- **Three phases:** (1) partial contributions, (2) prove (after global challenge), (3) aggregation. The first worker to finish phase 2 is selected as the aggregator.
- **MPI support:** Within a single process, MPI can distribute work across multiple ranks for multi-GPU/multi-node proving (`sdk/src/prover/mod.rs` — `mpi_broadcast`, `RankInfo`).
- **GPU support:** Proving can use GPU acceleration (`--features gpu`), with configurable GPU streams (`--max-streams`) and preallocation (`--preallocate`).
- **Timeouts:** Phase 1 default 300s, Phase 2 default 600s (configurable).

**Source files:** `distributed/`, `book/getting_started/distributed_execution.md`, `sdk/src/prover/backend.rs`, `sdk/src/prover/mod.rs`.

---

## 12. Trade-offs

| | RISC Zero | SP1 | Linea | Nexus | ZKsync Airbender | ZisK |
|---|---|---|---|---|---|---|
| **Segmentation complexity** | Low | Moderate | High | None | Low–moderate | Low–moderate |
| **Constraint specialization** | Low (one circuit for everything) | Medium (per-chip AIR) | High (per-module arithmetization) | Medium (main chips + extensions) | Medium (muxed main table + separate delegation AIRs) | Medium–high (one PIL airgroup, but each secondary SM has specialized constraints) |
| **Per-segment overhead** | High (table replication: 4,112+ rows) | Low (no table replication) | Low (no table replication) | N/A | To be investigated | Virtual table grouping in PIL; replication overhead to be investigated |
| **Parallelism** | Fully parallel | Fully parallel | Two-phase | Single-proof | Fully parallel (segments) | Chunk witness parallel; distributed proving with ≥1 barrier |
| **Aggregation tree depth** | Moderate (pairwise join) | Higher (many small shards) | Lower (fewer large segments) | N/A | Multiple recursion layers; exact depth to be investigated | Recursive aggregation via Proofman; depth to be investigated |
| **Final proof** | Groth16 (~260 bytes) | Groth16 | gnark Groth16 | Stwo StarkProof | FFLONK | Plonk or FFLONK (snarkJS-generated PlonkVerifier) |

### Why each approach fits its VM

- **RISC Zero:** RISC-V has a flat memory model → memory state is compactly committable via Merkle root → lookups stay local → simple segmentation. The single-table design means segment shape depends only on po2, not on which operations ran → only 13 lift programs needed.

- **SP1:** RISC-V with separate precompile chips → temporal segmentation keeps chip lookups local by construction → only memory spans shards → EC multiset hash handles memory without randomness → no cross-shard barriers.

- **Linea:** EVM with specialized modules per operation → cross-module lookups are intrinsic → large conflations overflow single-proof capacity → structural segmentation splits those lookups → shared randomness is necessary to close the log-derivative argument globally.

- **Nexus:** RISC-V with multi-chip main AIR + optional extensions → everything in one Stwo proof if it fits memory → no segmentation overhead, but no fallback for oversized traces. Modular per-component proving with LogUp-based cross-component consistency and recursive composition are planned for zkVM 4.0 (mid-2026). Sources: [zkVM 3.0 and Beyond](https://blog.nexus.xyz/zkvm-3-0-and-beyond-toward-modular-distributed-zero-knowledge-proofs/), [The Nexus Roadmap](https://blog.nexus.xyz/the-nexus-roadmap/).

- **ZKsync Airbender:** Proving layer for ZKsync OS (same RISC-V binary as the sequencer). Long runs (~2³⁶ cycle ceiling) → temporal segmentation. EVM-grade crypto → delegation circuits with unified RAM model. Cheap on-chain verification → STARK/FRI + FFLONK wrapper. Trusted-code model (no traps) keeps the main AIR smaller. Targets consumer GPUs.

- **ZisK:** Polygon's general-purpose zkVM. RISC-V rv64ima → transpiled to custom 64-bit Zisk ISA for compact trace representation. PIL-defined constraints allow a rich hierarchy of specialized secondary state machines (20+ SMs) while keeping them in a single airgroup for tight integration. Temporal segmentation (~2¹⁸ step chunks, up to ~2³⁶ total steps) enables parallel witness generation. Goldilocks field + Proofman (Vadcop) for intermediate proofs → Plonk/FFLONK SNARK wrapper for on-chain verification. Distributed Coordinator/Worker architecture for production-scale proving across multiple machines with GPU support.

---

## 13. Information Gaps

The following items would make this document fully self-contained and evidence-based.

### Across the board

1. **Performance benchmarks.** No proving times, memory usage, or proof sizes are compared across all five systems on comparable hardware.

2. **Proof sizes at each pipeline stage.** RISC Zero's final Groth16 is ~260 bytes, but intermediate sizes and other systems' sizes are not stated.

3. **Version pinning.** No specific commit or release tag pinned for any system.

### SP1

4. ~~**Proof system details.**~~ **RESOLVED** — See Section 5.2.1. Field is KoalaBear (not BabyBear); PCS is Jagged → Stacked Basefold; hash is Poseidon2 over KoalaBear; security target 100 bits.

5. ~~**Jagged PCS.**~~ **RESOLVED** — See Section 5.2.2. Variable-sized table commitment via sumcheck + jagged indicator polynomial + stacked Basefold.

6. ~~**SepticDigest curve details.**~~ **RESOLVED** — See Section 5.2.3. Curve y² = x³ + 45x + 41z³ over F_{p^7}; Poseidon2-based trial-and-increment hashing; send/receive sign convention.

7. ~~**Prover orchestration.**~~ **RESOLVED** — See Section 11.1. Controller/Worker async architecture with per-task-type channels, CompressTree for pipelined aggregation, no synchronization barriers.

### Linea Limitless

8. ~~**"GL" and "LPP" definitions.**~~ **RESOLVED** — See Section 5.3 and 6.3.

9. ~~**Codebase evidence.**~~ **PARTIALLY RESOLVED** — See Sections 5.3, 5.4, and 6.3.

10. ~~**Conflation sizing.**~~ **PARTIALLY RESOLVED (linea-monorepo):** The **coordinator** decides conflation boundaries (not a fixed constant in the prover). See `coordinator/.../ConflationConfig.kt`: optional `blocksLimit`, `conflationDeadline`, blob `blobSizeLimit` / `batchesLimit`, `tracesLimits`, aggregation `proofsLimit`, etc.; calculators such as `ConflationCalculatorByBlockLimit` enforce an upper bound on blocks per conflation when that trigger is used. The **prover** consumes a conflated execution request/trace; the outer execution circuit caps serialized execution metadata: `ExecDataBytes` is `1 << 17` bytes and `assign` panics if `len(execData)` exceeds that (`prover/circuits/execution/circuit.go`). Per-column **trace row caps** per module are in `prover/config/config-*-limitless.toml` under `[traces_limits]` (and `prover/config/traces_limit.go`). If a single module still exceeds one Plonk domain after segmentation, compilation uses a global `FixedNbRowPlonkCircuit` (e.g. `1 << 25` for HUB-A GL) in `prover/zkevm/limitless.go` — environment-specific.

11. ~~**Memory consistency mechanism.**~~ **PARTIALLY RESOLVED (linea-monorepo):** EVM memory is modeled in the arithmetization (e.g. **HUB-B** tables: `mxp.*`, `mmio.*`, `mmu.*`, `gas.*`) with permutation / lookup machinery inside the Wizard-IOP (LPP phase), not with RISC Zero–style Merkle RAM or SP1’s cross-shard `SepticDigest`. **Across vertical segments** of the same module, boundary-sensitive values are chained via `SentValuesGlobal` / `ReceivedValuesGlobal` for global constraints. **Across horizontal modules**, consistency is enforced globally when partial log-derivative sums, grand products, and Horner sums are aggregated to the values checked in the outer circuit (§5.4, §8). A full constraint-level walkthrough of MMU/MMIO would still add detail.

### Nexus  

12. **Maximum trace size.** Practical limit on trace size for a single Stwo proof.

13. **KeccakChip constraint completeness.** Minimal constraints with a TODO — production-ready?

14. **RAM/register consistency details.** `RamInitFinal` and `RegisterMemCheckChip` not explained.

### ZKsync Airbender

15. **Cross-segment proof linkage.** Precisely how adjacent segment STARKs are chained in the verifier (public inputs, hashing, recursion constraints) — requires reading recursion verifier circuits in detail.

16. **On-chain proof format.** ~~Unspecified~~ **Partially resolved:** FFLONK per [ZKsync Airbender Overview](https://docs.zksync.io/zk-stack/components/zksync-airbender). Still missing: proof sizes, curve, and aggregation policy.

17. **Production VK inventory.** Count of distinct machine setups and how they map to deployed verifiers.

18. **Delegation coverage.** Which delegation types are enabled for ZKsync OS production vs. tests.

### ZisK

19. **Cross-chunk memory linkage.** How the memory proxy state machines ensure consistency across chunk boundaries — the specific mechanism (Merkle root chaining, EC hash, shuffle argument, or something else) is not documented in the codebase.

20. **Proofman internals (PCS, hash, security level).** ZisK delegates proving to the external `pil2-proofman` library. The polynomial commitment scheme, hash function, FRI parameters (if any), and security target are internal to Proofman and not surfaced in the ZisK codebase.

21. **Shared randomness / synchronization barriers.** Whether the Proofman pipeline requires shared randomness across chunks (like Linea's GL→LPP barrier or Airbender's ExternalChallenges) or uses per-chunk Fiat-Shamir (like RISC Zero/SP1).

22. **Shape normalization / verifier key inventory.** How many distinct verifier key configurations exist and how variable SM instance sizes are normalized for the recursive aggregation circuit.

23. **Per-chunk overhead.** Whether lookup tables (dual_range, arith_table, binary_table, etc.) are replicated per chunk or shared, and the fixed cost per chunk.

24. **Proof sizes at each pipeline stage.** Vadcop Final proof size, compressed proof size, and final SNARK proof size are not documented.

25. **Prover orchestration internals (single-process).** The Proofman's internal witness-generation and proving parallelism model (thread pool, task graph, memory management) is external to the ZisK codebase.

26. **Production readiness.** The README explicitly states the project is "not production-ready" and "not fully tested."

---

## 14. SP1 LogUp Bus Architecture (Deep Dive)

This section provides a detailed look at SP1's LogUp bus mechanism for handling intra-proof relations, particularly memory checking and lookup consistency among shards.

### 14.1 Overview

SP1 uses **LogUp-GKR** (a GKR-based LogUp argument) to prove that all chip interactions within a shard balance correctly. The "bus" is not a single data structure but rather **the multiset of all `InteractionKind`-tagged messages**, proved globally for one shard by one LogUp-GKR proof attached to that `ShardProof`.

**Key insight:** Each shard has its own `logup_gkr_proof` — LogUp-GKR is **per-shard**, not one proof over all shards. Cross-shard consistency is achieved through **public values** (especially `global_cumulative_sum`) that the outer verifier composes.

### 14.2 Interaction Kinds (Bus Types)

Each interaction is tagged with an `InteractionKind` that partitions them into logical "buses":

| Kind | Purpose | Values |
|------|---------|--------|
| `Memory` | RAM read/write multiset | 9 values (timestamps, addr, value) |
| `Program` | Instruction fetch table | 16 values |
| `Byte` | Byte lookup operations | 4 values |
| `State` | CPU state interactions | 5 values |
| `Syscall` | Syscall handling | 9 values |
| `Global` | Global message digest | 11 values |
| `GlobalAccumulation` | Digest chaining | 15 values |
| `MemoryGlobalInitControl` | Memory init bookkeeping | 5 values |
| `MemoryGlobalFinalizeControl` | Memory finalize bookkeeping | 5 values |
| `InstructionFetch` | Instruction fetch table | 22 values |
| `InstructionDecode` | Instruction decode table | 19 values |
| `ShaExtend`, `ShaCompress`, `Keccak` | Precompile chips | varies |
| `PageProt*` | Page protection | varies |

**Source:** `crates/hypercube/src/lookup/interaction.rs`

### 14.3 Interaction Structure

An `Interaction<F>` contains:

```rust
pub struct Interaction<F: Field> {
    pub values: Vec<VirtualPairCol<F>>,      // The interaction payload
    pub multiplicity: VirtualPairCol<F>,     // +1 for send, -1 for receive
    pub kind: InteractionKind,               // Bus type tag
    pub scope: InteractionScope,             // Local or Global
}
```

### 14.4 Fingerprint Computation

For each interaction, the LogUp denominator (fingerprint) is computed as:

```
fingerprint = α + β₀·argument_index(kind) + Σⱼ βⱼ·valueⱼ
```

Where:
- `α` and `betas` are random challenges
- `argument_index(kind)` distinguishes different interaction kinds
- Sends contribute `+multiplicity`, receives contribute `-multiplicity`

**Source:** `Interaction::eval()` in `crates/hypercube/src/lookup/interaction.rs`

### 14.5 Chip Send/Receive Collection

When a `Chip` is constructed, it runs a symbolic evaluation of the AIR to collect all interactions:

```rust
impl<F, A> Chip<F, A> {
    pub fn new(air: A) -> Self {
        let mut builder = InteractionBuilder::new(...);
        air.eval(&mut builder);  // Symbolic evaluation
        let (sends, receives) = builder.interactions();
        // ...
    }
}
```

Each chip stores its `sends` and `receives` as `Arc<Vec<Interaction<F>>>`.

**Source:** `crates/hypercube/src/chip.rs`

### 14.6 Memory Checking (Intra-Shard)

The `MemoryAirBuilder` trait provides methods for memory read/write that use the `Memory` interaction kind:

**For a memory read:**
1. **Send** the previous tuple: `(prev_timestamp, addr, prev_value)` with multiplicity = 1
2. **Receive** the current tuple: `(current_timestamp, addr, prev_value)` with multiplicity = -1

**For a memory write:**
1. **Send** the previous tuple: `(prev_timestamp, addr, prev_value)` with multiplicity = 1
2. **Receive** the current tuple: `(current_timestamp, addr, new_value)` with multiplicity = -1

This creates a **permutation argument**: every memory state transition is "consumed" (previous state sent) and "produced" (new state received). If all transitions are valid, the multiset of sends equals the multiset of receives.

**Timestamp ordering** is enforced by range-checking `current_timestamp - prev_timestamp - 1 ∈ [0, 2²⁴)`.

**Source:** `crates/core/machine/src/air/memory.rs`

### 14.7 LogUp-GKR Proving Flow

1. **Symbolic extraction:** Each `MachineAir`'s `eval` calls `send`/`receive` on a builder. `InteractionBuilder` turns those into `Interaction` lists stored in `Chip`.

2. **First GKR layer:** For each trace row and each interaction of every included chip, compute `(numerator, denominator)` pairs:
   - `numerator = multiplicity`
   - `denominator = fingerprint(α, β, kind, values)`

3. **GKR + sumcheck:** Prove that the rational sum `Σ (numerator/denominator)` equals zero (or the expected value) using GKR with sumcheck rounds.

4. **Tie to trace openings:** The prover produces `LogUpEvaluations` — chip trace openings at a derived point — so the zerocheck can be coupled with the GKR batch opening.

5. **Public values hook:** The verifier's `verify_public_values` uses `Record::eval_public_values` to fold public values into the same LogUp balance via synthetic sends/receives.

**Source:** `crates/hypercube/src/logup_gkr/prover.rs`, `verifier.rs`, `execution.rs`

### 14.8 Cross-Shard Consistency

Shards are linked **not** by a single shared LogUp-GKR, but by **public values**:

#### 14.8.1 Global Cumulative Sum

The `GlobalChip` receives `InteractionKind::Global` messages and maintains a **septic curve digest** (`SepticDigest`). The `GlobalAccumulation` interaction kind chains these digests row-by-row.

The machine verifier **adds** each shard's `public_values.global_cumulative_sum` to `vk.initial_global_cumulative_sum` and requires the **sum to be zero** (in the septic curve group):

```rust
// In verify.rs
for shard_proof in shard_proofs {
    cumulative_sum += shard_proof.public_values.global_cumulative_sum;
}
assert!(cumulative_sum + vk.initial_global_cumulative_sum == SepticDigest::zero());
```

#### 14.8.2 Memory Init/Finalize Ranges

`PublicValues` carries:
- `previous_*` / `last_*` addresses
- Init/finalize counts

The `eval_global_memory_init` and `eval_global_memory_finalize` functions encode these into `MemoryGlobalInitControl` and `MemoryGlobalFinalizeControl` interactions, ensuring that:
- Shard N's final memory state matches Shard N+1's initial state
- The overall init/finalize ranges are consistent

**Source:** `crates/core/executor/src/record.rs`, `crates/prover/src/verify.rs`

### 14.9 Key Files Reference

| File | Role |
|------|------|
| `crates/hypercube/src/lookup/interaction.rs` | `Interaction`, `InteractionKind`, fingerprint eval |
| `crates/hypercube/src/lookup/builder.rs` | `InteractionBuilder`: records send/receive from AIR eval |
| `crates/hypercube/src/chip.rs` | `Chip { air, sends, receives }` |
| `crates/hypercube/src/logup_gkr/prover.rs` | LogUp-GKR prover |
| `crates/hypercube/src/logup_gkr/verifier.rs` | LogUp-GKR verifier |
| `crates/hypercube/src/logup_gkr/execution.rs` | `generate_interaction_vals` |
| `crates/core/machine/src/air/memory.rs` | `MemoryAirBuilder` trait |
| `crates/core/machine/src/memory/local.rs` | `MemoryLocalChip` |
| `crates/core/machine/src/memory/global.rs` | `MemoryGlobalInit/Finalize` |
| `crates/core/machine/src/global/mod.rs` | `GlobalChip` (septic digest accumulation) |
| `crates/core/executor/src/record.rs` | `eval_public_values`, `eval_global_sum` |
| `crates/prover/src/verify.rs` | Cross-shard cumulative sum verification |
| `sp1-gpu/crates/logup_gkr/` | GPU LogUp-GKR implementation |

### 14.10 Mental Model Summary

```
┌─────────────────────────────────────────────────────────────────┐
│                           SHARD N                                │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐            │
│  │  CPU    │  │ Memory  │  │  Byte   │  │ Global  │  ...       │
│  │  Chip   │  │ Local   │  │  Chip   │  │  Chip   │            │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘            │
│       │            │            │            │                   │
│       ▼            ▼            ▼            ▼                   │
│  ┌──────────────────────────────────────────────────┐           │
│  │         LogUp-GKR (per-shard proof)              │           │
│  │  • Memory bus: read/write multiset balance       │           │
│  │  • Byte bus: lookup table balance                │           │
│  │  • Global bus: septic digest accumulation        │           │
│  │  • ...all interaction kinds must balance         │           │
│  └──────────────────────────────────────────────────┘           │
│                          │                                       │
│                          ▼                                       │
│              public_values.global_cumulative_sum                 │
└─────────────────────────────┬───────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    OUTER VERIFIER                                │
│  Σ (shard.global_cumulative_sum) + vk.initial = 0               │
│  Memory init/finalize ranges chain correctly                     │
└─────────────────────────────────────────────────────────────────┘
```

**Key takeaways:**
1. LogUp-GKR proves **intra-shard** interaction balance (memory, lookups, etc.)
2. The `SepticDigest` provides a **succinct commitment** to global messages
3. Cross-shard consistency uses **public values** checked by the outer verifier
4. No cross-shard LogUp barriers — shards are independently provable
