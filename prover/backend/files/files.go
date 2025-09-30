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

	"github.com/fsnotify/fsnotify"
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

// WaitForAllFilesAtPath waits until all expected files exist or the context is done.
// It returns nil on success, or ctx.Err() / watcher error otherwise.
func WaitForAllFilesAtPath(ctx context.Context, files []string, reportMissing bool, msg string) error {

	logrus.Infoln(msg)

	// Map of expected files
	expected := make(map[string]bool)
	dirs := make(map[string]struct{})
	for _, f := range files {
		expected[f] = false
		dirs[filepath.Dir(f)] = struct{}{}
	}

	// Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	for dir := range dirs {
		if err := watcher.Add(dir); err != nil {
			return fmt.Errorf("adding watch on %s: %w", dir, err)
		}
	}

	// Initial scan
	total, count := len(files), 0
	for f := range expected {
		if _, err := os.Stat(f); err == nil {
			expected[f] = true
			count++
			logrus.Infof("found:%s", f)
			logrus.Infof("remaining files:%d", total-count)
		}
	}
	if count == len(expected) {
		return nil
	}

	done := make(chan struct{})

	// Run a simple event loop (watch loop)
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					if _, need := expected[event.Name]; need && !expected[event.Name] {
						expected[event.Name] = true
						count++
						logrus.Infof("found:%s", event.Name)
						logrus.Infof("remaining files:%d", total-count)
						if count == len(expected) {
							logrus.Infof("All %d file(s) have arrived", total)
							return
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Errorf("watcher error:%v", err)
			}
		}
	}()

	<-done

	// Did we finish because of success or timeout?
	if ctx.Err() != nil {
		if reportMissing {
			for f, ok := range expected {
				if !ok {
					logrus.Infof("missing file: %s", f)
				}
			}
		}
		return ctx.Err()
	}

	return nil
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

// MarkFiles attempts to rename each path -> path+suffix.
// Best-effort: log warnings on failure but do not return an error.
func MarkFiles(paths []string, suffix string) {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if renameErr := os.Rename(p, p+suffix); renameErr != nil {
			logrus.Warnf("could not mark %s with %s: %v", p, suffix, renameErr)
		} else {
			logrus.Infof("marked %s with %s", p, suffix)
		}
	}
}
