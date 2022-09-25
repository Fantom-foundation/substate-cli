package replay

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

const NumStateOperations = 2

// GetState POD datastructure
// encodes contract/storage addresses via index
type GetStateOperation struct {
	cidx uint32 // encoded contract address
	sidx uint32 // encoded storage address
}

// SetState POD datastructure
// encodes contract/storage addresses via index
type SetStateOperation struct {
	cidx  uint32      // encoded contract address
	sidx  uint32      // encoded storage address
	value common.Hash // stored value
}

// State opertion interface
type StateOperation interface {
	OperationIndex() int      // operation index (0..n-1) for n different operation
	Write(uint64, []*os.File) // write operation
}

func NewGetStateOperation(cidx uint32, sidx uint32) *GetStateOperation {
	return &GetStateOperation{cidx: cidx, sidx: sidx}
}

func NewSetStateOperation(cidx uint32, sidx uint32, value common.Hash) *SetStateOperation {
	return &SetStateOperation{cidx: cidx, sidx: sidx, value: value}
}

func Write(so StateOperation, opNum uint64, file []*os.File) {
	so.Write(opNum, file)
}

func (s *GetStateOperation) OperationIndex() int {
	return 0
}

func (s *GetStateOperation) Write(opNum uint64, files []*os.File) {
	// store get state operation in little endian byte format
	// | opNum (8 bytes) | contract-index (4 bytes) | storage-index (4 bytes) |
	opNumByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(opNumByte, opNum)
	cidxByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(cidxByte, s.cidx)
	sidxByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(sidxByte, s.sidx)
	data := append(append(opNumByte, cidxByte...), sidxByte...)

	// write data into state file of operation
	if n, err := files[s.OperationIndex()].Write(data); err != nil {
		log.Fatal(err)
	} else if n != len(data) {
		log.Fatalf("Data not written.")
	}
	fmt.Printf("GetState: operation number: %v\t contract idx: %v\t storage idx: %v\n", opNum, s.cidx, s.sidx)
}

func (s *SetStateOperation) OperationIndex() int {
	return 1
}

func (s *SetStateOperation) Write(opNum uint64, files []*os.File) {
	// store get state operation in little endian byte format
	// | opNum (8 bytes) | contract-index (4 bytes) | storage-index (4 bytes) | vale (32 bytes) |
	opNumByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(opNumByte, opNum)
	cidxByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(cidxByte, s.cidx)
	sidxByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(sidxByte, s.sidx)
	data := append(append(append(opNumByte, cidxByte...), sidxByte...), s.value.Bytes()...)

	// write data into state file of operation
	if n, err := files[s.OperationIndex()].Write(data); err != nil {
		log.Fatal(err)
	} else if n != len(data) {
		log.Fatalf("Data not written.")
	}
	fmt.Printf("SetState: operation number: %v\t contract idx: %v\t storage idx: %v\t value: %v\n", opNum, s.cidx, s.sidx, s.value.Hex())
}
