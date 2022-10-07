package replay

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/substate-cli/state"
	"github.com/Fantom-foundation/substate-cli/tracer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/substate"
	cli "gopkg.in/urfave/cli.v1"
)

const FPosBlockMultiple = 4

var TraceDirectoryFlag = cli.StringFlag{
	Name:  "trace-dir",
	Usage: "set storage trace's output directory",
	Value: "./",
}

// record-trace: substate-cli trace command
var TraceCommand = cli.Command{
	Action:    traceAction,
	Name:      "trace",
	Usage:     "executes full state transitions and check output consistency",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		substate.WorkersFlag,
		substate.SkipTransferTxsFlag,
		substate.SkipCallTxsFlag,
		substate.SkipCreateTxsFlag,
		substate.SubstateDirFlag,
		ChainIDFlag,
		TraceDirectoryFlag,
	},
	Description: `
The substate-cli trace command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to trace transactions.`,
}

type TraceConfig struct {
	vm_impl          string
	only_successful  bool
	use_in_memory_db bool
}

// traceTask generates storage traces for each transaction
func traceTask(config TraceConfig, block uint64, tx int, recording *substate.Substate, dCtx *tracer.DictionaryContext, ch chan tracer.StateOperation) error {

	if config.only_successful && recording.Result.Status != types.ReceiptStatusSuccessful {
		return nil
	}

	inputAlloc := recording.InputAlloc
	inputEnv := recording.Env
	inputMessage := recording.Message

	outputAlloc := recording.OutputAlloc
	outputResult := recording.Result

	var (
		vmConfig    vm.Config
		chainConfig *params.ChainConfig
	)

	vmConfig = opera.DefaultVMConfig
	vmConfig.NoBaseFee = true

	chainConfig = params.AllEthashProtocolChanges
	chainConfig.ChainID = big.NewInt(int64(chainID))
	chainConfig.LondonBlock = new(big.Int).SetUint64(37534833)
	chainConfig.BerlinBlock = new(big.Int).SetUint64(37455223)

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

	// TODO: drop slower OffTheChain DB??
	var statedb state.StateDB
	if config.use_in_memory_db {
		statedb = state.MakeInMemoryStateDB(&inputAlloc)
	} else {
		statedb = state.MakeOffTheChainStateDB(inputAlloc)
	}
	statedb = tracer.NewStateProxyDB(statedb, dCtx, ch)

	// Apply Message
	var (
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

	vmConfig.Tracer = nil
	vmConfig.Debug = false
	vmConfig.InterpreterImpl = config.vm_impl
	statedb.Prepare(txHash, txIndex)

	txCtx := evmcore.NewEVMTxContext(msg)

	evm := vm.NewEVM(blockCtx, txCtx, statedb, chainConfig, vmConfig)

	snapshot := statedb.Snapshot()
	start := time.Now()
	msgResult, err := evmcore.ApplyMessage(evm, msg, gaspool)
	addVmDuration(time.Since(start))

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
	evmAlloc := statedb.GetSubstatePostAlloc()

	r := outputResult.Equal(evmResult)
	a := outputAlloc.Equal(evmAlloc)
	if !(r && a) {
		fmt.Printf("Block: %v Transaction: %v\n", block, tx)
		if !r {
			fmt.Printf("inconsistent output: result\n")
			printResultDiffSummary(outputResult, evmResult)
		}
		if !a {
			fmt.Printf("inconsistent output: alloc\n")
			printAllocationDiffSummary(&outputAlloc, &evmAlloc)
		}
		return fmt.Errorf("inconsistent output")
	}

	return nil
}

// Read state operations from channel and write them into their files.
func StateOperationWriter(ctx context.Context, done chan struct{}, ch chan tracer.StateOperation, dCtx *tracer.DictionaryContext, iCtx *tracer.IndexContext) {
	defer close(done)

	// open trace file
	file, err := os.OpenFile("trace.dat", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Cannot open trace file. Error: %v", err)
	}

	// read from channel until receiving cancel signal
	for {
		select {
		case op := <-ch:
			// TODO: have flag for debug information

			// print debug information
			tracer.Debug(dCtx, op)

			// is it a pseudo operation?
			if op.GetOpId() >= tracer.NumWriteOperations {
				switch op.GetOpId() {
				case tracer.BeginBlockOperationID:
					tOp, ok := op.(*tracer.BeginBlockOperation)
					if !ok {
						log.Fatalf("Begin block operation downcasting failed")
					}
					offset, err := file.Seek(0, 1)
					if err != nil {
						log.Fatalf("Cannot retrieve current file position. Error: %v", err)
					}
					iCtx.BlockIndex.Add(tOp.BlockNumber, offset)
				case tracer.EndBlockOperationID:
					_, ok := op.(*tracer.EndBlockOperation)
					if !ok {
						log.Fatalf("Block operation downcasting failed")
					}
				default:
					log.Fatalf("Unhandled pseudo operation")
				}
				continue
			}
			tracer.Write(file, op)

		case <-ctx.Done():
			if len(ch) == 0 {
				return
			}
		}
	}

	// close trace file
	err = file.Close()
	if err != nil {
		log.Fatalf("Cannot close trace file. Error: %v", err)
	}
}

// record-trace: func traceAction for trace command
func traceAction(ctx *cli.Context) error {
	var err error

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli trace command requires exactly 2 arguments")
	}

	dCtx := tracer.NewDictionaryContext()
	iCtx := tracer.NewIndexContext()
	opChannel := make(chan tracer.StateOperation, 10000)

	cctx, cancel := context.WithCancel(context.Background())
	cancelChannel := make(chan struct{})
	go StateOperationWriter(cctx, cancelChannel, opChannel, dCtx, iCtx)
	defer func() {
		// cancel writers
		(cancel)()        // stop writer
		<-(cancelChannel) // wait for writer to finish
	}()

	chainID = ctx.Int(ChainIDFlag.Name)
	tracer.TraceDir = ctx.String(TraceDirectoryFlag.Name) + "/"

	first, last, argErr := SetBlockRange(ctx.Args().Get(0), ctx.Args().Get(1))
	if argErr != nil {
		return argErr
	}

	substate.SetSubstateFlags(ctx)
	substate.OpenSubstateDBReadOnly()
	defer substate.CloseSubstateDB()

	var config = TraceConfig{
		vm_impl:          ctx.String(InterpreterImplFlag.Name),
		only_successful:  ctx.Bool(OnlySuccessfulFlag.Name),
		use_in_memory_db: ctx.Bool(UseInMemoryStateDbFlag.Name),
	}

	iter := substate.NewSubstateIterator(first, ctx.Int(substate.WorkersFlag.Name))
	defer iter.Release()
	oldBlock := uint64(math.MaxUint64) // set to an infeasible block
	for iter.Next() {
		tx := iter.Value()
		// close off old block with an end-block operation
		if oldBlock != tx.Block {
			if oldBlock != math.MaxUint64 {
				opChannel <- tracer.NewEndBlockOperation(oldBlock)
			}
			if tx.Block > last {
				break
			}
			oldBlock = tx.Block
			// open new block with a begin-block operation
			opChannel <- tracer.NewBeginBlockOperation(tx.Block)
		}
		traceTask(config, tx.Block, tx.Transaction, tx.Substate, dCtx, opChannel)
		opChannel <- tracer.NewEndTransactionOperation()
	}

	// write dictionaries and indexes
	dCtx.Write()
	iCtx.Write()

	return err
}
