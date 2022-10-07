package replay

import (
	"fmt"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/Fantom-foundation/substate-cli/state"
	"github.com/Fantom-foundation/substate-cli/tracer"
	"github.com/ethereum/go-ethereum/substate"
)

// record-replay: substate-cli replay command
var TraceReplayCommand = cli.Command{
	Action:    traceReplayAction,
	Name:      "trace-replay",
	Usage:     "executes storage trace",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		substate.SubstateDirFlag,
	},
	Description: `
The substate-cli trace-replay command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay storage traces.`,
}

func storageDriver(first uint64, last uint64) {
	// load dictionaries
	dCtx := tracer.ReadDictionaryContext()

	// load indexes
	iCtx := tracer.ReadIndexContext()

	// Create dummy statedb to make it compile
	// TODO: plug-in real DBs and prime DB at block "first"

	// iterate substate (for in-membory state)
	stateIter := substate.NewSubstateIterator(first, 4)
	defer stateIter.Release()

	// replay storage trace
	traceIter := tracer.NewTraceIterator(iCtx, first, last)
	defer traceIter.Release()

	for stateIter.Next() {
		tx := stateIter.Value()
		if tx.Block > last {
			break
		}
		db := state.MakeOffTheChainStateDB(tx.Substate.InputAlloc)
		for traceIter.Next() {
			op := traceIter.Value()
			op.Execute(db, dCtx)
			tracer.Debug(dCtx, op)

			//find end of transaction
			if op.GetOpId() == tracer.EndTransactionOperationID {
				break
			}
		}

		db.Finalise(true)

		//Compare stateDB and OuputAlloc
		outputAlloc := db.GetSubstatePostAlloc()
		recordedAlloc := tx.Substate.OutputAlloc
		for account, xAlloc := range recordedAlloc {
			if yAlloc, exist := outputAlloc[account]; exist {
				for k, xv := range xAlloc.Storage {
					if yv, exist := yAlloc.Storage[k]; !exist || xv != yv {
						fmt.Printf("Error: mismatched value at storage key %v. want %v have %v\n", k, xv, yv)
					}

				}
			} else {
				fmt.Printf("Error: account %v doesn't exist\n", account)
			}

		}
	}
}

// record-replay: func traceReplayAction for replaying
func traceReplayAction(ctx *cli.Context) error {
	var err error

	tracer.TraceDir = ctx.String(TraceDirectoryFlag.Name)

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli replay-trace command requires exactly 2 arguments")
	}

	first, last, argErr := SetBlockRange(ctx.Args().Get(0), ctx.Args().Get(1))
	if argErr != nil {
		return argErr
	}
	substate.SetSubstateFlags(ctx)
	substate.OpenSubstateDBReadOnly()
	defer substate.CloseSubstateDB()

	storageDriver(first, last)

	return err
}
