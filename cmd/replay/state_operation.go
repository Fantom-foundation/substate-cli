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
const NumOperations = 4
const NumPseudoOperations = 2
const NumWriteOperations = NumOperations - NumPseudoOperations

// Operation IDs
// Pseudo Operations (not stored on file)
const BeginBlockOperationID = 0
const EndBlockOperationID = 1

// Stored Operations
const GetStateOperationID = 2
const SetStateOperationID = 3

// State operations' filenames
var idToFilename = [NumOperations]string{
	"sop-getstate.dat",
	"sop-setstate.dat",
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

// Get filename of a state operation that is written to a file
func GetFilename(i int) string {
	if i < 0 || i >= NumWriteOperations {
		log.Fatalf("GetFilename failed; id is out-of-bound")
	}
	return idToFilename[i]
}

////////////////////////////////////////////////////////////
// Begin Block Operation (Pseudo Operation)
////////////////////////////////////////////////////////////

// Block-operation data structure capturing the beginning of a block.
type BeginBlockOperation struct {
	blockNumber uint64 // block number
}

// Create a new begin-block operation.
func NewBeginBlockOperation(blockNumber uint64) *BeginBlockOperation {
	return &BeginBlockOperation{blockNumber: blockNumber}
}

// Return begin-block operation identifier.
func (s *BeginBlockOperation) GetOpId() int {
	return BeginBlockOperationID
}

// Write block operation (should never be invoked).
func (s *BeginBlockOperation) Write(opNum uint64, files []*os.File) {
	log.Fatalf("Begin-block operation for block %v attempted to be written", s.blockNumber)
}

////////////////////////////////////////////////////////////
// End Block Operation (Pseudo Operation)
////////////////////////////////////////////////////////////

// Block-operation data structure capturing the beginning of a block.
type EndBlockOperation struct {
	blockNumber uint64 // block number
}

// Create a new end-block operation.
func NewEndBlockOperation(blockNumber uint64) *EndBlockOperation {
	return &EndBlockOperation{blockNumber: blockNumber}
}

// Return end-block operation identifier.
func (s *EndBlockOperation) GetOpId() int {
	return EndBlockOperationID
}

// Write end-block operation (should never be invoked).
func (s *EndBlockOperation) Write(opNum uint64, files []*os.File) {
	log.Fatalf("End-block operation for block %v attempted to be written", s.blockNumber)
}

////////////////////////////////////////////////////////////
// GetState Operation
////////////////////////////////////////////////////////////

// GetState datastructure with encoded contract and storage addresses.
type GetStateOperation struct {
	contractIndex uint32 // encoded contract address
	storageIndex  uint32 // encoded storage address
}

// Create a new get-state operation.
func NewGetStateOperation(contractIndex uint32, storageIndex uint32) *GetStateOperation {
	return &GetStateOperation{contractIndex: contractIndex, storageIndex: storageIndex}
}

// Read get-state operation from a file.
func ReadGetStateOperation(file *os.File) (*GetStateOperation, error) {
	data := new(GetStateOperation)
	if err := binary.Read(file, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	return data, nil
}

// Return get-state operation identifier.
func (s *GetStateOperation) GetOpId() int {
	return GetStateOperationID
}

// Write a get-state operation.
func (s *GetStateOperation) Write(opNum uint64, files []*os.File) {
	// group information into data slice
	var data = []any{opNum, s.contractIndex, s.storageIndex}

	// write data to file
	idx := s.GetOpId() - NumPseudoOperations
	for _, value := range data {
		if err := binary.Write(files[idx], binary.LittleEndian, value); err != nil {
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

// Create a new set-state operation.
func NewSetStateOperation(contractIndex uint32, storageIndex uint32, value common.Hash) *SetStateOperation {
	return &SetStateOperation{contractIndex: contractIndex, storageIndex: storageIndex, value: value}
}

// Read set-state operation from a file.
func ReadSetStateOperation(file *os.File) (*SetStateOperation, error) {
	data := new(SetStateOperation)
	if err := binary.Read(file, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	return data, nil
}

// Return set-state identifier
func (s *SetStateOperation) GetOpId() int {
	return SetStateOperationID
}

// Write a set-state operation.
func (s *SetStateOperation) Write(opNum uint64, files []*os.File) {
	// group information into data slice
	var data = []any{opNum, s.contractIndex, s.storageIndex, s.value.Bytes()}

	// write data to file
	idx := s.GetOpId() - NumPseudoOperations
	for _, value := range data {
		if err := binary.Write(files[idx], binary.LittleEndian, value); err != nil {
			log.Fatal(err)
		}
	}

	// debug message
	fmt.Printf("SetState: operation number: %v\t contract idx: %v\t storage idx: %v\t value: %v\n", opNum, s.contractIndex, s.storageIndex, s.value.Hex())
}
