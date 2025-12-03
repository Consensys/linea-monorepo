package files

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/consensys/linea-monorepo/prover/config"
	"github.com/sirupsen/logrus"
)

// --------- Helper ----

// parseSbrEbr extracts sbr/ebr from names like 22504197-22504198-...-getZkProof.json
var (
	regReq = regexp.MustCompile(
		`(^|.*/)(\d+)-(\d+)-.*-(getZkProof|getZkBlobCompressionProof|getZkAggregatedProof)\.json.*$`,
	)

	regWitness = regexp.MustCompile(
		`(^|.*/)(\d+)-(\d+)-seg-(\d+)-mod-(\d+)-(gl|lpp)-wit\.bin.*$`,
	)
)

func ParseReqFile(reqFilePath string) (sbr, ebr string, _ error) {
	m := regReq.FindStringSubmatch(reqFilePath)
	if m == nil {
		return "", "", fmt.Errorf("unable to parse sbr/ebr from %s", reqFilePath)
	}
	return m[2], m[3], nil
}

func ParseWitnessFile(filePath string) (sb, eb string, segID int, err error) {
	m := regWitness.FindStringSubmatch(filePath)
	if m == nil {
		return "", "", 0, fmt.Errorf("unable to parse sb/eb/segID from %s", filePath)
	}

	sb = m[2]
	eb = m[3]

	segID, err = strconv.Atoi(m[4])
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid segID in %s: %w", filePath, err)
	}

	return sb, eb, segID, nil
}

// CheckFilePath checks whether the provided filePath points to an existing file.
func CheckFilePath(filePath string) error {
	// Use os.Stat to get information about the file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		// this error is either a os.ErrNotExist or something else if the Stat function failed.
		return err
	}

	// Check if it's a regular file (not a directory)
	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("%q is not a file", filePath)
	}

	return nil
}

// CheckDirPath checks whether the provided dirPath points to an existing directory.
func CheckDirPath(dirPath string) error {
	// Use os.Stat to get information about the directory
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		// this error is either a os.ErrNotExist or something else if the Stat function failed.
		return err
	}

	// Check if it's a directory
	if !fileInfo.IsDir() {
		return fmt.Errorf("%q is not a directory", dirPath)
	}

	return nil
}

// TODO @gbotrel most of this "MustXXX" functions must go away and be replaced by proper error handling

// opens an existing file with read access
func MustRead(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		logrus.Panicf("Error opening %v - %v", path, err)
	}

	return f
}

// create a file and the parent's folder if necessary. Overwrite the file
// if it already exists.
func MustOverwrite(p string) *os.File {

	/*
		Ensures the parent directory exists
	*/
	dir := path.Dir(p)
	err := os.MkdirAll(dir, 0770)
	if err != nil {
		logrus.Panicf("Could not create directory - %v, error: %v, filepath: %v", dir, err, p)
	}

	/*
		Create or overwrite a file
	*/
	f, err := os.Create(p)
	if err != nil {
		logrus.Panicf("Could not create directory - %v, error: %v, filepath: %v", dir, err, p)
	}

	return f
}

// Wrapper for os.File allowing to close cleanly
type ZipFile struct {
	f *os.File
	*gzip.Writer
	*gzip.Reader
}

// Read a .gz archive or panic
func MustReadCompressed(path string) *ZipFile {
	f := MustRead(path)
	unzipped, err := gzip.NewReader(f)
	if err != nil {
		logrus.Panicf("Error decompressing %v - %v", path, err)
	}

	return &ZipFile{
		f:      f,
		Reader: unzipped,
	}
}

// Close unzipped file
func (z *ZipFile) Close() error {
	// Close the writer if any
	if z.Writer != nil {
		// For some reason, forgetting to call `Flush` here causes
		// `EOF`` errors.
		if err := z.Writer.Flush(); err != nil {
			return err
		}

		if err := z.Writer.Close(); err != nil {
			return err
		}
	}

	// Close the reader if any
	if z.Reader != nil {
		if err := z.Reader.Close(); err != nil {
			return err
		}
	}

	// Close the file
	if err := z.f.Close(); err != nil {
		return err
	}

	return nil
}

// WaitForFileAtPath : Polls every configured time interval for the file and
// waits until the file is found or context is done.
func WaitForFileAtPath(ctx context.Context, file string, pollInterval time.Duration, reportMissing bool, msg string) error {
	logrus.Infoln(msg)

	// Quick initial stat
	if _, err := os.Stat(file); err == nil {
		logrus.Infof("found (initial stat): %s ", file)
		return nil
	}

	pollInterval = pollInterval * time.Second
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	foundCh := make(chan struct{})
	go func() {
		defer close(foundCh)
		for {
			select {
			case <-ctx.Done():
				return // context cancelled
			case <-ticker.C:
				if _, err := os.Stat(file); err == nil {
					logrus.Infof("found (poll): %s ", file)
					return // file found
				}
			}
		}
	}()

	// Wait for goroutine to finish
	<-foundCh

	// Distinguish reason for exit
	if _, err := os.Stat(file); err == nil {
		return nil // file is actually there → definite success
	}

	// file not there → the only remaining reason is context cancellation
	if ctx.Err() != nil {
		if reportMissing {
			logrus.Infof("missing file: %s", file)
		}
		return ctx.Err()
	}

	// fallback (should never happen, but safe)
	return fmt.Errorf("unexpected exit: file %s not found", file)
}

// ReadRequest reads and decodes a request from a file
func ReadRequest(path string, into any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(into); err != nil {
		return fmt.Errorf("could not decode input file: %w", err)
	}

	return nil
}

// OutcomeSuffix maps an error to a suffix used for marking files.
// - nil -> ".success"
// - context.DeadlineExceeded -> ".timeout"
// - otherwise -> ".failure"
func OutcomeSuffix(err error) string {
	if err == nil {
		return "." + config.SuccessSuffix
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "." + config.TimeoutSuffix
	}
	// requires manual investigation
	return ".failure." + config.FailSuffix + "_2"
}

// MarkAndMoveToDone first attempts to rename each path -> path+suffix,
// then moves the marked file into ".../requests-done/filename+suffix".
// - Marking is best-effort: warnings are logged, no error returned.
// - Moving is strict: if any move fails, the function returns an error.
func MarkAndMoveToDone(cfg *config.Config, filePaths []string, suffix string) error {
	for _, filePath := range filePaths {
		if filePath == "" {
			continue
		}

		// Step 1: Mark (best-effort)
		markedPath := filePath + suffix
		if err := os.Rename(filePath, markedPath); err != nil {
			logrus.Warnf("could not mark %s with %s: %v", filePath, suffix, err)
			// if marking fails, skip moving since the file doesn’t exist under marked name
			continue
		} else {
			logrus.Infof("marked %s with %s", filePath, suffix)
		}

		// Step 2: Move to requests-done
		dir := filepath.Dir(markedPath)
		base := filepath.Base(markedPath)

		if !strings.Contains(dir, config.RequestsFromSubDir) {
			return fmt.Errorf("path %q does not contain '%s'", markedPath, config.RequestsFromSubDir)
		}
		doneDir := strings.Replace(dir, config.RequestsFromSubDir, config.RequestsDoneSubDir, 1)

		if err := os.MkdirAll(doneDir, 0o755); err != nil {
			return fmt.Errorf("failed to create done dir: %w", err)
		}

		dest := filepath.Join(doneDir, base)
		if err := os.Rename(markedPath, dest); err != nil {
			return fmt.Errorf("failed to move %q to %q: %w", markedPath, dest, err)
		}

		logrus.Infof("moved %s to %s", markedPath, dest)
	}

	return nil
}

// globMatches returns the list of filepaths matching the given pattern.
// It wraps filepath.Glob but gives consistent error formatting.
func globMatches(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob failed for pattern %q: %w", pattern, err)
	}
	return matches, nil
}

// HasWildcardMatch returns true if at least one file whose name starts with
// basePattern exists in the directory.
func HasWildcardMatch(dir, basePattern string) (bool, error) {
	pattern := filepath.Join(dir, basePattern)

	matches, err := globMatches(pattern)
	if err != nil {
		return false, err
	}

	return len(matches) > 0, nil
}

// RemoveMatchingFiles deletes all files matching the given pattern.
// Returns (true, nil) if any were deleted, (false, nil) if none matched.
func RemoveMatchingFiles(pattern string, log bool) (bool, error) {
	matches, err := globMatches(pattern)
	if err != nil {
		return false, err
	}

	if len(matches) == 0 {
		if log {
			logrus.Infof("No file found matching pattern: %s", pattern)
		}
		return false, nil
	}

	if log {
		logrus.Infof("Removing %d file(s) matching pattern: %s", len(matches), pattern)
	}

	for _, f := range matches {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			return true, fmt.Errorf("failed to remove %s: %w", f, err)
		}
	}

	return true, nil
}

// LocateBaseBySuffix searches the given directory for a single file
// belonging to a specific job whose name contains a given suffix. Filenames
// follow the pattern:
//
//	<start>-<end>-*.suffix[...]
//
// Example match:
//
//	10889556-10889655-etv2.1.0-stv2.3.0-getZkProof.json.inprogress.conglomerator
//
// The function performs:
//
//  1. Construct a glob pattern using start/end and the given suffix.
//  2. Ensure exactly one filename matches.
//  3. Find the first occurrence of "." + suffix.
//  4. Return both:
//     - baseFile: the portion of the filename before the suffix
//     - oldFile:  the full path of the matched file
//
// baseFile is what you should append a new suffix to when finalizing status.
//
// Returns an error if:
//   - Glob fails
//   - No matching file exists
//   - More than one file matches
//   - The suffix is not found in the filename
func LocateBaseBySuffix(start, end int, dir, suffix string) (baseFile string, oldFile string, err error) {

	// Build glob pattern: "<dir>/<start>-<end>-*.suffix*"
	pattern := filepath.Join(dir, fmt.Sprintf("%d-%d-*.%s*", start, end, suffix))

	matches, err := globMatches(pattern)
	if err != nil {
		return "", "", err
	}

	switch len(matches) {
	case 0:
		return "", "", fmt.Errorf("no matches for pattern %s", pattern)
	case 1:
	default:
		return "", "", fmt.Errorf("multiple matches for pattern %s: %#v", pattern, matches)
	}

	oldFile = matches[0]

	// find index of ".suffix" inside the filename
	marker := "." + suffix
	idx := strings.Index(oldFile, marker)
	if idx == -1 {
		return "", "", fmt.Errorf("unexpected filename format: %s (missing %s)", oldFile, marker)
	}

	// baseFile = everything before ".suffix"
	return oldFile[:idx], oldFile, nil
}
