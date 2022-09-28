package replay

import (
	"errors"
)

// Block map stores the first operation number of a block and stores
type BlockMap struct {
	blockToOperation     map[uint64]uint64   // block number -> operation number
	blockToFilePositions map[uint64][]uint64 // block number -> array of file positions of operations
}

// Create new block map.
func NewBlockMap() *BlockMap {
	p := new(BlockMap)
	p.blockToOperation = map[uint64]uint64{}
	p.blockToFilePositions = map[uint64][]uint64{}
	return p
}

// Add operation number to the block map
func (bm *BlockMap) addOperation(block uint64, operation uint64) error {
	var err error = nil
	if _, ok := bm.blockToOperation[block]; ok {
		err = errors.New("block number exists in operation map")
	}
	bm.blockToOperation[block] = operation
	return err
}

// Add file index positions of state operations to the block map
func (bm *BlockMap) addFilePositions(block uint64, filePositions []uint64) error {
	var err error = nil
	if _, ok := bm.blockToFilePositions[block]; ok {
		err = errors.New("block number exists in file-positions map")
	}
	bm.blockToFilePositions[block] = filePositions
	return err
}

// Write block map to file
func (bm *BlockMap) Write(string) {
}
