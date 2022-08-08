package replay

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ethereum/go-ethereum/substate"
	cli "gopkg.in/urfave/cli.v1"
)

// substate-cli dump command
var SubstateDumpCommand = cli.Command{
	Action:    substateDumpAction,
	Name:      "dump",
	Usage:     "returns content in substates in json format",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		substate.SubstateDirFlag,
	},
	Description: `
The substate-cli dump command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay transactions.`,
}

// replayTask replays a transaction substate
func substateDumpTask(block uint64, tx int, recording *substate.Substate, taskPool *substate.SubstateTaskPool) error {

	inputAlloc := recording.InputAlloc
	inputEnv := recording.Env
	inputMessage := recording.Message

	outputAlloc := recording.OutputAlloc
	outputResult := recording.Result

	fmt.Printf("block: %v Transaction: %v\n", block, tx)
	var jbytes []byte
	jbytes, _ = json.MarshalIndent(inputAlloc, "", " ")
	fmt.Printf("Recorded input substate:\n%s\n", jbytes)
	jbytes, _ = json.MarshalIndent(inputEnv, "", " ")
	fmt.Printf("Recorded input environmnet:\n%s\n", jbytes)
	jbytes, _ = json.MarshalIndent(inputMessage, "", " ")
	fmt.Printf("Recorded input message:\n%s\n", jbytes)
	jbytes, _ = json.MarshalIndent(outputAlloc, "", " ")
	fmt.Printf("Recorded output substate:\n%s\n", jbytes)
	jbytes, _ = json.MarshalIndent(outputResult, "", " ")
	fmt.Printf("Recorded output result:\n%s\n", jbytes)

	return nil
}

// record-replay: func replayAction for replay command
func substateDumpAction(ctx *cli.Context) error {
	var err error

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli dump cammand requires exactly 2 arguments")
	}

	first, ferr := strconv.ParseInt(ctx.Args().Get(0), 10, 64)
	last, lerr := strconv.ParseInt(ctx.Args().Get(1), 10, 64)
	if ferr != nil || lerr != nil {
		return fmt.Errorf("substate-cli dump: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("substate-cli dump: error: block number must be greater than 0")
	}
	if first > last {
		return fmt.Errorf("substate-cli dump: error: first block has larger number than last block")
	}

	substate.SetSubstateFlags(ctx)
	substate.OpenSubstateDBReadOnly()
	defer substate.CloseSubstateDB()

	taskPool := substate.NewSubstateTaskPool("substate-cli dump", substateDumpTask, uint64(first), uint64(last), ctx)
	taskPool.Workers = 1 //force sequential order
	err = taskPool.Execute()
	return err
}
