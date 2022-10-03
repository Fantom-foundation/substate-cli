package tracer

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/Fantom-foundation/substate-cli/state"
	"github.com/ethereum/go-ethereum/common"
)

// Number of pseudo operations that don't write to files
const NumPseudoOperations = 2

// Operation IDs
// Pseudo Operations (not stored on file)
const BeginBlockOperationID = 0
const EndBlockOperationID = 1

// Stored Operations
const GetStateOperationID = NumPseudoOperations
const SetStateOperationID = NumPseudoOperations + 1
const GetCommittedStateOperationID = NumPseudoOperations + 2
const SnapshotOperationID = NumPseudoOperations + 3
const RevertToSnapshotOperationID = NumPseudoOperations + 4
const EndOfTransactionOperationID = NumPseudoOperations + 5 //last

// Number of state operation identifiers
const NumOperations = EndOfTransactionOperationID + 1 //last op + 1

// Number of write operations
const NumWriteOperations = NumOperations - NumPseudoOperations

// State operations' filenames
var idToFilename = [NumOperations]string{
	"sop-getstate.dat",
	"sop-setstate.dat",
	"sop-getcommittedstate.dat",
	"sop-snapshot.dat",
	"sop-reverttosnapshot.dat",
	"sop-endoftransaction.dat",
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
// Writable State Operations
////////////////////////////////////////////////////////////

// Writable is the base class of writeable state operations.
// State operations whose base class is Writable can
// be written to disk and have a sequence number for
// sequencing operations on disk.
type Writable struct {
	SequenceNumber uint64 // operation number
}

// Set operation number.
func (w *Writable) Set(opNum uint64) {
	w.SequenceNumber = opNum
}

// Get operation number.
func (w *Writable) Get() uint64 {
	return w.SequenceNumber
}

////////////////////////////////////////////////////////////
// State Operation Interface
////////////////////////////////////////////////////////////

// TODO: Perhaps have in future two interfaces
//       1) Pseudo Operations
//       2) Writable Operations

// State-opertion interface
type StateOperation interface {
	GetOpId() int                                    // obtain operation identifier
	GetWritable() *Writable                          // obtain writeable interface
	Write(*os.File)                                  // write operation
	Execute(*state.StateDB, *ExecutionContext) error // execute operation
}

// Read a state operation from file
func Read(f *os.File, ID int) *StateOperation {
	var (
		sop StateOperation
		err error = nil
	)

	switch ID + NumPseudoOperations {
	case GetStateOperationID:
		sop, err = ReadGetStateOperation(f)
	case SetStateOperationID:
		sop, err = ReadSetStateOperation(f)
	case GetCommittedStateOperationID:
		sop, err = ReadGetCommittedStateOperation(f)
	case SnapshotOperationID:
		sop, err = ReadSnapshotOperation(f)
	case RevertToSnapshotOperationID:
		sop, err = ReadRevertToSnapshotOperation(f)
	case EndOfTransactionOperationID:
		sop, err = ReadEndOfTransactionOperation(f)
	}
	if err != nil {
		log.Fatal(err)
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
func (bb *BeginBlockOperation) GetWritable() *Writable {
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
func (eb *EndBlockOperation) GetWritable() *Writable {
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
	Writable
	ContractIndex uint32 // encoded contract address
	StorageIndex  uint32 // encoded storage address
}

// Return get-state operation identifier.
func (sop *GetStateOperation) GetOpId() int {
	return GetStateOperationID
}

// Create a new get-state operation.
func NewGetStateOperation(ContractIndex uint32, StorageIndex uint32) *GetStateOperation {
	return &GetStateOperation{ContractIndex: ContractIndex, StorageIndex: StorageIndex}
}

// Read get-state operation from a file.
func ReadGetStateOperation(file *os.File) (*GetStateOperation, error) {
	data := new(GetStateOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (sop *GetStateOperation) GetWritable() *Writable {
	return &sop.Writable
}

// Write a get-state operation.
func (sop *GetStateOperation) Write(f *os.File) {
	// group information into data slice
	var data = []any{sop.Writable.Get(), sop.ContractIndex, sop.StorageIndex}

	// write data to file
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatalf("GetStateOperation: %v", err)
		}
	}

	// debug message
	fmt.Printf("GetState: operation number: %v\t contract idx: %v\t storage idx: %v\n", sop.Writable.Get(), sop.ContractIndex, sop.StorageIndex)
}

// Execute state operation
func (sop *GetStateOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	contract, cerr := ctx.ContractDictionary.Decode(sop.ContractIndex)
	if cerr != nil {
		return cerr
	}
	storage, serr := ctx.StorageDictionary.Decode(sop.StorageIndex)
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
	Writable
	ContractIndex uint32      // encoded contract address
	StorageIndex  uint32      // encoded storage address
	Value         common.Hash // stored value
}

// Return set-state identifier
func (sop *SetStateOperation) GetOpId() int {
	return SetStateOperationID
}

// Create a new set-state operation.
func NewSetStateOperation(ContractIndex uint32, StorageIndex uint32, value common.Hash) *SetStateOperation {
	return &SetStateOperation{ContractIndex: ContractIndex, StorageIndex: StorageIndex, Value: value}
}

// Read set-state operation from a file.
func ReadSetStateOperation(file *os.File) (*SetStateOperation, error) {
	data := new(SetStateOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (sop *SetStateOperation) GetWritable() *Writable {
	return &sop.Writable
}

// Write a set-state operation.
func (sop *SetStateOperation) Write(f *os.File) {
	// group information into data slice
	var data = []any{sop.Writable.Get(), sop.ContractIndex, sop.StorageIndex, sop.Value.Bytes()}

	// write data to file
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatalf("SetStateOperation: %v", err)
		}
	}

	// debug message
	fmt.Printf("SetState: operation number: %v\t contract idx: %v\t storage idx: %v\t value: %v\n", sop.Writable.Get(), sop.ContractIndex, sop.StorageIndex, sop.Value.Hex())
}

// Execute state operation
func (sop *SetStateOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	contract, cerr := ctx.ContractDictionary.Decode(sop.ContractIndex)
	if cerr != nil {
		return cerr
	}
	storage, serr := ctx.StorageDictionary.Decode(sop.StorageIndex)
	if serr != nil {
		return serr
	}
	(*db).SetState(contract, storage, sop.Value)
	return nil
}

////////////////////////////////////////////////////////////
// GetCommittedState Operation
////////////////////////////////////////////////////////////

// GetCommittedState datastructure with encoded contract and storage addresses.
type GetCommittedStateOperation struct {
	Writable
	ContractIndex uint32 // encoded contract address
	StorageIndex  uint32 // encoded storage address
}

// Return get commited state operation identifier.
func (sop *GetCommittedStateOperation) GetOpId() int {
	return GetCommittedStateOperationID
}

// Create a new get commited state operation.
func NewGetCommittedStateOperation(ContractIndex uint32, StorageIndex uint32) *GetCommittedStateOperation {
	return &GetCommittedStateOperation{ContractIndex: ContractIndex, StorageIndex: StorageIndex}
}

// Read get commited state operation from a file.
func ReadGetCommittedStateOperation(file *os.File) (*GetCommittedStateOperation, error) {
	data := new(GetCommittedStateOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (sop *GetCommittedStateOperation) GetWritable() *Writable {
	return &sop.Writable
}

// Write a get commited state operation.
func (sop *GetCommittedStateOperation) Write(f *os.File) {
	// group information into data slice
	var data = []any{sop.Writable.Get(), sop.ContractIndex, sop.StorageIndex}

	// write data to file
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatalf("GetCommittedStateOperation: %v", err)
		}
	}

	// debug message
	fmt.Printf("GetCommittedState: operation number: %v\t contract idx: %v\t storage idx: %v\n", sop.Writable.Get(), sop.ContractIndex, sop.StorageIndex)
}

// Execute state operation
func (sop *GetCommittedStateOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	contract, cerr := ctx.ContractDictionary.Decode(sop.ContractIndex)
	if cerr != nil {
		return cerr
	}
	storage, serr := ctx.StorageDictionary.Decode(sop.StorageIndex)
	if serr != nil {
		return serr
	}
	(*db).GetCommittedState(contract, storage)
	return nil
}

////////////////////////////////////////////////////////////
// Snapshot Operation
////////////////////////////////////////////////////////////

// Snapshot datastructure with returned snapshot id
type SnapshotOperation struct {
	Writable
}

// Return snapshot operation identifier.
func (sop *SnapshotOperation) GetOpId() int {
	return SnapshotOperationID
}

// Create a new snapshot operation.
func NewSnapshotOperation() *SnapshotOperation {
	return &SnapshotOperation{}
}

// Read snapshot operation from a file.
func ReadSnapshotOperation(file *os.File) (*SnapshotOperation, error) {
	data := new(SnapshotOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (sop *SnapshotOperation) GetWritable() *Writable {
	return &sop.Writable
}

// Write a snapshot operation.
func (sop *SnapshotOperation) Write(f *os.File) {
	// group information into data slice
	var data = []any{sop.Writable.Get()}

	// write data to file
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatalf("SnapshotOperation: %v", err)
		}
	}

	// debug message
	fmt.Printf("Snapshot: operation number: %v\n", sop.Writable.Get())
}

// Execute state operation
func (sop *SnapshotOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	(*db).Snapshot()
	return nil
}

////////////////////////////////////////////////////////////
// RevertToSnapshot Operation
////////////////////////////////////////////////////////////

// RevertToSnapshot datastructure with returned snapshot id
type RevertToSnapshotOperation struct {
	Writable
	SnapshotID int
}

// Return snapshot operation identifier.
func (sop *RevertToSnapshotOperation) GetOpId() int {
	return RevertToSnapshotOperationID
}

// Create a new snapshot operation.
func NewRevertToSnapshotOperation(SnapshotID int) *RevertToSnapshotOperation {
	return &RevertToSnapshotOperation{SnapshotID: SnapshotID}
}

// Read snapshot operation from a file.
func ReadRevertToSnapshotOperation(file *os.File) (*RevertToSnapshotOperation, error) {
	var data struct {
		Writable
		SnapshotID int32
	}
	err := binary.Read(file, binary.LittleEndian, &data)
	rtso := &RevertToSnapshotOperation{Writable: data.Writable, SnapshotID: int(data.SnapshotID)}
	return rtso, err
}

// Return writeable interface
func (sop *RevertToSnapshotOperation) GetWritable() *Writable {
	return &sop.Writable
}

// Write a snapshot operation.
func (sop *RevertToSnapshotOperation) Write(f *os.File) {
	// group information into data slice
	var data = []any{sop.Writable.Get(), int32(sop.SnapshotID)}

	// write data to file
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatalf("RevertToSnapshotOperation: %v", err)
		}
	}

	// debug message
	fmt.Printf("RevertToSnapshot: operation number: %v\t snapshot id: %v\n", sop.Writable.Get(), sop.SnapshotID)
}

// Execute state operation
func (sop *RevertToSnapshotOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	(*db).RevertToSnapshot(int(sop.SnapshotID))
	return nil
}

////////////////////////////////////////////////////////////
// End of transaction Operation
////////////////////////////////////////////////////////////

// EndOfTransaction datastructure with returned snapshot id
type EndOfTransactionOperation struct {
	Writable
}

// Return snapshot operation identifier.
func (sop *EndOfTransactionOperation) GetOpId() int {
	return EndOfTransactionOperationID
}

// Create a new snapshot operation.
func NewEndOfTransactionOperation() *EndOfTransactionOperation {
	return &EndOfTransactionOperation{}
}

// Read snapshot operation from a file.
func ReadEndOfTransactionOperation(file *os.File) (*EndOfTransactionOperation, error) {
	data := new(EndOfTransactionOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (sop *EndOfTransactionOperation) GetWritable() *Writable {
	return &sop.Writable
}

// Write a snapshot operation.
func (sop *EndOfTransactionOperation) Write(f *os.File) {
	// group information into data slice
	var data = []any{sop.Writable.Get()}

	// write data to file
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatalf("EndOfTransactionOperation: %v", err)
		}
	}

	// debug message
	fmt.Printf("EndOfTransaction: operation number: %v\n", sop.Writable.Get())
}

// Execute state operation
func (sop *EndOfTransactionOperation) Execute(db *state.StateDB, ctx *ExecutionContext) error {
	return nil
}
