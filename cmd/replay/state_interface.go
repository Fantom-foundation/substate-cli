package replay

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

// numer of different state operations
const NumStateOperations = 3

// IDs of state operations
const BlockOperationID = 0
const GetStateOperationID = 1
const SetStateOperationID = 2

// State opertion interface
type StateOperation interface {
	GetOpId() int             // static operation identifier in the range from 0..n-1 for n different operation
	Write(uint64, []*os.File) // write operation
}

// polymorphic call to the Write() function of a state operation
func Write(so StateOperation, opNum uint64, file []*os.File) {
	so.Write(opNum, file)
}

// New Transaction POD datastructure
type BlockOperation struct {
	blockNumber uint64 // block number
}

func NewBlockOperation(blockNumber uint64) *BlockOperation {
	return &BlockOperation{blockNumber: blockNumber}
}

func (s *BlockOperation) GetOpId() int {
	return 0
}

func (s *BlockOperation) Write(opNum uint64, files []*os.File) {
	fmt.Printf("New Transaction: operation number: %v\t block number: %v\t\n", opNum, s.blockNumber)
}

// GetState POD datastructure
// Encodes contract/storage addresses via index
type GetStateOperation struct {
	contractIndex uint32 // encoded contract address
	storageIndex  uint32 // encoded storage address
}

func NewGetStateOperation(contractIndex uint32, storageIndex uint32) *GetStateOperation {
	return &GetStateOperation{contractIndex: contractIndex, storageIndex: storageIndex}
}

func (s *GetStateOperation) GetOpId() int {
	return 1
}

func (s *GetStateOperation) Write(opNum uint64, files []*os.File) {
	// store get state operation in little endian byte format
	// | opNum (8 bytes) | contract-index (4 bytes) | storage-index (4 bytes) |
	opNumByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(opNumByte, opNum)
	contractIndexByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(contractIndexByte, s.contractIndex)
	storageIndexByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(storageIndexByte, s.storageIndex)
	data := append(append(opNumByte, contractIndexByte...), storageIndexByte...)

	// write data into state file of operation
	if n, err := files[s.GetOpId()].Write(data); err != nil {
		log.Fatal(err)
	} else if n != len(data) {
		log.Fatalf("Data not written.")
	}
	fmt.Printf("GetState: operation number: %v\t contract idx: %v\t storage idx: %v\n", opNum, s.contractIndex, s.storageIndex)
}

// SetState POD datastructure
// Encodes contract/storage addresses via index
type SetStateOperation struct {
	contractIndex uint32      // encoded contract address
	storageIndex  uint32      // encoded storage address
	value         common.Hash // stored value
}

func NewSetStateOperation(contractIndex uint32, storageIndex uint32, value common.Hash) *SetStateOperation {
	return &SetStateOperation{contractIndex: contractIndex, storageIndex: storageIndex, value: value}
}

func (s *SetStateOperation) GetOpId() int {
	return 2
}

func (s *SetStateOperation) Write(opNum uint64, files []*os.File) {
	// store get state operation in little endian byte format
	// | opNum (8 bytes) | contract-index (4 bytes) | storage-index (4 bytes) | vale (32 bytes) |
	opNumByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(opNumByte, opNum)
	contractIndexByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(contractIndexByte, s.contractIndex)
	storageIndexByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(storageIndexByte, s.storageIndex)
	data := append(append(append(opNumByte, contractIndexByte...), storageIndexByte...), s.value.Bytes()...)

	// write data into state file of operation
	if n, err := files[s.GetOpId()].Write(data); err != nil {
		log.Fatal(err)
	} else if n != len(data) {
		log.Fatalf("Data not written.")
	}
	fmt.Printf("SetState: operation number: %v\t contract idx: %v\t storage idx: %v\t value: %v\n", opNum, s.contractIndex, s.storageIndex, s.value.Hex())
}
