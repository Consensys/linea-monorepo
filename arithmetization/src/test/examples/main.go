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
// segment are added only when needed to reproduce the loader memory image over
// non-zero RAM. In case we can assume the RAM to be zero-initialized, we can skip the gaps between sections
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
		for _, s := range sections {
			if s.Size == 0 || s.Flags&elf.SHF_ALLOC == 0 {
				continue
			}
			sectionEnd := s.Addr + s.Size
			if s.Addr < p.Vaddr || sectionEnd > progEnd {
				continue
			}
			sectionBlobs = append(sectionBlobs, memoryBlob{offset: s.Addr, data: readSectionBytes(s), name: s.Name})
		}
		sort.Slice(sectionBlobs, func(i, j int) bool { return sectionBlobs[i].offset < sectionBlobs[j].offset })

		cursor := p.Vaddr
		for _, blob := range sectionBlobs {
			if cursor < blob.offset {
				data := readProgBytes(p, cursor-p.Vaddr, blob.offset-cursor)
				blobs = append(blobs, memoryBlob{
					offset: cursor,
					data:   data,
					name:   gapName(data),
				})
			}
			blobs = append(blobs, blob)
			cursor = blob.offset + uint64(len(blob.data))
		}
		if cursor < progEnd {
			data := readProgBytes(p, cursor-p.Vaddr, progEnd-cursor)
			blobs = append(blobs, memoryBlob{
				offset: cursor,
				data:   data,
				name:   gapName(data),
			})
		}
	}

	if len(blobs) == 0 {
		panic("no loadable program sections found.")
	}

	return blobs
}

func readProgBytes(p *elf.Prog, offset, size uint64) []byte {
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

func readSectionBytes(s *elf.Section) []byte {
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

func gapName(data []byte) string {
	for _, b := range data {
		if b != 0 {
			return "load gap"
		}
	}
	return "zero gap"
}

func writeSectionsFile(file *os.File, blobs []memoryBlob) {
	fmt.Fprintln(file, "# index offset size name")
	for i, blob := range blobs {
		fmt.Fprintf(file, "%d 0x%016x 0x%016x %s\n", i, blob.offset, len(blob.data), blob.name)
	}
}

func printJson(blobs []memoryBlob, entryPoint uint64) {
	var (
		entryPointString   = fmt.Sprintf("%016x", entryPoint)
		blobsCountString   = fmt.Sprintf("%016x", len(blobs))
		entryPointAndBlobs = entryPointString + blobsCountString
		blobMetadata       []string
		blobData           []string
	)

	for _, blob := range blobs {
		blobMetadata = append(blobMetadata, fmt.Sprintf("%016x%016x", blob.offset, len(blob.data)))
		if len(blob.data) > 0 {
			blobData = append(blobData, hex.EncodeToString(blob.data))
		}
	}

	fmt.Println("{")
	fmt.Printf("\t\"%s\": \"0x%s\",\n", ENTRY_POINT_AND_BLOBS_COUNT, entryPointAndBlobs)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", BLOBS_OFFSET_AND_SIZE, strings.Join(blobMetadata, "_"))
	fmt.Printf("\t\"%s\": \"0x%s\"\n", BLOBS_DATA, strings.Join(blobData, "_"))
	fmt.Println("}")
}
