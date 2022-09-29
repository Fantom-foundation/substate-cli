package replay

import (
	"log"
	"os"
)

// IndexContext keeps all index data strutures
type IndexContext struct {
	FilePositionIndex *FilePositionIndex
	OperationIndex    *OperationIndex
}

// Iterator data structure for storage traces
type TraceIterator struct {
	lastBlock uint64
	iCtx      *IndexContext
	files     []*os.File
	nextOp    []*StateOperation
	currentOp *StateOperation
}

// Create new storage trace iterator.
func NewStorageTraceIterator(iCtx *IndexContext, first uint64, last uint64) *TraceIterator {
	p := new(TraceIterator)
	p.iCtx = iCtx
	p.lastBlock = last
	p.files = make([]*os.File, NumWriteOperations)
	for i := 0; i < NumWriteOperations; i++ {
		f, err := os.OpenFile(GetFilename(i), os.O_RDONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Cannot open state operation file %v", i)
		}
		p.files[i] = f
		p.nextOp[i] = Read(f, i)
	}

	// TODO: skipping the first blocks (using the file-position and operation index)
	return p
}

// Release the storage trace iterator. This closes all associated files
func (ti *TraceIterator) Release() {
	for i := 0; i < NumWriteOperations; i++ {
		err := ti.files[i].Close()
		if err != nil {
			log.Fatalf("Cannot close state operation file %v", i)
		}
	}
}

// Get next state operation. The next state
// operation is found by searching over all state operation
// types and finding the operation that has the smallest
// operation sequence number.
func (ti *TraceIterator) Next() bool {
	// TODO: make this more efficient
	minIdx := -1
	for i := 0; i < NumWriteOperations; i++ {
		if ti.nextOp[i] != nil {
			if minIdx == -1 {
				minIdx = i
			} else if (*ti.nextOp[i]).GetWriteable().Get() < (*ti.nextOp[minIdx]).GetWriteable().Get() {
				minIdx = i
			}
		}
	}
	if minIdx == -1 {
		return false
	} else {
		ti.currentOp = ti.nextOp[minIdx]
		ti.nextOp[minIdx] = Read(ti.files[minIdx], minIdx)
		return true
	}
}

func (ti *TraceIterator) Value() *StateOperation {
	return ti.currentOp
}
