package serialization

import "strings"

var (
	delimiter = "_"
)

// --- trie (token-based, '_' delimiter) ------------------------------------
type TrieNode struct {
	Children map[string]*TrieNode `cbor:"c,omitempty"`
	Backref  int                  `cbor:"b"` // -1 means none; >=0 is leaf backref
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		Children: map[string]*TrieNode{},
		Backref:  -1,
	}
}

// Insert tokenized string (split by '_') and set leaf backref.
// If leaf already exists, it overwrites Backref with the provided one.
func (t *TrieNode) InsertByTokens(tokens []string, backref int) {
	node := t
	for _, tok := range tokens {
		ch, ok := node.Children[tok]
		if !ok {
			ch = NewTrieNode()
			node.Children[tok] = ch
		}
		node = ch
	}
	node.Backref = backref
}

// Convenience: insert whole string using '_' delimiter
func (t *TrieNode) Insert(s string, backref int) {
	if s == "" {
		// treat empty as root leaf (rare). assign backref to root.
		t.Backref = backref
		return
	}
	tokens := strings.Split(s, delimiter)
	t.InsertByTokens(tokens, backref)
}

// Lookup returns (backref, true) if the exact id is present; otherwise (0,false)
func (t *TrieNode) Lookup(s string) (int, bool) {
	if t == nil {
		return 0, false
	}
	if s == "" {
		if t.Backref >= 0 {
			return t.Backref, true
		}
		return 0, false
	}
	tokens := strings.Split(s, delimiter)
	node := t
	for _, tok := range tokens {
		ch, ok := node.Children[tok]
		if !ok {
			return 0, false
		}
		node = ch
	}
	if node.Backref >= 0 {
		return node.Backref, true
	}
	return 0, false
}

// walkTrie collects mapping backref -> reconstructed string.
// It expects no duplicate backrefs (if duplicates present, later wins).
func buildLeafMap(root *TrieNode) map[int]string {
	out := map[int]string{}
	var dfs func(n *TrieNode, path []string)
	dfs = func(n *TrieNode, path []string) {
		if n == nil {
			return
		}
		// If node is a leaf (has a backref), record it
		if n.Backref >= 0 {
			// join currently collected path (root->...->this)
			out[n.Backref] = strings.Join(path, delimiter)
		}
		for tok, child := range n.Children {
			dfs(child, append(path, tok))
		}
	}
	dfs(root, []string{})
	return out
}

/*
// PackColumnID serializes an ifaces.ColID (string), returning a BackReference to its index in PackedObject.ColumnIDs.
func (ser *Serializer) PackColumnID(c ifaces.ColID) (BackReference, *serdeError) {
	if _, ok := ser.columnIdMap[string(c)]; !ok {
		ser.PackedObject.ColumnIDs = append(ser.PackedObject.ColumnIDs, string(c))
		ser.columnIdMap[string(c)] = len(ser.PackedObject.ColumnIDs) - 1
	}

	return BackReference(ser.columnIdMap[string(c)]), nil
}

// UnpackColumnID deserializes an ifaces.ColID from a BackReference.
func (de *Deserializer) UnpackColumnID(v BackReference) (reflect.Value, *serdeError) {
	if v < 0 || int(v) >= len(de.PackedObject.ColumnIDs) {
		return reflect.Value{}, newSerdeErrorf("invalid column-ID backreference: %v", v)
	}

	res := ifaces.ColID(de.PackedObject.ColumnIDs[v])
	return reflect.ValueOf(res), nil
}


// PackCoinID serializes a coin.Name (string), returning a BackReference to its index in PackedObject.CoinIDs.
func (ser *Serializer) PackCoinID(c coin.Name) (BackReference, *serdeError) {
	if _, ok := ser.coinIdMap[string(c)]; !ok {
		ser.PackedObject.CoinIDs = append(ser.PackedObject.CoinIDs, string(c))
		ser.coinIdMap[string(c)] = len(ser.PackedObject.CoinIDs) - 1
	}

	return BackReference(ser.coinIdMap[string(c)]), nil
}

// UnpackCoinID deserializes a coin.Name from a BackReference.
func (de *Deserializer) UnpackCoinID(v BackReference) (reflect.Value, *serdeError) {
	if v < 0 || int(v) >= len(de.PackedObject.CoinIDs) {
		return reflect.Value{}, newSerdeErrorf("invalid coin ID back reference: %v", v)
	}

	res := coin.Name(de.PackedObject.CoinIDs[v])
	return reflect.ValueOf(res), nil
}

// PackQueryID serializes an ifaces.QueryID (string), returning a BackReference to its index in PackedObject.QueryIDs.
func (ser *Serializer) PackQueryID(q ifaces.QueryID) (BackReference, *serdeError) {
	if _, ok := ser.queryIDMap[string(q)]; !ok {
		ser.PackedObject.QueryIDs = append(ser.PackedObject.QueryIDs, string(q))
		ser.queryIDMap[string(q)] = len(ser.PackedObject.QueryIDs) - 1
	}

	return BackReference(ser.queryIDMap[string(q)]), nil
}

// UnpackQueryID deserializes an ifaces.QueryID from a BackReference.
func (de *Deserializer) UnpackQueryID(v BackReference) (reflect.Value, *serdeError) {
	if v < 0 || int(v) >= len(de.PackedObject.QueryIDs) {
		return reflect.Value{}, newSerdeErrorf("invalid query-ID backreference: %v", v)
	}

	res := ifaces.QueryID(de.PackedObject.QueryIDs[v])
	return reflect.ValueOf(res), nil
}
*/
