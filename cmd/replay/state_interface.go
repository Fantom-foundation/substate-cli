package replay

import (
	"binary"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
)

type GetStateOperation struct {
	cidx  uint32
	sidx  uint32
}

type SetStateOperation struct {
	cidx  uint32
	sidx  uint32
	value common.Hash
}

type StateOperation interface {
	Write(uint64, []os.File)
}

func NewGetStateOperation(cidx uint32, sidx uint32) *GetStateOperation {
	return &GetStateOperation{cidx: cidx, sidx: sidx}
}

func NewSetStateOperation(cidx uint32, sidx uint32, value common.Hash) *SetStateOperation {
	return &SetStateOperation{cidx: cidx, sidx: sidx, value: value}
}

func Write(so StateOperation, opNum uint64, file []os.File) {
	so.Write(opNum, file)
}

func (s *GetStateOperation) Write(opNum uint64, file []os.File) {
	opNumByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(opNumByte, opNum)
	cidxByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(cidxByte, cidx)
	sidxByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(sidxByte, sidx)
	data := append(opNumByte,cidxByte,sidxByte...)
	if n, err := f[0].Write(data); err != nil {
		log.Fatal(err)
        } else if n != len(data) {
		log.Fatalf("Data not written.")
	}
	fmt.Printf("GetState: operation number: %v\t contract idx: %v\t storage idx: %v\n", opNum, s.cidx, s.sidx)
}

func (s *SetStateOperation) Write(opNum uint64, file []os.File) {
	opNumByte := make([]byte, 8)
	binary.LittleEndian.PutUint64(opNumByte, opNum)
	cidxByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(cidxByte, cidx)
	sidxByte := make([]byte, 4)
	binary.LittleEndian.PutUint32(sidxByte, sidx)
	data := append(opNumByte,cidxByte,sidxByte...)
	if n, err := f[1].Write(data); err != nil {
		log.Fatal(err)
        } else if n != len(data) {
		log.Fatalf("Data not written.")
	}
	fmt.Printf("SetState: operation number: %v\t contract idx: %v\t storage idx: %v\t value: %v\n", opNum, s.cidx, s.sidx, s.value.Hex())
}
