package tracer

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// OperationIndex maps a block number to the first operation number in the block.
type OperationIndex struct {
	blockToOperation map[uint64]uint64 // block number -> operation number
}

// Initialize an operation index.
func (oIdx *OperationIndex) Init() {
	oIdx.blockToOperation = make(map[uint64]uint64)
}

// Create new OperationIndex data structure.
func NewOperationIndex() *OperationIndex {
	p := new(OperationIndex)
	p.Init()
	return p
}

// Add new entry.
func (oIdx *OperationIndex) Add(block uint64, operation uint64) error {
	var err error = nil
	if _, ok := oIdx.blockToOperation[block]; ok {
		err = errors.New("block number already exists")
	}
	oIdx.blockToOperation[block] = operation
	return err
}

// Get operation number.
func (oIdx *OperationIndex) Get(block uint64) (uint64, error) {
	operation, ok := oIdx.blockToOperation[block]
	if !ok {
		return 0, errors.New("block number does not exist")
	}
	return operation, nil
}

// Write index to a binary file.
func (oIdx *OperationIndex) Write(filename string) error {
	// open index file for writing
	f, err := os.OpenFile(TraceDir + filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()

	// write all dictionary entries
	for block, operation := range oIdx.blockToOperation {
		var data = []any{block, operation}
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
func (oIdx *OperationIndex) Read(filename string) error {
	// clear storage dictionary
	oIdx.Init()

	// open storage dictionary file for reading
	f, err := os.OpenFile(TraceDir + filename, os.O_CREATE|os.O_RDONLY, 0644)
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
			Block     uint64
			Operation uint64
		}
		err := binary.Read(f, binary.LittleEndian, &data)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = oIdx.Add(data.Block, data.Operation)
		if err != nil {
			return err
		}
	}
	return nil
}
