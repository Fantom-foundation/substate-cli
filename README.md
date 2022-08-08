# Fantom Substate Recorder/Replayer

Fantom's off-the-chain testing framework based on [Transaction Substate Recorder/Replayer](https://github.com/verovm/record-replay).

## Building the Recorder/Replayer
To build all substate-cli, run ``` make all ```.   You can find ```substate-cli``` in the build directory.

It should be noted that this version of ```opera``` client requires  [a modified version of go-ethereum](https://github.com/fantom-foundation/go-ethereum/tree/rr1.10.8-ftm-rc4-vm) which contains ```substate``` package. 

## Running the replayer
To replay substrate in a given block range,
```shell
substate-cli replay 0 41000000
```
The command replays transactions from block 0 to block 41000000. If the execution result in the replayer does not match the recorder result, it is immediately terminated with an error message.
By default, ```substate-cli``` executes using 4 worker threads and reads the substate database from ./substate.fantom directory. These parameters can be adjusted. The number of worker threads can be set via ```--workers``` option.
 ```shell
substate-cli replay --workers 32 0 41000000
```

If substate database is located in another location, the directory path can be explicitly specified via the``` --substatedir``` option.
```shell
substate-cli replay --substatedir /path/to/substate_directory 0 41000000
```
Substates of the first 41 million blocks can be downloaded from [here](https://drive.google.com/file/d/1oAJAdD3Sv9ALuPnEGoMfJTRBv0hYFBc8/view?usp=sharing)

More details of ```substate-cli``` can be found in [substate-cli/README.md](cmd/substate-cli/README.md)

# Substate-cli
In addition to functionality testing, the recorder/replayer is useful to obtain execution metrics for smart contracts such as EVM call runtime or opcode/instruction statistics. ```substate-cli``` is a command-line interface for the replayer and applications to obtain mentioned metrics. 

```
NAME:
   substate-cli - Fantom substate command line interface

USAGE:
   substate-cli [global options] command [command options] [arguments...]

VERSION:
   1.10.8-stable

COMMANDS:
     replay        executes full state transitions and checks output consistency
     storage-size  returns changes in storage size by transactions in the specified block range
     code-size     reports code size and nonce of smart contracts in the specified block range
     dump          returns content in substates in json format
     db            A set of commands on substate DB
     help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### Substate Replayer
To execute substrate in a given block range,
```shell
substate-cli replay 0 41000000
```

 
### EVM Call Runtime
To measure EVM call runtime of transactions in a given block range,
```shell
substate-cli replay --profiling-call  0 41000000
```

Output format
```
call: <Block>, <Nonce>, <Account>, <runtime in ns>
```

### EVM Micro Profiling
To get micro-profiling statistics,
```shell
substate-cli replay --profiling-opcode 0 41000000
```
The returned statistics include

 - **opcode frequency:** the total number of times an opcode is used.
 - **opcode runtime:** the total runtime of an opcode in nanoseconds.
 - **instruction frequency:** a distribution showing number of times an instruction is executed.
 - **step length frequency:** a distribution showing number of steps per smart contract invocation.

Output format
```
opcode-freq: <opcode>, <execution count>
opcode-runtime: <opcode>, <total runtime in ns>
instruction-freq: <execution frequency>, <instruction count>
steplen-freq: <execution steps>, <number of contracts>
```

### Blockchain Storage
To profile storage update size after each transaction in a given block range,
```shell
substate-cli storage-size 0 41000000
```

Output format
```
metric: <Block>, <Transaction>, <Unix timestamp>, <Account>, <Storage size change> ,<Size in input substate>, <Size in output substate>
```

### Smart Contract Code Size 
To profile smart contract code size and nonce in a given block range,
```shell
substate-cli code-size 0 41000000
```

Output format
```
metric: <Block>, <Transaction>, <Unix timestamp>, <Account>, <Code size> ,<Nonce>, <Transaction type>
```
