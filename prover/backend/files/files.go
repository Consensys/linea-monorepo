package files

import (
	"compress/gzip"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

// Wrapper for os.File allowing to close cleanly
type ZipFile struct {
	f *os.File
	*gzip.Writer
	*gzip.Reader
}

// opens an existing file with read access
func MustRead(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		logrus.Panicf("error opening %v - %v", path, err)
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
		logrus.Panicf("could not create directory - %v, error: %v, filepath: %v", dir, err, p)
	}

	/*
		Create or overwrite a file
	*/
	f, err := os.Create(p)
	if err != nil {
		logrus.Panicf("could not create directory - %v, error: %v, filepath: %v", dir, err, p)
	}

	return f
}

// Read a .gz archive or panic
func MustReadCompressed(path string) *ZipFile {
	f := MustRead(path)
	unzipped, err := gzip.NewReader(f)
	if err != nil {
		logrus.Panicf("error decompressing %v - %v", path, err)
	}

	return &ZipFile{
		f:      f,
		Reader: unzipped,
	}
}

// Write into a .gz file or panic
func MustWriteCompressed(p string) *ZipFile {
	f := MustOverwrite(p)
	return &ZipFile{
		f:      f,
		Writer: gzip.NewWriter(f),
	}
}

// Close unzipped file
func (z *ZipFile) Close() {
	// Close the writer if any
	if z.Writer != nil {
		// For some reason, forgetting to call `Flush` here causes
		// `EOF`` errors.
		if err := z.Writer.Flush(); err != nil {
			panic(err)
		}

		if err := z.Writer.Close(); err != nil {
			panic(err)
		}
	}

	// Close the reader if any
	if z.Reader != nil {
		if err := z.Reader.Close(); err != nil {
			panic(err)
		}
	}

	// Close the file
	if err := z.f.Close(); err != nil {
		panic(err)
	}
}
