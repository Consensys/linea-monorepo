// The package symbolic provides a tool to generate and manipulate symbolic
// expressions. A symbolic expression is a directed-acyclic-graph (DAG)
// representating a succession of arithmetic operations (addition,
// multiplication, squaring, negation, subtraction etc...) over symbolic
// variables or constants. These expressions can in turn be optimized and
// evaluated vector-wise by providing an assignment.
package symbolic

// # High-level overview — what this `symbolic` package does

// Short version: the package builds, optimizes, and evaluates **symbolic polynomial expressions** (DAGs) over a finite
// field. You construct expressions from variables/constants and arithmetic operators (linear combinations, products,
// polynomial evaluations). Expressions are normalized with an **expression-sensitive hash (ESHash)** so structurally-different
// but algebraically-equivalent nodes can be deduplicated. To evaluate efficiently the package converts an expression into
// a **board** (topologically-sorted DAG with deduped nodes) and evaluates that board vector-wise in parallel and in fixed-size chunks.

// # Core types & concepts

// * `Expression`
//   Node in the symbolic DAG. Contains:

//   * `ESHash` — a `field.Element` that encodes *what* the expression computes (not how). Used for deduplication and regrouping.
//   * `Children` — operand sub-expressions.
//   * `Operator` — one of several operator implementations (see below).

// * `Operator` interface
//   Every operator implements:

//   * `Evaluate([]sv.SmartVector) sv.SmartVector` — vector evaluation helper used for computing ESHash and some tests.
//   * `Validate(e *Expression) error` — sanity checks.
//   * `Degree([]int) int` — polynomial degree from children's degrees.
//   * `GnarkEval(frontend.API, []frontend.Variable) frontend.Variable` — how to evaluate in a gnark circuit.

// * Provided operator implementations:

//   * `Constant` — stores a `field.Element`.
//   * `Variable` — wraps `Metadata` (anything implementing `Metadata.String()`); `metadataToESH` hashes the metadata string to get ESHash.
//   * `LinComb` — linear combination (addition/subtraction) represented by integer coefficients.
//   * `Product` — product with integer exponents.
//   * `PolyEval` — special operator to evaluate a polynomial with coefficients that are themselves expressions (Horner method implemented).

// * `Metadata` / `StringVar` / `NewDummyVar`
//   Used to represent variable identity. `metadataToESH` uses blake2b over the `String()` to produce the variable ESHash.

// ---

// # Building & simplifying expressions

// * Top-level builders: `Add`, `Mul`, `Sub`, `Neg`, `Square`, `Pow`, `NewPolyEval`, `intoExpr`, etc. These accept either `*Expression`, `Metadata`, or literal field-like values and convert inputs into expression nodes.

// * `NewLinComb` and `NewProduct` are the canonical constructors that apply multiple simplifications:

//   * Flattening (expand nested LCs/products).
//   * Regroup terms with same `ESHash` (`regroupTerms`).
//   * Pull out constant terms and fold them together.
//   * Remove zero coefficients / zero exponents (`removeZeroCoeffs`).
//   * Potentially collapse to a single child or to a constant (so constructor may return a `Constant` or `Variable` rather than a `LinComb`/`Product`).
//   * `expandTerms` expands nested products / linear combos when appropriate.

// * Important restrictions:

//   * No inversion (only polynomials).
//   * Negative exponents are disallowed.
//   * The constructors will panic on invalid input shapes (by design).

// ---

// # ESHash — expression-sensitive hash

// * Purpose: canonicalize expressions semantically so equal expressions (e.g., `a+a` and `2*a` when represented appropriately) can be detected/merged.
// * Rules used by operators: e.g., `ESHash(a + b) = ESHash(a) + ESHash(b)`, `ESHash(a * b) = ESHash(a) * ESHash(b)`, constants/variables set their ESHash from value or hashed metadata.
// * Used extensively for deduplication when building a board.

// ---

// # From Expression → Board (anchoring / topological layout)

// * `Expression.Board()` (via `anchor`) converts a tree-like `Expression` into an `ExpressionBoard`:

//   * Produces `Nodes [][]Node` where `Nodes[0]` are leaves (variables/constants), final node is the head/root.
//   * Maintains `EsHashesToPos map[field.Element]nodeID` to deduplicate nodes (if a node with the same ESHash already exists it reuses it).
//   * `nodeID` packs `(level << 32) | pos` so you can get level and pos cheaply.

// * Result: an acyclic, topologically-sorted, deduplicated DAG — ready for efficient evaluation and degree computations.

// ---

// # Evaluation (vectorized, chunked, parallel)

// * `ExpressionBoard.Evaluate(inputs []sv.SmartVector)` evaluates the board on provided inputs (smartvectors) and returns a `sv.SmartVector` result.

// * Key points:

//   * Input vectors must all have the same length `totalSize`.
//   * Evaluation runs chunk-by-chunk where `chunkSize = min(ChunkSize(), totalSize)`. ChunkSize is chosen based on available memory (via `debug.SetMemoryLimit(-1)`), to balance memory/parallelism.
//   * `parallel.Execute(numChunks, func(start, stop int) { ... })` runs chunks in parallel across chunk ranges.
//   * Within each chunk, an `evaluation` struct holds `evaluationNode`s, a scratch buffer, and a `vectorArena` used to allocate contiguous memory for intermediate vectors.
//   * Constants are handled specially (stored as single-element vectors) — many optimizations in `evalProduct`, `evalNode` and LinComb code — if nodes are constant the node result becomes constant and allocation can be avoided.
//   * Final result: either a `sv.Constant` (if all inputs are constants) or a `sv.Regular` vector.

// * Memory/perf helpers:

//   * `arena.VectorArena` is used to allocate contiguous slices into pre-allocated memory to avoid lots of small allocations during evaluation.
//   * `ChunkSize()` determines chunking based on memory limit to avoid OOM.

// ---

// # Degree, Gnark integration, and utilities

// * `ExpressionBoard.Degree(getdeg GetDegree)` computes polynomial degree across the board by dynamic programming on levels. `getdeg` maps `Variable` metadata to a degree (caller-provided).
// * `ExpressionBoard.GnarkEval(api, inputs)` performs same topological evaluation but using gnark `frontend.API` and `frontend.Variable`s so the symbolic expression can be realized in a zk-circuit.
// * Each operator implements `GnarkEval`, e.g.:

//   * `LinComb.GnarkEval` uses `api.Add`/`api.Mul`.
//   * `Product.GnarkEval` uses exponentiation helper `gnarkutil.Exp`.
//   * `PolyEval.GnarkEval` uses Horner method.

// ---

// # Traversal & transformation helpers

// * `Replay(translationMap)` — substitute variables according to a mapping, returning a new expression or original if nothing changed.
// * `ReconstructBottomUp` (parallel) & `ReconstructBottomUpSingleThreaded` (single threaded) — walk the expression bottom-up applying a constructor to rebuild/transform expressions (useful for optimizations or rewrites).
// * `SameWithNewChildren` — helper to rebuild a node with different children (preserving operator semantics).
// * `MarshalJSONString()` — debug helper to JSON pretty-print an expression.

// ---

// # Validation & safety

// * `Validate()` checks ESHash correctness (by computing operator.Evaluate on child ESHs), operator-specific validation, and recursively validates children. `AssertValid()` panics on invalid expression.
// * Many functions intentionally `panic` on misuse (invalid shapes, wrong counts, negative exponents) — the package is defensive and prefers early failures.

// ---

// # Typical lifecycle / example usage

// 1. Build expressions with `NewDummyVar`, `Add`, `Mul`, `NewConstant`, `NewPolyEval`, etc.
// 2. Optionally call `AssertValid()` to ensure the tree is sane.
// 3. Call `Board()` to get a deduplicated, topologically-sorted `ExpressionBoard`.
// 4. Use `ListVariableMetadata()` to determine the order of variable inputs.
// 5. Evaluate with `Evaluate(inputs ...)` for fast vector evaluation, or `GnarkEval` to build constraints in a circuit.
// 6. Use `ReconstructBottomUp` or constructors to simplify/transform expressions.

// ---

// # Limitations & notes

// * This system targets polynomial arithmetic only (no division/inversion).
// * Negative exponents are rejected.
// * Many operations panic on malformed inputs (intentional design).
// * ESHash-driven deduplication assumes consistent ESHash rules; changing ESHash semantics affects deduplication and simplification behavior.
// * Parallel reconstruction/evaluation introduces concurrency — some helpers use goroutines (e.g., `ReconstructBottomUp`) and arenas to optimize allocations.
