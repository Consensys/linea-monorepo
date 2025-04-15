package mempool

import (
	"errors"
	"fmt"
	"github.com/consensys/linea-monorepo/prover/maths/field/fext"
	"runtime"
	"strconv"
	"unsafe"

	"github.com/consensys/linea-monorepo/prover/maths/field"
	"github.com/consensys/linea-monorepo/prover/utils"
)

type DebuggeableCall struct {
	Parent MemPool
	Logs   map[uintptr]*[]Record
}

func NewDebugPool(p MemPool) *DebuggeableCall {
	return &DebuggeableCall{
		Parent: p,
		Logs:   make(map[uintptr]*[]Record),
	}
}

type Record struct {
	Where string
	What  recordType
}

func (m *DebuggeableCall) Prewarm(nbPrewarm int) MemPool {
	m.Parent.Prewarm(nbPrewarm)
	return m
}

type recordType string

const (
	AllocRecordBase recordType = "allocBase"
	FreeRecordBase  recordType = "freeBase"
	AllocRecordExt  recordType = "allocExt"
	FreeRecordExt   recordType = "freeExt"
)

func (m *DebuggeableCall) Alloc() *[]field.Element {

	var (
		v                = m.Parent.Alloc()
		uptr             = uintptr(unsafe.Pointer(v))
		logs             *[]Record
		_, file, line, _ = runtime.Caller(2)
	)

	logs, found := m.Logs[uptr]

	if !found {
		logs = &[]Record{}
		m.Logs[uptr] = logs
	}

	*logs = append(*logs, Record{
		Where: file + ":" + strconv.Itoa(line),
		What:  AllocRecordBase,
	})

	return v
}

func (m *DebuggeableCall) AllocBase() *[]field.Element {

	var (
		v                = m.Parent.AllocBase()
		uptr             = uintptr(unsafe.Pointer(v))
		logs             *[]Record
		_, file, line, _ = runtime.Caller(2)
	)

	logs, found := m.Logs[uptr]

	if !found {
		logs = &[]Record{}
		m.Logs[uptr] = logs
	}

	*logs = append(*logs, Record{
		Where: file + ":" + strconv.Itoa(line),
		What:  AllocRecordBase,
	})

	return v
}

func (m *DebuggeableCall) AllocExt() *[]fext.Element {

	var (
		v                = m.Parent.AllocExt()
		uptr             = uintptr(unsafe.Pointer(v))
		logs             *[]Record
		_, file, line, _ = runtime.Caller(2)
	)

	logs, found := m.Logs[uptr]

	if !found {
		logs = &[]Record{}
		m.Logs[uptr] = logs
	}

	*logs = append(*logs, Record{
		Where: file + ":" + strconv.Itoa(line),
		What:  AllocRecordExt,
	})

	return v
}

func (m *DebuggeableCall) Free(v *[]field.Element) error {

	var (
		uptr             = uintptr(unsafe.Pointer(v))
		logs             *[]Record
		_, file, line, _ = runtime.Caller(2)
	)

	logs, found := m.Logs[uptr]

	if !found {
		logs = &[]Record{}
		m.Logs[uptr] = logs
	}

	*logs = append(*logs, Record{
		Where: file + ":" + strconv.Itoa(line),
		What:  FreeRecordBase,
	})

	return m.Parent.Free(v)
}

func (m *DebuggeableCall) FreeBase(v *[]field.Element) error {

	var (
		uptr             = uintptr(unsafe.Pointer(v))
		logs             *[]Record
		_, file, line, _ = runtime.Caller(2)
	)

	logs, found := m.Logs[uptr]

	if !found {
		logs = &[]Record{}
		m.Logs[uptr] = logs
	}

	*logs = append(*logs, Record{
		Where: file + ":" + strconv.Itoa(line),
		What:  FreeRecordBase,
	})

	return m.Parent.Free(v)
}

func (m *DebuggeableCall) FreeExt(v *[]fext.Element) error {

	var (
		uptr             = uintptr(unsafe.Pointer(v))
		logs             *[]Record
		_, file, line, _ = runtime.Caller(2)
	)

	logs, found := m.Logs[uptr]

	if !found {
		logs = &[]Record{}
		m.Logs[uptr] = logs
	}

	*logs = append(*logs, Record{
		Where: file + ":" + strconv.Itoa(line),
		What:  FreeRecordExt,
	})

	return m.Parent.FreeExt(v)
}

func (m *DebuggeableCall) Size() int {
	return m.Parent.Size()
}

func (m *DebuggeableCall) TearDown() {
	if p, ok := m.Parent.(*SliceArena); ok {
		p.TearDown()
	}
}

func (m *DebuggeableCall) Errors() error {

	var err error

	for _, logs_ := range m.Logs {

		if logs_ == nil || len(*logs_) == 0 {
			utils.Panic("got a nil entry")
		}

		logs := *logs_

		for i := range logs {
			if i == 0 && logs[i].What == FreeRecordBase {
				err = errors.Join(err, fmt.Errorf("freed a base vector that was not from the pool: where=%v", logs[i].Where))
			}

			if i == 0 && logs[i].What == FreeRecordExt {
				err = errors.Join(err, fmt.Errorf("freed an extension vector that was not from the pool: where=%v", logs[i].Where))
			}

			if i == len(logs)-1 && logs[i].What == AllocRecordBase {
				err = errors.Join(err, fmt.Errorf("leaked a base vector out of the pool: where=%v", logs[i].Where))
			}

			if i == len(logs)-1 && logs[i].What == AllocRecordExt {
				err = errors.Join(err, fmt.Errorf("leaked an extension vector out of the pool: where=%v", logs[i].Where))
			}

			if i == 0 {
				continue
			}

			errorGeneration := func(recordType recordType, verbString string) {
				if logs[i-1].What == recordType && logs[i].What == recordType {
					wheres := []string{logs[i-1].Where, logs[i].Where}
					for k := i + 1; k < len(logs) && logs[k].What == recordType; k++ {
						wheres = append(wheres, logs[k].Where)
					}

					err = errors.Join(
						err,
						fmt.Errorf("vector was %s multiple times concurrently where=%v", verbString, wheres),
					)
				}
			}

			errorGeneration(AllocRecordBase, "allocated")
			errorGeneration(AllocRecordExt, "allocated")
			errorGeneration(FreeRecordBase, "freed")
			errorGeneration(FreeRecordExt, "freed")
		}
	}

	return err
}
