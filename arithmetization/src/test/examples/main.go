package main

import (
	"debug/elf"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	ENTRY_POINT_AND_BLOBS_COUNT = "entry_point_and_blobs_count"
	BLOBS_OFFSET_AND_SIZE       = "blobs_offset_and_size"
	BLOBS_DATA                  = "blobs_data"
)

type memoryBlob struct {
	offset uint64
	data   []byte
	name   string
}

// The purpose of this program is simply to generate a suitable ZkC json input
// file for a given RISC-V binary program.
func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: go run main.go <elfFile> <inBytes> <inBytesOffset>")
		os.Exit(1)
	}

	elfFile, err := elf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening ELF file: %v\n", err)
		os.Exit(1)
	}
	defer elfFile.Close()
	// Parse inBytes
	var inBytes []byte
	inBytesString := os.Args[2]
	if strings.HasPrefix(inBytesString, "0x") || strings.HasPrefix(inBytesString, "0X") {
		inBytes, err = hex.DecodeString(inBytesString[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error decoding hex input bytes: %v\n", err)
			os.Exit(1)
		}
	} else {
		inBytes = []byte(inBytesString)
	}
	// Parse inBytesOffset
	var inBytesOffset uint64
	inBytesOffset, err = strconv.ParseUint(os.Args[3], 0, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input bytes offset: %v\n", err)
		os.Exit(1)
	}
	// The entry point, program blob offsets and program blob sizes are taken
	// directly from the ELF. Only the optional input bytes offset is external.
	var blobs = extractProgramBlobs(elfFile.Progs, elfFile.Sections)
	if len(inBytes) > 0 {
		blobs = append(blobs, memoryBlob{offset: inBytesOffset, data: inBytes, name: "in_bytes"})
	}
	// Optionally write a .sections file with the indexes, offsets, sizes and names of the blobs for debugging purposes.
	// This is controlled by the ELF2JSON_WRITE_SECTIONS environment variable, which must be set to "true" to enable this feature.
	switch writeSections := os.Getenv("ELF2JSON_WRITE_SECTIONS"); writeSections {
	case "", "false":
	case "true":
		sectionsFile, err := os.Create(strings.TrimSuffix(os.Args[1], ".elf") + ".sections")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating ELF sections file: %v\n", err)
			os.Exit(1)
		}
		writeSectionsFile(sectionsFile, blobs)
		if err := sectionsFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error writing ELF sections file: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "ELF2JSON_WRITE_SECTIONS must be true or false, got %q\n", writeSections)
		os.Exit(1)
	}
	printJson(blobs, elfFile.Entry)
}

// Extract sparse memory blobs from allocated sections. Gaps inside a PT_LOAD
// segment are emitted explicitly to reproduce the ELF loader memory image
// without assuming RAM is zero-initialized.
//
// Our own tests contain .text, .rodata, .data and .bss sections.
// ACT4 tests contain .text.init, .text.rvtest, .text.rvmodel, .data,
// and .tohost sections. We do not filter by section names here.
func extractProgramBlobs(progs []*elf.Prog, sections []*elf.Section) []memoryBlob {
	var blobs []memoryBlob

	for _, p := range progs {
		if p.Type != elf.PT_LOAD || p.Memsz == 0 {
			continue
		}
		if p.Filesz > p.Memsz {
			panic(fmt.Sprintf("loadable segment at %#x has file size larger than memory size", p.Vaddr))
		}

		var sectionBlobs []memoryBlob
		progEnd := p.Vaddr + p.Memsz
		if progEnd < p.Vaddr {
			panic(fmt.Sprintf("loadable segment address overflow at %#x", p.Vaddr))
		}
		for _, s := range sections {
			if s.Size == 0 || s.Flags&elf.SHF_ALLOC == 0 {
				continue
			}
			sectionEnd := s.Addr + s.Size
			if sectionEnd < s.Addr {
				panic(fmt.Sprintf("section %s address overflow at %#x", s.Name, s.Addr))
			}
			if s.Addr < p.Vaddr || sectionEnd > progEnd {
				continue
			}
			sectionBlobs = append(sectionBlobs, memoryBlob{offset: s.Addr, data: readAllocSectionBytes(s), name: s.Name})
		}
		sort.Slice(sectionBlobs, func(i, j int) bool { return sectionBlobs[i].offset < sectionBlobs[j].offset })

		appendGap := func(offset, size uint64) {
			data := readLoadSegmentBytes(p, offset-p.Vaddr, size)
			name := "zero gap"
			for _, b := range data {
				if b != 0 {
					name = "load gap"
					break
				}
			}
			blobs = append(blobs, memoryBlob{offset: offset, data: data, name: name})
		}

		cursor := p.Vaddr
		for _, blob := range sectionBlobs {
			if cursor < blob.offset {
				appendGap(cursor, blob.offset-cursor)
			}
			blobs = append(blobs, blob)
			cursor = blob.offset + uint64(len(blob.data))
		}
		if cursor < progEnd {
			appendGap(cursor, progEnd-cursor)
		}
	}

	if len(blobs) == 0 {
		panic("no loadable program sections found.")
	}

	return blobs
}

// readLoadSegmentBytes reads a byte range from a PT_LOAD segment memory image.
// Bytes past Filesz but inside Memsz are returned as zeroes, matching ELF
// loader semantics for uninitialized segment memory.
func readLoadSegmentBytes(p *elf.Prog, offset, size uint64) []byte {
	data := make([]byte, size)
	if offset < p.Filesz {
		sizeFromFile := min(size, p.Filesz-offset)
		n, err := p.ReadAt(data[:sizeFromFile], int64(offset))
		if err != nil && err != io.EOF {
			panic(fmt.Sprintf("error reading loadable segment at %#x: %v", p.Vaddr+offset, err))
		}
		if uint64(n) != sizeFromFile {
			panic(fmt.Sprintf("short read for loadable segment at %#x: got %d bytes, expected %d", p.Vaddr+offset, n, sizeFromFile))
		}
	}
	return data
}

// readAllocSectionBytes reads the bytes for an allocated ELF section. SHT_NOBITS
// sections, such as .bss, occupy memory but have no file bytes, so they are
// returned as explicit zeroes.
func readAllocSectionBytes(s *elf.Section) []byte {
	data := make([]byte, s.Size)
	if s.Type == elf.SHT_NOBITS {
		return data
	}
	n, err := s.ReadAt(data, 0)
	if err != nil && err != io.EOF {
		panic(fmt.Sprintf("error reading section %s: %v", s.Name, err))
	}
	if uint64(n) != s.Size {
		panic(fmt.Sprintf("short read for section %s: got %d bytes, expected %d", s.Name, n, s.Size))
	}
	return data
}

func writeSectionsFile(file *os.File, blobs []memoryBlob) {
	fmt.Fprintln(file, "index, offset,             size,               name")
	for i, blob := range blobs {
		fmt.Fprintf(file, "%-5d, 0x%016x, 0x%016x, %s\n", i, blob.offset, len(blob.data), blob.name)
	}
}

func printJson(blobs []memoryBlob, entryPoint uint64) {
	var (
		entryPointString   = fmt.Sprintf("%016x", entryPoint)
		blobsCountString   = fmt.Sprintf("%016x", len(blobs))
		entryPointAndBlobs = entryPointString + "_" + blobsCountString
		blobMetadata       []string
		blobData           []string
	)

	for _, blob := range blobs {
		blobMetadata = append(blobMetadata, fmt.Sprintf("%016x_%016x", blob.offset, len(blob.data)))
		if len(blob.data) > 0 {
			blobData = append(blobData, hex.EncodeToString(blob.data))
		}
	}

	fmt.Println("{")
	fmt.Printf("\t\"%s\": \"0x%s\",\n", ENTRY_POINT_AND_BLOBS_COUNT, entryPointAndBlobs)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", BLOBS_OFFSET_AND_SIZE, strings.Join(blobMetadata, "____"))
	fmt.Printf("\t\"%s\": \"0x%s\"\n", BLOBS_DATA, strings.Join(blobData, "____"))
	fmt.Println("}")
}
