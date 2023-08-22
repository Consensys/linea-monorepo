package smt

import (
	"github.com/consensys/accelerated-crypto-monorepo/crypto/state-management/hashtypes"
	"github.com/consensys/accelerated-crypto-monorepo/utils"
	"github.com/consensys/accelerated-crypto-monorepo/utils/parallel"
)

type Digest = hashtypes.Digest

// choic of hash function. Implicitly the tree is always binary.
type Config struct {
	HashFunc func() hashtypes.Hasher
	Depth    int
}

// Represents a Merkle-tree
type Tree struct {
	// Config of the hash function
	Config *Config
	// The root of the tree
	Root Digest
	// Continuous list of the occupied leaves. For the toy implementation
	// we track all the leaves.
	OccupiedLeaves []Digest
	/*
		Does not include the root. (So there are 39 levels and not 40).
		Returns a node at a given level:
			- 0 is the level just above the leaves
			- 38 is the level just above the root
	*/
	OccupiedNodes [][]Digest
	// Empty nodes does not include the "empty root" nor the empty leaf
	// so the first position contains the empty leaf for the level one.
	// So there are 39, and not 40 levels. That way, the indexing stays
	// consistent with "OccupiedNode"
	EmptyNodes []Digest
}

// Returns an empty leaf
func EmptyLeaf() Digest {
	return Digest{}
}

// It is used for hashing the leaf-right children
func HashLR(config *Config, nodeL, nodeR Digest) Digest {
	hasher := config.HashFunc()
	nodeL.WriteTo(hasher)
	nodeR.WriteTo(hasher)
	d := hashtypes.BytesToDigest(hasher.Sum(nil))
	return d
}

// Creates an empty tree
func NewEmptyTree(conf *Config) *Tree {
	// Computes the empty nodes
	emptyNodes := make([]Digest, conf.Depth-1)
	prevNode := EmptyLeaf()

	for i := range emptyNodes {
		newNode := HashLR(conf, prevNode, prevNode)
		emptyNodes[i] = newNode
		prevNode = newNode
	}

	// Stores the initial root separately
	root := HashLR(conf, prevNode, prevNode)

	return &Tree{
		Config:         conf,
		Root:           root,
		OccupiedLeaves: make([]Digest, 0),
		OccupiedNodes:  make([][]Digest, conf.Depth-1),
		EmptyNodes:     emptyNodes,
	}
}

// Returns a leaf by position
func (t *Tree) GetLeaf(pos int) Digest {
	// Check that the accessed node is within the bounds of the SMT
	maxPos := 1 << t.Config.Depth
	if pos >= maxPos {
		utils.Panic("nodeID is out of bound")
	}
	// Check if this is an empty leaf
	if pos >= len(t.OccupiedLeaves) {
		return EmptyLeaf()
	}
	// Return the leaf if occupied
	return t.OccupiedLeaves[pos]
}

/*
Returns a node at a given level:
  - 0 is a leaf
  - 1 - 39 : is an intermediate node
  - 40 is the root
*/
func (t *Tree) getNode(level, posInLevel int) Digest {
	switch {
	case level == t.Config.Depth:
		// The only logical posInLevels value is zero in this case
		if posInLevel > 0 {
			utils.Panic("there is only one root but posInLevel was %v", posInLevel)
		}
		// Opportunistic sanity-checks. Parenthesis are necessary because
		// of a hole in my linter.
		if t.Root == (Digest{}) {
			utils.Panic("sanity-check failed : the root is zero.")
		}
		return t.Root
	case level >= 1 && level <= t.Config.Depth-1:
		// Check that the accessed node is within the bounds of the SMT
		maxPos := 1 << (t.Config.Depth - level)
		if posInLevel >= maxPos {
			utils.Panic("nodeID is out of bound")
		}
		// Check if this is an empty node
		if posInLevel >= len(t.OccupiedNodes[level-1]) {
			return t.EmptyNodes[level-1]
		}
		// Or return an non-empty one
		res := t.OccupiedNodes[level-1][posInLevel]
		if res == (Digest{}) {
			utils.Panic("sanity-check : intermediary node is 0")
		}
		return res
	case level == 0:
		return t.GetLeaf(posInLevel)
	default:
		utils.Panic("Got level %v", level)
	}
	panic("unreachable")
}

/*
Returns a node at a given level:
  - 0 is a leaf
  - 1 - 39 : is an intermediate node
  - 40 is the root
*/
func (t *Tree) updateNode(level, posInLevel int, newVal Digest) {
	switch {
	case level == t.Config.Depth:
		// The only logical posInLevels value is zero in this case
		if posInLevel > 0 {
			utils.Panic("there is only one root but posInLevel was %v", posInLevel)
		}
		// Opportunistic sanity-checks. Parenthesis are necessary because
		// of a hole in my linter.
		if t.Root == (Digest{}) {
			utils.Panic("sanity-check failed : the root is zero.")
		}
		t.Root = newVal
	case level >= 1 && level < t.Config.Depth:
		// Check that the accessed node is within the bounds of the SMT
		maxPos := 1 << (t.Config.Depth - level)
		if posInLevel >= maxPos {
			utils.Panic("node is out of bound level %v (maxPos %v), pos=%v", level, maxPos, posInLevel)
		}
		// Check if this is an empty node : and reserve if necessary
		if posInLevel >= len(t.OccupiedNodes[level-1]) {
			t.reserveLevel(level, posInLevel+1)
		}
		// Or return an non-empty one
		t.OccupiedNodes[level-1][posInLevel] = newVal
	case level == 0:
		// Check that the accessed node is within the bounds of the SMT
		maxPos := 1 << t.Config.Depth
		if posInLevel >= maxPos {
			utils.Panic("nodeID is out of bound")
		}
		// Check if this is an empty leaf
		if posInLevel >= len(t.OccupiedLeaves) {
			t.reserveLevel(0, posInLevel+1)
		}
		// Return the leaf if occupied
		t.OccupiedLeaves[posInLevel] = newVal
	default:
		utils.Panic("Got level %v", level)
	}
}

/*
Level == 0 : reserve in the leaves
Level == 1..=39 : reserve in the nodes
Level == 40 : panic, can't reserve at the leaf level
*/
func (t *Tree) reserveLevel(level, newSize int) {

	// Edge-case, level out of bound
	if level > t.Config.Depth {
		utils.Panic("level out of bound %v", level)
	}

	// Edge : case root level
	if level == t.Config.Depth {
		utils.Panic("tried extending the root level %v", newSize)
	}

	// Will work for both the leaves and the intermediate nodes
	if newSize > 1<<t.Config.Depth-level {
		utils.Panic("overextending the tree %v, %v", level, newSize)
	}

	if level == 0 {
		if newSize <= len(t.OccupiedLeaves) {
			// already big enough
			return
		}
		// else, we add extra "empty leaves" at the end of the slice
		padding := make([]Digest, newSize-len(t.OccupiedLeaves))
		for i := range padding {
			padding[i] = EmptyLeaf()
		}
		t.OccupiedLeaves = append(t.OccupiedLeaves, padding...)
		return
	}

	/*
		Else we want to reserve within the occupiedNodes
	*/
	if newSize <= len(t.OccupiedNodes[level-1]) {
		// already big enough
		return
	}
	// else, we add extra "empty leaves" at the end of the slice
	padding := make([]Digest, newSize-len(t.OccupiedNodes[level-1]))
	for i := range padding {
		padding[i] = EmptyLeaf()
	}
	t.OccupiedNodes[level-1] = append(t.OccupiedNodes[level-1], padding...)
}

// Builds from scratch builds a complete Merkle-tree. Requires that
// the input leaves are powers of 2.
func BuildComplete(leaves []Digest, hashFunc func() hashtypes.Hasher) *Tree {

	numLeaves := len(leaves)

	// Sanity check : there should be a power of two number of leaves
	if !utils.IsPowerOfTwo(numLeaves) || numLeaves == 0 {
		utils.Panic("expected power of two number of leaves, got %v", numLeaves)
	}

	depth := utils.Log2Ceil(numLeaves)
	config := &Config{HashFunc: hashFunc, Depth: depth}

	// Builds an empty tree and passes the leaves
	tree := NewEmptyTree(config)
	tree.OccupiedLeaves = leaves

	// Builds the tree bottom-up
	currLevels := leaves

	for i := 0; i < depth-1; i++ {
		nextLevel := make([]Digest, len(currLevels)/2)
		parallel.Execute(len(nextLevel), func(start, stop int) {
			for k := start; k < stop; k++ {
				nextLevel[k] = HashLR(config, currLevels[2*k], currLevels[2*k+1])
			}
		})
		tree.OccupiedNodes[i] = nextLevel
		currLevels = nextLevel
	}

	// sanity-check : len(currLevels) == 2 is an invariant
	if len(currLevels) != 2 {
		utils.Panic("broken invariant : len(currLevels) != 2, =%v", len(currLevels))
	}

	// And overwrite the root
	tree.Root = HashLR(config, currLevels[0], currLevels[1])
	return tree
}
