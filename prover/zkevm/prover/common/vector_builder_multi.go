package common

// MultiVectorBuilder is a convenience structure to assign groups of columns
// by appending rows on top of rows.
type MultiVectorBuilder struct {
	T []*VectorBuilder
}
