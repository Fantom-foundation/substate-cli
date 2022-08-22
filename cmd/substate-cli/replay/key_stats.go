package replay

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/substate"
	"gopkg.in/urfave/cli.v1"
)

// record-replay: substate-cli key-stats command
var GetKeyStatsCommand = cli.Command{
	Action:    getKeyStatsAction,
	Name:      "key-stats",
	Usage:     "computes usage statistics of accessed storage locations",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		substate.WorkersFlag,
		substate.SubstateDirFlag,
		ChainIDFlag,
	},
	Description: `
The substate-cli key-stats command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to be analysed.

Statistics on the usage of accessed storage locations are printed to the console.
`,
}

// getKeyStatsAction collects statistical information on the usage
// of keys (=addresses of storage locations) in transactions.
func getKeyStatsAction(ctx *cli.Context) error {
	return getReferenceStatsAction(ctx, "key-stats", func(info *TransactionInfo) []common.Hash {
		keys := []common.Hash{}
		for _, account := range info.st.InputAlloc {
			for key := range account.Storage {
				keys = append(keys, key)
			}
		}
		for _, account := range info.st.OutputAlloc {
			for key := range account.Storage {
				keys = append(keys, key)
			}
		}
		return keys
	})
}
