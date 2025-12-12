package serialization

import "strings"

var (
	delimiter = '_'
)

// PackedFlatBoard represents the flattened Trie.
type PackedFlatBoard struct {
	// 1. Dictionary of unique segments (e.g. "BIGRANGE", "ACCUMULATOR", "HI")
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

type BoardSerializer struct {
	// Output structure
	Packed *PackedFlatBoard

	// Deduplication Maps (Only used during serialization, not serialized!)
	segmentMap map[string]int32  // "BIGRANGE" -> 0
	nodeMap    map[nodeKey]int32 // {Parent: 0, Seg: 5} -> NodeIndex
}

type nodeKey struct {
	parentIdx int32
	segIdx    int32
}

func newFlatBoardSerializer() *BoardSerializer {
	return &BoardSerializer{
		Packed: &PackedFlatBoard{
			Segments:    make([]string, 0, 1024),
			Parents:     make([]int32, 0, 1024),
			SegmentRefs: make([]int32, 0, 1024),
			LeafToNode:  make([]int32, 0, 1024),
		},
		segmentMap: make(map[string]int32),
		nodeMap:    make(map[nodeKey]int32),
	}
}

// addID adds a full ID string (e.g. "A_B_C") and returns the leaf index (BackReference).
// It performs 0 allocations for substrings using index math.
func (h *BoardSerializer) addID(fullID string) int {
	parentIdx := int32(-1) // Root

	// Efficient Tokenization Loop
	start := 0
	for i := 0; i <= len(fullID); i++ {
		// Detect delimiter or end of string
		if i == len(fullID) || fullID[i] == byte(delimiter) {
			segment := fullID[start:i]

			// 1. Deduplicate Segment
			segIdx, ok := h.segmentMap[segment]
			if !ok {
				segIdx = int32(len(h.Packed.Segments))
				h.Packed.Segments = append(h.Packed.Segments, segment)
				h.segmentMap[segment] = segIdx
			}

			// 2. Deduplicate Node (Parent + Segment pair)
			key := nodeKey{parentIdx, segIdx}
			nodeIdx, ok := h.nodeMap[key]
			if !ok {
				nodeIdx = int32(len(h.Packed.Parents))
				h.Packed.Parents = append(h.Packed.Parents, parentIdx)
				h.Packed.SegmentRefs = append(h.Packed.SegmentRefs, segIdx)
				h.nodeMap[key] = nodeIdx
			}

			parentIdx = nodeIdx
			start = i + 1
		}
	}

	// The final parentIdx represents the Leaf Node for this ID
	leafID := len(h.Packed.LeafToNode)
	h.Packed.LeafToNode = append(h.Packed.LeafToNode, parentIdx)

	return leafID
}

type BoardDeserializer struct {
	Packed *PackedFlatBoard
	// Cache for fully reconstructed strings to avoid rebuilding them
	cache []string
}

func newBoardDeserializer(p *PackedFlatBoard) *BoardDeserializer {
	return &BoardDeserializer{
		Packed: p,
		cache:  make([]string, len(p.LeafToNode)),
	}
}

func (de *Deserializer) reconstructStringFromLeaf(leaf int) (string, *serdeError) {
	if de.HierarchyDes == nil {
		return "", newSerdeErrorf("hierarchy deserializer not initialized")
	}
	s := de.HierarchyDes.getString(leaf)
	if s == "" && leaf != 0 {
		// Depending on your logic, leaf 0 might be valid empty or root.
		// If GetString returns empty for non-zero leaf, likely an index error or invalid data
		return "", newSerdeErrorf("invalid leaf index (back reference): %v", leaf)
	}
	return s, nil
}

func (hd *BoardDeserializer) getString(leafID int) string {
	if leafID < 0 || leafID >= len(hd.cache) {
		return ""
	}
	// Return cached if available
	if hd.cache[leafID] != "" {
		return hd.cache[leafID]
	}

	// Reconstruct
	nodeIdx := hd.Packed.LeafToNode[leafID]

	// We build the string backwards then reverse, or build a list of parts.
	var parts []string

	for nodeIdx != -1 {
		segIdx := hd.Packed.SegmentRefs[nodeIdx]
		parts = append(parts, hd.Packed.Segments[segIdx])
		nodeIdx = hd.Packed.Parents[nodeIdx]
	}

	// Reverse parts to get correct order
	// (Opt: Use a pre-allocated buffer on the struct to avoid allocs here)
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	// Join and Cache
	fullID := strings.Join(parts, "_")
	hd.cache[leafID] = fullID
	return fullID
}

// compactFlatBoard performs a path-compression pass on a PackedFlatBoard.
// It merges linear chains of single-child nodes into single "joined" segments.
// Input: original packed board (Segments, Parents, SegmentRefs, LeafToNode).
// Output: a new PackedFlatBoard with fewer nodes and deduped/merged segments.
func (p *PackedFlatBoard) compactFlatBoard() *PackedFlatBoard {
	if p == nil || len(p.Parents) == 0 {
		return p
	}

	n := len(p.Parents)
	// build children list for each node
	children := make([][]int, n)
	for i := 0; i < n; i++ {
		parent := int(p.Parents[i])
		if parent >= 0 && parent < n {
			children[parent] = append(children[parent], i)
		}
	}

	// old->new node index mapping
	oldToNew := make([]int32, n)
	for i := range oldToNew {
		oldToNew[i] = -1
	}

	newSegments := make([]string, 0, len(p.Segments))
	newParents := make([]int32, 0, n)
	newSegRefs := make([]int32, 0, n)

	// Deduplicate combined segments in newSegments
	segMapNew := make(map[string]int32, len(p.Segments))

	var addNewSegment = func(seg string) int32 {
		if idx, ok := segMapNew[seg]; ok {
			return idx
		}
		idx := int32(len(newSegments))
		newSegments = append(newSegments, seg)
		segMapNew[seg] = idx
		return idx
	}

	// recursive DFS that compacts linear chains starting at old node "start"
	var process func(start int, parentNew int32)
	process = func(start int, parentNew int32) {
		// follow chain while there is exactly one child
		cur := start
		// collect parts for the chain
		var parts []string

		// note: we loop at least once and include the start node's segment
		for {
			parts = append(parts, p.Segments[p.SegmentRefs[cur]])
			ch := children[cur]
			if len(ch) != 1 {
				// stop: either leaf or branching
				break
			}
			// single child → continue the chain
			cur = ch[0]
		}

		// combined segment
		combined := strings.Join(parts, "_")
		segIdx := addNewSegment(combined)

		newNodeIdx := int32(len(newParents))
		newParents = append(newParents, parentNew)
		newSegRefs = append(newSegRefs, segIdx)

		// map all old nodes belonging to this chain (from start up to cur) to newNodeIdx
		// we need to walk again from start until we hit cur
		w := start
		for {
			oldToNew[w] = newNodeIdx
			if w == cur {
				break
			}
			// since earlier we determined chain children length==1, safe to index
			w = children[w][0]
		}

		// now process children of cur (these are branching children or terminals)
		for _, child := range children[cur] {
			process(child, newNodeIdx)
		}
	}

	// identify root nodes (parent == -1); process each root
	for oldNode := 0; oldNode < n; oldNode++ {
		if p.Parents[oldNode] == -1 {
			process(oldNode, -1)
		}
	}

	// Build new LeafToNode mapping by mapping old leaf->new node
	newLeaf := make([]int32, len(p.LeafToNode))
	for i, oldNode := range p.LeafToNode {
		if oldNode >= 0 && int(oldNode) < len(oldToNew) {
			newLeaf[i] = oldToNew[oldNode]
		} else {
			// invalid or missing mapping: keep -1 (or choose to error)
			newLeaf[i] = -1
		}
	}

	return &PackedFlatBoard{
		Segments:    newSegments,
		Parents:     newParents,
		SegmentRefs: newSegRefs,
		LeafToNode:  newLeaf,
	}
}
