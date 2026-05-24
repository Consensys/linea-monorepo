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
	RISCV_PROGRAM               = "riscV_program"
	IN_BYTES                    = "in_bytes"
	RISCV_PROGRAM_OFFSET        = "riscV_program_offset"
	IN_BYTES_OFFSET             = "in_bytes_offset"
	ENTRY_POINT                 = "entry_point"
	RISCV_PROGRAM_LENGTH        = "riscV_program_length"
	IN_BYTES_LENGTH             = "in_bytes_length"
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
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "usage: go run main.go <elfFile> <inBytes> <programOffset> <inputsOffset> <entryPoint>")
		os.Exit(1)
	}

	elfFile, err := elf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening ELF file: %v\n", err)
		os.Exit(1)
	}
	defer elfFile.Close()
	// inBytes
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
	// offsets
	var programOffset uint64
	var inputsOffset uint64
	var entryPoint uint64
	programOffset, err = strconv.ParseUint(os.Args[3], 0, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading program offset: %v\n", err)
		os.Exit(1)
	}
	inputsOffset, err = strconv.ParseUint(os.Args[4], 0, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input bytes offset: %v\n", err)
		os.Exit(1)
	}
	entryPoint, err = strconv.ParseUint(os.Args[5], 0, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading entry point: %v\n", err)
		os.Exit(1)
	}
	// extract loadable program segments
	var loadBlobs = extractLoadBlobs(elfFile.Progs)
	var program = mergeProgramBytes(loadBlobs, programOffset)
	var blobs = extractProgramBlobs(elfFile.Progs, elfFile.Sections)
	if len(inBytes) > 0 {
		blobs = append(blobs, memoryBlob{offset: inputsOffset, data: inBytes, name: IN_BYTES})
	}
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
	printJson(program, inBytes, blobs, programOffset, inputsOffset, entryPoint)
}

// Extract full PT_LOAD memory images for the legacy contiguous riscV_program
// field.
func extractLoadBlobs(progs []*elf.Prog) []memoryBlob {
	var blobs []memoryBlob

	for _, p := range progs {
		if p.Type != elf.PT_LOAD || p.Memsz == 0 {
			continue
		}
		if p.Filesz > p.Memsz {
			panic(fmt.Sprintf("loadable segment at %#x has file size larger than memory size", p.Vaddr))
		}
		// Vaddr is the RAM address where the segment is loaded; Memsz is the
		// number of RAM bytes it occupies, including zero-initialized bytes
		// not present in the ELF file
		blobs = append(blobs, memoryBlob{offset: p.Vaddr, data: readProgBytes(p, 0, p.Memsz), name: "PT_LOAD"})
	}

	if len(blobs) == 0 {
		panic("no loadable program segments found.")
	}

	return blobs
}

// Extract sparse memory blobs from allocated sections. Gaps inside a PT_LOAD
// segment are added only when needed to reproduce the loader memory image over
// non-zero RAM.
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

		for i, blob := range sectionBlobs {
			if i > 0 {
				gapStart := sectionBlobs[i-1].offset + uint64(len(sectionBlobs[i-1].data))
				if gapStart < blob.offset {
					blobs = append(blobs, memoryBlob{
						offset: gapStart,
						data:   readProgBytes(p, gapStart-p.Vaddr, blob.offset-gapStart),
						name:   "zero gap",
					})
				}
			}
			blobs = append(blobs, blob)
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

func writeSectionsFile(file *os.File, blobs []memoryBlob) {
	fmt.Fprintln(file, "# index offset size name")
	for i, blob := range blobs {
		fmt.Fprintf(file, "%d 0x%016x 0x%016x %s\n", i, blob.offset, len(blob.data), blob.name)
	}
}

func mergeProgramBytes(blobs []memoryBlob, programOffset uint64) []byte {
	var maxAddr uint64 = 0
	for _, blob := range blobs {
		if blob.offset < programOffset {
			panic(fmt.Sprintf("loadable segment starts before program offset: segment=%#x programOffset=%#x", blob.offset, programOffset))
		}
		end := blob.offset + uint64(len(blob.data))
		if end < blob.offset {
			panic(fmt.Sprintf("loadable segment address overflow at %#x", blob.offset))
		}
		if end > maxAddr {
			maxAddr = end
		}
	}

	buf := make([]byte, maxAddr-programOffset)

	for _, blob := range blobs {
		offset := blob.offset - programOffset
		copy(buf[offset:], blob.data)
	}

	// If needed pad buffer to multiple of 4 bytes (add at most 3 zero bytes)
	for len(buf)%4 != 0 {
		buf = append(buf, 0)
	}

	return buf
}

func printJson(program, inBytes []byte, blobs []memoryBlob, programOffset, inputsOffset, entryPoint uint64) {
	var (
		// programString       = hex.EncodeToString(program)
		// inBytesString       = hex.EncodeToString(inBytes)
		// programOffsetString = fmt.Sprintf("%016x", programOffset)
		// inputsOffsetString  = fmt.Sprintf("%016x", inputsOffset)
		entryPointString = fmt.Sprintf("%016x", entryPoint)
		// programLenString    = fmt.Sprintf("%016x", len(program))
		// inputLenString      = fmt.Sprintf("%016x", len(inBytes))
		blobsCountString   = fmt.Sprintf("%016x", len(blobs))
		entryPointAndBlobs = entryPointString + blobsCountString
		blobMetadata       []string
		blobData           []string
	)

	if len(program)%4 != 0 {
		panic("program length not multiple of 4")
	}
	for _, blob := range blobs {
		blobMetadata = append(blobMetadata, fmt.Sprintf("%016x%016x", blob.offset, len(blob.data)))
		if len(blob.data) > 0 {
			blobData = append(blobData, hex.EncodeToString(blob.data))
		}
	}

	fmt.Println("{")
	// fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_PROGRAM, programString)
	// fmt.Printf("\t\"%s\": \"0x%s\",\n", IN_BYTES, inBytesString)
	// fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_PROGRAM_OFFSET, programOffsetString)
	// fmt.Printf("\t\"%s\": \"0x%s\",\n", IN_BYTES_OFFSET, inputsOffsetString)
	// fmt.Printf("\t\"%s\": \"0x%s\",\n", ENTRY_POINT, entryPointString)
	// fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_PROGRAM_LENGTH, programLenString)
	// fmt.Printf("\t\"%s\": \"0x%s\",\n", IN_BYTES_LENGTH, inputLenString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", ENTRY_POINT_AND_BLOBS_COUNT, entryPointAndBlobs)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", BLOBS_OFFSET_AND_SIZE, strings.Join(blobMetadata, "_"))
	fmt.Printf("\t\"%s\": \"0x%s\"\n", BLOBS_DATA, strings.Join(blobData, "_"))
	fmt.Println("}")
}
