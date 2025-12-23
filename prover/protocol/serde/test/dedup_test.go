package serde_test

import (
	"os"
	"testing"

	"github.com/consensys/linea-monorepo/prover/protocol/column"
	"github.com/consensys/linea-monorepo/prover/protocol/serde"
	"github.com/consensys/linea-monorepo/prover/protocol/wizard"
	"github.com/stretchr/testify/require"
)

func serializeSize(t *testing.T, v any) int64 {
	t.Helper()

	data, err := serde.Serialize(v)
	require.NoError(t, err)
	return int64(len(data))
}

func storeSize(t *testing.T, v any) int64 {
	t.Helper()

	f, err := os.CreateTemp("", "serde-dedup-*")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	err = serde.StoreToDisk(f.Name(), v, false)
	require.NoError(t, err)

	info, err := os.Stat(f.Name())
	require.NoError(t, err)
	return info.Size()
}

func requireDedup(
	t *testing.T,
	single any,
	multi any,
) {
	t.Helper()

	sz1 := serializeSize(t, single)
	sz2 := serializeSize(t, multi)

	// allow tiny overhead for slice/struct headers
	require.LessOrEqual(
		t,
		sz2,
		sz1*2,
		"expected dedup but size grew too much: single=%d multi=%d",
		sz1,
		sz2,
	)
}

func requireNoDedup(
	t *testing.T,
	single any,
	multi any,
) {
	t.Helper()

	sz1 := serializeSize(t, single)
	sz2 := serializeSize(t, multi)

	require.Greater(
		t,
		sz2,
		sz1*2/1, // linear-ish growth
		"expected no dedup but size did not grow: single=%d multi=%d",
		sz1,
		sz2,
	)
}

func TestSerdeDedup_NaturalColumn(t *testing.T) {
	comp := wizard.NewCompiledIOP()
	col := comp.InsertColumn(0, "a", 16, column.Committed)

	type S struct {
		A any
		B any
	}

	requireDedup(
		t,
		S{A: col},
		S{A: col, B: col},
	)

}

func TestSerdeDedup_NaturalColumn_SameUUIDDifferentPointer(t *testing.T) {
	comp1 := wizard.NewCompiledIOP()
	comp2 := wizard.NewCompiledIOP()

	c1 := comp1.InsertColumn(0, "a", 16, column.Committed)
	c2 := comp2.InsertColumn(0, "a", 16, column.Committed)

	type S struct {
		A any
		B any
	}

	requireDedup(
		t,
		S{A: c1},
		S{A: c1, B: c2},
	)
}

func TestSerdeDedup_ColumnStore_PointerOnly(t *testing.T) {
	//comp := wizard.NewCompiledIOP()
	store := column.NewStore()

	type S struct {
		A any
		B any
	}

	requireDedup(
		t,
		S{A: store},
		S{A: store, B: store},
	)
}
