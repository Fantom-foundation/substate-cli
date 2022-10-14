package replay

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"strconv"
)

// chain id
var chainID int
var (
	gitCommit = "" // Git SHA1 commit hash of the release (set via linker flags)
	gitDate   = ""
)

// command line options
var (
	ChainIDFlag = cli.IntFlag{
		Name:  "chainid",
		Usage: "ChainID for replayer",
		Value: 250,
	}
	ProfileEVMCallFlag = cli.BoolFlag{
		Name:  "profiling-call",
		Usage: "enable profiling for EVM call",
	}
	MicroProfilingFlag = cli.BoolFlag{
		Name:  "micro-profiling",
		Usage: "enable micro-profiling of EVM",
	}
	BasicBlockProfilingFlag = cli.BoolFlag{
		Name:  "basic-block-profiling",
		Usage: "enable profiling of basic block",
	}
	OnlySuccessfulFlag = cli.BoolFlag{
		Name:  "only-successful",
		Usage: "only runs transactions that have been successful",
	}
	InterpreterImplFlag = cli.StringFlag{
		Name:  "interpreter",
		Usage: "select the interpreter version to be used",
	}
	CpuProfilingFlag = cli.StringFlag{
		Name:  "cpuprofile",
		Usage: "the file name where to write a CPU profile of the evaluation step to",
	}
	UseInMemoryStateDbFlag = cli.BoolFlag{
		Name:  "faststatedb",
		Usage: "enables a faster, yet still experimental StateDB implementation",
	}
	DatabaseNameFlag = cli.StringFlag{
		Name:  "db",
		Usage: "set a database name for storing micro-profiling results",
		Value: "./profiling.db",
	}
	ChannelBufferSizeFlag = cli.IntFlag{
		Name:  "buffer-size",
		Usage: "set a buffer size for profiling channel",
		Value: 100000,
	}
	// contract-db filename
	ContractDBFlag = cli.StringFlag{
		Name:  "contractdb",
		Usage: "Contract database name for smart contracts",
		Value: "./contracts.db",
	}
)

func SetBlockRange(firstArg string, lastArg string) (uint64, uint64, error) {
	first, ferr := strconv.ParseUint(firstArg, 10, 64)
	last, lerr := strconv.ParseUint(lastArg, 10, 64)
	if ferr != nil || lerr != nil {
		return first, last, fmt.Errorf("substate-cli replay: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return first, last, fmt.Errorf("substate-cli replay: error: block number must be greater than 0")
	}
	if first > last {
		return first, last, fmt.Errorf("substate-cli replay: error: first block has larger number than last block")
	}
	return first, last, nil
}
