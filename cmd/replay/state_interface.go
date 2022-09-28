package replay

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

////////////////////////////////////////////////////////////
// State Operation Interface
////////////////////////////////////////////////////////////

// Number of state operations identifiers
const NumOperations = 3

// State operations' IDs
const BlockOperationID = 0
const GetStateOperationID = 1
const SetStateOperationID = 2

// State operations' filenames
var idToFilename = [NumOperations]string{
	"", // Pseudo operation (has no filename)
	"op-getstate.dat",
	"op-setstate.dat",
}

// State-opertion interface
type StateOperation interface {
	GetOpId() int             // obtain operation identifier
	Write(uint64, []*os.File) // write operation
}

// Polymorphic call for writing a state operation to its file
func Write(so StateOperation, opNum uint64, file []*os.File) {
	so.Write(opNum, file)
}

// Get filename of state operation for given identifier id
func GetFilename(id int) string {
	return idToFilename[id]
}

////////////////////////////////////////////////////////////
// Block Operation (Pseudo Operation)
////////////////////////////////////////////////////////////

// Block-operation data structure capturing the block number only.
type BlockOperation struct {
	blockNumber uint64 // block number
}

// Create a new block-operation.
func NewBlockOperation(blockNumber uint64) *BlockOperation {
	return &BlockOperation{blockNumber: blockNumber}
}

// Return the block-operation's identifier
func (s *BlockOperation) GetOpId() int {
	return 0
}

// Write block operation, which is a pseudo operation and never invoked.
func (s *BlockOperation) Write(opNum uint64, files []*os.File) {
	log.Fatalf("Block operation for block %v attempted to be written", s.blockNumber)
}

////////////////////////////////////////////////////////////
// GetState Operation
////////////////////////////////////////////////////////////

// GetState datastructure with encoded contract and storage addresses.
type GetStateOperation struct {
	contractIndex uint32 // encoded contract address
	storageIndex  uint32 // encoded storage address
}

// Create a new GetState operation.
func NewGetStateOperation(contractIndex uint32, storageIndex uint32) *GetStateOperation {
	return &GetStateOperation{contractIndex: contractIndex, storageIndex: storageIndex}
}

// Return the GetState-operation's identifier
func (s *GetStateOperation) GetOpId() int {
	return 1
}

// Write a GetState operation
func (s *GetStateOperation) Write(opNum uint64, files []*os.File) {
	// group information into data slice
	var data = []any{opNum, s.contractIndex, s.storageIndex}

	// write data to file
	for _, value := range data {
		if err := binary.Write(files[s.GetOpId()], binary.LittleEndian, value); err != nil {
			log.Fatal(err)
		}
	}

	// debug message
	fmt.Printf("GetState: operation number: %v\t contract idx: %v\t storage idx: %v\n", opNum, s.contractIndex, s.storageIndex)
}

////////////////////////////////////////////////////////////
// SetState Operation
////////////////////////////////////////////////////////////

// SetState datastructure with encoded contract and storage addresses, and value.
type SetStateOperation struct {
	contractIndex uint32      // encoded contract address
	storageIndex  uint32      // encoded storage address
	value         common.Hash // stored value
}

// Create a new SetState operation.
func NewSetStateOperation(contractIndex uint32, storageIndex uint32, value common.Hash) *SetStateOperation {
	return &SetStateOperation{contractIndex: contractIndex, storageIndex: storageIndex, value: value}
}

// Return the SetState operation's identifier
func (s *SetStateOperation) GetOpId() int {
	return 2
}

func (s *SetStateOperation) Write(opNum uint64, files []*os.File) {
	// group information into data slice
	var data = []any{opNum, s.contractIndex, s.storageIndex, s.value.Bytes()}

	// write data to file
	for _, value := range data {
		if err := binary.Write(files[s.GetOpId()], binary.LittleEndian, value); err != nil {
			log.Fatal(err)
		}
	}

	// debug message
	fmt.Printf("SetState: operation number: %v\t contract idx: %v\t storage idx: %v\t value: %v\n", opNum, s.contractIndex, s.storageIndex, s.value.Hex())
}
