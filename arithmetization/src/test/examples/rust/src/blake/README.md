# Blake2f zkVM tests

This folder contains Rust zkVM tests for the Blake2b-F compression function.

## `blake2f_all_to_in_bytes.py`

Converts each runnable row from `blake2f.all` into the `IN_BYTES` format expected by `blake_with_in_bytes.rs`.

The emitted value is:

```text
0x<213 bytes Blake input><64 bytes expected output>
```

Blank lines and commented rows starting with `;;` are skipped.

## `run_blake2f_all.sh`

Runs `blake_with_in_bytes.rs` once per converted test case:

```bash
./run_blake2f_all.sh
```

It always uses `blake/blake2f.all`.

Run only one original `.all` line:

```bash
./run_blake2f_all.sh 344
```

Run an inclusive range of original `.all` lines:

```bash
./run_blake2f_all.sh 340-350
```

A case is considered passing only when zkc reports:

```text
Program exited successfully (exit with code 0).
```

At the end, the script prints either `all cases passed` or the failing case numbers.
