//go:build cuda

package plonk

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

func TestG1MSMPointsPinnedLayoutCompactRegression(t *testing.T) {
	require.Equal(t, 96, g1TEPointSize, "compact TE points must stay 96 bytes")
	require.Equal(t, 144, g1TEPrecompPointSize, "precomputed TE points are larger")

	points := []G1TEPoint{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		{13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24},
	}

	pinned, err := PinG1TEPoints(points)
	require.NoError(t, err)
	defer pinned.Free()

	got := unsafe.Slice((*G1TEPoint)(pinned.pinned), len(points))
	require.Equal(t, points, append([]G1TEPoint(nil), got...),
		"pinned MSM points must preserve the compact X,Y layout")
}
