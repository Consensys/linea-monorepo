package serialization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/fxamacker/cbor/v2"
)

var (
	cborEncMode cbor.EncMode
	cborDecMode cbor.DecMode
	initOnce    sync.Once

	bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	readerPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Reader)
		},
	}
)

func initCBOREncDecModes() {
	var err error
	cborEncMode, err = cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		panic(fmt.Errorf("failed to create CBOR EncMode: %w", err))
	}
	cborDecMode, err = cbor.DecOptions{
		MaxArrayElements: 134217728,
		MaxMapPairs:      134217728,
	}.DecMode()
	if err != nil {
		panic(fmt.Errorf("failed to create CBOR DecMode: %w", err))
	}
}

// serializeAnyWithCborPkg serializes an interface{} object into CBOR using a pooled buffer.
func serializeAnyWithCborPkg(x any) (json.RawMessage, error) {
	initOnce.Do(initCBOREncDecModes)

	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	enc := cborEncMode.NewEncoder(buf)
	if err := enc.Encode(x); err != nil {
		return nil, err
	}
	// Copy the bytes out before returning buffer to pool
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out, nil
}

// deserializeAnyWithCborPkg deserializes CBOR data into an object using a pooled reader.
func deserializeAnyWithCborPkg(data json.RawMessage, x any) error {
	initOnce.Do(initCBOREncDecModes)

	r := readerPool.Get().(*bytes.Reader)
	r.Reset(data)
	defer readerPool.Put(r)

	dec := cborDecMode.NewDecoder(r)
	if err := dec.Decode(x); err != nil {
		return fmt.Errorf("cbor.Unmarshal failed: %w", err)
	}
	return nil
}
