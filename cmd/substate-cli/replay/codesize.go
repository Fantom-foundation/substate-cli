package replay

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/substate"
	"github.com/urfave/cli/v2"
)

// substate-cli code-size command
var GetCodeSizeCommand = cli.Command{
	Action:    getCodeSizeAction,
	Name:      "code-size",
	Usage:     "reports code size and nonce of smart contracts in the specified block range",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&substate.WorkersFlag,
		&substate.SubstateDirFlag,
		&ChainIDFlag,
	},
	Description: `
The substate-cli code-size command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay transactions.

Output log format: (block, timestamp, transaction, account, code size, nonce, transaction type)`,
}

func GetTxType(to *common.Address, alloc substate.SubstateAlloc) string {
	if to == nil {
		return "create"
	}
	account, hasReceiver := alloc[*to]
	if to != nil && (!hasReceiver || len(account.Code) == 0) {
		return "transfer"
	}
	if to != nil && (hasReceiver && len(account.Code) > 0) {
		return "call"
	}
	return "unknown"
}

// getCodeSizeTask returns codesize and nonce of accounts in a substate
func getCodeSizeTask(block uint64, tx int, st *substate.Substate, taskPool *substate.SubstateTaskPool) error {
	to := st.Message.To
	timestamp := st.Env.Timestamp
	txType := GetTxType(to, st.InputAlloc)
	for account, accountInfo := range st.OutputAlloc {
		fmt.Printf("metric: %v,%v,%v,%v,%v,%v,%v\n",
			block,
			timestamp,
			tx,
			account.Hex(),
			len(accountInfo.Code),
			accountInfo.Nonce,
			txType)
	}
	for account, accountInfo := range st.InputAlloc {
		if _, found := st.OutputAlloc[account]; !found {
			fmt.Printf("metric: %v,%v,%v,%v,%v,%v,%v\n",
				block,
				timestamp,
				tx,
				account.Hex(),
				len(accountInfo.Code),
				accountInfo.Nonce,
				txType)
		}
	}
	return nil
}

// func getCodeSizeAction for GetCodeSizeCommand
func getCodeSizeAction(ctx *cli.Context) error {
	var err error

	if ctx.Args().Len() != 2 {
		return fmt.Errorf("substate-cli code-size command requires exactly 2 arguments")
	}

	chainID = ctx.Int(ChainIDFlag.Name)
	fmt.Printf("chain-id: %v\n", chainID)
	fmt.Printf("git-date: %v\n", gitDate)
	fmt.Printf("git-commit: %v\n", gitCommit)

	first, last, argErr := SetBlockRange(ctx.Args().Get(0), ctx.Args().Get(1))
	if argErr != nil {
		return argErr
	}

	substate.SetSubstateFlags(ctx)
	substate.OpenSubstateDBReadOnly()
	defer substate.CloseSubstateDB()

	taskPool := substate.NewSubstateTaskPool("substate-cli storage", getCodeSizeTask, first, last, ctx)
	err = taskPool.Execute()
	return err
}
