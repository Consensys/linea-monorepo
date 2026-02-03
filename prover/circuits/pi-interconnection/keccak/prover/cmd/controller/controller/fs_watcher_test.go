package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLsName(t *testing.T) {

	dir := t.TempDir()

	// When the dir doesn't exist we should return an error
	_, err := lsname("/dir-that-does-not-exist-45464343434")
	assert.Errorf(t, err, "no error on non-existing directory")

	// When the dir exists and is non-empty (cwd will be non-empty)
	ls, err := lsname(".")
	assert.NoErrorf(t, err, "error on current directory")
	assert.NotEmptyf(t, ls, "empty on cwd")

	// When the directory is empty
	ls, err = lsname(dir)
	assert.NoErrorf(t, err, "error on tmp directory")
	assert.Emptyf(t, ls, "non empty dir")
}
