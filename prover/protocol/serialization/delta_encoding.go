package serialization

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
)

// FrontCodedIDs is the CBOR-serializable structure for ID lists.
type FrontCodedIDs struct {
	Count   uint64 `cbor:"c"` // Total number of IDs
	Payload []byte `cbor:"p"` // The delta-encoded stream
}

// helper struct for sorting
type idEntry struct {
	val string
	id  int
}

// buildFrontCoded takes the raw list of IDs (insertion order), sorts them,
// and generates the FrontCodedIDs struct with embedded original IDs.
func buildFrontCoded(rawIDs []string) FrontCodedIDs {
	if len(rawIDs) == 0 {
		return FrontCodedIDs{}
	}

	// 1. Wrap with original index and Sort
	entries := make([]idEntry, len(rawIDs))
	for i, s := range rawIDs {
		entries[i] = idEntry{val: s, id: i}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].val < entries[j].val
	})

	// 2. Encode
	var payload bytes.Buffer
	var scratch [binary.MaxVarintLen64]byte

	writeU := func(v uint64) {
		n := binary.PutUvarint(scratch[:], v)
		payload.Write(scratch[:n])
	}

	prev := ""
	for _, entry := range entries {
		// Calculate common prefix
		cp := 0
		minLen := len(prev)
		if len(entry.val) < minLen {
			minLen = len(entry.val)
		}
		for i := 0; i < minLen; i++ {
			if prev[i] != entry.val[i] {
				break
			}
			cp++
		}

		suffix := entry.val[cp:]

		// Stream format: [PrefixLen] [SuffixLen] [SuffixBytes] [OriginalID]
		writeU(uint64(cp))
		writeU(uint64(len(suffix)))
		payload.WriteString(suffix)
		writeU(uint64(entry.id))

		prev = entry.val
	}

	return FrontCodedIDs{
		Count:   uint64(len(rawIDs)),
		Payload: payload.Bytes(),
	}
}

// inflateFrontCoded decodes the stream entirely into a slice mapped by Original ID.
func inflateFrontCoded(fc FrontCodedIDs) ([]string, error) {
	if fc.Count == 0 {
		return []string{}, nil
	}

	out := make([]string, fc.Count)
	br := bytes.NewReader(fc.Payload)

	prev := ""
	for i := uint64(0); i < fc.Count; i++ {
		// Read Prefix Len
		cp, err := binary.ReadUvarint(br)
		if err != nil {
			return nil, fmt.Errorf("read prefix: %w", err)
		}

		// Read Suffix Len
		sl, err := binary.ReadUvarint(br)
		if err != nil {
			return nil, fmt.Errorf("read suffix len: %w", err)
		}

		// Read Suffix Bytes
		suffixBuf := make([]byte, sl)
		if _, err := br.Read(suffixBuf); err != nil {
			return nil, fmt.Errorf("read suffix bytes: %w", err)
		}

		// Read Original ID
		origID, err := binary.ReadUvarint(br)
		if err != nil {
			return nil, fmt.Errorf("read origID: %w", err)
		}

		// Reconstruct
		current := ""
		if cp == 0 {
			current = string(suffixBuf)
		} else {
			if uint64(len(prev)) < cp {
				return nil, fmt.Errorf("corrupt stream: prefix len > prev string")
			}
			current = prev[:cp] + string(suffixBuf)
		}

		if origID >= fc.Count {
			return nil, fmt.Errorf("origID out of bounds")
		}
		out[origID] = current
		prev = current
	}

	return out, nil
}
