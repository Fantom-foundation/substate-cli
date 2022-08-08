package replay

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/substate"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	gitCommit = "" // Git SHA1 commit hash of the release (set via linker flags)
	gitDate   = ""
)
// chain id
var chainID int
var ChainIDFlag = cli.IntFlag{
		Name:  "chainid",
		Usage: "ChainID for replayer",
		Value: 250,
	}

var ProfileEVMCallFlag = cli.BoolFlag{
		Name:  "profiling-call",
		Usage: "enable profiling for EVM call",
	}

var ProfileEVMOpCodeFlag = cli.BoolFlag{
		Name:  "profiling-opcode",
		Usage: "enable profiling for EVM opcodes",
	}

// record-replay: substate-cli replay command
var ReplayCommand = cli.Command{
	Action:    replayAction,
	Name:      "replay",
	Usage:     "executes full state transitions and check output consistency",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		substate.WorkersFlag,
		substate.SkipTransferTxsFlag,
		substate.SkipCallTxsFlag,
		substate.SkipCreateTxsFlag,
		substate.SubstateDirFlag,
		ChainIDFlag,
		ProfileEVMCallFlag,
		ProfileEVMOpCodeFlag,
	},
	Description: `
The substate-cli replay command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay transactions.`,
}

// replayTask replays a transaction substate
func replayTask(block uint64, tx int, recording *substate.Substate, taskPool *substate.SubstateTaskPool) error {

	inputAlloc := recording.InputAlloc
	inputEnv := recording.Env
	inputMessage := recording.Message

	outputAlloc := recording.OutputAlloc
	outputResult := recording.Result

	var (
		vmConfig    vm.Config
		chainConfig *params.ChainConfig
		getTracerFn func(txIndex int, txHash common.Hash) (tracer vm.Tracer, err error)
	)

	vmConfig = opera.DefaultVMConfig
	vmConfig.NoBaseFee = true

	chainConfig = params.AllEthashProtocolChanges
	chainConfig.ChainID = big.NewInt(int64(chainID))
	chainConfig.LondonBlock = new(big.Int).SetUint64(37534833)
	chainConfig.BerlinBlock = new(big.Int).SetUint64(37455223)

	getTracerFn = func(txIndex int, txHash common.Hash) (tracer vm.Tracer, err error) {
		return nil, nil
	}

	var hashError error
	getHash := func(num uint64) common.Hash {
		if inputEnv.BlockHashes == nil {
			hashError = fmt.Errorf("getHash(%d) invoked, no blockhashes provided", num)
			return common.Hash{}
		}
		h, ok := inputEnv.BlockHashes[num]
		if !ok {
			hashError = fmt.Errorf("getHash(%d) invoked, blockhash for that block not provided", num)
		}
		return h
	}

	// Apply Message
	var (
		statedb   = MakeOffTheChainStateDB(inputAlloc)
		gaspool   = new(evmcore.GasPool)
		blockHash = common.Hash{0x01}
		txHash    = common.Hash{0x02}
		txIndex   = tx
	)

	gaspool.AddGas(inputEnv.GasLimit)
	blockCtx := vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    inputEnv.Coinbase,
		BlockNumber: new(big.Int).SetUint64(inputEnv.Number),
		Time:        new(big.Int).SetUint64(inputEnv.Timestamp),
		Difficulty:  inputEnv.Difficulty,
		GasLimit:    inputEnv.GasLimit,
		GetHash:     getHash,
	}
	// If currentBaseFee is defined, add it to the vmContext.
	if inputEnv.BaseFee != nil {
		blockCtx.BaseFee = new(big.Int).Set(inputEnv.BaseFee)
	}

	msg := inputMessage.AsMessage()

	tracer, err := getTracerFn(txIndex, txHash)
	if err != nil {
		return err
	}
	vmConfig.Tracer = tracer
	vmConfig.Debug = (tracer != nil)
	statedb.Prepare(txHash, txIndex)

	txCtx := evmcore.NewEVMTxContext(msg)

	evm := vm.NewEVM(blockCtx, txCtx, statedb, chainConfig, vmConfig)

	snapshot := statedb.Snapshot()
	msgResult, err := evmcore.ApplyMessage(evm, msg, gaspool)

	if err != nil {
		statedb.RevertToSnapshot(snapshot)
		return err
	}

	if hashError != nil {
		return hashError
	}

	if chainConfig.IsByzantium(blockCtx.BlockNumber) {
		statedb.Finalise(true)
	} else {
		statedb.IntermediateRoot(chainConfig.IsEIP158(blockCtx.BlockNumber))
	}

	evmResult := &substate.SubstateResult{}
	if msgResult.Failed() {
		evmResult.Status = types.ReceiptStatusFailed
	} else {
		evmResult.Status = types.ReceiptStatusSuccessful
	}
	evmResult.Logs = statedb.GetLogs(txHash, blockHash)
	evmResult.Bloom = types.BytesToBloom(types.LogsBloom(evmResult.Logs))
	if to := msg.To(); to == nil {
		evmResult.ContractAddress = crypto.CreateAddress(evm.TxContext.Origin, msg.Nonce())
	}
	evmResult.GasUsed = msgResult.UsedGas

	evmAlloc := statedb.SubstatePostAlloc

	r := outputResult.Equal(evmResult)
	a := outputAlloc.Equal(evmAlloc)
	if !(r && a) {
		fmt.Printf("block: %v Transaction: %v\n", block, tx)
		if !r {
			fmt.Printf("inconsistent output: result\n")
		}
		if !a {
			fmt.Printf("inconsistent output: alloc\n")
		}
		var jbytes []byte
		jbytes, _ = json.MarshalIndent(inputAlloc, "", " ")
		fmt.Printf("Recorded input substate:\n%s\n", jbytes)
		jbytes, _ = json.MarshalIndent(inputEnv, "", " ")
		fmt.Printf("Recorded input environmnet:\n%s\n", jbytes)
		jbytes, _ = json.MarshalIndent(inputMessage, "", " ")
		fmt.Printf("Recorded input message:\n%s\n", jbytes)
		jbytes, _ = json.MarshalIndent(outputAlloc, "", " ")
		fmt.Printf("Recorded output substate:\n%s\n", jbytes)
		jbytes, _ = json.MarshalIndent(evmAlloc, "", " ")
		fmt.Printf("Replayed output substate:\n%s\n", jbytes)
		jbytes, _ = json.MarshalIndent(outputResult, "", " ")
		fmt.Printf("Recorded output result:\n%s\n", jbytes)
		jbytes, _ = json.MarshalIndent(evmResult, "", " ")
		fmt.Printf("Replayed output result:\n%s\n", jbytes)
		return fmt.Errorf("inconsistent output")
	}

	return nil
}

// record-replay: func replayAction for replay command
func replayAction(ctx *cli.Context) error {
	var err error

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli replay command requires exactly 2 arguments")
	}

	chainID = ctx.Int(ChainIDFlag.Name)
	fmt.Printf("chain-id: %v\n",chainID)
	fmt.Printf("git-date: %v\n", gitDate)
	fmt.Printf("git-commit: %v\n",gitCommit)

	first, ferr := strconv.ParseInt(ctx.Args().Get(0), 10, 64)
	last, lerr := strconv.ParseInt(ctx.Args().Get(1), 10, 64)
	if ferr != nil || lerr != nil {
		return fmt.Errorf("substate-cli replay: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("substate-cli replay: error: block number must be greater than 0")
	}
	if first > last {
		return fmt.Errorf("substate-cli replay: error: first block has larger number than last block")
	}

	if ctx.Bool(ProfileEVMCallFlag.Name) {
		vm.ProfileEVMCall = true;
	}
	if ctx.Bool(ProfileEVMOpCodeFlag.Name) {
		vm.ProfileEVMOpCode = true;
	}

	substate.SetSubstateFlags(ctx)
	substate.OpenSubstateDBReadOnly()
	defer substate.CloseSubstateDB()

	taskPool := substate.NewSubstateTaskPool("substate-cli replay", replayTask, uint64(first), uint64(last), ctx)
	err = taskPool.Execute()

	if ctx.Bool(ProfileEVMOpCodeFlag.Name) {
		vm.PrintStatistics()
	}
	return err
}
