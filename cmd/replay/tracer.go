package replay

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/Fantom-foundation/go-opera/evmcore"
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	//"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/substate"
	cli "gopkg.in/urfave/cli.v1"
)

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

// traceTask traces a transaction substate
func traceTask(config TraceConfig, block uint64, tx int, recording *substate.Substate, contractDict *ContractDictionary, storageDict *StorageDictionary, ch chan StateOperation) error {

	// If requested, skip failed transactions.
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

	var statedb StateDB
	if config.use_in_memory_db {
		statedb = MakeInMemoryStateDB(&inputAlloc)
	} else {
		statedb = MakeOffTheChainStateDB(inputAlloc)
	}
	statedb = NewStateProxyDB(statedb, contractDict, storageDict, ch)

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
		fmt.Printf("block: %v Transaction: %v\n", block, tx)
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

func StateOperationWriter(ctx context.Context, done chan struct{}, ch chan StateOperation) {
	files := []*os.File{nil, nil}
	fn := []string{"GetState.bin", "SetState.bin"}
	for i := 0; i < NumStateOperations; i++ {
		f, err := os.OpenFile(fn[i], os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalf("Cannot open state file %v", i)
		}
		files[i] = f
	}
	defer close(done)
	var opNum uint64 = 0
	for {
		select {
		case op := <-ch:
			op.Write(opNum, files)
			opNum++
		case <-ctx.Done():
			if len(ch) == 0 {
				return
			}
		}
	}
	for i := 0; i < NumStateOperations; i++ {
		err := files[i].Close()
		if err != nil {
			log.Fatalf("Cannot close state file %v", i)
		}
	}
}

// record-trace: func traceAction for trace command
func traceAction(ctx *cli.Context) error {
	var err error

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli trace command requires exactly 2 arguments")
	}

	contractDict := NewContractDictionary()
	storageDict := NewStorageDictionary()
	opChannel := make(chan StateOperation, 10000)

	cctx, cancel := context.WithCancel(context.Background())
	cancelChannel := make(chan struct{})
	go StateOperationWriter(cctx, cancelChannel, opChannel)
	defer func() {
		// cancel writers
		(cancel)()        // stop writer
		<-(cancelChannel) // wait for writer to finish
	}()

	chainID = ctx.Int(ChainIDFlag.Name)

	first, ferr := strconv.ParseUint(ctx.Args().Get(0), 10, 64)
	last, lerr := strconv.ParseUint(ctx.Args().Get(1), 10, 64)
	if ferr != nil || lerr != nil {
		return fmt.Errorf("substate-cli trace: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("substate-cli trace: error: block number must be greater than 0")
	}
	if first > last {
		return fmt.Errorf("substate-cli trace: error: first block has larger number than last block")
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
	for iter.Next() {
		tx := iter.Value()
		traceTask(config, tx.Block, tx.Transaction, tx.Substate, contractDict, storageDict, opChannel)
		if tx.Block >= last {
			break
		}
	}

	return err
}
