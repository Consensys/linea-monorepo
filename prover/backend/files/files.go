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

// RemoveMatchingFiles deletes all files matching the given pattern (if exists).
// The pattern can include wildcards like "*.tmp.*" or "filename*".
func RemoveMatchingFiles(pattern string, isLog bool) (bool, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false, fmt.Errorf("glob pattern failed for %q: %w", pattern, err)
	}
	// Nothing to delete
	if len(matches) == 0 {
		if isLog {
			logrus.Infof("No file found matching pattern:%s", pattern)
		}
		return false, nil
	}

	logrus.Infof("Removing file(s) found matching pattern:%s", pattern)
	for _, file := range matches {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return true, fmt.Errorf("failed to remove %s: %w", file, err)
		}
	}
	return true, nil
}

// WaitForFileAtPath waits until file exists or context done.
// Uses fsnotify when possible but falls back to polling every pollInterval.
func WaitForFileAtPath(ctx context.Context, file string, pollInterval time.Duration, reportMissing bool, msg string) error {
	logrus.Infoln(msg)

	// Quick initial stat
	if _, err := os.Stat(file); err == nil {
		logrus.Infof("found: %s (initial stat)", file)
		return nil
	}

	pollInterval = pollInterval * time.Second
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := os.Stat(file); err == nil {
					logrus.Infof("found: %s (poll)", file)
					return
				}
			}
		}
	}()

	<-done

	if ctx.Err() != nil {
		if reportMissing {
			if _, err := os.Stat(file); err != nil {
				logrus.Infof("missing file: %s", file)
			}
		}

		return ctx.Err()
	}

	return nil
}

// DoneFilePath replaces "requests" -> "requests-done" and
// ".inprogress.<anything>" -> to given suffix (eg. ".success.parital.bootstrap")
func DoneFilePath(inProgressPath, suffix string) string {
	dir := filepath.Dir(inProgressPath)
	base := filepath.Base(inProgressPath)

	// Change dir from "requests" → "requests-done"
	doneDir := strings.Replace(dir, "requests", "requests-done", 1)

	// Strip everything after ".inprogress.*" and add the suffix
	if idx := strings.Index(base, config.InProgressSufix); idx != -1 {
		base = base[:idx] + suffix
	}

	return filepath.Join(doneDir, base)
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
		return ".success"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ".timeout"
	}
	return ".failure"
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
