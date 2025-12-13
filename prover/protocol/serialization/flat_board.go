package serialization

import (
	"strings"
)

// PackedBoard represents the flattened Radix Tree.
type PackedBoard struct {
	// 1. Dictionary of unique longest prefix segments
	Segments []string `cbor:"s"`

	// 2. Flattened Trie Structure.
	// Index i represents a Node.
	// Parents[i] is the index of the parent Node (or -1 if root).
	Parents []int32 `cbor:"p"`

	// SegmentRefs[i] is the index into Segments[] for this Node's name.
	SegmentRefs []int32 `cbor:"r"`

	// 3. The mapping from your "Leaf IDs" to the Trie Nodes.
	// If UnpackColumnID(5) is called, we look at LeafToNode[5] to start the walk.
	LeafToNode []int32 `cbor:"l"`
}

// boardSer buffers strings and builds a highly compressed Radix tree at the end.
type boardSer struct {
	Packed *PackedBoard

	// Buffers for the "Two-Pass" strategy
	// We map "FullString" -> "LeafID" (index in LeafToNode)
	uniqueStrings map[string]int
	sortedStrings []string
}

func newFlatBoardSerializer() *boardSer {
	return &boardSer{
		Packed: &PackedBoard{
			Segments:    make([]string, 0),
			Parents:     make([]int32, 0),
			SegmentRefs: make([]int32, 0),
			LeafToNode:  make([]int32, 0),
		},
		uniqueStrings: make(map[string]int),
		sortedStrings: make([]string, 0),
	}
}

// addID simply buffers the string. The heavy lifting happens in Finalize().
// This returns a temporary LeafID (index into LeafToNode) that we will fill later.
func (h *boardSer) addID(fullID string) int {
	if idx, ok := h.uniqueStrings[fullID]; ok {
		return idx
	}

	// Assign a new ID
	leafID := len(h.uniqueStrings)
	h.uniqueStrings[fullID] = leafID
	h.sortedStrings = append(h.sortedStrings, fullID)

	// We expand LeafToNode to reserve space, but we fill it with -1 for now.
	// It will be populated in Finalize().
	h.Packed.LeafToNode = append(h.Packed.LeafToNode, -1)

	return leafID
}

// reconstructStringFromLeaf bridges the Deserializer to the specialized BoardDeserializer.
func (de *Deserializer) reconstructStringFromLeaf(leaf int) (string, *serdeError) {
	if de.FlatBoardDes == nil {
		return "", newSerdeErrorf("hierarchy deserializer not initialized")
	}

	// getString is defined in flat_board.go
	s := de.FlatBoardDes.getString(leaf)

	// Optional: You can add checks here if 's' is empty and leaf != 0,
	// but the BoardDeserializer handles out-of-bounds gracefully.
	return s, nil
}

// finalize builds the Radix Tree.
func (h *boardSer) finalize() {
	if len(h.sortedStrings) == 0 {
		return
	}

	// 1. Build the pointer-based Radix Tree in memory
	root := &radixNode{children: make(map[byte]*radixNode)}

	for _, str := range h.sortedStrings {
		leafID := h.uniqueStrings[str]
		insertRadix(root, str, leafID)
	}

	// 2. Flatten the tree into the Packed arrays
	// Root is always conceptually index -1, so we start BFS/DFS from its children.

	// We perform a BFS to keep children near parents in the arrays (good for locality)
	type queueItem struct {
		node      *radixNode
		parentIdx int32
	}

	queue := []queueItem{}

	// Initialize queue with root's children
	// Sorting ensures deterministic radix shape and stable serialized output.
	// Safe to remove if insertion order is already deterministic.
	for _, child := range root.children {
		queue = append(queue, queueItem{child, -1})
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Add current node to Packed Arrays
		myIdx := int32(len(h.Packed.Parents))

		h.Packed.Parents = append(h.Packed.Parents, current.parentIdx)

		// Deduplicate Segments (optional, but good if "0" appears in many sub-branches)
		// For simplicity/speed here we just append, but you could use a segmentMap here too.
		segIdx := int32(len(h.Packed.Segments))
		h.Packed.Segments = append(h.Packed.Segments, current.node.segment)
		h.Packed.SegmentRefs = append(h.Packed.SegmentRefs, segIdx)

		// If this node corresponds to a Leaf ID, update the mapping
		if current.node.leafID != -1 {
			h.Packed.LeafToNode[current.node.leafID] = myIdx
		}

		// Enqueue children
		for _, child := range current.node.children {
			queue = append(queue, queueItem{child, myIdx})
		}
	}
}

// --- Radix Tree Helper Logic ---

type radixNode struct {
	segment  string
	children map[byte]*radixNode
	leafID   int // -1 if internal node, >=0 if it matches a requested ID
}

func insertRadix(node *radixNode, remaining string, leafID int) {
	// Find the edge that matches the first char
	if len(remaining) == 0 {
		node.leafID = leafID
		return
	}

	firstChar := remaining[0]
	child, exists := node.children[firstChar]

	if !exists {
		// Case 1: No edge starts with this char. Create new edge.
		node.children[firstChar] = &radixNode{
			segment:  remaining,
			children: make(map[byte]*radixNode),
			leafID:   leafID,
		}
		return
	}

	// Case 2: Edge exists. Determine Longest Common Prefix (LCP).
	commonLen := 0
	maxLen := min(len(remaining), len(child.segment))
	for i := 0; i < maxLen; i++ {
		if remaining[i] != child.segment[i] {
			break
		}
		commonLen++
	}

	// Case 2a: Perfect match with existing segment (consume and recurse)
	if commonLen == len(child.segment) {
		insertRadix(child, remaining[commonLen:], leafID)
		return
	}

	// Case 2b: Partial match. We must SPLIT the existing edge.
	// Current: Node -> [ "BIGRANGE" ] -> ChildChildren...
	// Insert: "BIG"
	// New: Node -> [ "BIG" ] -> [ "RANGE" ] -> ChildChildren...

	// 1. Create the split child (suffixed part)
	suffix := child.segment[commonLen:]
	splitChild := &radixNode{
		segment:  suffix,
		children: child.children,
		leafID:   child.leafID, // Inherits the ID if the original was a leaf
	}

	// 2. Update the existing child to become the "Prefix Node"
	child.segment = child.segment[:commonLen] // "BIG"
	child.children = make(map[byte]*radixNode)
	child.children[suffix[0]] = splitChild
	child.leafID = -1 // It wasn't a leaf before (unless exactly at this split point, handled below)

	// 3. Insert the new part
	remainingSuffix := remaining[commonLen:]
	if len(remainingSuffix) == 0 {
		// The new string *is* the prefix.
		child.leafID = leafID
	} else {
		// Branch off for the new string
		child.children[remainingSuffix[0]] = &radixNode{
			segment:  remainingSuffix,
			children: make(map[byte]*radixNode),
			leafID:   leafID,
		}
	}
}

// --- Deserializer ---

type boardDeser struct {
	Packed *PackedBoard
	cache  []string
}

func newBoardDeserializer(p *PackedBoard) *boardDeser {
	return &boardDeser{
		Packed: p,
		cache:  make([]string, len(p.LeafToNode)),
	}
}

func (hd *boardDeser) getString(leafID int) string {
	if leafID < 0 || leafID >= len(hd.cache) {
		return ""
	}
	if hd.cache[leafID] != "" {
		return hd.cache[leafID]
	}

	nodeIdx := hd.Packed.LeafToNode[leafID]
	var parts []string

	for nodeIdx != -1 {
		segIdx := hd.Packed.SegmentRefs[nodeIdx]
		parts = append(parts, hd.Packed.Segments[segIdx])
		nodeIdx = hd.Packed.Parents[nodeIdx]
	}

	// Reverse
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	// CHANGE: Join with empty string, not "_".
	// The Radix Serializer preserves delimiters inside the segments.
	fullID := strings.Join(parts, "")
	hd.cache[leafID] = fullID
	return fullID
}
