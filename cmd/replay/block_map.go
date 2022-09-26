package replay

import (
	"errors"
)

// Block map data structure
type BlockMap struct {
	blockToOperation     map[uint64]uint64   // maps a block number to an operation number
	blockToFilePositions map[uint64][]uint64 // maps a block number to an array of file positions
}

// Create new block map
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
