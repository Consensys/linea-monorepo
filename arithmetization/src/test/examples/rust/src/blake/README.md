# Blake2f zkVM tests

This folder contains Rust zkVM tests for the Blake2b-F compression function.

## `blake2f.all`

`blake2f.all` now contains one `IN_BYTES` value per line.

Each line is:

```text
0x<213 bytes Blake input><64 bytes expected output>
```

## `Makefile`

Runs `blake_with_in_bytes.rs` once per line in `blake2f.all`:

```bash
make
```

Run only one case:

```bash
make SELECTOR=344
```

Run an inclusive range of cases:

```bash
make SELECTOR=340-350
```

A case is considered passing only when zkc reports:

```text
Program exited successfully (exit with code 0).
```

At the end, the Makefile prints either `all cases passed` or the failing case numbers.
