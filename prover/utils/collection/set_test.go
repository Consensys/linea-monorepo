package collection_test

import (
	"testing"

	"github.com/consensys/accelerated-crypto-monorepo/utils/collection"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {

	s := collection.NewSet[string]()

	assert.False(t, s.Exists("a"))
	assert.False(t, s.Exists("a", "b"))

	s.InsertNew("a")

	assert.True(t, s.Exists("a"))
	assert.False(t, s.Exists("b"))
	assert.False(t, s.Exists("a", "b"))

}
