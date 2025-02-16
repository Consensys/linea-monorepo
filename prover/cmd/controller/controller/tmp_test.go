package controller

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestTmpLimitlessRun(t *testing.T) {
	var (
		// exit0 int = 0
		// exit2   int = 2
		// exit10  int = 10
		// exit12 int = 12
		exit77 int = 77
		// exit137 int = 137
	)

	_, confL := setupLimitlessFsTest(t)

	// Dirs
	execBootstrapFrom := []string{confL.ExecBootstrap.DirFrom(0)}

	// Populate the filesystem with job files

	// Bootstrap
	// createLimitlessTestInputFiles(execBootstrapFrom, 0, 1, execBootstrapPriority, exit0)
	// createLimitlessTestInputFiles(execBootstrapFrom, 1, 2, execBootstrapPriority, exit12, forLarge)
	createLimitlessTestInputFiles(execBootstrapFrom, 2, 3, execBootstrapPriority, exit77)
	// createLimitlessTestInputFiles(execBootstrapFrom, 3, 4, execBootstrapPriority, exit77, forLarge)
	// createLimitlessTestInputFiles(execBootstrapFrom, 4, 5, execBootstrapPriority, exit137)
	// createLimitlessTestInputFiles(execBootstrapFrom, 5, 6, execBootstrapPriority, exit137, forLarge)
	// createLimitlessTestInputFiles(execBootstrapFrom, 6, 7, execBootstrapPriority, exit2)
	// createLimitlessTestInputFiles(execBootstrapFrom, 7, 8, execBootstrapPriority, exit2)
	// createLimitlessTestInputFiles(execBootstrapFrom, 8, 9, execBootstrapPriority, exit10)
	// createLimitlessTestInputFiles(execBootstrapFrom, 9, 10, execBootstrapPriority, exit12)

	// ctxM, stopM := context.WithCancel(context.Background())
	ctxL, stopL := context.WithCancel(context.Background())

	// go runController(ctxM, confM)
	// go runController(ctxL, confL)

	// For Debug mode only
	runController(ctxL, confL)

	// Wait for a few secs, for the test to complete
	// <-time.After(2 * time.Second)

	// Shutdown the controller
	// stopM()
	stopL()

	expectedStructure := []struct {
		Path    string
		Entries []string
	}{
		{
			Path:    confL.ExecBootstrap.DirFrom(0),
			Entries: []string{}, // all files should be processed
		},
		{
			Path: confL.ExecBootstrap.DirDone(0),
			Entries: []string{
				// "0-1-etv0.1.2-stv1.2.3-getZkProof.json.success",
				// "1-2-etv0.1.2-stv1.2.3-getZkProof.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof.json.failure.code_67",
			},
		},
		// {
		// 	Path: confL.ExecBootstrap.DirTo(0),
		// 	Entries: []string{
		// 		"0-1-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json",
		// 		"1-2-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json",
		// 	},
		// },
		// {
		// 	Path: confL.ExecBootstrap.DirTo(1),
		// 	Entries: []string{
		// 		"0-1-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
		// 		"1-2-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
		// 	},
		// },
	}

	for _, dirVal := range expectedStructure {
		dir, err := os.Open(dirVal.Path)
		require.NoErrorf(t, err, "dir %v", dirVal.Path)
		filesFound, err := dir.Readdirnames(-1)
		require.NoErrorf(t, err, "dir %v", dirVal.Path)
		slices.Sort(filesFound)
		assert.Equalf(t, dirVal.Entries, filesFound, "dir %v", dirVal.Path)
	}
}
