# L2 Execution Guest

This package contains the RISC-V guest program for vanilla EVM execution. The guest reads a length-prefixed list of stateless execution witnesses from guest memory and executes each witnessed block with Zesu's EVM/stateless executor. Rollup-specific validation is intentionally out of scope for this iteration.

## Scope

- Executes witnessed EVM blocks using canonical block RLP plus `debug_executionWitness`-style state, code, key, and header payloads.
- Provides native Zig tests and checked-in fixtures for local replay.
- Does not include blob compression, recursive proof aggregation, or Rollup-specific public-input validation.
- Keeps cryptographic precompile and signature-recovery acceleration behind Zesu's `accel_impl` boundary. Native tests wire secp256k1 through the system library; the RISC-V guest exposes stable `zkvm_*` hook symbols for future zkVM precompile interception and uses local hash fallbacks until that layer exists.

## Development

The Zig version, dependency checkout, build manifest, and ZKC helper commands are shared by all guests at `riscv-guests/`.

Run from the parent directory:

```bash
cd riscv-guests
make compile GUEST=l2-execution ZIG=/path/to/zig
make test ZIG=/path/to/zig
make fixture-exec GUEST=l2-execution ZIG=/path/to/zig
```

The compiled ELF and JSON integration artifact are written under `riscv-guests/zig-out/bin/`.
