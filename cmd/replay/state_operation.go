package replay

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

// Number of state operations identifiers
const NumOperations = 4

// Number of pseudo operations that don't write to files
const NumPseudoOperations = 2

// Number of write operations
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

// Get filename of a state operation that is written to a file
func GetFilename(i int) string {
	if i < 0 || i >= NumWriteOperations {
		log.Fatalf("GetFilename failed; index is out-of-bound")
	}
	return idToFilename[i]
}

////////////////////////////////////////////////////////////
// Writeable State Operations
////////////////////////////////////////////////////////////

// Writeable is the base class of writeable state operations.
// State operations whose base class is Writeable can
// be written to disk and have an operation number for
// sequecing all operations on disk.
type Writeable struct {
	OperationNumber uint64 // operation number
}

// Set operation number.
func (w *Writeable) Set(opNum uint64) {
	w.OperationNumber = opNum
}

// Get operation number.
func (w *Writeable) Get() uint64 {
	return w.OperationNumber
}

////////////////////////////////////////////////////////////
// State Operation Interface
////////////////////////////////////////////////////////////

// State-opertion interface
type StateOperation interface {
	GetOpId() int             // obtain operation identifier
	GetWriteable() *Writeable // obtain writeable interface
	Write(*os.File)           // write operation
}

// Polymorphic call for writing a writeable state operation to its file
func Write(so StateOperation, files []*os.File) {
	// compute index
	idx := so.GetOpId() - NumPseudoOperations

	// write object to its file
	so.Write(files[idx])
}

////////////////////////////////////////////////////////////
// Begin Block Operation (Pseudo Operation)
////////////////////////////////////////////////////////////

// Block-operation data structure capturing the beginning of a block.
type BeginBlockOperation struct {
	blockNumber uint64 // block number
}

// Return begin-block operation identifier.
func (bb *BeginBlockOperation) GetOpId() int {
	return BeginBlockOperationID
}

// Create a new begin-block operation.
func NewBeginBlockOperation(blockNumber uint64) *BeginBlockOperation {
	return &BeginBlockOperation{blockNumber: blockNumber}
}

// Return writeable interface
func (bb *BeginBlockOperation) GetWriteable() *Writeable {
	return nil
}

// Write block operation (should never be invoked).
func (bb *BeginBlockOperation) Write(files *os.File) {
	log.Fatalf("Begin-block operation for block %v attempted to be written", bb.blockNumber)
}

////////////////////////////////////////////////////////////
// End Block Operation (Pseudo Operation)
////////////////////////////////////////////////////////////

// Block-operation data structure capturing the beginning of a block.
type EndBlockOperation struct {
	blockNumber uint64 // block number
}

// Return end-block operation identifier.
func (eb *EndBlockOperation) GetOpId() int {
	return EndBlockOperationID
}

// Create a new end-block operation.
func NewEndBlockOperation(blockNumber uint64) *EndBlockOperation {
	return &EndBlockOperation{blockNumber: blockNumber}
}

// Return writeable interface
func (eb *EndBlockOperation) GetWriteable() *Writeable {
	return nil
}

// Write end-block operation (should never be invoked).
func (eb *EndBlockOperation) Write(files *os.File) {
	log.Fatalf("End-block operation for block %v attempted to be written", eb.blockNumber)
}

////////////////////////////////////////////////////////////
// GetState Operation
////////////////////////////////////////////////////////////

// GetState datastructure with encoded contract and storage addresses.
type GetStateOperation struct {
	Writeable
	ContractIndex uint32 // encoded contract address
	StorageIndex  uint32 // encoded storage address
}

// Return get-state operation identifier.
func (gso *GetStateOperation) GetOpId() int {
	return GetStateOperationID
}

// Create a new get-state operation.
func NewGetStateOperation(ContractIndex uint32, StorageIndex uint32) *GetStateOperation {
	return &GetStateOperation{ContractIndex: ContractIndex, StorageIndex: StorageIndex}
}

// Read get-state operation from a file.
func ReadGetStateOperation(file *os.File) (*GetStateOperation, error) {
	data := new(GetStateOperation)
	if err := binary.Read(file, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	return data, nil
}

// Return writeable interface
func (gso *GetStateOperation) GetWriteable() *Writeable {
	return &gso.Writeable
}

// Write a get-state operation.
func (gso *GetStateOperation) Write(f *os.File) {
	// group information into data slice
	var data = []any{gso.Writeable.Get(), gso.ContractIndex, gso.StorageIndex}

	// write data to file
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatal(err)
		}
	}

	// debug message
	fmt.Printf("GetState: operation number: %v\t contract idx: %v\t storage idx: %v\n", gso.Writeable.Get(), gso.ContractIndex, gso.StorageIndex)
}

////////////////////////////////////////////////////////////
// SetState Operation
////////////////////////////////////////////////////////////

// SetState datastructure with encoded contract and storage addresses, and value.
type SetStateOperation struct {
	Writeable
	ContractIndex uint32      // encoded contract address
	StorageIndex  uint32      // encoded storage address
	Value         common.Hash // stored value
}

// Return set-state identifier
func (sso *SetStateOperation) GetOpId() int {
	return SetStateOperationID
}

// Create a new set-state operation.
func NewSetStateOperation(ContractIndex uint32, StorageIndex uint32, value common.Hash) *SetStateOperation {
	return &SetStateOperation{ContractIndex: ContractIndex, StorageIndex: StorageIndex, Value: value}
}

// Read set-state operation from a file.
func ReadSetStateOperation(file *os.File) (*SetStateOperation, error) {
	data := new(SetStateOperation)
	if err := binary.Read(file, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	return data, nil
}

// Return writeable interface
func (sso *SetStateOperation) GetWriteable() *Writeable {
	return &sso.Writeable
}

// Write a set-state operation.
func (sso *SetStateOperation) Write(f *os.File) {
	// group information into data slice
	var data = []any{sso.Writeable.Get(), sso.ContractIndex, sso.StorageIndex, sso.Value.Bytes()}

	// write data to file
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatal(err)
		}
	}

	// debug message
	fmt.Printf("SetState: operation number: %v\t contract idx: %v\t storage idx: %v\t value: %v\n", sso.Writeable.Get(), sso.ContractIndex, sso.StorageIndex, sso.Value.Hex())
}
