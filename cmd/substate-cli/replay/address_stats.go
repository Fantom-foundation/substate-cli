package replay

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/substate"
	"gopkg.in/urfave/cli.v1"
)

// record-replay: substate-cli address stats command
var GetAddressStatsCommand = cli.Command{
	Action:    getAddressStatsAction,
	Name:      "address-stats",
	Usage:     "computes usage statistics of addresss",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		substate.WorkersFlag,
		substate.SubstateDirFlag,
		ChainIDFlag,
	},
	Description: `
The substate-cli address-stats command requires two arguments:
<blockNumFirst> <blockNumLast>

<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to be analysed.

Statistics on the usage of addresses are printed to the console.
`,
}

type addressStatistics struct {
	accesses map[common.Address]int
}

func newStatistics() addressStatistics {
	return addressStatistics{accesses: map[common.Address]int{}}
}

func (a *addressStatistics) RegisterAccess(address *common.Address) {
	a.accesses[*address]++
}

func (a *addressStatistics) PrintSummary() {
	var count = len(a.accesses)
	var sum int64 = 0
	list := make([]int, 0, len(a.accesses))
	for _, count := range a.accesses {
		sum += int64(count)
		list = append(list, count)
	}
	sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })

	var prefix_sum = 0
	for i := range list {
		list[i] = prefix_sum + list[i]
		prefix_sum = list[i]
	}

	fmt.Printf("Reference frequency distribution:\n")
	for i := 0; i < 100; i++ {
		fmt.Printf("%d,%d\n", i, list[i*len(list)/100])
	}
	fmt.Printf("100,%d\n", list[len(list)-1])
	fmt.Printf("Number of addresses:        %15d\n", count)
	fmt.Printf("Number of references:       %15d\n", sum)
	fmt.Printf("Average references/address: %15.2f\n", float32(sum)/float32(count))

}

func runAddressStatCollector(stats *addressStatistics, src <-chan common.Address, done chan<- int) {
	for address := range src {
		stats.RegisterAccess(&address)
	}
	close(done)
}

// collectAddressStats collects statistical information on address usage.
func collectAccressStats(dest chan<- common.Address, block uint64, tx int, st *substate.Substate, taskPool *substate.SubstateTaskPool) error {
	// Collect all addresses accessed by this transaction in a set.
	accessed_addresses := map[common.Address]int{}
	for address := range st.OutputAlloc {
		accessed_addresses[address] = 0
	}
	for address := range st.InputAlloc {
		accessed_addresses[address] = 0
	}
	// Report accessed addresses to statistics collector.
	for address := range accessed_addresses {
		dest <- address
	}
	return nil
}

// getAddressStatsAction collects statistical information on the usage
// of addresss in transactions.
func getAddressStatsAction(ctx *cli.Context) error {
	var err error

	if len(ctx.Args()) != 2 {
		return fmt.Errorf("substate-cli address-stats command requires exactly 2 arguments")
	}

	chainID = ctx.Int(ChainIDFlag.Name)
	fmt.Printf("chain-id: %v\n", chainID)
	fmt.Printf("git-date: %v\n", gitDate)
	fmt.Printf("git-commit: %v\n", gitCommit)
	fmt.Printf("contract-db: %v\n", ContractDB)

	first, ferr := strconv.ParseInt(ctx.Args().Get(0), 10, 64)
	last, lerr := strconv.ParseInt(ctx.Args().Get(1), 10, 64)
	if ferr != nil || lerr != nil {
		return fmt.Errorf("substate-cli code: error in parsing parameters: block number not an integer")
	}
	if first < 0 || last < 0 {
		return fmt.Errorf("substate-cli code: error: block number must be greater than 0")
	}
	if first > last {
		return fmt.Errorf("substate-cli code: error: first block has larger number than last block")
	}

	substate.SetSubstateFlags(ctx)
	substate.OpenSubstateDBReadOnly()
	defer substate.CloseSubstateDB()

	// Start Collector.
	stats := newStatistics()
	done := make(chan int)
	addr := make(chan common.Address, 100)
	go runAddressStatCollector(&stats, addr, done)

	// Create per-transaction task.
	task := func(block uint64, tx int, st *substate.Substate, taskPool *substate.SubstateTaskPool) error {
		return collectAccressStats(addr, block, tx, st, taskPool)
	}

	// Process all transactions in parallel, out-of-order.
	taskPool := substate.NewSubstateTaskPool("substate-cli code", task, uint64(first), uint64(last), ctx)
	err = taskPool.Execute()
	if err != nil {
		return err
	}

	// Signal the end of the processed addresses.
	close(addr)

	// Wait for the collector to finish.
	for {
		if _, open := <-done; !open {
			break
		}
	}

	// Print the statistics.
	fmt.Printf("\n\n----- Summary: -------\n")
	stats.PrintSummary()
	fmt.Printf("----------------------\n")
	return nil
}
