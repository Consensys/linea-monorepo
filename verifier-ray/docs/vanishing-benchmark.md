# Vanishing Benchmarking

This document explains how verifier-ray measures the RISC-V cost of the vanishing verifier. The generated single-case result lives in `bench/vanishing-benchmark.md`, and the generated comparison report lives in `bench/vanishing-benchmark-comparison.md`.

## Technique

The benchmark builds a tiny R5 guest from `bench/vanishing_main.zig`. That guest imports one generated vanishing fixture, calls `vanishing.verify(selected_case.system, selected_case.input)`, and exits with code `0` on success or `1` on verifier rejection.

The build path is:

```text
Zig benchmark guest
  -> RISC-V ELF
  -> zkc JSON input
  -> zkc RISC-V interpreter
  -> cycle_count output
```

`make bench-vanishing-doc` drives that path. Zig builds the guest for the freestanding RISC-V target, `elf_to_json_gen` turns the ELF loadable sections into zkc input blobs, and `bench/riscv_main_bench.zkc` interprets the guest. The zkc runner increments `cycle_count` once per interpreted RISC-V instruction and exposes the final count as a public output.

## Separate Fixture

The ordinary vanishing test fixture in `testdata/generated/vanishing.zig` is shaped for unit tests. It stores proof values as raw `u32` limbs and the tests convert them into `runtime.RoundMessage`, `field.Element`, and `ext.Ext` values using allocator-backed helpers.

The benchmark fixture in `bench/generated/vanishing.zig` is shaped for measurement. It emits typed `vanishing.CheckInput` constants directly, so the guest does not spend measured RISC-V cycles parsing proof bytes, allocating arrays, or converting raw limbs into field elements. The selected proof/system data is compiled into the RISC-V ELF as immutable constants and loaded into guest memory before the interpreter starts counting RISC-V instructions.

## Measurement Scope

The reported value is an instruction-count proxy for verifier cost in the R5 environment. It includes the benchmark guest entry path, reads from embedded fixture data, `vanishing.verify`, and the guest exit path.

It does not include generating the fixtures, converting raw test fixtures into typed inputs, serializing proof bytes, parsing proof bytes, or copying ELF blobs from the zkc JSON into RAM before the RISC-V loop starts.

One benchmark run measures one selected case:

```bash
make bench-vanishing-doc VANISHING_BENCH_CASE=1 VANISHING_BENCH_RELEASE=small
```

The generated catalog currently has 83 honest benchmark cases. `VANISHING_BENCH_CASE` selects which one is compiled into the guest.

## Comparison Runs

Use `bench-vanishing-compare-doc` to benchmark more than one honest case and write a comparison table:

```bash
make bench-vanishing-compare-doc VANISHING_BENCH_CASES=0-10
```

`VANISHING_BENCH_CASES` accepts `all`, an inclusive range, a comma-separated list, or a mix:

```bash
make bench-vanishing-compare-doc VANISHING_BENCH_CASES=all
make bench-vanishing-compare-doc VANISHING_BENCH_CASES=0,2,5-8
```

`make bench-vanishing-all-doc` is a shortcut for the full generated catalog. The comparison target rebuilds the tiny R5 guest once per selected case, stores logs under `zig-out/vanishing-bench/case-<index>.log`, and renders the table from those logs plus `bench/generated/vanishing.zig` metadata.

## Invalid Sanity Check

The invalid-input path is a rejection smoke test, not a cost report. It builds the same benchmark guest with `-Dvanishing-invalid=true`, selects an invalid variant for the chosen case, and expects zkc to fail with guest exit code `1`.

```bash
make bench-vanishing-zkc-failing-expected VANISHING_BENCH_CASE=0
```

Only scenarios that have source invalid assignments can be used this way. The current catalog has invalid variants for cases `0..45`, so there are 46 invalid inputs. Cases `46..82` do not have invalid variants and will fail at compile time if selected with `VANISHING_INVALID=true`.
