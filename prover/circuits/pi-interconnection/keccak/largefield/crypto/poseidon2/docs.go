// poseidon2 wraps the [github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2] and
// provides low-level utility methods to use the block compression of Poseidon2
// directly. This is used in several occasions in the wizard package to
// implement and test the arithmetization of the Poseidon2 block compression function
// and to have interfaces that are friendly to field elements.
package poseidon2
