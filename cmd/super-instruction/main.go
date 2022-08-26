package main

import (
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

var (
	Code                     = map[common.Address][]byte{}
	JumpDestFrequency        = map[common.Address]map[uint64]uint64{}
	ContractDB        string = "./contracts.db"
)

// read from contract db to populate the map for contracts
func readContracts() {
	// open leveldb database
	db, err := leveldb.OpenFile(ContractDB, nil)
	if err != nil {
		log.Fatal("Cannot open codedb!")
	}
	defer db.Close()
	// read all contracts and populate map
	iter := db.NewIterator(nil, nil)
	ctr := 0
	for iter.Next() {
		address := string(iter.Key())
		code := iter.Value()
		Code[common.HexToAddress(address)] = code
		ctr++
	}
	fmt.Printf("Read %v contracts.\n", ctr)
}

// read the JUMPDEST frequency database
func readJumpDestFrequency() {
	// open sqlite3 database
	db, err := sql.Open("sqlite3", "./jumpdest.db")
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
	ctr := 1
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

// write basic blocks into a sqlite3 database
func writeBasicBlocks() {
	// open sqlite3 database
	db, err := sql.Open("sqlite3", "./basicblock.db") // Open the created SQLite File
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	// drop jump-dest table
	const dropBasicBlock string = `DROP TABLE IF EXISTS BasicBlock;`
	statement, err := db.Prepare(dropBasicBlock)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err.Error())
	}

	// create new basic block table
	const createBasicBlock string = `
	CREATE TABLE BasicBlock (
	 contract TEXT, 
	 pc NUMERIC,
	 frequency NUMERIC,
	 instructions BLOB
	);`
	statement, err = db.Prepare(createBasicBlock)
	if err != nil {
		log.Fatal(err.Error())
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err.Error())
	}

	// populate values
	insertFrequency := `INSERT INTO BasicBlock(contract, pc, frequency, instructions) VALUES (?, ?, ?, ?)`
	statement, err = db.Prepare(insertFrequency)
	if err != nil {
		log.Fatal(err.Error())
	}
	for contract, freqMap := range JumpDestFrequency {
		for start, freq := range freqMap {
			pc := start
			instructions := []vm.OpCode{}
			length := uint64(len(Code[contract]))
			for {

				// exceed code segment
				if pc >= length {
					break
				}

				// fetch op-code
				op := vm.OpCode(Code[contract][pc])
				instructions = append(instructions, op)

				// end of basic block?
				if op == vm.JUMP ||
					op == vm.JUMPI ||
					op == vm.STOP ||
					op == vm.RETURN ||
					op == vm.REVERT ||
					op == vm.SELFDESTRUCT {
					break
				}

				// skip constant of a push operation
				if op >= vm.PUSH1 && op <= vm.PUSH32 {
					numbits := op - vm.PUSH1 + 1
					if numbits >= 8 {
						for ; numbits >= 16; numbits -= 16 {
							pc += 16
						}
						for ; numbits >= 8; numbits -= 8 {
							pc += 8
						}
					}
					switch numbits {
					case 1:
						pc += 1
					case 2:
						pc += 2
					case 3:
						pc += 3
					case 4:
						pc += 4
					case 5:
						pc += 5
					case 6:
						pc += 6
					case 7:
						pc += 7
					}
				}

				// skip to next instruction
				pc++
			}
			_, err = statement.Exec(contract.String(), pc, freq, instructions)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}
}

func main() {
	fmt.Printf("Read contracts database ...\n")
	readContracts()

	fmt.Printf("Read JUMPDEST frequencies database ...\n")
	readJumpDestFrequency()

	fmt.Printf("Write basic blocks ...\n")
	writeBasicBlocks()

	fmt.Printf("Done.\n")
}
