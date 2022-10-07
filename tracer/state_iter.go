package tracer

import (
	"log"
	"os"
)

// IndexContext keeps all index data strutures for the iterator.
type IndexContext struct {
	BlockIndex *BlockIndex
}

// Iterator data structure for storage traces
type TraceIterator struct {
	lastBlock uint64
	iCtx      *IndexContext
	file      *os.File
	currentOp StateOperation
}

// Create new trace iterator.
func NewTraceIterator(iCtx *IndexContext, first uint64, last uint64) *TraceIterator {
	p := new(TraceIterator)
	p.iCtx = iCtx
	p.lastBlock = last

	// TODO: Add trace directory to filename
	var err error
	p.file, err = os.OpenFile("trace.dat", os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Cannot open trace file.")
	}

	// TODO: set file position to the first position using seek

	return p
}

// Get next state operation from trace file.
func (ti *TraceIterator) Next() bool {
	// TODO: if file position succeeds last block, return false.
	ti.currentOp = Read(ti.file)
	return ti.currentOp == nil
}

// Retrieve current state operation of the iterator.
func (ti *TraceIterator) Value() StateOperation {
	return ti.currentOp
}

// Release the storage trace iterator.
func (ti *TraceIterator) Release() {
	// close trace file
	err := ti.file.Close()
	if err != nil {
		log.Fatalf("Cannot close trace file")
	}
}
