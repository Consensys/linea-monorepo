# Roadmap

## Specifying public inputs in the wizard

* [ ] Specifying a commitment as a public input in the wizard
* [ ] Passing a lot of public inputs in the outer-proof system
	* Nikolasgg strategy https://hackmd.io/@nikkolasg/S1I11Zr2w
	* KoE + Polynomial commitment opening at a random point
	* (Research) More optimized protocol
	* Pass a Keccak hash of the data
	* Pass a SNARK-Friendly hash
	* Plonk
* [ ] Public input arrangement for the outer-proof system

## Self-recursion

* [ ] Implement the Horner trick
* [ ] Framework evolution for enabling the self-recursion
* [ ] Proving the erasure-code
* [ ] Proving SIS
* [ ] Proving ring-SIS
* [ ] Proving the linear combination
* [ ] Proving the column selection

## Vortex improvement

* [ ] Use a better erasure code
* [ ] Optimized ring-SIS implementation

## Commitment splitting

* [ ] Optimize memory consumption for global constraints and MPTS
* [ ] Changes in the wizard framework
* [ ] Translation of the special queries constraints
* [ ] Translation of the global constraints
* [ ] Vortex : commit to the Lagrange representation
* [ ] Add possibility to specify at runtime the practical size of a commitment
* [ ] Optimize Vortex for zero-prepended columns
* [ ] Optimize Global constraints for zero-columns

## Switch to Goldilock

* [ ] Ring-SIS with Goldilocks
* [ ] Repeat random linear combinations in Arcane
* [ ] Repeat random linear combinations in Vortex
* [ ] Non-native field arithmetic in the outer-proof system

## Optimized Keccak

* [ ] (Research) Bit-slicing Wizard-IOP approach
* [ ] (Research) Gather some litterature
* [ ] (Temporary?) Integrate as a Plonk component
* [ ] Framework update for Keccak 

## Optimized ECDSA verification

* [ ] (Research) Technical solution
* [ ] (Research) Changes in the framework

## Merkle-tree verification in wizard

* [ ] Design the changes in the framework
* [ ] Technical solution for verifying many Keccak
* [ ] Sequencing the hash verification
* [ ] Choose the hash function

## Integrating Plonk component

* [ ] PoC convert a Plonk circuit into a Wizard-system
* [ ] Framework update in the Wizard
* [ ] Optimization for repeatedly calling the circuit
* [ ] Framework arrangement for the precompiles

## General implementation improvement

* [ ] CI/CD
* [ ] Githooks
* [ ] Proving keys on S3
* [ ] Actual backend for the prover
* [ ] General all purpose Memory Pool
* [ ] Serialization format for the incoming proofs
* [ ] Testcase automation maintainance

## General Arcane compiler improvement 

* Range checks
	* [ ] (Research) Better subprotocol for range-check
	* [ ] [BCGM18](https://eprint.iacr.org/2018/380.pdf) for range-check
* Lookup
	* [ ] Optimized implementation
	* [ ] Lookups in parallelization
	* [ ] (Research) Structured lookup-table
	* [ ] (Research) Alternative design for lookup-table
* Permutation argument
	* [ ] Optimized implementation
	* [ ] Parallelization
* Local constraint
	* [ ] Use evaluation at a fixed point
* Global constraint
	* [ ] Optimize for the degree of the constraint
	* [ ] Tune for memory consumption
* MPTS
	* [ ] Optimize the parallelization
	* [ ] Optimize the memory usage