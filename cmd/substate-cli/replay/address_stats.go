package replay

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/substate"
	"gopkg.in/urfave/cli.v1"
)

const cli_command = "address-stats"

// record-replay: substate-cli address stats command
var GetAddressStatsCommand = cli.Command{
	Action:    getAddressStatsAction,
	Name:      cli_command,
	Usage:     "computes usage statistics of addresss",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		substate.WorkersFlag,
		substate.SubstateDirFlag,
		ChainIDFlag,
	},
	Description: fmt.Sprintf(`
The substate-cli %v command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to be analysed.

Statistics on the usage of addresses are printed to the console.
`, cli_command),
}

// getAddressStatsAction collects statistical information on the usage
// of addresses in transactions.
func getAddressStatsAction(ctx *cli.Context) error {
	return getReferenceStatsAction(ctx, cli_command, func(info *TransactionInfo) []common.Address {
		addresses := []common.Address{}
		for address := range info.st.InputAlloc {
			addresses = append(addresses, address)
		}
		for address := range info.st.OutputAlloc {
			addresses = append(addresses, address)
		}
		return addresses
	})
}
