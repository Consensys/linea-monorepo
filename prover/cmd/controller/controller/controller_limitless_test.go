package controller

// func TestLimitlessRun(t *testing.T) {
// 	var (
// 		exit0 int = 0
// 		// exit2   int = 2
// 		// exit10  int = 10
// 		// exit12  int = 12
// 		// exit77  int = 77
// 		// exit137 int = 137
// 	)

// 	_, confL := setupLimitlessFsTest(t)

// 	// Dirs
// 	execBootstrapFrom := []string{confL.ExecBootstrap.DirFrom(0)}

// 	// Populate the filesystem with job files

// 	// Bootstrap
// 	createLimitlessTestInputFiles(execBootstrapFrom, 0, 1, Bootstrap, exit0)
// 	// createLimitlessTestInputFiles(execBootstrapFrom, 1, 2, execBootstrapPriority, exit12, forLarge)
// 	// createLimitlessTestInputFiles(execBootstrapFrom, 2, 3, execBootstrapPriority, exit77)
// 	// createLimitlessTestInputFiles(execBootstrapFrom, 3, 4, execBootstrapPriority, exit77, forLarge)
// 	// createLimitlessTestInputFiles(execBootstrapFrom, 4, 5, execBootstrapPriority, exit137)
// 	// createLimitlessTestInputFiles(execBootstrapFrom, 5, 6, execBootstrapPriority, exit137, forLarge)
// 	// createLimitlessTestInputFiles(execBootstrapFrom, 6, 7, execBootstrapPriority, exit2)
// 	// createLimitlessTestInputFiles(execBootstrapFrom, 7, 8, execBootstrapPriority, exit2)
// 	// createLimitlessTestInputFiles(execBootstrapFrom, 8, 9, execBootstrapPriority, exit10)
// 	// createLimitlessTestInputFiles(execBootstrapFrom, 9, 10, execBootstrapPriority, exit12)

// 	ctxL, stopL := context.WithCancel(context.Background())

// 	go runController(ctxL, confL)

// 	// Give one sec, for the test to complete
// 	<-time.After(4 * time.Second)

// 	// Shutdown the controller
// 	stopL()

// 	expectedStructure := []struct {
// 		Path    []string
// 		Entries [][]string
// 	}{
// 		{
// 			Path:    []string{confL.ExecBootstrap.DirFrom(0)},
// 			Entries: [][]string{}, // all files should be processed
// 		},
// 	}

// 	for _, dirVal := range expectedStructure {
// 		for i, dirPath := range dirVal.Path {
// 			dir, err := os.Open(dirPath)
// 			require.NoErrorf(t, err, "dir %v", dirPath)
// 			filesFound, err := dir.Readdirnames(-1)
// 			require.NoErrorf(t, err, "dir %v", dirPath)
// 			slices.Sort(filesFound)
// 			assert.Equalf(t, dirVal.Entries[i], filesFound, "dir %v", dirVal.Path)
// 		}
// 	}
// }
