# Profiling Scripts
Scripts for running metric collections using the Recoder/Repalyer framework. 

### Usage:
```run_profiling``` requires exactly 1 action and at least 1 metric
```shell
./run_profiling <action> <metrics>
```
For example, the following runs end-to-end profiling on two metrics: ```storage-update``` and ```evm-call```. This includes running the replayer, extracting raw data, storing metrics to a database and generating metric graphs.
```
./run_profiling all storage-update evm-call
```
### Actions
List of available actions
- ```record```  records substates and generate evm-call metrics data in a log file.
- ```replay```  replays substates and generate metrics data in a log file.
- ```extract``` extracts data from a log file and store in a database.
- ```plot```    plots graphs for specified metrics.
- ```all```     performs replay, extract and plot.
 
List of available metrics
- ```storage-update``` gets storage update size after each transaction
- ```code-size```      gets code size of smart contracts
- ```evm-call```       gets runtime of evm calls
- ```evm-opcode```     gets statistics of evm opcodes

## Requirements
- ```sqlite3``` database for storing metrics.
- ```gnuplot``` for plotting metric graphs.

## Configurations
Before start the profiling script, parameters such as preferred file locations, block range and event file should be configured in ```./run_profiling```.
### General configuration
```DATABASE_FILE```: Sqlite3 database location where metrics are stored.
```LOG_PATH```:  Output path for generated logs from substate-cli command.
```PNG_PATH```: Output path for metric graphs.

### The Replayer configuration
These parameters are required to execute ```replay``` command.
```WORKERS```: Number of worker threads that execute substates in parallel.
```FIRST_BLOCK```: The first block to replay transactions.
```LAST_BLOCK```: The last block to replay transaction (inclusive range).

### The Recorder configuration
These parameters are required to execute ```record``` command.
```EVENT_FILE```: Fantom event files exported from Fantom chain via ```opera export```.
```GENESIS```: Fantom genesis file (optional).
```OPERA_DATADIR```: Chaindata directory (optional).
