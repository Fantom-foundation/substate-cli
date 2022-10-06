package tracer

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Fantom-foundation/substate-cli/state"
	"github.com/ethereum/go-ethereum/common"
)

// Stored Operations
const GetStateOperationID = 0
const SetStateOperationID = 1
const GetCommittedStateOperationID = 2
const SnapshotOperationID = 3
const RevertToSnapshotOperationID = 4
const CreateAccountOperationID = 5
const EndTransactionOperationID = 6 //last

// Number of write operations
const NumWriteOperations = EndTransactionOperationID + 1

// Operation IDs
// Pseudo Operations (not stored on file but generated while recording)
const BeginBlockOperationID = NumWriteOperations
const EndBlockOperationID = NumWriteOperations + 1

// Number of state operation identifiers
const NumOperations = EndBlockOperationID + 1 //last op + 1

// Number of pseudo operations that are not written to a file
const NumPseudoOperations = NumOperations - NumWriteOperations

// Output directory
var TraceDir string = "./"

// State operations' filenames
var idToFilename = [NumWriteOperations]string{
	"op-getstate.dat",
	"op-setstate.dat",
	"op-getcommittedstate.dat",
	"op-snapshot.dat",
	"op-reverttosnapshot.dat",
	"op-createaccount.dat",
	"op-endoftransaction.dat",
}

// State operations' names
var idToLabel = [NumOperations]string{
	"GetState",
	"SetState",
	"GetCommittedState",
	"Snapshot",
	"RevertToSnapshot",
	"CreateAccount",
	"EndOfTransaction",
	// Pseudo Operations
	"BeginBlock",      
	"EndBlock",
}

// Get filename of a state operation that is written to a file
func GetFilename(i int) string {
	if i < 0 || i >= NumWriteOperations {
		log.Fatalf("GetFilename failed; index is out-of-bound")
	}
	return TraceDir + idToFilename[i]
}

// Get a label of a state operation
func GetLabel(i int) string {
	if i < 0 || i >= NumOperations {
		log.Fatalf("GetLabel failed; index is out-of-bound")
	}
	return idToLabel[i]
}

////////////////////////////////////////////////////////////
// Execution Context
////////////////////////////////////////////////////////////

// ExecutionContext contains the contract/storage dictionaries
// so that a recorded StateDB operation can be executed.
type ExecutionContext struct {
	ContractDictionary *ContractDictionary // dictionary to compact contract addresses
	StorageDictionary  *StorageDictionary  // dictionary to compact storage addresses
	ValueDictionary    *ValueDictionary    // dictionary to compact storage values
}

// Create new execution context.
func NewExecutionContext(cDict *ContractDictionary, sDict *StorageDictionary, vDict *ValueDictionary) *ExecutionContext {
	return &ExecutionContext{ContractDictionary: cDict, StorageDictionary: sDict, ValueDictionary: vDict}
}

// Get the contract address for a given contract index.
func getContract(ctx *ExecutionContext, cIdx uint32) common.Address {
	contract, err := ctx.ContractDictionary.Decode(cIdx)
	if err != nil {
		log.Fatalf("Contract index could not be decoded, error: %v", err)
	}
	return contract
}

// Get the storage address for a given storage address index.
func getStorage(ctx *ExecutionContext, sIdx uint32) common.Hash {
	storage, err := ctx.StorageDictionary.Decode(sIdx)
	if err != nil {
		log.Fatalf("Storage index could not be decoded, error: %v", err)
	}
	return storage
}

// Get the storage value for a given value index.
func getValue(ctx *ExecutionContext, vIdx uint64) common.Hash {
	value, err := ctx.ValueDictionary.Decode(vIdx)
	if err != nil {
		log.Fatalf("Value index could not be decoded, error: %v", err)
	}
	return value
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

// Print debug message.
func (w *Writable) Debug() {
	fmt.Printf("(%v)", w.Get())
}

////////////////////////////////////////////////////////////
// State Operation Interface
////////////////////////////////////////////////////////////

// TODO: Perhaps have in future two interfaces for
//       1) Pseudo Operations
//       2) Writable Operations

// State-operation interface
type StateOperation interface {
	GetOpId() int                             // obtain operation identifier
	GetWritable() *Writable                   // obtain writeable interface
	Write(*os.File)                           // write operation
	Execute(state.StateDB, *ExecutionContext) // execute operation
	Debug(*ExecutionContext)                  // print debug message for operation
}

// Read a state operation from file.
func Read(f *os.File, ID int) *StateOperation {
	var (
		op  StateOperation
		err error = nil
	)

	switch ID {
	case GetStateOperationID:
		op, err = ReadGetStateOperation(f)
	case SetStateOperationID:
		op, err = ReadSetStateOperation(f)
	case GetCommittedStateOperationID:
		op, err = ReadGetCommittedStateOperation(f)
	case SnapshotOperationID:
		op, err = ReadSnapshotOperation(f)
	case RevertToSnapshotOperationID:
		op, err = ReadRevertToSnapshotOperation(f)
	case CreateAccountOperationID:
		op, err = ReadCreateAccountOperation(f)
	case EndTransactionOperationID:
		op, err = ReadEndTransactionOperation(f)
	default:
		if ID >= 0 && ID < NumWriteOperations {
			log.Fatalf("Read operation %v not implemented", GetLabel(op.GetOpId()))
		} else if ID >= NumWriteOperations && ID < NumOperations {
			log.Fatalf("Cannot read pseudo-operation %v from file", GetLabel(op.GetOpId()))
		} else {
			log.Fatalf("ID out of range", GetLabel(op.GetOpId()))
		}
	}
	if err == io.EOF {
		return nil
	} else if err != nil {
		log.Fatalf("Failed to read operation %v. Error %v", op.GetOpId(), err)
	}
	return &op
}

// Write slice in little-endian format to file.
func WriteSlice(f *os.File, data []any) {
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatalf("Failed to write binary data: %v", err)
		}
	}
}

// Print debug information of a state operation.
func Debug(ctx *ExecutionContext, op *StateOperation) {
	fmt.Printf("%v: ", GetLabel((*op).GetOpId()))
	w := (*op).GetWritable()
	if w != nil {
		w.Debug()
	}
	fmt.Println()
	(*op).Debug(ctx)
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
func NewBeginBlockOperation(bbNum uint64) *BeginBlockOperation {
	return &BeginBlockOperation{BlockNumber: bbNum}
}

// Return writeable interface (not implemented for pseudo operations).
func (bb *BeginBlockOperation) GetWritable() *Writable {
	return nil
}

// Write block operation (should never be invoked).
func (bb *BeginBlockOperation) Write(files *os.File) {
	log.Fatalf("Begin-block operation for block %v attempted to be written", bb.BlockNumber)
}

// Execute state operation.
func (bb *BeginBlockOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	log.Fatalf("Begin-block operation for block %v attempted to be executed", bb.BlockNumber)
}

// Print a debug message.
func (bb *BeginBlockOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tblock number: %v\n", bb.BlockNumber)
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
func NewEndBlockOperation(ebNum uint64) *EndBlockOperation {
	return &EndBlockOperation{BlockNumber: ebNum}
}

// Return writeable interface (not implemented for pseudo operations).
func (eb *EndBlockOperation) GetWritable() *Writable {
	return nil
}

// Write end-block operation (should never be invoked).
func (eb *EndBlockOperation) Write(files *os.File) {
	log.Fatalf("End-block operation for block %v attempted to be written", eb.BlockNumber)
}

// Execute state operation.
func (eb *EndBlockOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	log.Fatalf("End-block operation for block %v attempted to be executed", eb.BlockNumber)
}

// Print a debug message
func (eb *EndBlockOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tblock number: %v\n", eb.BlockNumber)
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
func (op *GetStateOperation) GetOpId() int {
	return GetStateOperationID
}

// Create a new get-state operation.
func NewGetStateOperation(cIdx uint32, sIdx uint32) *GetStateOperation {
	return &GetStateOperation{ContractIndex: cIdx, StorageIndex: sIdx}
}

// Read get-state operation from a file.
func ReadGetStateOperation(file *os.File) (*GetStateOperation, error) {
	data := new(GetStateOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface.
func (op *GetStateOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a get-state operation in binary format to a file.
func (op *GetStateOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get(), op.ContractIndex, op.StorageIndex}
	WriteSlice(f, data)
}

// Execute get-state operation.
func (op *GetStateOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	contract := getContract(ctx, op.ContractIndex)
	storage := getStorage(ctx, op.StorageIndex)
	db.GetState(contract, storage)
}

// Print a debug message.
func (op *GetStateOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tcontract: %v\t storage: %v\n", getContract(ctx, op.ContractIndex), getStorage(ctx, op.StorageIndex))
}

////////////////////////////////////////////////////////////
// SetState Operation
////////////////////////////////////////////////////////////

// SetState datastructure with encoded contract and storage addresses, and value.
type SetStateOperation struct {
	Writable
	ContractIndex uint32 // encoded contract address
	StorageIndex  uint32 // encoded storage address
	ValueIndex    uint64 // encoded storage value
}

// Return set-state identifier
func (op *SetStateOperation) GetOpId() int {
	return SetStateOperationID
}

// Create a new set-state operation.
func NewSetStateOperation(cIdx uint32, sIdx uint32, vIdx uint64) *SetStateOperation {
	return &SetStateOperation{ContractIndex: cIdx, StorageIndex: sIdx, ValueIndex: vIdx}
}

// Read set-state operation from a file.
func ReadSetStateOperation(file *os.File) (*SetStateOperation, error) {
	data := new(SetStateOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface.
func (op *SetStateOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a set-state operation in binary format to a file.
func (op *SetStateOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get(), op.ContractIndex, op.StorageIndex, op.ValueIndex}
	WriteSlice(f, data)
}

// Execute set-state operation.
func (op *SetStateOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	contract := getContract(ctx, op.ContractIndex)
	storage := getStorage(ctx, op.StorageIndex)
	value := getValue(ctx, op.ValueIndex)
	db.SetState(contract, storage, value)
}

// Print a debug message.
func (op *SetStateOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tcontract: %v\t storage: %v\t value: %v\n", getContract(ctx, op.ContractIndex), getStorage(ctx, op.StorageIndex), getValue(ctx, op.ValueIndex))
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
func (op *GetCommittedStateOperation) GetOpId() int {
	return GetCommittedStateOperationID
}

// Create a new get commited state operation.
func NewGetCommittedStateOperation(cIdx uint32, sIdx uint32) *GetCommittedStateOperation {
	return &GetCommittedStateOperation{ContractIndex: cIdx, StorageIndex: sIdx}
}

// Read get commited state operation from a file.
func ReadGetCommittedStateOperation(file *os.File) (*GetCommittedStateOperation, error) {
	data := new(GetCommittedStateOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (op *GetCommittedStateOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a get commited state operation in binary format to file.
func (op *GetCommittedStateOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get(), op.ContractIndex, op.StorageIndex}
	WriteSlice(f, data)
}

// Execute state operation.
func (op *GetCommittedStateOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	contract := getContract(ctx, op.ContractIndex)
	storage := getStorage(ctx, op.StorageIndex)
	db.GetCommittedState(contract, storage)
}

// Print a debug message.
func (op *GetCommittedStateOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tcontract: %v\t storage: %v\n", getContract(ctx, op.ContractIndex), getStorage(ctx, op.StorageIndex))
}

////////////////////////////////////////////////////////////
// Snapshot Operation
////////////////////////////////////////////////////////////

// Snapshot datastructure with returned snapshot id
type SnapshotOperation struct {
	Writable
}

// Return snapshot operation identifier.
func (op *SnapshotOperation) GetOpId() int {
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

// Return writeable interface.
func (op *SnapshotOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a snapshot operation in binary format to file.
func (op *SnapshotOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get()}
	WriteSlice(f, data)
}

// Execute state operation.
func (op *SnapshotOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	db.Snapshot()
}

// Print a debug message for snapshot operation.
func (op *SnapshotOperation) Debug(*ExecutionContext) {
}

////////////////////////////////////////////////////////////
// RevertToSnapshot Operation
////////////////////////////////////////////////////////////

// Revert-to-snapshot operation's datastructure with returned snapshot id
type RevertToSnapshotOperation struct {
	Writable
	SnapshotID int
}

// Return snapshot operation identifier.
func (op *RevertToSnapshotOperation) GetOpId() int {
	return RevertToSnapshotOperationID
}

// Create a new snapshot operation.
func NewRevertToSnapshotOperation(SnapshotID int) *RevertToSnapshotOperation {
	return &RevertToSnapshotOperation{SnapshotID: SnapshotID}
}

// Read snapshot operation in binary format from file.
func ReadRevertToSnapshotOperation(file *os.File) (*RevertToSnapshotOperation, error) {
	var data struct {
		Writable
		SnapshotID int32
	}
	err := binary.Read(file, binary.LittleEndian, &data)
	rtso := &RevertToSnapshotOperation{Writable: data.Writable, SnapshotID: int(data.SnapshotID)}

	return rtso, err
}

// Return writeable interface.
func (op *RevertToSnapshotOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a snapshot operation in binary format to file.
func (op *RevertToSnapshotOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get(), int32(op.SnapshotID)}
	WriteSlice(f, data)
}

// Execute revert-to-snapshot operation.
func (op *RevertToSnapshotOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	db.RevertToSnapshot(op.SnapshotID)
}

// Print a debug message for revert-to-snapshot operation.
func (op *RevertToSnapshotOperation) Debug(*ExecutionContext) {
	fmt.Printf("RevertToSnapshot: operation number: %v\t snapshot id: %v\n", op.Writable.Get(), op.SnapshotID)
}

////////////////////////////////////////////////////////////
// CreateAccount Operation
////////////////////////////////////////////////////////////

// Create-account operation's datastructure with returned snapshot id
type CreateAccountOperation struct {
	Writable
	ContractIndex uint32 // encoded contract address
}

// Return snapshot operation identifier.
func (op *CreateAccountOperation) GetOpId() int {
	return CreateAccountOperationID
}

// Create a new snapshot operation.
func NewCreateAccountOperation(cIdx uint32) *CreateAccountOperation {
	return &CreateAccountOperation{ContractIndex: cIdx}
}

// Read snapshot operation in binary format from a file.
func ReadCreateAccountOperation(file *os.File) (*CreateAccountOperation, error) {
	data := new(CreateAccountOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface.
func (op *CreateAccountOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a snapshot operation in binary format to file.
func (op *CreateAccountOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get(), op.ContractIndex}
	WriteSlice(f, data)
}

// Execute snapshot operation.
func (op *CreateAccountOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	contract := getContract(ctx, op.ContractIndex)
	db.CreateAccount(contract)
}

// Print a debug message for snapshot operation.
func (op *CreateAccountOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tcontract: %v\n", getContract(ctx, op.ContractIndex))
}

////////////////////////////////////////////////////////////
// End of transaction Operation
////////////////////////////////////////////////////////////

// End-transaction operation's datastructure
type EndTransactionOperation struct {
	Writable
}

// Return end-transaction operation identifier.
func (op *EndTransactionOperation) GetOpId() int {
	return EndTransactionOperationID
}

// Create a new end-transaction operation.
func NewEndTransactionOperation() *EndTransactionOperation {
	return &EndTransactionOperation{}
}

// Read snapshot operation in binary format from a file.
func ReadEndTransactionOperation(file *os.File) (*EndTransactionOperation, error) {
	data := new(EndTransactionOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface.
func (op *EndTransactionOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a end-transaction operation in binary format to file.
func (op *EndTransactionOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get()}
	WriteSlice(f, data)
}

// Execute end-transaction operation.
func (op *EndTransactionOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
}

// Print a debug message for end-transaction.
func (op *EndTransactionOperation) Debug(*ExecutionContext) {
}
