package replay

import (
	"encoding/binary"
	"errors"
	"os"
)

// FilePositionIndex maps a block number to the first operation number in the block.
type FilePositionIndex struct {
	blockToFilePos map[uint64][NumOperations-1]uint64 // block number -> operation number
}

// Initialize an operation index.
func (fposIdx *FilePositionIndex) Init() {
	fposIdx.blockToFilePos = make(map[uint64][NumOperations-1]uint64)
}

// Create new FilePositionIndex data structure.
func NewFilePositionIndex() *FilePositionIndex {
	p := new(FilePositionIndex)
	p.Init()
	return p
}

// Add new entry.
func (fposIdx *FilePositionIndex) Add(block uint64, filepos [NumOperations-1]uint64) error {
	var err error = nil
	if _, ok := fposIdx.blockToFilePos[block]; ok {
		err = errors.New("block number already exists")
	}
	fposIdx.blockToFilePos[block] = filepos
	return err
}

// Get operation number.
func (fposIdx *FilePositionIndex) Get(block uint64) ([NumOperations-1]uint64, error) {
	filepos, ok := fposIdx.blockToFilePos[block]
	if !ok {
		return [NumOperations-1]uint64{}, errors.New("block number does not exist")
	}
	return filepos, nil
}

// Write index to a binary file.
func (fposIdx *FilePositionIndex) Write(filename string) error {
	// open index file for writing
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()

	// write all dictionary entries
	for block, fpos := range fposIdx.blockToFilePos {
		var data = []any{block, fpos}
		for _, value := range data {
			err := binary.Write(f, binary.LittleEndian, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Read dictionary from a binary file.
func (fposIdx *FilePositionIndex) Read(filename string) error {
	// clear storage dictionary
	fposIdx.Init()

	// open storage dictionary file for reading
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()

	// read entries from file
	for {
		// read next entry
		var data struct {
			block     uint64
			fpos [NumOperations-1] uint64
		}
		err := binary.Read(f, binary.LittleEndian, &data)
		if err != nil {
			return err
		}
		err = fposIdx.Add(data.block, data.fpos)
		if err != nil {
			return err
		}
	}
	return nil
}
