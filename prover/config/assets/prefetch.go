package assets

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type PrefetchOptions struct {
	Parallelism        int   // 0 -> runtime.GOMAXPROCS(0)
	ChunkSize          int64 // bytes chunk for mmap fallback, e.g. 64<<20
	LargeFileThreshold int64 // bytes; if file >= threshold, only fadvise is used
	PerFileTimeout     time.Duration
}

// Default ChunkSize: 64MiB
// Default LargeFileThreshold: 10GiB
func DefaultPrefetchOptions() *PrefetchOptions {
	return &PrefetchOptions{
		Parallelism:        4,
		ChunkSize:          64 << 20, // 64 MiB
		LargeFileThreshold: 10 << 30, // 10 GiB
		PerFileTimeout:     90 * time.Second,
	}
}

// PrefetchFiles warms the kernel page cache for the given paths.
// It uses posix_fadvise(FADV_WILLNEED) first and falls back to chunked mmap + touch.
// ctx can cancel the prefetching.
func PrefetchFiles(ctx context.Context, paths []string, opts *PrefetchOptions) error {
	if len(paths) == 0 {
		return nil
	}
	if opts == nil {
		opts = DefaultPrefetchOptions()
	}
	if opts.Parallelism <= 0 {
		opts.Parallelism = runtime.GOMAXPROCS(0)
	}
	type job struct{ path string }
	type res struct {
		path string
		err  error
	}
	jobs := make(chan job)
	results := make(chan res)

	var wg sync.WaitGroup
	for i := 0; i < opts.Parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				rctx := ctx
				var cancel func()
				if opts.PerFileTimeout > 0 {
					rctx, cancel = context.WithTimeout(ctx, opts.PerFileTimeout)
				}
				err := prefetchSingle(rctx, j.path, opts)
				if cancel != nil {
					cancel()
				}
				select {
				case results <- res{path: j.path, err: err}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
	outer1:
		for _, p := range paths {
			select {
			case jobs <- job{path: p}:
			case <-ctx.Done():
				// break out of the loop entirely if context canceled
				break outer1
			}
		}
		close(jobs)
	}()

	var aggErr error
	expected := len(paths)
outer2:
	for i := 0; i < expected; i++ {
		select {
		case r := <-results:
			if r.err != nil {
				if aggErr == nil {
					aggErr = fmt.Errorf("%s: %w", r.path, r.err)
				} else {
					aggErr = fmt.Errorf("%v; %s: %w", aggErr, r.path, r.err)
				}
			}
		case <-ctx.Done():
			aggErr = fmt.Errorf("prefetch canceled")
			// break out of the loop entirely if context canceled
			break outer2
		}
	}
	wg.Wait()
	close(results)
	return aggErr
}

// prefetchSingle is a best-effort file prefetcher and warms the OS page cache for a given file
// It requests OS(Linux) to read the file’s data into memory now, so that later reads will hit RAM instead of disk
func prefetchSingle(ctx context.Context, path string, opts *PrefetchOptions) error {
	if path == "" {
		return errors.New("empty path")
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	size := fi.Size()
	if size == 0 {
		return nil
	}

	fd := int(f.Fd())

	// Try FADV_WILLNEED. Best-effort: The underscore _ = ignores any error because it’s best-effort:
	// some filesystems (or container environments) may not support it.
	// This tells the kernel: the user plans to read this file soon and requests reading it into page cache in background
	// It is asynchronous and non-blocking — it just hints to the kernel.
	// Effect: the kernel may start prefetching pages in background via readahead.
	_ = unix.Fadvise(fd, 0, 0, unix.FADV_WILLNEED)

	// If file is very large, avoid doing mapping-touch in controller; rely on fadvise
	if opts.LargeFileThreshold > 0 && size >= opts.LargeFileThreshold {
		// Return successfully even if fadvise failed; child will fetch pages on demand.
		return nil
	}

	// Fallback: chunked mmap + touch pages
	chunk := opts.ChunkSize
	if chunk <= 0 {
		chunk = 64 << 20 // default to 64MIB
	}
	page := int64(os.Getpagesize())

	for offset := int64(0); offset < size; offset += chunk {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := offset + chunk
		if end > size {
			end = size
		}
		length := int(end - offset)

		// map must be page-aligned: align offset down to page boundary
		aligned := (offset / page) * page
		prefix := offset - aligned
		// map length must cover prefix + length
		mapLen := int(prefix) + length
		if mapLen <= 0 {
			continue
		}

		// Map a segment of the file into the process’s address space (read-only).
		// OS won’t actually read data yet — it will lazily page-fault when you touch those pages.
		data, err := unix.Mmap(fd, aligned, mapLen, unix.PROT_READ, unix.MAP_SHARED)
		if err != nil {
			return fmt.Errorf("m-map fallback failed for %s offset %d len %d: %w", path, aligned, mapLen, err)
		}

		// Touch pages by reading one byte per page starting at prefix
		// Accessing one byte per page forces the kernel to fault in that page (read it from disk into memory).
		// Doing this once per page is enough to bring the entire chunk into page cache.
		// It doesn’t copy or store data — just reads, so it’s fast (bounded by disk speed).
		for i := int64(prefix); i < int64(mapLen); i += page {
			_ = data[i]
		}

		// Releases the mapped region from process address space.
		// The data remains cached in the kernel, so future reads are still fast.
		if err := unix.Munmap(data); err != nil {
			return fmt.Errorf("m-unmap failed for %s: %w", path, err)
		}
	}
	return nil
}
