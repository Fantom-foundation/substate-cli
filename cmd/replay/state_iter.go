package replay

type IndexContext struct {
	FilePositionIndex *FilePositionIndex
	OperationIndex *OperationIndex
}

type TraceIterator struct {
	currentBlock uint64
	lastBlock uint64
	iCtx *IndexContext
}

func NewStorageTraceIterator(iCtx *IndexContext, first uint64, last uint64) *TraceIterator {
	p:= new(TraceIterator)
	p.iCtx = iCtx
	p.currentBlock = first
	p.lastBlock = last
	return p
}

func (ti *TraceIterator) Release() {
}

func (ti *TraceIterator) Next() bool {
	return false
}

func (ti *TraceIterator) Value() *StateOperation {
	return nil
}
