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

// record-replay: func traceReplayAction for replay command
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
