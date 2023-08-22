package fiatshamir

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFiatShamirSafeguardUpdate(t *testing.T) {

	fs := NewMiMCFiatShamir()

	a := fs.RandomField()
	b := fs.RandomField()

	// Two consecutive call to fs do not return the same result
	require.NotEqual(t, a.String(), b.String())
}

func TestFiatShamirRandomVec(t *testing.T) {

	fs := NewMiMCFiatShamir()

	a := fs.RandomManyIntegers(15, 16)
	require.Equal(t, 15, len(a))
}
