# Blake2f zkVM tests

This folder contains Rust zkVM tests for the Blake2b-F compression function.

## `blake2f.all`

`blake2f.all` contains one `IN_BYTES` value per line.

Each line is:

```text
0x<213 bytes Blake input><64 bytes expected output>
```

## Run

Run every line in `blake2f.all` with the examples Makefile:

```bash
make -f arithmetization/src/test/examples/Makefile blake2f-all
```

A case is considered passing only when zkc reports:

```text
Program exited successfully (exit with code 0).
```
