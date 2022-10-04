package replay

import (
	"fmt"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/substate-cli/tracer"
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
	contractDict := tracer.NewContractDictionary()
	storageDict := tracer.NewStorageDictionary()
	opIndex := tracer.NewOperationIndex()
	fposIndex := tracer.NewFilePositionIndex()

	// load dictionaries and indexes from file
	contractDict.Read("contract-dictionary.dat")
	storageDict.Read("storage-dictionary.dat")
	opIndex.Read("operation-index.dat")
	fposIndex.Read("filepos-index.dat")

	// create index and execution context
	// eCtx := &ExecutionContext{ContractDictionary: contractDict, StorageDictionary: storageDict}
	iCtx := &tracer.IndexContext{OperationIndex: opIndex, FilePositionIndex: fposIndex}

	// Create dummy statedb to make it compile
	// TODO: plug-in real DBs and prime DB at block "first"
	// var db *StateDB = nil

	// replay storage trace
	iter := tracer.NewStorageTraceIterator(iCtx, first, last)
	defer iter.Release()
	for iter.Next() {
		// op := iter.Value()
		// (*op).Execute(db, eCtx)
	}
}

// record-replay: func traceReplayAction for replaying
func traceReplayAction(ctx *cli.Context) error {
	var err error

	tracer.TraceDir = ctx.String(TraceDirectoryFlag.Name) + "/"

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli replay-trace command requires exactly 2 arguments")
	}

	_ , _, argErr := SetBlockRange(ctx.Args().Get(0), ctx.Args().Get(1))
	if argErr != nil {
		return argErr
	}

	return err
}
