package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRename(t *testing.T) {

	dir, err := os.ReadDir(".")
	require.NoError(t, err)
	for _, entry := range dir {
		div := strings.Split(entry.Name(), "-")
		if len(div) < 2 {
			t.Log("skipping", entry.Name())
			continue
		}
		require.NoError(t, os.Rename(entry.Name(), fmt.Sprintf("%s-%s-getZkProof.json", div[0], div[1])))

	}
}
