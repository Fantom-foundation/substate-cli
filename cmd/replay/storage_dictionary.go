package replay

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math"
	"os"
)

// Dictioanary data structure
type StorageDictionary struct {
	storageToIdx map[common.Hash]uint32  // storage address to index map for encoding
	idxToStorage []common.Hash         // storage address slice for decoding 
}

// Create new dictionary
func NewStorageDictionary() *StorageDictionary {
	p := new(StorageDictionary)
	p.storageToIdx = map[common.Hash]uint32{}
	p.idxToStorage = []common.Hash{}
	return p
}

// Encode an address in the dictionary to an index
func (cd *StorageDictionary) Encode(key common.Hash) (uint32, error) {
	var (
		idx uint32
		ok  bool
		err error = nil
	)
	if idx, ok = cd.storageToIdx[key]; !ok {
		idx = uint32(len(cd.idxToStorage))
		if idx != math.MaxUint32 {
			cd.storageToIdx[key] = idx
			cd.idxToStorage = append(cd.idxToStorage, key)
		} else {
			idx = 0
			err = errors.New("Storage dictionary exhausted")
		}
	}
	return idx, err
}

// Decode a dictionary index to an address
func (cd *StorageDictionary) Decode(idx uint32) (common.Hash, error) {
	var (
		key common.Hash
		err  error
	)
	if idx < uint32(len(cd.idxToStorage)) {
		key = cd.idxToStorage[idx]
		err = nil
	} else {
		key = common.Hash{}
		err = errors.New("Index out-of-bound")
	}
	return key, err
}

// Write dictionary to a binary file
func (cd *StorageDictionary) Write(filename string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	for _, key := range cd.idxToStorage {
		data := key.Bytes()
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

// Read dictionary from a binary file
func (cd *StorageDictionary) Read(filename string) {
	cd.storageToIdx = map[common.Hash]uint32{}
	cd.idxToStorage = []common.Hash{}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	data := common.Hash{}.Bytes()
	for {
		n, err := f.Read(data)
		if n == 0 {
			break
		} else if n < len(data) || err != nil {
			log.Fatalf("Storage dictionary file is corrupted")
		}
		key := common.BytesToHash(data)
		idx := uint32(len(cd.idxToStorage))
		if idx == math.MaxUint32 {
			log.Fatalf("Too many entries in dictionary; file corrupted")
		}
		cd.storageToIdx[key] = uint32(len(cd.idxToStorage))
		cd.idxToStorage = append(cd.idxToStorage, key)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
