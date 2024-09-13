package smartvectors

import (
	"sort"

	"github.com/consensys/linea-monorepo/prover/utils"
)

// circularInterval represents an interval over a discretized circle. The
// discretized circle is assumed to be equipped with an origin point; thus
// allowing to set a unique coordinate for each point.
//
//   - The intervals are "cardinal": meaning that the largest possible interval
//     is the full-circuit
//   - The empty interval is considered as invalid and should never be constructed
type circularInterval struct {
	// circleSize is the size of the circle
	circleSize int
	// istart is the starting point of the interval (included in the interval).
	//
	// istart must always be within the bound of the circle (can't be negative
	// or be larger or equal to `circleSize`.
	istart int
	// intervalLen is length of the interval. Meaning the number of points in
	// the interval
	intervalLen int
}

// ivalWithFullLen returns an interval representing the full-circle.
func ivalWithFullLen(n int) circularInterval {
	if n <= 0 {
		panic("zero or negative length interval is not allowed")
	}
	return circularInterval{
		istart:      0,
		intervalLen: n,
		circleSize:  n,
	}
}

// ivalWithStartLen constructs an interval by passing the start, the len and n
// being the size of the circle.
func ivalWithStartLen(start, len, n int) circularInterval {
	// empty length is forbidden
	if len == 0 {
		panic("empty interval")
	}
	// ensures that start is within bounds
	if 0 > start || start >= n {
		panic("start out of bounds")
	}
	// full length is forbidden
	if len >= n {
		panic("full length is forbidden")
	}
	return circularInterval{
		circleSize:  n,
		istart:      start,
		intervalLen: len,
	}
}

// ivalWithStartStop constructs a [circularInterval] by using its starting and
// stopping points.
func ivalWithStartStop(start, stop, n int) circularInterval {
	// empty interval is forbidden
	if start == stop {
		panic("empty interval")
	}
	// ensures that start is within bounds
	if 0 > start || start >= n {
		panic("start out of bounds")
	}
	// full length is forbidden
	if 0 > stop || stop >= n {
		panic("stop out of bound")
	}
	return circularInterval{
		circleSize:  n,
		istart:      start,
		intervalLen: utils.PositiveMod(stop-start, n),
	}
}

// Start returns the starting point (included) of the interval
func (c circularInterval) start() int {
	return c.istart
}

// Stop returns the stopping point (excluded) of the interval of the interval
func (c circularInterval) stop() int {
	return utils.PositiveMod(c.istart+c.intervalLen, c.circleSize)
}

// doesWrapAround returns true iff the interval rolls over
func (c circularInterval) doesWrapAround() bool {
	return c.stop() < c.start()
}

// isFullCircle returns true of the interval is the full circle
func (c circularInterval) isFullCircle() bool {
	return c.intervalLen == c.circleSize
}

// Returns true iff `p` is included in the receiver interval
func (c circularInterval) doesInclude(p int) bool {

	// forbidden : the point does not belong on the circle
	if p < 0 || p > c.circleSize {
		panic("point does not belong to the circle")
	}

	// edge-case
	if c.isFullCircle() {
		return true
	}

	// if the interval wraps around the origin point
	if c.doesWrapAround() {
		return p < c.stop() || p >= c.start()
	}

	// "normal" case
	return p >= c.start() && p < c.stop()
}

// doesFullyContain returns true if `c` fully contains `other`
func (c circularInterval) doesFullyContain(other circularInterval) bool {

	// edge case : c is the complete circle
	if c.isFullCircle() {
		return true
	}

	// edge case : c is not the complete circle but other is
	if !c.isFullCircle() && other.isFullCircle() {
		return false
	}

	if !c.doesWrapAround() {
		return c.doesInclude(other.start()) &&
			c.doesInclude(other.stop()-1) &&
			!other.doesWrapAround()
	}

	// Here, we can assume that c wraps around

	// Case : 1, other is on the left arm
	if !other.doesWrapAround() && other.stop() <= c.stop() {
		return true
	}

	// Case : 2, other is on the right arm
	if !other.doesWrapAround() && other.start() >= c.start() {
		return true
	}

	// Case 3 : other also wraps around
	if other.doesWrapAround() && other.start() >= c.start() && other.stop() <= c.stop() {
		return true
	}

	return false
}

/*
tryOverlapWith returns true if the left of `c` touches the right of `other`

			        c.start-------------c.stop
							other.start---------other.stop

										OR

		|c.start|-------------|c.stop|

	    							 |other.start|---------|other.stop|

This also include the edge cases where `other.stop`. Also
returns the resulting circular interval obtained by connecting
the two.
*/
func (c circularInterval) tryOverlapWith(other circularInterval) (ok bool, connected circularInterval) {

	// Sanity-check, both sides should have the same circle size
	if c.circleSize != other.circleSize {
		panic("not the same circle size")
	}

	// Size of the circle
	n := c.circleSize

	// There are still edge-cases for when either c or other are the full circle.
	// Once these cases are eliminated, we process by case enumeration.
	if c.isFullCircle() || other.isFullCircle() {
		return true, ivalWithFullLen(n)
	}

	/*
		Order to simplify the function, we reason on normalized coordinates. This
		reduces the (still huge) combinatoric of the function.

		Namely,
			[0, c1) represents the interval of 'c'
			[o0, o1) represents the interval of 'other'
	*/

	c1 := utils.PositiveMod(c.stop()-c.start(), n)
	o0 := utils.PositiveMod(other.start()-c.start(), n)
	o1 := utils.PositiveMod(other.stop()-c.start(), n)

	/*
		|-----------------c1
			o0-------o1
	*/
	if 0 <= o0 && o0 < o1 && o1 <= c1 {
		return true, c
	}

	/*
		|-----------------c1
		--------o1     o0---------------
	*/
	if 0 <= o1 && o1 < o0 && o0 <= c1 {
		return true, ivalWithFullLen(n)
	}

	/*
		|-----------------c1
		         o0----------------o1
	*/
	if 0 <= o0 && o0 <= c1 && c1 <= o1 {
		return true, ivalWithStartStop(c.start(), other.stop(), n)
	}

	/*
		|-----------------c1
		--------o1              o0--------
	*/
	if 0 <= o1 && o1 <= c1 && c1 < o0 {
		return true, ivalWithStartStop(other.start(), c.stop(), n)
	}

	/*
		|-----------------c1
		----------------------o1    o0----
	*/
	if 0 <= c1 && c1 <= o1 && o1 < o0 {
		return true, other
	}

	return false, circularInterval{}
}

// Returns the smallest windows that covers the entire set
func smallestCoverInterval(intervals []circularInterval) circularInterval {
	// Deep-copy the inputs to prevent any side effect
	intervals = append([]circularInterval{}, intervals...)

	if len(intervals) == 0 {
		panic("no windows passed")
	}

	// Assumption : the length of all arguments windows are equals.
	// This is asserted later in the function
	circleSize := intervals[0].circleSize

	// First step, we aggregate the windows whose union is a circle arc
	// into disjoints buckets. Thereafter, we take the complements of the
	// largest gap between buckets as our result.
	sort.Slice(intervals, func(i, j int) bool { return intervals[i].start() <= intervals[j].start() })
	overlaps := []circularInterval{}

	// Then we group the intervals whose union is still an interval. Since
	// the intervals are now sorted by their "start" argument, it suffices
	// to try and merge each with the next one. It they are not connected on
	// the right, then the following ones won't either.
	for i, interval := range intervals {

		if i == 0 {
			overlaps = append(overlaps, interval)
			continue
		}

		if intervals[0].circleSize != circleSize {
			panic("inconsistent sizes")
		}

		// Since the input intervals are sorted by their start at the beginning,
		// it suffices to try to merge with the last one.
		last := overlaps[len(overlaps)-1]

		if ok, newW := last.tryOverlapWith(interval); ok {
			overlaps[len(overlaps)-1] = newW
		} else {
			// Else create a new bucket
			overlaps = append(overlaps, interval)
		}

	}

	// Try to merge the last one into the first one (and possibly
	// the following ones). Indeed there is a possibility that
	// overlap[0] and overlap[1] cannot be connected by themselves
	// but are both connected to last
	for {
		// Everything was merged
		if len(overlaps) == 1 {
			break
		}

		last := overlaps[len(overlaps)-1]

		if ok, newW := last.tryOverlapWith(overlaps[0]); ok {
			overlaps[len(overlaps)-1] = newW
			overlaps = overlaps[1:]
		} else {
			break
		}
	}

	if len(overlaps) == 1 {
		// If there is only one group, just return the union
		return overlaps[0]
	}

	maxGap := 0    // size of the largest gap
	posMaxGap := 0 // pos of the window left to the gap in the overlaps
	for i, w := range overlaps {

		nextW := overlaps[(i+1)%len(overlaps)]
		gap := utils.PositiveMod(nextW.start()-w.stop(), circleSize)

		if gap > maxGap {
			maxGap = gap
			posMaxGap = i
		}
	}

	// Sanity-check, the max gap cannot be zero
	if maxGap < 1 {
		utils.Panic("Max gap is %v", maxGap)
	}

	start := overlaps[(posMaxGap+1)%len(overlaps)].start()
	stop := overlaps[posMaxGap].stop()
	return ivalWithStartStop(start, stop, circleSize)

}
