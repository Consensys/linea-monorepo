package mempool

import (
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"strings"
	"testing"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/stretchr/testify/assert"
)

func TestDebugPool(t *testing.T) {

	t.Run("leak-detection base", func(t *testing.T) {

		pool := NewDebugPool(CreateFromSyncPool(32))

		for i := 0; i < 16; i++ {
			func() {
				_ = pool.Alloc()
			}()
		}

		err := pool.Errors().Error()
		assert.True(t, strings.HasPrefix(err, "leaked a base vector out of the pool"))
	})

	t.Run("leak-detection extension", func(t *testing.T) {

		pool := NewDebugPool(CreateFromSyncPool(32))

		for i := 0; i < 16; i++ {
			func() {
				_ = pool.AllocExt()
			}()
		}

		err := pool.Errors().Error()
		assert.True(t, strings.HasPrefix(err, "leaked an extension vector out of the pool"))
	})

	t.Run("double-free base", func(t *testing.T) {

		pool := NewDebugPool(CreateFromSyncPool(32))

		v := pool.Alloc()

		for i := 0; i < 16; i++ {
			pool.Free(v)
		}

		err := pool.Errors().Error()
		assert.Truef(t, strings.HasPrefix(err, "vector was freed multiple times concurrently"), err)
	})

	t.Run("double-free base", func(t *testing.T) {

		pool := NewDebugPool(CreateFromSyncPool(32))

		v := pool.AllocExt()

		for i := 0; i < 16; i++ {
			pool.FreeExt(v)
		}

		err := pool.Errors().Error()
		assert.Truef(t, strings.HasPrefix(err, "vector was freed multiple times concurrently"), err)
	})

	t.Run("foreign-free", func(t *testing.T) {

		pool := NewDebugPool(CreateFromSyncPool(32))

		v := make([]field.Element, 32)
		pool.Free(&v)

		err := pool.Errors().Error()
		assert.Truef(t, strings.HasPrefix(err, "freed a base vector that was not from the pool"), err)
	})

	t.Run("foreign-free extension", func(t *testing.T) {

		pool := NewDebugPool(CreateFromSyncPool(32))

		v := make([]fext.Element, 32)
		pool.FreeExt(&v)

		err := pool.Errors().Error()
		assert.Truef(t, strings.HasPrefix(err, "freed an extension vector that was not from the pool"), err)
	})

}
