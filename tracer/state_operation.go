package tracer

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/Fantom-foundation/substate-cli/state"
)

// Number of state operation identifiers
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
// Execution Context
////////////////////////////////////////////////////////////

// ExecutionContext contains the contract/storage dictionaries
// so that a recorded StateDB operation can be executed.
type ExecutionContext struct {
	ContractDictionary *ContractDictionary
	StorageDictionary  *StorageDictionary
}

////////////////////////////////////////////////////////////
// Writeable State Operations
////////////////////////////////////////////////////////////

// Writeable is the base class of writeable state operations.
// State operations whose base class is Writeable can
// be written to disk and have a sequence number for
// sequencing operations on disk.
type Writeable struct {
	SequenceNumber uint64 // operation number
}

// Set operation number.
func (w *Writeable) Set(opNum uint64) {
	w.SequenceNumber = opNum
}

// Get operation number.
func (w *Writeable) Get() uint64 {
	return w.SequenceNumber
}

////////////////////////////////////////////////////////////
// State Operation Interface
////////////////////////////////////////////////////////////

// TODO: Perhaps have in future two interfaces
//       1) Pseudo Operations
//       2) Writeable Operations

// State-opertion interface
type StateOperation interface {
	GetOpId() int                              // obtain operation identifier
	GetWriteable() *Writeable                  // obtain writeable interface
	Write(*os.File)                            // write operation
	Execute(*state.StateDB, *ExecutionContext) error // execute operation
}

// Read a state operation from file
func Read(f *os.File, ID int) *StateOperation {
	var sop StateOperation
	switch ID + NumPseudoOperations {
	case GetStateOperationID:
		sop = ReadGetStateOperation(f)
	case SetStateOperationID:
		sop = ReadSetStateOperation(f)
	}
	return &sop
}

////////////////////////////////////////////////////////////
// Begin Block Operation (Pseudo Operation)
////////////////////////////////////////////////////////////

// Block-operation data structure capturing the beginning of a block.
type BeginBlockOperation struct {
	BlockNumber uint64 // block number
}

// Return begin-block operation identifier.
func (bb *BeginBlockOperation) GetOpId() int {
	return BeginBlockOperationID
}

// Create a new begin-block operation.
func NewBeginBlockOperation(blockNumber uint64) *BeginBlockOperation {
	return &BeginBlockOperation{BlockNumber: blockNumber}
}

// Return writeable interface
func (bb *BeginBlockOperation) GetWriteable() *Writeable {
	return nil
}

// Write block operation (should never be invoked).
func (bb *BeginBlockOperation) Write(files *os.File) {
	log.Fatalf("Begin-block operation for block %v attempted to be written", bb.BlockNumber)
}

// Execute state operation
func (bb *BeginBlockOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	log.Fatalf("Begin-block operation for block %v attempted to be executed", bb.BlockNumber)
	return nil
}

////////////////////////////////////////////////////////////
// End Block Operation (Pseudo Operation)
////////////////////////////////////////////////////////////

// Block-operation data structure capturing the beginning of a block.
type EndBlockOperation struct {
	BlockNumber uint64 // block number
}

// Return end-block operation identifier.
func (eb *EndBlockOperation) GetOpId() int {
	return EndBlockOperationID
}

// Create a new end-block operation.
func NewEndBlockOperation(blockNumber uint64) *EndBlockOperation {
	return &EndBlockOperation{BlockNumber: blockNumber}
}

// Return writeable interface
func (eb *EndBlockOperation) GetWriteable() *Writeable {
	return nil
}

// Write end-block operation (should never be invoked).
func (eb *EndBlockOperation) Write(files *os.File) {
	log.Fatalf("End-block operation for block %v attempted to be written", eb.BlockNumber)
}

// Execute state operation
func (eb *EndBlockOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	log.Fatalf("End-block operation for block %v attempted to be executed", eb.BlockNumber)
	return nil
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
func ReadGetStateOperation(file *os.File) *GetStateOperation {
	data := new(GetStateOperation)
	if err := binary.Read(file, binary.LittleEndian, data); err != nil {
		log.Fatal(err)
	}
	return data
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

// Execute state operation
func (gso *GetStateOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	contract, cerr := ctx.ContractDictionary.Decode(gso.ContractIndex)
	if cerr != nil {
		return cerr
	}
	storage, serr := ctx.StorageDictionary.Decode(gso.StorageIndex)
	if serr != nil {
		return serr
	}
	(*db).GetState(contract, storage)
	return nil
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
func ReadSetStateOperation(file *os.File) *SetStateOperation {
	data := new(SetStateOperation)
	if err := binary.Read(file, binary.LittleEndian, data); err != nil {
		log.Fatal(err)
	}
	return data
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

// Execute state operation
func (sso *SetStateOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	contract, cerr := ctx.ContractDictionary.Decode(sso.ContractIndex)
	if cerr != nil {
		return cerr
	}
	storage, serr := ctx.StorageDictionary.Decode(sso.StorageIndex)
	if serr != nil {
		return serr
	}
	(*db).SetState(contract, storage, sso.Value)
	return nil
}
