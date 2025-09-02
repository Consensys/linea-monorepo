package serialization

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/fxamacker/cbor/v2"
)

var (
	cborEncMode cbor.EncMode
	cborDecMode cbor.DecMode
	encInitOnce sync.Once
	decInitOnce sync.Once
)

func initCBOREncMode() {
	var err error
	cborEncMode, err = cbor.CoreDetEncOptions().EncMode()
	if err != nil {
		panic(fmt.Errorf("failed to create CBOR EncMode: %w", err))
	}
}

func initCBORDecMode() {
	var err error
	cborDecMode, err = cbor.DecOptions{
		MaxArrayElements: 134217728,
		MaxMapPairs:      134217728,
		MaxNestedLevels:  256,
	}.DecMode()
	if err != nil {
		panic(fmt.Errorf("failed to create CBOR DecMode: %w", err))
	}
}

// encodeWithCBORToBuffer uses a user buffer (if supported by the library) to reduce realloc/copies.
func encodeWithCBORToBuffer(buf *bytes.Buffer, x any) error {
	encInitOnce.Do(initCBOREncMode)
	// If the EncMode supports MarshalToBuffer (fxamacker/cbor v2.7+), use it to encode x directly into buf,
	// avoiding an intermediate allocation. The encoding is performed in a single pass directly into the buffer,
	// eliminating the extra copy operation. This reduces CPU usage, particularly for frequent or large serializations.
	if ub, ok := any(cborEncMode).(interface {
		MarshalToBuffer(v any, b *bytes.Buffer) error
	}); ok {
		return ub.MarshalToBuffer(x, buf)
	}
	// Fallback to std. Marshal then write to the provided buffer. Note The encoding process is followed by a manual buf.Write(b),
	// which involves a memory copy from the allocated slice to the buffer. This additional copy adds CPU cycles.
	b, err := cborEncMode.Marshal(x)
	if err != nil {
		return err
	}
	_, err = buf.Write(b)
	return err
}

// decodeWithCBOR: single-step Unmarshal (no reader pool)
func decodeWithCBOR(data []byte, x any) error {
	decInitOnce.Do(initCBORDecMode)
	return cborDecMode.Unmarshal(data, x)
}
