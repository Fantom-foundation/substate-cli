package replay

import (
	"fmt"
	cli "gopkg.in/urfave/cli.v1"
	"strconv"
)

// record-replay: substate-cli replay command
var TraceReplayCommand = cli.Command{
	Action:    traceReplayAction,
	Name:      "trace-replay",
	Usage:     "executes storage trace",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags:     []cli.Flag{},
	Description: `
The substate-cli trace-replay command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay storage traces.`,
}

func storageDriver(first uint64, last uint64) {
	// create new dictionaries and indices
	contractDict := NewContractDictionary()
	storageDict := NewStorageDictionary()
	opIndex := NewOperationIndex()
	fposIndex := NewFilePositionIndex()

	// load dictionaries and indexes from file
	contractDict.Read("contract-dictionary.dat")
	storageDict.Read("storage-dictionary.dat")
	opIndex.Read("operation-index.dat")
	fposIndex.Read("filepos-index.dat")

	// create index and execution context
	// eCtx := &ExecutionContext{ContractDictionary: contractDict, StorageDictionary: storageDict}
	iCtx := &IndexContext{OperationIndex: opIndex, FilePositionIndex: fposIndex}

	// Create dummy statedb to make it compile
	// TODO: plug-in real DBs and prime DB at block "first"
	// var db *StateDB = nil

	// replay storage trace
	iter := NewStorageTraceIterator(iCtx, first, last)
	defer iter.Release()
	for iter.Next() {
		// op := iter.Value()
		// (*op).Execute(db, eCtx)
	}
}

// record-replay: func traceReplayAction for replaying
func traceReplayAction(ctx *cli.Context) error {
	var err error

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli replay-trace command requires exactly 2 arguments")
	}

	first, ferr := strconv.ParseInt(ctx.Args().Get(0), 10, 64)
	last, lerr := strconv.ParseInt(ctx.Args().Get(1), 10, 64)
	if ferr != nil || lerr != nil {
		return fmt.Errorf("substate-cli replay: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("substate-cli replay-trace: error: block number must be greater than 0")
	}
	if first > last {
		return fmt.Errorf("substate-cli replay-trace: error: first block has larger number than last block")
	}

	return err
}
