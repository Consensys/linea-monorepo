package main

import (
	"debug/elf"
	"encoding/hex"
	"fmt"
	"io"
	"os"
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
	var programBlobs = extractProgramBlobs(elfFile.Progs)
	var program = mergeProgramBytes(programBlobs, programOffset)
	var blobs = append(programBlobs, memoryBlob{offset: inputsOffset, data: inBytes})
	switch writeSections := os.Getenv("ELF2JSON_WRITE_SECTIONS"); writeSections {
	case "", "false":
	case "true":
		sectionsFile, err := os.Create(strings.TrimSuffix(os.Args[1], ".elf") + ".sections")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating ELF sections file: %v\n", err)
			os.Exit(1)
		}
		for _, s := range elfFile.Sections {
			if s.Name != "" && s.Size > 0 && s.Flags&elf.SHF_ALLOC != 0 {
				fmt.Fprintln(sectionsFile, s.Name)
			}
		}
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

// Extract loadable ELF segments as sparse memory blobs. The legacy
// riscV_program field below still merges them into a single contiguous blob for
// compatibility.
//
// Our own tests contain .text, .rodata, .data and .bss sections.
// ACT4 tests contain .text.init, .text.rvtest, .text.rvmodel, .data,
// and .tohost sections. We do not filter by section names here:
// PT_LOAD segments are what the ELF loader actually maps into memory.
func extractProgramBlobs(progs []*elf.Prog) []memoryBlob {
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
		data := make([]byte, p.Memsz)
		if p.Filesz > 0 {
			n, err := p.ReadAt(data[:p.Filesz], 0)
			if err != nil && err != io.EOF {
				panic(fmt.Sprintf("error reading loadable segment at %#x: %v", p.Vaddr, err))
			}
			if uint64(n) != p.Filesz {
				panic(fmt.Sprintf("short read for loadable segment at %#x: got %d bytes, expected %d", p.Vaddr, n, p.Filesz))
			}
		}
		blobs = append(blobs, memoryBlob{offset: p.Vaddr, data: data})
	}

	if len(blobs) == 0 {
		panic("no loadable program segments found.")
	}

	return blobs
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
		programString       = hex.EncodeToString(program)
		inBytesString       = hex.EncodeToString(inBytes)
		programOffsetString = fmt.Sprintf("%016x", programOffset)
		inputsOffsetString  = fmt.Sprintf("%016x", inputsOffset)
		entryPointString    = fmt.Sprintf("%016x", entryPoint)
		programLenString    = fmt.Sprintf("%016x", len(program))
		inputLenString      = fmt.Sprintf("%016x", len(inBytes))
		blobsCountString    = fmt.Sprintf("%016x", len(blobs))
		entryPointAndBlobs  = entryPointString + blobsCountString
		blobMetadata        []string
		blobData            []string
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
	fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_PROGRAM, programString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", IN_BYTES, inBytesString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_PROGRAM_OFFSET, programOffsetString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", IN_BYTES_OFFSET, inputsOffsetString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", ENTRY_POINT, entryPointString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", RISCV_PROGRAM_LENGTH, programLenString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", IN_BYTES_LENGTH, inputLenString)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", ENTRY_POINT_AND_BLOBS_COUNT, entryPointAndBlobs)
	fmt.Printf("\t\"%s\": \"0x%s\",\n", BLOBS_OFFSET_AND_SIZE, strings.Join(blobMetadata, "_"))
	fmt.Printf("\t\"%s\": \"0x%s\"\n", BLOBS_DATA, strings.Join(blobData, "_"))
	fmt.Println("}")
}
