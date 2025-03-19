package wizard

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/consensys/linea-monorepo/prover/backend/files"
	"github.com/sirupsen/logrus"
)

// artefactDir is the directory used to store the artefacts. The directory is
// hardcoded into `/tmp` this means that when the prover is started, the artefacts
// will not be present and the first run of the prover will regenerate the artefacts
// so that the following runs can have access to it and instead of wasting time
// regenerating the artefacts.
var artefactDir string = "/tmp/prover-artefacts"

// Artefact is an ad-hoc interface characterizing serializable objects. The
// interface should be implemented over a pointer type as it is used for reading
// the object from a blob of bytes.
type Artefact interface {
	io.ReaderFrom
	io.WriterTo
}

// artefactCache is a generic data-store that can be used to serialize
// compilation data. The implementation is file-based, meaning that the artefacts
// are written on the filesystem.
//
// The artefacts are dumped/read from the directory indicated by [artefactDir]
// and the corresponding filenames match the artefacts keys.
type artefactCache struct{}

// TryLoad attempts finding a key. The boolean indicates whether the corresponding
// file was found and the error indicates whether the file was successfully read.
func (a artefactCache) TryLoad(key string, obj Artefact) (found bool, parseErr error) {

	var (
		fpath     = path.Join(artefactDir, key)
		fCheckErr = files.CheckFilePath(fpath)
	)

	if errors.Is(fCheckErr, os.ErrNotExist) {
		logrus.Infof("attempted to open the cache-key=%v, was missing", fpath)
		return false, nil
	}

	if fCheckErr != nil {
		// This can happen if the directory does not exists
		logrus.Infof("attempted to open the cache-key=%v err=%v", fpath, fCheckErr.Error())
		return false, fmt.Errorf("CheckFilePath failed: %w", fCheckErr)
	}

	f, readErr := os.Open(fpath)

	if readErr != nil {
		logrus.Infof("attempted to open the cache-key=%v err=read-file-failed:%v", fpath, readErr.Error())
		return false, fmt.Errorf("ReadFile failed: %w", readErr)
	}

	_, parseErr = obj.ReadFrom(f)

	if parseErr != nil {
		logrus.Infof("attempted to open the cache-key=%v err=read-from-failed:%v", fpath, parseErr.Error())
		return false, fmt.Errorf("ReadFrom failed: %w", parseErr)
	}

	logrus.Debugf("cache-key found cache-key=%v", fpath)

	return true, nil
}

// Store stores a new object in the cache. It will return an error if the file
// already exists.
func (a artefactCache) Store(key string, obj Artefact) error {

	var (
		fpath       = path.Join(artefactDir, key)
		writingPath = fpath + ".tmp"
		statErr     = files.CheckFilePath(writingPath)
	)

	if statErr == nil {
		return fmt.Errorf("the file %q already exists", fpath)
	}

	logrus.Infof("Started writing the global constraint in the cache")
	defer logrus.Infof("Done writing the global constraint in the cache")

	f := files.MustOverwrite(writingPath)
	if _, writeErr := obj.WriteTo(f); writeErr != nil {
		return fmt.Errorf("error writing artefact in file %q : %w", writingPath, writeErr)
	}
	f.Close()

	// This is to attempt cleaning. The error handing should be safe.
	defer os.Remove(writingPath)

	if mvErr := os.Rename(writingPath, fpath); mvErr != nil {
		return fmt.Errorf("could not rename %q into %q : %w", writingPath, fpath, mvErr)
	}

	return nil
}
