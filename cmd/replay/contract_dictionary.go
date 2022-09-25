package replay

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math"
	"os"
)

// Dictioanary data structure
type ContractDictionary struct {
	contractToIdx map[common.Address]uint32 // contract to index map for encoding
	idxToContract []common.Address          // contract address slice for decoding
}

// Create new dictionary
func NewContractDictionary() *ContractDictionary {
	p := new(ContractDictionary)
	p.contractToIdx = map[common.Address]uint32{}
	p.idxToContract = []common.Address{}
	return p
}

// Encode an address in the dictionary to an index
func (cd *ContractDictionary) Encode(addr common.Address) (uint32, error) {
	var (
		idx uint32
		ok  bool
		err error = nil
	)
	if idx, ok = cd.contractToIdx[addr]; !ok {
		idx = uint32(len(cd.idxToContract))
		if idx != math.MaxUint32 {
			cd.contractToIdx[addr] = idx
			cd.idxToContract = append(cd.idxToContract, addr)
		} else {
			idx = 0
			err = errors.New("Contract dictionary exhausted")
		}
	}
	return idx, err
}

// Decode a dictionary index to an address
func (cd *ContractDictionary) Decode(idx uint32) (common.Address, error) {
	var (
		addr common.Address
		err  error
	)
	if idx < uint32(len(cd.idxToContract)) {
		addr = cd.idxToContract[idx]
		err = nil
	} else {
		addr = common.Address{}
		err = errors.New("Index out-of-bound")
	}
	return addr, err
}

// Write dictionary to a binary file
func (cd *ContractDictionary) Write(filename string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	for _, addr := range cd.idxToContract {
		data := addr.Bytes()
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

// Read dictionary from a binary file
func (cd *ContractDictionary) Read(filename string) {
	cd.contractToIdx = map[common.Address]uint32{}
	cd.idxToContract = []common.Address{}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	data := common.Address{}.Bytes()
	for {
		n, err := f.Read(data)
		if n == 0 {
			break
		} else if n < len(data) || err != nil {
			log.Fatalf("Contract dictionary file is corrupted")
		}
		addr := common.BytesToAddress(data)
		idx := uint32(len(cd.idxToContract))
		if idx == math.MaxUint32 {
			log.Fatalf("Too many entries in dictionary; file corrupted")
		}
		cd.contractToIdx[addr] = uint32(len(cd.idxToContract))
		cd.idxToContract = append(cd.idxToContract, addr)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
