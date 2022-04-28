# Fuzz

## Install

### Build from source

Preconditions

```bash
# version >= 1.18
go version
```

build xfuzz

```bash
make
```

## Usage

```bash
Usage of ./bin/xfuzz:
  -cmd string
        command
  -cmd-template
        {{.File}}
  -corpus string
        corpus dir path
  -deadline int
        maximum amount of time a command can run (second) (default 1)
  -signal string
        regex exception predicate (default ".*panic.*")
  -test.bench regexp
        run only benchmarks matching regexp
  -test.benchmem
        print memory allocations for benchmarks
  -test.benchtime d
        run each benchmark for duration d (default 1s)
  -test.blockprofile file
        write a goroutine blocking profile to file
  -test.blockprofilerate rate
        set blocking profile rate (see runtime.SetBlockProfileRate) (default 1)
  -test.count n
        run tests and benchmarks n times (default 1)
  -test.coverprofile file
        write a coverage profile to file
  -test.cpu list
        comma-separated list of cpu counts to run each test with
  -test.cpuprofile file
        write a cpu profile to file
  -test.failfast
        do not start new tests after the first test failure
  -test.fuzz regexp
        run the fuzz test matching regexp
  -test.fuzzcachedir string
        directory where interesting fuzzing inputs are stored (for use only by cmd/go)
  -test.fuzzminimizetime value
        time to spend minimizing a value after finding a failing input (default 1m0s)
  -test.fuzztime value
        time to spend fuzzing; default is to run indefinitely
  -test.fuzzworker
        coordinate with the parent process to fuzz random values (for use only by cmd/go)
  -test.list regexp
        list tests, examples, and benchmarks matching regexp then exit
  -test.memprofile file
        write an allocation profile to file
  -test.memprofilerate rate
        set memory allocation profiling rate (see runtime.MemProfileRate)
  -test.mutexprofile string
        write a mutex contention profile to the named file after execution
  -test.mutexprofilefraction int
        if >= 0, calls runtime.SetMutexProfileFraction() (default 1)
  -test.outputdir dir
        write profiles to dir
  -test.paniconexit0
        panic on call to os.Exit(0)
  -test.parallel n
        run at most n tests in parallel (default 8)
  -test.run regexp
        run only tests and examples matching regexp
  -test.short
        run smaller test suite to save time
  -test.shuffle string
        randomize the execution order of tests and benchmarks (default "off")
  -test.testlogfile file
        write test action log to file (for use only by cmd/go)
  -test.timeout d
        panic test binary after duration d (default 0, timeout disabled)
  -test.trace file
        write an execution trace to file
  -test.v
        verbose: print additional output
  -tmp string
        tmp storage dir path
```

## Example

compile cases

```bash
make build_cases
```

```bash
# Use xfuzz to test some commands without errors within 30 seconds
xfuzz -test.fuzz=FuzzProcess -test.v -test.fuzztime 30s -test.fuzzcachedir ./cache -cmd "cat"
xfuzz -test.fuzz=FuzzProcess -test.v -test.fuzztime 30s -test.fuzzcachedir ./cache -cmd "ls -al"
# Use xfuzz to test some commands with errors with predefined corpus within 30 seconds
xfuzz -test.fuzz=FuzzProcess -test.v -test.fuzztime 30s -test.fuzzcachedir ./cache -corpus "./case/panic/corpus" -cmd "./bin/panic"
xfuzz -test.fuzz=FuzzProcess -test.v -test.fuzztime 30s -test.fuzzcachedir ./cache -corpus "./case/panic/corpus" -cmd "./bin/segmentation_fault"
# Use xfuzz to test some commands that read configuration from the filesystem
xfuzz -test.fuzz=FuzzProcess -test.v -test.fuzztime 30s -test.fuzzcachedir ./cache -cmd-template -tmp ./tmp -cmd "cat {{.File}}"
# Use xfuzz to test some commands that may run for a long time
xfuzz -test.fuzz=FuzzProcess -test.v -test.fuzztime 30s -test.fuzzcachedir ./cache -deadline 1 -cmd "sleep 10"
# Reproduce bugs with toxic configs already found by xfuzz
xfuzz -test.fuzz=FuzzProcess -test.v -test.fuzztime 30s -test.fuzzcachedir ./cache -cmd "./bin/segmentation_fault" -test.run=FuzzProcess/54e5e70cc98e9c3f19f3e43f660f167332c6aaf07fa02eda098cc6b211e37b79
```

## Development Environment configuration

[Install](https://go.dev/doc/install)

```bash
# version >= 1.18
go version
```

## Links

- <https://mijailovic.net/2017/07/29/go-fuzz/>
- <https://gitlab.com/akihe/radamsa>
- <https://github.com/secfigo/Awesome-Fuzzing#tools>
- <https://github.com/AFLplusplus/AFLplusplus>
- <https://github.com/dvyukov/go-fuzz>
