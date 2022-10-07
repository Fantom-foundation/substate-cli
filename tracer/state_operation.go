package tracer

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Fantom-foundation/substate-cli/state"
)

// Stored Operations
// TODO: Remove word Operation from ID constants
const GetStateOperationID = 0
const SetStateOperationID = 1
const GetCommittedStateOperationID = 2
const SnapshotOperationID = 3
const RevertToSnapshotOperationID = 4
const CreateAccountOperationID = 5
const GetBalanceOperationID = 6
const GetCodeHashOperationID = 7
const SuicideOperationID = 8
const ExistOperationID = 9
const FinaliseOperationID = 10
const EndTransactionOperationID = 11 //last
const BeginBlockOperationID = 12
const EndBlockOperationID = 13

// Number of state operation identifiers
const NumOperations = EndBlockOperationID + 1 //last op + 1


// Output directory
var TraceDir string = "./"

// State operations' names
var idToLabel = [NumOperations]string{
	"GetState",
	"SetState",
	"GetCommittedState",
	"Snapshot",
	"RevertToSnapshot",
	"CreateAccount",
	"GetBalance",
	"GetCodeHash",
	"Suicide",
	"Exist",
	"Finalise",
	"EndTransaction",
	// Pseudo Operations
	"BeginBlock",
	"EndBlock",
}

// State operation's read functions
var readFunction = [NumOperations]func(*os.File) (StateOperation, error){
	ReadGetStateOperation,
	ReadSetStateOperation,
	ReadGetCommittedStateOperation,
	ReadSnapshotOperation,
	ReadRevertToSnapshotOperation,
	ReadCreateAccountOperation,
	ReadEndTransactionOperation,
}

// Get a label of a state operation
func GetLabel(i byte) string {
	if i < 0 || i >= NumOperations {
		log.Fatalf("GetLabel failed; index is out-of-bound")
	}
	return idToLabel[i]
}

////////////////////////////////////////////////////////////
// State Operation Interface
////////////////////////////////////////////////////////////

// State-operation interface
// TODO: Rename StateOperation to Operation
type StateOperation interface {
	GetOpId() byte                             // obtain operation identifier
	Write(*os.File)                            // write operation
	Execute(state.StateDB, *DictionaryContext) // execute operation
	Debug(*DictionaryContext)                  // print debug message for operation
}

// Read a state operation from file.
// TODO: Rename Read to ReadOperation
func Read(f *os.File) StateOperation {
	var (
		op StateOperation
		ID byte
	)

	// read ID from file
	err := binary.Read(f, binary.LittleEndian, &ID)
	if err == io.EOF {
		return nil
	} else if err != nil {
		log.Fatalf("Cannot read ID from file. Error:%v", err)
	}
	if ID >= NumOperations {
		log.Fatalf("ID out of range %v", ID)
	}

	// read state operation in binary format from file
	op, err = readFunction[ID](f)
	if err != nil {
		log.Fatalf("Failed to read operation %v. Error %v", GetLabel(ID), err)
	}
	if op.GetOpId() != ID {
		log.Fatalf("Generated object of type %v has wrong ID (%v) ", GetLabel(op.GetOpId()), GetLabel(ID))
	}
	return op
}

// Write state operation to file.
// TODO: Rename Write to WriteOperation
func Write(f *os.File, op StateOperation) {
	// write ID to file
	ID := op.GetOpId()
	if err := binary.Write(f, binary.LittleEndian, &ID); err != nil {
		log.Fatalf("Failed to write ID for operation %v. Error: %v", GetLabel(ID), err)
	}

	// write details of operation to file
	op.Write(f)
}

// Write slice in little-endian format to file (helper Function).
func writeSlice(f *os.File, data []any) {
	for _, val := range data {
		if err := binary.Write(f, binary.LittleEndian, val); err != nil {
			log.Fatalf("Failed to write binary data: %v", err)
		}
	}
}

// Print debug information of a state operation.
func Debug(ctx *DictionaryContext, op StateOperation) {
	fmt.Printf("%v:\n", GetLabel(op.GetOpId()))
	op.Debug(ctx)
}

// TODO: Remove from Operation from following structs 

////////////////////////////////////////////////////////////
// Begin Block Operation (Pseudo Operation)
////////////////////////////////////////////////////////////

// Block-operation data structure capturing the beginning of a block.
type BeginBlockOperation struct {
	BlockNumber uint64 // block number
}

// Return begin-block operation identifier.
func (bb *BeginBlockOperation) GetOpId() byte {
	return BeginBlockOperationID
}

// Create a new begin-block operation.
func NewBeginBlockOperation(bbNum uint64) *BeginBlockOperation {
	return &BeginBlockOperation{BlockNumber: bbNum}
}

// Write block operation (should never be invoked).
func (bb *BeginBlockOperation) Write(files *os.File) {
}

// Execute state operation.
func (bb *BeginBlockOperation) Execute(db state.StateDB, ctx *DictionaryContext) {
}

// Print a debug message.
func (bb *BeginBlockOperation) Debug(ctx *DictionaryContext) {
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
func (eb *EndBlockOperation) GetOpId() byte {
	return EndBlockOperationID
}

// Create a new end-block operation.
func NewEndBlockOperation(ebNum uint64) *EndBlockOperation {
	return &EndBlockOperation{BlockNumber: ebNum}
}

// Write end-block operation (should never be invoked).
func (eb *EndBlockOperation) Write(files *os.File) {
}

// Execute state operation.
func (eb *EndBlockOperation) Execute(db state.StateDB, ctx *DictionaryContext) {
}

// Print a debug message
func (eb *EndBlockOperation) Debug(ctx *DictionaryContext) {
	fmt.Printf("\tblock number: %v\n", eb.BlockNumber)
}

////////////////////////////////////////////////////////////
// GetState Operation
////////////////////////////////////////////////////////////

// GetState datastructure with encoded contract and storage addresses.
type GetStateOperation struct {
	ContractIndex uint32 // encoded contract address
	StorageIndex  uint32 // encoded storage address
}

// Return get-state operation identifier.
func (op *GetStateOperation) GetOpId() byte {
	return GetStateOperationID
}

// Create a new get-state operation.
func NewGetStateOperation(cIdx uint32, sIdx uint32) *GetStateOperation {
	return &GetStateOperation{ContractIndex: cIdx, StorageIndex: sIdx}
}

// Read get-state operation from a file.
func ReadGetStateOperation(file *os.File) (StateOperation, error) {
	data := new(GetStateOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Write a get-state operation in binary format to a file.
func (op *GetStateOperation) Write(f *os.File) {
	var data = []any{op.ContractIndex, op.StorageIndex}
	writeSlice(f, data)
}

// Execute get-state operation.
func (op *GetStateOperation) Execute(db state.StateDB, ctx *DictionaryContext) {
	contract := ctx.getContract(op.ContractIndex)
	storage := ctx.getStorage(op.StorageIndex)
	db.GetState(contract, storage)
}

// Print a debug message.
func (op *GetStateOperation) Debug(ctx *DictionaryContext) {
	fmt.Printf("\tcontract: %v\t storage: %v\n",
		ctx.getContract(op.ContractIndex),
		ctx.getStorage(op.StorageIndex))
}

////////////////////////////////////////////////////////////
// SetState Operation
////////////////////////////////////////////////////////////

// SetState datastructure with encoded contract and storage addresses, and value.
type SetStateOperation struct {
	ContractIndex uint32 // encoded contract address
	StorageIndex  uint32 // encoded storage address
	ValueIndex    uint64 // encoded storage value
}

// Return set-state identifier
func (op *SetStateOperation) GetOpId() byte {
	return SetStateOperationID
}

// Create a new set-state operation.
func NewSetStateOperation(cIdx uint32, sIdx uint32, vIdx uint64) *SetStateOperation {
	return &SetStateOperation{ContractIndex: cIdx, StorageIndex: sIdx, ValueIndex: vIdx}
}

// Read set-state operation from a file.
func ReadSetStateOperation(file *os.File) (StateOperation, error) {
	data := new(SetStateOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Write a set-state operation in binary format to a file.
func (op *SetStateOperation) Write(f *os.File) {
	var data = []any{op.ContractIndex, op.StorageIndex, op.ValueIndex}
	writeSlice(f, data)
}

// Execute set-state operation.
func (op *SetStateOperation) Execute(db state.StateDB, ctx *DictionaryContext) {
	contract := ctx.getContract(op.ContractIndex)
	storage := ctx.getStorage(op.StorageIndex)
	value := ctx.getValue(op.ValueIndex)
	db.SetState(contract, storage, value)
}

// Print a debug message.
func (op *SetStateOperation) Debug(ctx *DictionaryContext) {
	fmt.Printf("\tcontract: %v\t storage: %v\t value: %v\n",
		ctx.getContract(op.ContractIndex),
		ctx.getStorage(op.StorageIndex),
		ctx.getValue(op.ValueIndex))
}

////////////////////////////////////////////////////////////
// GetCommittedState Operation
////////////////////////////////////////////////////////////

// GetCommittedState datastructure with encoded contract and storage addresses.
type GetCommittedStateOperation struct {
	ContractIndex uint32 // encoded contract address
	StorageIndex  uint32 // encoded storage address
}

// Return get commited-state-operation identifier.
func (op *GetCommittedStateOperation) GetOpId() byte {
	return GetCommittedStateOperationID
}

// Create a new get-commited-state operation.
func NewGetCommittedStateOperation(cIdx uint32, sIdx uint32) *GetCommittedStateOperation {
	return &GetCommittedStateOperation{ContractIndex: cIdx, StorageIndex: sIdx}
}

// Read get-commited-state operation from a file.
func ReadGetCommittedStateOperation(file *os.File) (StateOperation, error) {
	data := new(GetCommittedStateOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Write a get-commited-state operation in binary format to file.
func (op *GetCommittedStateOperation) Write(f *os.File) {
	var data = []any{op.ContractIndex, op.StorageIndex}
	writeSlice(f, data)
}

// Execute get-committed-state operation.
func (op *GetCommittedStateOperation) Execute(db state.StateDB, ctx *DictionaryContext) {
	contract := ctx.getContract(op.ContractIndex)
	storage := ctx.getStorage(op.StorageIndex)
	db.GetCommittedState(contract, storage)
}

// Print details of get-committed-state operation
func (op *GetCommittedStateOperation) Debug(ctx *DictionaryContext) {
	fmt.Printf("\tcontract: %v\t storage: %v\n",
		ctx.getContract(op.ContractIndex),
		ctx.getStorage(op.StorageIndex))
}

////////////////////////////////////////////////////////////
// Snapshot Operation
////////////////////////////////////////////////////////////

// Snapshot datastructure with returned snapshot id
type SnapshotOperation struct {
}

// Return snapshot operation identifier.
func (op *SnapshotOperation) GetOpId() byte {
	return SnapshotOperationID
}

// Create a new snapshot operation.
func NewSnapshotOperation() *SnapshotOperation {
	return &SnapshotOperation{}
}

// Read a snapshot operation from a file.
func ReadSnapshotOperation(file *os.File) (StateOperation, error) {
	return NewSnapshotOperation(), nil
}

// Write the snapshot operation in binary format to file.
func (op *SnapshotOperation) Write(f *os.File) {
}

// Execute the snapshot operation.
func (op *SnapshotOperation) Execute(db state.StateDB, ctx *DictionaryContext) {
	db.Snapshot()
}

// Print the details for the snapshot operation.
func (op *SnapshotOperation) Debug(*DictionaryContext) {
}

////////////////////////////////////////////////////////////
// RevertToSnapshot Operation
////////////////////////////////////////////////////////////

// Revert-to-snapshot operation's datastructure with returned snapshot id
type RevertToSnapshotOperation struct {
	SnapshotID int
}

// Return revert-to-snapshot operation identifier.
func (op *RevertToSnapshotOperation) GetOpId() byte {
	return RevertToSnapshotOperationID
}

// Create a new revert-to-snapshot operation.
func NewRevertToSnapshotOperation(SnapshotID int) *RevertToSnapshotOperation {
	return &RevertToSnapshotOperation{SnapshotID: SnapshotID}
}

// Read a revert-to-snapshot operation in binary format from file.
func ReadRevertToSnapshotOperation(file *os.File) (StateOperation, error) {
	data := new(RevertToSnapshotOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Write a revert-to-snapshot operation in binary format to file.
func (op *RevertToSnapshotOperation) Write(f *os.File) {
	var data = []any{int32(op.SnapshotID)}
	writeSlice(f, data)
}

// Execute revert-to-snapshot operation.
func (op *RevertToSnapshotOperation) Execute(db state.StateDB, ctx *DictionaryContext) {
	db.RevertToSnapshot(op.SnapshotID)
}

// Print a debug message for revert-to-snapshot operation.
func (op *RevertToSnapshotOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tsnapshot id: %v\n", op.SnapshotID)
}

////////////////////////////////////////////////////////////
// CreateAccount Operation
////////////////////////////////////////////////////////////

// Create-account operation's datastructure with returned snapshot id
type CreateAccountOperation struct {
	ContractIndex uint32 // encoded contract address
}

// Return snapshot operation identifier.
func (op *CreateAccountOperation) GetOpId() byte {
	return CreateAccountOperationID
}

// Create a new snapshot operation.
func NewCreateAccountOperation(cIdx uint32) *CreateAccountOperation {
	return &CreateAccountOperation{ContractIndex: cIdx}
}

// Read snapshot operation in binary format from a file.
func ReadCreateAccountOperation(file *os.File) (StateOperation, error) {
	data := new(CreateAccountOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Write a snapshot operation in binary format to file.
func (op *CreateAccountOperation) Write(f *os.File) {
	var data = []any{op.ContractIndex}
	writeSlice(f, data)
}

// Execute snapshot operation.
func (op *CreateAccountOperation) Execute(db state.StateDB, ctx *DictionaryContext) {
	contract := ctx.getContract(op.ContractIndex)
	db.CreateAccount(contract)
}

// Print a debug message for snapshot operation.
func (op *CreateAccountOperation) Debug(ctx *DictionaryContext) {
	fmt.Printf("\tcontract: %v\n", ctx.getContract(op.ContractIndex))
}

////////////////////////////////////////////////////////////
// GetBalance Operation
////////////////////////////////////////////////////////////

// GetBalance datastructure with returned snapshot id
type GetBalanceOperation struct {
	Writable
	ContractIndex uint32
}

// Return snapshot operation identifier.
func (op *GetBalanceOperation) GetOpId() int {
	return GetBalanceOperationID
}

// Create a new snapshot operation.
func NewGetBalanceOperation(cIdx uint32) *GetBalanceOperation {
	return &GetBalanceOperation{ContractIndex: cIdx}
}

// Read snapshot operation from a file.
func ReadGetBalanceOperation(file *os.File) (*GetBalanceOperation, error) {
	data := new(GetBalanceOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (op *GetBalanceOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a snapshot operation.
func (op *GetBalanceOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get(), op.ContractIndex}
	WriteSlice(f, data)
}

// Execute state operation
func (op *GetBalanceOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	contract := getContract(ctx, op.ContractIndex)
	db.GetBalance(contract)
}

// Print a debug message
func (op *GetBalanceOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tcontract: %v\n", getContract(ctx,op.ContractIndex))
}

////////////////////////////////////////////////////////////
// GetCodeHash Operation
////////////////////////////////////////////////////////////

// GetCodeHash datastructure with returned snapshot id
type GetCodeHashOperation struct {
	Writable
	ContractIndex uint32
}

// Return snapshot operation identifier.
func (op *GetCodeHashOperation) GetOpId() int {
	return GetCodeHashOperationID
}

// Create a new snapshot operation.
func NewGetCodeHashOperation(cIdx uint32) *GetCodeHashOperation {
	return &GetCodeHashOperation{ContractIndex: cIdx}
}

// Read snapshot operation from a file.
func ReadGetCodeHashOperation(file *os.File) (*GetCodeHashOperation, error) {
	data := new(GetCodeHashOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (op *GetCodeHashOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a snapshot operation.
func (op *GetCodeHashOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get(), op.ContractIndex}
	WriteSlice(f, data)
}

// Execute state operation
func (op *GetCodeHashOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	contract := getContract(ctx, op.ContractIndex)
	db.GetCodeHash(contract)
}

// Print a debug message
func (op *GetCodeHashOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tcontract: %v\n", getContract(ctx,op.ContractIndex))
}

////////////////////////////////////////////////////////////
// Suicide Operation
////////////////////////////////////////////////////////////

// Suicide datastructure with returned snapshot id
type SuicideOperation struct {
	Writable
	ContractIndex uint32
}

// Return snapshot operation identifier.
func (op *SuicideOperation) GetOpId() int {
	return SuicideOperationID
}

// Create a new snapshot operation.
func NewSuicideOperation(cIdx uint32) *SuicideOperation {
	return &SuicideOperation{ContractIndex: cIdx}
}

// Read snapshot operation from a file.
func ReadSuicideOperation(file *os.File) (*SuicideOperation, error) {
	data := new(SuicideOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (op *SuicideOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a snapshot operation.
func (op *SuicideOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get(), op.ContractIndex}
	WriteSlice(f, data)
}

// Execute state operation
func (op *SuicideOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	contract := getContract(ctx, op.ContractIndex)
	db.Suicide(contract)
}

// Print a debug message
func (op *SuicideOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tcontract: %v\n", getContract(ctx, op.ContractIndex))
}

////////////////////////////////////////////////////////////
// Exist Operation
////////////////////////////////////////////////////////////

// Exist datastructure with returned snapshot id
type ExistOperation struct {
	Writable
	ContractIndex uint32
}

// Return snapshot operation identifier.
func (op *ExistOperation) GetOpId() int {
	return ExistOperationID
}

// Create a new snapshot operation.
func NewExistOperation(cIdx uint32) *ExistOperation {
	return &ExistOperation{ContractIndex: cIdx}
}

// Read snapshot operation from a file.
func ReadExistOperation(file *os.File) (*ExistOperation, error) {
	data := new(ExistOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (op *ExistOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a snapshot operation.
func (op *ExistOperation) Write(f *os.File) {
	var data = []any{op.Writable.Get(), op.ContractIndex}
	WriteSlice(f, data)
}

// Execute state operation
func (op *ExistOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	contract := getContract(ctx, op.ContractIndex)
	db.Exist(contract)
}

// Print a debug message
func (op *ExistOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tcontract: %v\n", getContract(ctx, op.ContractIndex))
}

////////////////////////////////////////////////////////////
// Finalise Operation
////////////////////////////////////////////////////////////

// Finalise datastructure with returned snapshot id
type FinaliseOperation struct {
	Writable
	DeleteEmptyObjects bool // encoded contract address
}

// Return snapshot operation identifier.
func (op *FinaliseOperation) GetOpId() int {
	return FinaliseOperationID
}

// Create a new snapshot operation.
func NewFinaliseOperation(deleteEmptyObjects bool) *FinaliseOperation {
	return &FinaliseOperation{DeleteEmptyObjects: deleteEmptyObjects}
}

// Read snapshot operation from a file.
func ReadFinaliseOperation(file *os.File) (*FinaliseOperation, error) {
	data := new(FinaliseOperation)
	err := binary.Read(file, binary.LittleEndian, data)
	return data, err
}

// Return writeable interface
func (op *FinaliseOperation) GetWritable() *Writable {
	return &op.Writable
}

// Write a snapshot operation.
func (op *FinaliseOperation) Write(f *os.File) {
	// group information into data slice
	var data = []any{op.Writable.Get(), op.DeleteEmptyObjects}
	WriteSlice(f, data)
}

// Execute state operation
func (op *FinaliseOperation) Execute(db state.StateDB, ctx *ExecutionContext) {
	db.Finalise(op.DeleteEmptyObjects)
}

// Print a debug message
func (op *FinaliseOperation) Debug(ctx *ExecutionContext) {
	fmt.Printf("\tdelete empty objects: %v\n",op.DeleteEmptyObjects)
}

////////////////////////////////////////////////////////////
// End of transaction Operation
////////////////////////////////////////////////////////////

// End-transaction operation's datastructure
type EndTransactionOperation struct {
}

// Return end-transaction operation identifier.
func (op *EndTransactionOperation) GetOpId() byte {
	return EndTransactionOperationID
}

// Create a new end-transaction operation.
func NewEndTransactionOperation() *EndTransactionOperation {
	return &EndTransactionOperation{}
}

// Read snapshot operation in binary format from a file.
func ReadEndTransactionOperation(file *os.File) (StateOperation, error) {
	return NewEndTransactionOperation(), nil
}

// Write a end-transaction operation in binary format to file.
func (op *EndTransactionOperation) Write(f *os.File) {
}

// Execute end-transaction operation.
func (op *EndTransactionOperation) Execute(db state.StateDB, ctx *DictionaryContext) {
}

// Print a debug message for end-transaction.
func (op *EndTransactionOperation) Debug(*DictionaryContext) {
}
