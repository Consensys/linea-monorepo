package smt

import "github.com/consensys/accelerated-crypto-monorepo/utils"

/*
Overwrite a leaf in the tree and update the tree in consequence
*/
func (t *Tree) Update(pos int, newVal Digest) {
	depth := t.Config.Depth
	current := newVal
	idx := pos

	if pos >= 1<<depth {
		utils.Panic("out of bound %v", pos)
	}

	for level := 0; level < t.Config.Depth; level++ {
		// store the newly computed node
		t.updateNode(level, idx, current)
		sibling := t.getNode(level, idx^1) // xor 1, switch the last bits
		left, right := current, sibling
		if idx&1 == 1 {
			left, right = right, left
		}
		current = HashLR(t.Config, left, right)
		idx >>= 1
	}

	// Sanity-check: the idx should be zero
	if idx != 0 {
		panic("idx should be zero")
	}

	t.Root = current
}
