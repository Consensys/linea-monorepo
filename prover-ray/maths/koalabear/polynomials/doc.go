// Package polynomials provides native polynomial evaluation utilities over the
// KoalaBear field and its degree-4 extension, using the union types
// [field.Vec] and [field.Gen] for type-aware dispatch.
//
// # Univariate polynomials
//
// Two evaluation bases are supported:
//   - Canonical (coefficient) form: P(X) = Σᵢ p[i]·Xⁱ  (see [EvalCanonical])
//   - Lagrange (evaluation) form: P(X) given by its values P(ωⁱ) at roots of unity (see [EvalLagrange])
//
// # Multilinear polynomials
//
// A multilinear polynomial in n variables is stored as a length-2ⁿ slice of
// evaluations on the boolean hypercube {0,1}ⁿ. The index x ∈ [0, 2ⁿ) encodes
// the evaluation point (x₀, x₁, …, x_{n-1}) with x₀ being the most-significant
// bit, matching the convention used by [FoldInto] and [EvalMultilin].
//
// Core operations:
//   - [FoldInto]: fold a table on its first variable (single step of evaluation)
//   - [EvalMultilin]: evaluate a multilinear polynomial at an arbitrary point
//
// Equality polynomial:
//   - [EvalEq], [EvalEqBase], [EvalEqExt]: compute Π Eq(qᵢ, hᵢ) where Eq(x,y) = 1−x−y+2xy
//   - [FoldedEqTableBase], [FoldedEqTableExt]: build the dense Eq(coords, ·) table
//   - [ChunkOfEqTableBase], [ChunkOfEqTableExt]: build a chunk of the Eq table for parallelism
//
// All functions support base-field ([field.Element]), extension-field ([field.Ext]), and
// mixed inputs via the union type [field.Gen] / [field.Vec]; the result type tracks
// whether computation stayed in the base field.
package polynomials
