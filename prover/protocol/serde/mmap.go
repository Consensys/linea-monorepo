package serde

import (
	"os"
	"syscall"
)

func Open(path string) (*View, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, _ := f.Stat()
	size := int(stat.Size())

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return NewView(data)
}
