package replay

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/substate"
	"github.com/syndtr/goleveldb/leveldb"
	cli "gopkg.in/urfave/cli.v1"
)

// record-replay: substate-cli code command
var GetCodeCommand = cli.Command{
	Action:    getCodeAction,
	Name:      "code",
	Usage:     "write all contracts into a database",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		substate.WorkersFlag,
		substate.SubstateDirFlag,
		ContractDBFlag,
		ChainIDFlag,
	},
	Description: `
The substate-cli code command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay transactions.

The contracts of the block range are written into a levelDB database.
`,
}

// contract-db filename
var ContractDBFlag = cli.StringFlag{
	Name:  "contractdb",
	Usage: "Contract database name for smart contracts",
	Value: "./contracts.db",
}
var ContractDB = ContractDBFlag.Value

// registry to keep track the bytecode of a smart contract
var (
	CodeRegistryMutex sync.Mutex
	CodeRegistry      map[common.Address][]byte
)

// register the code of an account
func registerContract(contract common.Address, code []byte) {
	if len(code) > 0 {
		CodeRegistryMutex.Lock()
		if _, ok := CodeRegistry[contract]; !ok {
			CodeRegistry[contract] = code
		}
		CodeRegistryMutex.Unlock()
	}
}

// write code registry to a levelDB instance
func writeContracts() {
	db, err := leveldb.OpenFile(ContractDB, nil)
	if err != nil {
		log.Fatal("Cannot open codedb!")
	}
	defer db.Close()

	for account, code := range CodeRegistry {
		err = db.Put([]byte(account.String()), code, nil)
		if err != nil {
			log.Fatal("Writing of code into levelDB failed.")
		}
	}
}

// getCodeTask finds the code block of alll contracts
func getCodeTask(block uint64, tx int, st *substate.Substate, taskPool *substate.SubstateTaskPool) error {
	for account, accountInfo := range st.OutputAlloc {
		registerContract(account, accountInfo.Code)
	}
	for account, accountInfo := range st.InputAlloc {
		registerContract(account, accountInfo.Code)
	}
	return nil
}

// func getStorageUpdateSizeAction for replay-storage command
func getCodeAction(ctx *cli.Context) error {
	var err error

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli storage command requires exactly 2 arguments")
	}

	chainID = ctx.Int(ChainIDFlag.Name)
	fmt.Printf("chain-id: %v\n", chainID)
	fmt.Printf("git-date: %v\n", gitDate)
	fmt.Printf("git-commit: %v\n", gitCommit)
	fmt.Printf("contract-db: %v\n", ContractDB)

	first, ferr := strconv.ParseInt(ctx.Args().Get(0), 10, 64)
	last, lerr := strconv.ParseInt(ctx.Args().Get(1), 10, 64)
	if ferr != nil || lerr != nil {
		return fmt.Errorf("substate-cli code: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("substate-cli code: error: block number must be greater than 0")
	}
	if first > last {
		return fmt.Errorf("substate-cli code: error: first block has larger number than last block")
	}

	substate.SetSubstateFlags(ctx)
	substate.OpenSubstateDBReadOnly()
	defer substate.CloseSubstateDB()

	CodeRegistry = make(map[common.Address][]byte)

	taskPool := substate.NewSubstateTaskPool("substate-cli code", getCodeTask, uint64(first), uint64(last), ctx)
	err = taskPool.Execute()

	writeContracts()
	return err
}
