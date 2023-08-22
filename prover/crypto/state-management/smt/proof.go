package smt

import (
	"fmt"

	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
)

// ProvedClaim is the composition of a proof with the claim it proves.
type ProvedClaim struct {
	Proof      Proof
	Root, Leaf Digest
}

// Merkle proof of membership for the Merkle-tree
type Proof struct {
	Path     int                // Position of the leaf
	Siblings []hashtypes.Digest // length 40
}

// Update a leaf in the tree
func (t *Tree) Prove(pos int) Proof {
	depth := t.Config.Depth
	siblings := make([]hashtypes.Digest, depth)
	idx := pos

	if pos >= 1<<depth {
		utils.Panic("pos %v is larger than the tree width %v", pos, 1<<depth)
	}

	for level := 0; level < depth; level++ {
		sibling := t.getNode(level, idx^1) // xor 1, switch the last bits
		siblings[level] = sibling
		idx >>= 1 // erase the last bit
	}

	// Sanity-check: the idx should be zero
	if idx != 0 {
		panic("idx should be zero")
	}

	return Proof{
		Siblings: siblings,
		Path:     pos,
	}
}

// Returns the root as an output of the Merkle verification
func (p *Proof) RecoverRoot(conf *Config, leaf hashtypes.Digest) hashtypes.Digest {
	current := leaf
	idx := p.Path

	for _, sibling := range p.Siblings {
		left, right := current, sibling
		if idx&1 == 1 {
			left, right = right, left
		}
		current = HashLR(conf, left, right)
		idx >>= 1
	}

	// Sanity-check: the idx should be zero
	if idx != 0 {
		panic("idx should be zero")
	}

	return current
}

// Verify the Merkle-proof against a hash and a root
func (p *Proof) Verify(conf *Config, leaf, root hashtypes.Digest) bool {
	actual := p.RecoverRoot(conf, leaf)
	return actual == root
}

// Pretty-print a proof
func (p *Proof) String() string {
	return fmt.Sprintf("&smt.Proof{Path: %d, Siblings: %x}", p.Path, p.Siblings)
}
