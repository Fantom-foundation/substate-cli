package main

import (
	"database/sql"
	"fmt"
	"log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/syndtr/goleveldb/leveldb"
	_ "github.com/mattn/go-sqlite3"
)

var (
	CodeRegistry map[common.Address][]byte = make(map[common.Address][]byte)
	JumpDestFrequency    map[common.Address]map[uint64]uint64 = make(map[common.Address]map[uint64]uint64)
	ContractDB   string = "./contracts.db"
)

// read from contract db to populate the code registry
func readContracts() {
	db, err := leveldb.OpenFile(ContractDB, nil)
	if err != nil {
		log.Fatal("Cannot open codedb!")
	}
	defer db.Close()
	iter := db.NewIterator(nil, nil)
	ctr := 0;
	for iter.Next() {
		address := string(iter.Key())
		code := iter.Value()
		CodeRegistry[common.HexToAddress(address)] = code
		ctr++
	}
	fmt.Printf("Read %v contracts.\n", ctr)
}

func readJumpDestFrequency() {
	// open sqlite3 database
	db, err := sql.Open("sqlite3", "./jumpdest.db") // Open the created SQLite File
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM JumpDestFrequency;")
        if err != nil {
          log.Fatal(err)
        }
        defer rows.Close()
	// read from table and populate the JumpDestFrequency map
	ctr:=1
	for rows.Next() {
          var address string
          var jumpdestpc uint64
          var frequency uint64
          err = rows.Scan(&address, &jumpdestpc, &frequency)
          if err != nil {
            log.Fatal(err)
          }
	  contract := common.HexToAddress(address)
  	  if JumpDestFrequency[contract] == nil {
		JumpDestFrequency[contract] = make(map[uint64]uint64)
	  }
	  JumpDestFrequency[contract][jumpdestpc] = frequency
	  ctr++
        }
	fmt.Printf("Read %v JUMPDEST frequencies.\n", ctr)
}


func main() {
	fmt.Printf("Read contracts database ...\n")
	readContracts()

	fmt.Printf("Read JUMPDEST frequencies database ...\n")
	readJumpDestFrequency()
	fmt.Printf("Done.\n")
}
