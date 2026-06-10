# Running the l2-execution guest in zkc — findings

## Why a link step is needed
`zig build` (default) emits `zig-out/lib/evm_execution_guest.o`, which is
`Type: REL` (relocatable): no entry point, no PT_LOAD segments, and unresolved
references (`_start`, `memcpy/memmove/memset`, `__udivti3`). The prover normally
supplies `_start` + the final link. For a **standalone** run in the ZkC
interpreter we must link it into an `EXEC` ELF first — the JSON converter reads
PT_LOAD segments + entry point, which a `.o` lacks.

## Run the assembled ELF in zkc with an input file
Run from this directory (`riscv-guests/l2-execution/`):

```bash
# 1. Build the runnable ELF (use the pinned Zig; -> zig-out/bin/evm_execution_guest)
zig build elf

# 2. Get a VALID SSZ input and frame it. The decoder requires the SSZ to begin
#    with the v0.4.1 schema-id 0x00 0x01 (optionally behind a 4-byte LE length);
#    valid bytes live in a zkevm fixture's `statelessInputBytes` field. Then two
#    transport gotchas, both handled below:
#      - the guest reads [8-byte BIG-ENDIAN length][SSZ bytes] at 0x08800000
#      - the go converter byte-REVERSES 0x-hex before RAM, so we pre-reverse it.
FX=$(find zig-pkg -path "*bal_empty_block_no_coinbase.json" | head -1)
python3 - "$FX" <<'PY'
import sys, json
def walk(o, acc):
    if isinstance(o, dict):
        for k, v in o.items(): acc.append(v) if k == "statelessInputBytes" else walk(v, acc)
    elif isinstance(o, list):
        for v in o: walk(v, acc)
acc = []; walk(json.load(open(sys.argv[1])), acc)
ssz = bytes.fromhex(acc[0][2:])              # valid SSZ (starts 00 01 …)
ram = len(ssz).to_bytes(8, 'big') + ssz      # guest frame: [BE u64 len][ssz]
open('fixture_input.in.hex', 'w').write('0x' + ram[::-1].hex())  # pre-reverse for converter
PY

# 3. ELF -> JSON (go converter) + run in zkc, in one step. Exit 0 = validated.
make -C ../../arithmetization/src/test/examples elf-exec \
    BIN_EXT=$PWD/zig-out/bin/evm_execution_guest \
    IN_BYTES=@$PWD/fixture_input.in.hex \
    IN_BYTES_OFFSET=0x08800000 \
    ZKC_EXEC_FLAGS=-q
```
Input-format note: a valid `SszStatelessInput` begins with schema-id `0x00 0x01`,
then a 16-byte all-variable container whose first offset is `16`. The loose
`stateless_input.ssz` in this dir is **not** valid — it has a stray 8-byte prefix
(`25 09 …`) before the real SSZ, so the decoder rejects it (`InvalidSsz`) and the
guest exits 1. Use a fixture's `statelessInputBytes` (as above).

Verify the input landed: in `zig-out/bin/evm_execution_guest.json`, the blob at
offset `0x08800000` starts with the 8-byte big-endian length, then `0001…`.

## Pipeline (what the above does, 3 steps)
1. **link → ELF**: `zig build elf` — opt-in step; adds entry `_start` via
   `src/start.s`, the rv64im memory layout via `linker_script.ld`, and
   `__udivti3`/`mem*` from Zig's soft-float compiler_rt.
2. **ELF → JSON**: `go run arithmetization/src/test/examples/scripts/elf_to_json_gen/main.go <elf> <IN_BYTES> <IN_BYTES_OFFSET>`
3. **run**: `zkc exec [-q] <elf>.json arithmetization/src/main/riscv/main.zkc`

## Status
Working end-to-end. `zig build elf` produces a fully-linked `EXEC` (entry
`_start` @0x0, zero undefined symbols, `__udivti3` from
`compiler_rt.udivmod.__udivti3`); the custom-1 keccak op is a single 32-bit
instruction `00c5852b` (`.insn 4`) inside `<zkvm_keccak256>`. Fed a fixture's
`statelessInputBytes`, the guest decodes the SSZ, executes the block (keccak runs
via the custom op), checks the post-state/receipts roots, and exits **0**
(successful_validation).

The earlier exit-1 was a bad input, not a build/opcode bug: `stateless_input.ssz`
carried a stray 8-byte prefix, so SSZ decode rejected it (`InvalidSsz`) after
~2796 cycles, before any execution. The heap-at-`0x50000000` worry was a red
herring (allocation worked fine).

## Notes / caveats
- `linker_script.ld`'s `IN` origin assumes the default `-Dinput-offset`
  (0x08800000); a custom offset needs a matching linker script.
- The `elf` target is opt-in so the normal `.o` build (what the prover consumes)
  stays unchanged.
- Fallback (no Zig link): hand-link the `.o` with the GNU toolchain — but
  `-lgcc` is unusable (its libgcc is a double-float multilib, won't link against
  this soft-float `lp64` object), so `__udivti3` + `mem*` must be supplied by a
  small soft-float `builtins.c` (constant-shift `__udivti3`):
  `riscv64-unknown-elf-gcc -march=rv64im -mabi=lp64 -nostdlib -T linker_script.ld src/start.s builtins.o <obj> -o guest.elf`
