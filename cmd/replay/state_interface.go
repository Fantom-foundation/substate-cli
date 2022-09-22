package replay

import (
	"fmt"
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
//	GetContractIndex() uint32
//	GetStateIndex() uint32
//	GetValue() common.Hash
	Write(uint64)
}

func NewGetStateOperation(cidx uint32, sidx uint32) *GetStateOperation {
	return &GetStateOperation{cidx: cidx, sidx: sidx}
}

func NewSetStateOperation(cidx uint32, sidx uint32, value common.Hash) *SetStateOperation {
	return &SetStateOperation{cidx: cidx, sidx: sidx, value: value}
}

func Write(so StateOperation, prnr uint64) {
	so.Write(prnr)
}

func (s *GetStateOperation) Write(prnr uint64) {
	fmt.Printf("GetState: operation number: %v\t contract idx: %v\t storage idx: %v\n", prnr, s.cidx, s.sidx)
}

func (s *SetStateOperation) Write(prnr uint64) {
	fmt.Printf("SetState: operation number: %v\t contract idx: %v\t storage idx: %v\t value: %v\n", prnr, s.cidx, s.sidx, s.value.Hex())
}
