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
