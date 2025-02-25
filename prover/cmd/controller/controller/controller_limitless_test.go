package controller

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestLimitlessRun(t *testing.T) {
	var (
		exit0   int = 0
		exit2   int = 2
		exit10  int = 10
		exit12  int = 12
		exit77  int = 77
		exit137 int = 137
	)

	confM, confL := setupLimitlessFsTest(t)

	// Dirs
	execBootstrapFrom := []string{confL.ExecBootstrap.DirFrom(0)}
	execGLFrom := []string{confL.ExecGL.DirFrom(0)}
	execRndBeaconFrom := []string{
		confL.ExecRndBeacon.DirFrom(0),
		confL.ExecRndBeacon.DirFrom(1),
	}
	execLPPFrom := []string{confL.ExecLPP.DirFrom(0)}
	execConglomerationFrom := []string{
		confL.ExecConglomeration.DirFrom(0),
		confL.ExecConglomeration.DirFrom(1),
		confL.ExecConglomeration.DirFrom(2),
	}

	// Populate the filesystem with job files

	// Bootstrap
	createLimitlessTestInputFiles(execBootstrapFrom, 0, 1, execBootstrapPriority, exit0)
	createLimitlessTestInputFiles(execBootstrapFrom, 1, 2, execBootstrapPriority, exit12, forLarge)
	createLimitlessTestInputFiles(execBootstrapFrom, 2, 3, execBootstrapPriority, exit77)
	createLimitlessTestInputFiles(execBootstrapFrom, 3, 4, execBootstrapPriority, exit77, forLarge)
	createLimitlessTestInputFiles(execBootstrapFrom, 4, 5, execBootstrapPriority, exit137)
	createLimitlessTestInputFiles(execBootstrapFrom, 5, 6, execBootstrapPriority, exit137, forLarge)
	createLimitlessTestInputFiles(execBootstrapFrom, 6, 7, execBootstrapPriority, exit2)
	createLimitlessTestInputFiles(execBootstrapFrom, 7, 8, execBootstrapPriority, exit2)
	createLimitlessTestInputFiles(execBootstrapFrom, 8, 9, execBootstrapPriority, exit10)
	createLimitlessTestInputFiles(execBootstrapFrom, 9, 10, execBootstrapPriority, exit12)

	// GL
	createLimitlessTestInputFiles(execGLFrom, 0, 1, execGLPriority, exit0)
	createLimitlessTestInputFiles(execGLFrom, 1, 2, execGLPriority, exit12, forLarge)
	createLimitlessTestInputFiles(execGLFrom, 2, 3, execGLPriority, exit77)
	createLimitlessTestInputFiles(execGLFrom, 3, 4, execGLPriority, exit77, forLarge)
	createLimitlessTestInputFiles(execGLFrom, 4, 5, execGLPriority, exit137)
	createLimitlessTestInputFiles(execGLFrom, 5, 6, execGLPriority, exit137, forLarge)
	createLimitlessTestInputFiles(execGLFrom, 6, 7, execGLPriority, exit2)
	createLimitlessTestInputFiles(execGLFrom, 7, 8, execGLPriority, exit2)
	createLimitlessTestInputFiles(execGLFrom, 8, 9, execGLPriority, exit10)
	createLimitlessTestInputFiles(execGLFrom, 9, 10, execGLPriority, exit12)

	// RndBeacon
	createLimitlessTestInputFiles(execRndBeaconFrom, 0, 1, execRndBeaconPriority, exit0)
	createLimitlessTestInputFiles(execRndBeaconFrom, 1, 2, execRndBeaconPriority, exit12, forLarge)
	createLimitlessTestInputFiles(execRndBeaconFrom, 2, 3, execRndBeaconPriority, exit77)
	createLimitlessTestInputFiles(execRndBeaconFrom, 3, 4, execRndBeaconPriority, exit77, forLarge)
	createLimitlessTestInputFiles(execRndBeaconFrom, 4, 5, execRndBeaconPriority, exit137)
	createLimitlessTestInputFiles(execRndBeaconFrom, 5, 6, execRndBeaconPriority, exit137, forLarge)
	createLimitlessTestInputFiles(execRndBeaconFrom, 6, 7, execRndBeaconPriority, exit2)
	createLimitlessTestInputFiles(execRndBeaconFrom, 7, 8, execRndBeaconPriority, exit2)
	createLimitlessTestInputFiles(execRndBeaconFrom, 8, 9, execRndBeaconPriority, exit10)
	createLimitlessTestInputFiles(execRndBeaconFrom, 9, 10, execRndBeaconPriority, exit12)

	// LPP
	createLimitlessTestInputFiles(execLPPFrom, 0, 1, execLPPPriority, exit0)
	createLimitlessTestInputFiles(execLPPFrom, 1, 2, execLPPPriority, exit12, forLarge)
	createLimitlessTestInputFiles(execLPPFrom, 2, 3, execLPPPriority, exit77)
	createLimitlessTestInputFiles(execLPPFrom, 3, 4, execLPPPriority, exit77, forLarge)
	createLimitlessTestInputFiles(execLPPFrom, 4, 5, execLPPPriority, exit137)
	createLimitlessTestInputFiles(execLPPFrom, 5, 6, execLPPPriority, exit137, forLarge)
	createLimitlessTestInputFiles(execLPPFrom, 6, 7, execLPPPriority, exit2)
	createLimitlessTestInputFiles(execLPPFrom, 7, 8, execLPPPriority, exit2)
	createLimitlessTestInputFiles(execLPPFrom, 8, 9, execLPPPriority, exit10)
	createLimitlessTestInputFiles(execLPPFrom, 9, 10, execLPPPriority, exit12)

	// Conglomeration
	createLimitlessTestInputFiles(execConglomerationFrom, 0, 1, execConglomerationPriority, exit0)
	createLimitlessTestInputFiles(execConglomerationFrom, 1, 2, execConglomerationPriority, exit12, forLarge)
	createLimitlessTestInputFiles(execConglomerationFrom, 2, 3, execConglomerationPriority, exit77)
	createLimitlessTestInputFiles(execConglomerationFrom, 3, 4, execConglomerationPriority, exit77, forLarge)
	createLimitlessTestInputFiles(execConglomerationFrom, 4, 5, execConglomerationPriority, exit137)
	createLimitlessTestInputFiles(execConglomerationFrom, 5, 6, execConglomerationPriority, exit137, forLarge)
	createLimitlessTestInputFiles(execConglomerationFrom, 6, 7, execConglomerationPriority, exit2)
	createLimitlessTestInputFiles(execConglomerationFrom, 7, 8, execConglomerationPriority, exit2)
	createLimitlessTestInputFiles(execConglomerationFrom, 8, 9, execConglomerationPriority, exit10)
	createLimitlessTestInputFiles(execConglomerationFrom, 9, 10, execConglomerationPriority, exit12)

	ctxM, stopM := context.WithCancel(context.Background())
	ctxL, stopL := context.WithCancel(context.Background())

	go runController(ctxM, confM)
	go runController(ctxL, confL)

	// For DEBUG TEST mode only
	// runController(ctxL, confL)

	// Wait for a few secs, for the test to complete
	<-time.After(4 * time.Second)

	// Shutdown the controller
	stopM()
	stopL()

	expectedStructure := []struct {
		Path    string
		Entries []string
	}{
		// Bootstrap
		{
			Path:    confL.ExecBootstrap.DirFrom(0),
			Entries: []string{}, // all files should be processed
		},
		{
			Path: confL.ExecBootstrap.DirDone(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof.json.success",
				"1-2-etv0.1.2-stv1.2.3-getZkProof.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof.json.failure.code_67",
				"3-4-etv0.1.2-stv1.2.3-getZkProof.json.large.failure.code_65",
				"4-5-etv0.1.2-stv1.2.3-getZkProof.json.large.failure.code_125",
				"5-6-etv0.1.2-stv1.2.3-getZkProof.json.large.failure.code_125",
				"6-7-etv0.1.2-stv1.2.3-getZkProof.json.failure.code_2",
				"7-8-etv0.1.2-stv1.2.3-getZkProof.json.failure.code_2",
				"8-9-etv0.1.2-stv1.2.3-getZkProof.json.success",
				"9-10-etv0.1.2-stv1.2.3-getZkProof.json.large.success",
			},
		},
		{
			Path: confL.ExecBootstrap.DirTo(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json",
			},
		},
		{
			Path: confL.ExecBootstrap.DirTo(1),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json",
			},
		},

		// GL
		{
			Path:    confL.ExecGL.DirFrom(0),
			Entries: []string{}, // all files should be processed
		},
		{
			Path: confL.ExecGL.DirDone(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.success",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_67",
				"3-4-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_65",
				"4-5-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_125",
				"5-6-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.failure.code_125",
				"6-7-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_2",
				"7-8-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.failure.code_2",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.success",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_GLSubmodule.json.large.success",
			},
		},
		{
			Path: confL.ExecGL.DirTo(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json",
			},
		},
		{
			Path: confL.ExecGL.DirTo(1),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_GL.json",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_GL.json",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_GL.json",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_GL.json",
			},
		},

		// RndBeacon
		{
			Path:    confL.ExecRndBeacon.DirFrom(0),
			Entries: []string{}, // all files should be processed
		},
		{
			Path:    confL.ExecRndBeacon.DirFrom(1),
			Entries: []string{}, // all files should be processed
		},
		{
			Path: confL.ExecRndBeacon.DirDone(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_67",
				"3-4-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_65",
				"4-5-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_125",
				"5-6-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_125",
				"6-7-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2",
				"7-8-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success",
			},
		},
		{
			Path: confL.ExecRndBeacon.DirDone(1),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.success",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_67",
				"3-4-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_65",
				"4-5-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_125",
				"5-6-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.large.failure.code_125",
				"6-7-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2",
				"7-8-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.failure.code_2",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.success",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_GL_RndBeacon.json.large.success",
			},
		},
		{
			Path: confL.ExecRndBeacon.DirTo(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json",
			},
		},

		// LPP
		{
			Path:    confL.ExecLPP.DirFrom(0),
			Entries: []string{}, // all files should be processed
		},
		{
			Path: confL.ExecLPP.DirDone(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.success",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.failure.code_67",
				"3-4-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_65",
				"4-5-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_125",
				"5-6-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.large.failure.code_125",
				"6-7-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.failure.code_2",
				"7-8-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.failure.code_2",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.success",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_RndBeacon.json.large.success",
			},
		},
		{
			Path: confL.ExecLPP.DirTo(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_LPP.json",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_LPP.json",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_LPP.json",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_LPP.json",
			},
		},

		// Conglomeration
		// (Assumed no of segments=1)
		{
			Path:    confL.ExecConglomeration.DirFrom(0),
			Entries: []string{}, // all files should be processed
		},
		{
			Path:    confL.ExecConglomeration.DirFrom(1),
			Entries: []string{}, // all files should be processed
		},
		{
			Path:    confL.ExecConglomeration.DirFrom(2),
			Entries: []string{}, // all files should be processed
		},
		{
			Path: confL.ExecConglomeration.DirDone(0),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_67",
				"3-4-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_65",
				"4-5-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_125",
				"5-6-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.failure.code_125",
				"6-7-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2",
				"7-8-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.failure.code_2",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.success",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_Bootstrap_DistMetadata.json.large.success",
			},
		},
		{
			Path: confL.ExecConglomeration.DirDone(1),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_GL.json.success",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_GL.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof_GL.json.failure.code_67",
				"3-4-etv0.1.2-stv1.2.3-getZkProof_GL.json.large.failure.code_65",
				"4-5-etv0.1.2-stv1.2.3-getZkProof_GL.json.large.failure.code_125",
				"5-6-etv0.1.2-stv1.2.3-getZkProof_GL.json.large.failure.code_125",
				"6-7-etv0.1.2-stv1.2.3-getZkProof_GL.json.failure.code_2",
				"7-8-etv0.1.2-stv1.2.3-getZkProof_GL.json.failure.code_2",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_GL.json.success",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_GL.json.large.success",
			},
		},
		{
			Path: confL.ExecConglomeration.DirDone(2),
			Entries: []string{
				"0-1-etv0.1.2-stv1.2.3-getZkProof_LPP.json.success",
				"1-2-etv0.1.2-stv1.2.3-getZkProof_LPP.json.large.success",
				"2-3-etv0.1.2-stv1.2.3-getZkProof_LPP.json.failure.code_67",
				"3-4-etv0.1.2-stv1.2.3-getZkProof_LPP.json.large.failure.code_65",
				"4-5-etv0.1.2-stv1.2.3-getZkProof_LPP.json.large.failure.code_125",
				"5-6-etv0.1.2-stv1.2.3-getZkProof_LPP.json.large.failure.code_125",
				"6-7-etv0.1.2-stv1.2.3-getZkProof_LPP.json.failure.code_2",
				"7-8-etv0.1.2-stv1.2.3-getZkProof_LPP.json.failure.code_2",
				"8-9-etv0.1.2-stv1.2.3-getZkProof_LPP.json.success",
				"9-10-etv0.1.2-stv1.2.3-getZkProof_LPP.json.large.success",
			},
		},
		{
			Path: confM.ExecConglomeration.DirTo(0),
			Entries: []string{
				"0-1-getZkProof.json",
				"1-2-getZkProof.json",
				"8-9-getZkProof.json",
				"9-10-getZkProof.json",
			},
		},
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
