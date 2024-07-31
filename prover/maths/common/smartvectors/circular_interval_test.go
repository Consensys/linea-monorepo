package smartvectors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircularIntervalConstructors(t *testing.T) {

	t.Run("for a normal interval", func(t *testing.T) {
		i := ivalWithStartLen(2, 5, 10)
		assert.Equal(t, 2, i.start(), "start")
		assert.Equal(t, 7, i.stop(), "stop")
		assert.Equal(t, 5, i.intervalLen, "interval length")
		assert.False(t, i.doesWrapAround(), "wrap around")
		assert.False(t, i.isFullCircle(), "full circle")

		assert.True(t, i.doesInclude(5), "in the middle of the interval")
		assert.True(t, i.doesInclude(2), "it should be closed on the left")
		assert.False(t, i.doesInclude(7), "it should be open on the right")

		assert.False(t, i.doesInclude(0), "point on the left")
		assert.False(t, i.doesInclude(8), "point on the right")

		assert.Equal(t, ivalWithStartStop(2, 7, 10), i)
	})

	t.Run("for a wrapped around vector", func(t *testing.T) {

		i := ivalWithStartLen(7, 5, 10)

		assert.Equal(t, 7, i.start(), "start")
		assert.Equal(t, 2, i.stop(), "stop")
		assert.Equal(t, 5, i.intervalLen, "interval length")
		assert.True(t, i.doesWrapAround(), "wrap around")
		assert.False(t, i.isFullCircle(), "full circle")

		assert.False(t, i.doesInclude(5), "in the middle of the interval")
		assert.True(t, i.doesInclude(7), "it should be closed on the left")
		assert.False(t, i.doesInclude(2), "it should be open on the right")

		assert.True(t, i.doesInclude(0), "point on the left")
		assert.True(t, i.doesInclude(8), "point on the right")

		assert.Equal(t, ivalWithStartStop(7, 2, 10), i)
	})

	t.Run("for a full vector", func(t *testing.T) {
		i := ivalWithFullLen(10)
		assert.Equal(t, 0, i.start(), "start")
		assert.Equal(t, 0, i.stop(), "stop")
		assert.Equal(t, 10, i.intervalLen, "interval length")
		assert.False(t, i.doesWrapAround(), "wrap around")
		assert.True(t, i.isFullCircle(), "full circle")

		assert.True(t, i.doesInclude(5), "in the middle of the interval")
		assert.True(t, i.doesInclude(7), "it should be closed on the left")
		assert.True(t, i.doesInclude(2), "it should be open on the right")

		assert.True(t, i.doesInclude(0), "point on the left")
		assert.True(t, i.doesInclude(8), "point on the right")
	})
}

func TestDoesFullyContain(t *testing.T) {

	t.Run("for a normal vector", func(t *testing.T) {
		i := ivalWithStartStop(5, 10, 15)

		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 3, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 5, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 8, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 10, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 13, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 1, 15)))

		assert.True(t, i.doesFullyContain(ivalWithStartStop(5, 8, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(5, 10, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(5, 13, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(5, 3, 15)))

		assert.True(t, i.doesFullyContain(ivalWithStartStop(7, 8, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(7, 10, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 13, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 3, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 5, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 6, 15)))

		assert.False(t, i.doesFullyContain(ivalWithStartStop(10, 13, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(10, 3, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(10, 5, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(10, 8, 15)))

		assert.False(t, i.doesFullyContain(ivalWithStartStop(12, 13, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(12, 3, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(12, 5, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(12, 8, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(12, 10, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(12, 11, 15)))

		assert.False(t, i.doesFullyContain(ivalWithFullLen(15)))
	})

	t.Run("for a wrap around", func(t *testing.T) {
		i := ivalWithStartStop(10, 5, 15)

		assert.True(t, i.doesFullyContain(ivalWithStartStop(2, 3, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(2, 5, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 8, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 10, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 13, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(2, 1, 15)))

		assert.False(t, i.doesFullyContain(ivalWithStartStop(5, 8, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(5, 10, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(5, 13, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(5, 3, 15)))

		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 8, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 10, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 13, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 3, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 5, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(7, 6, 15)))

		assert.True(t, i.doesFullyContain(ivalWithStartStop(10, 13, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(10, 3, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(10, 5, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(10, 8, 15)))

		assert.True(t, i.doesFullyContain(ivalWithStartStop(12, 13, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(12, 3, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(12, 5, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(12, 8, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(12, 10, 15)))
		assert.False(t, i.doesFullyContain(ivalWithStartStop(12, 11, 15)))

		assert.False(t, i.doesFullyContain(ivalWithFullLen(15)))

	})

	t.Run("for a wrap around", func(t *testing.T) {
		i := ivalWithFullLen(15)

		assert.True(t, i.doesFullyContain(ivalWithStartStop(2, 3, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(2, 5, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(2, 8, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(2, 10, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(2, 13, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(2, 1, 15)))

		assert.True(t, i.doesFullyContain(ivalWithStartStop(5, 8, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(5, 10, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(5, 13, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(5, 3, 15)))

		assert.True(t, i.doesFullyContain(ivalWithStartStop(7, 8, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(7, 10, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(7, 13, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(7, 3, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(7, 5, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(7, 6, 15)))

		assert.True(t, i.doesFullyContain(ivalWithStartStop(10, 13, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(10, 3, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(10, 5, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(10, 8, 15)))

		assert.True(t, i.doesFullyContain(ivalWithStartStop(12, 13, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(12, 3, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(12, 5, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(12, 8, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(12, 10, 15)))
		assert.True(t, i.doesFullyContain(ivalWithStartStop(12, 11, 15)))

		assert.True(t, i.doesFullyContain(ivalWithFullLen(15)))

	})

}

func TestTryOverlap(t *testing.T) {

	var ok bool
	var res circularInterval

	t.Run("for a normal vector", func(t *testing.T) {
		i := ivalWithStartStop(5, 10, 15)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 3, 15))
		assert.False(t, ok)
		assert.Equal(t, circularInterval{}, res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(2, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(2, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(2, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(2, 13, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 1, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(2, 1, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 13, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 3, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 13, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 3, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 6, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 13, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(5, 3, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 13, 15))
		assert.False(t, ok)
		assert.Equal(t, circularInterval{}, res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 3, 15))
		assert.False(t, ok)
		assert.Equal(t, circularInterval{}, res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(12, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(12, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(12, 10, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 11, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(12, 11, 15), res)

		ok, res = i.tryOverlapWith(ivalWithFullLen(15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)
	})

	t.Run("for a wrap around", func(t *testing.T) {
		i := ivalWithStartStop(10, 5, 15)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 8, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 1, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 8, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 8, 15))
		assert.False(t, ok)
		assert.Equal(t, circularInterval{}, res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(7, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(7, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(7, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(7, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 6, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(7, 6, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 8, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 5, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithStartStop(10, 8, 15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 11, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithFullLen(15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

	})

	t.Run("for a wrap around", func(t *testing.T) {
		i := ivalWithFullLen(15)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(2, 1, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(5, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(7, 6, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(10, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 13, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 3, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 5, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 8, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 10, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithStartStop(12, 11, 15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

		ok, res = i.tryOverlapWith(ivalWithFullLen(15))
		assert.True(t, ok)
		assert.Equal(t, ivalWithFullLen(15), res)

	})

}
