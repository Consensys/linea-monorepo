package serialization

import "strings"

// PackedHierarchy represents the flattened Trie.
type PackedHierarchy struct {
	// 1. Dictionary of unique segments (e.g. "BIGRANGE", "ACCUMULATOR", "10")
	Segments []string `cbor:"s"`

	// 2. The Trie Structure.
	// Index i represents a Node.
	// Parents[i] is the index of the parent Node (or -1 if root).
	Parents []int32 `cbor:"p"`

	// SegmentRefs[i] is the index into Segments[] for this Node's name.
	SegmentRefs []int32 `cbor:"r"`

	// 3. The mapping from your "Leaf IDs" to the Trie Nodes.
	// If UnpackColumnID(5) is called, we look at LeafToNode[5] to start the walk.
	LeafToNode []int32 `cbor:"l"`
}

type HierarchySerializer struct {
	// Output structure
	Packed *PackedHierarchy

	// Deduplication Maps (Only used during serialization, not serialized!)
	segmentMap map[string]int32  // "BIGRANGE" -> 0
	nodeMap    map[nodeKey]int32 // {Parent: 0, Seg: 5} -> NodeIndex
}

type nodeKey struct {
	parentIdx int32
	segIdx    int32
}

func NewHierarchySerializer() *HierarchySerializer {
	return &HierarchySerializer{
		Packed: &PackedHierarchy{
			Segments:    make([]string, 0, 1024),
			Parents:     make([]int32, 0, 1024),
			SegmentRefs: make([]int32, 0, 1024),
			LeafToNode:  make([]int32, 0, 1024),
		},
		segmentMap: make(map[string]int32),
		nodeMap:    make(map[nodeKey]int32),
	}
}

// AddID adds a full ID string (e.g. "A_B_C") and returns the leaf index (BackReference).
// It performs 0 allocations for substrings using index math.
func (h *HierarchySerializer) AddID(fullID string) int {
	parentIdx := int32(-1) // Root

	// Efficient Tokenization Loop
	start := 0
	for i := 0; i <= len(fullID); i++ {
		// Detect delimiter or end of string
		if i == len(fullID) || fullID[i] == '_' { // Assuming '_' is delimiter
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

type HierarchyDeserializer struct {
	Packed *PackedHierarchy
	// Cache for fully reconstructed strings to avoid rebuilding them
	cache []string
}

func NewHierarchyDeserializer(p *PackedHierarchy) *HierarchyDeserializer {
	return &HierarchyDeserializer{
		Packed: p,
		cache:  make([]string, len(p.LeafToNode)),
	}
}

func (hd *HierarchyDeserializer) GetString(leafID int) string {
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
	// Since depth is usually small (10-20), a small slice is fine.
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
