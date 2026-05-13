#!/usr/bin/env python3
"""Build a single batched IN_BYTES blob for ``keccak/keccak.rs``.

The Rust program (``rust/src/keccak/keccak.rs``) now expects the input
region to be the concatenation of ``N_VECTORS`` (compile-time constant)
test vectors laid out back-to-back, each 720 bytes long::

    | 680 B padded message (u5440, BE, left-zero-padded) |
    |   8 B msg_len_bits (u64, LE)                       |
    |  32 B expected Keccak-256 digest                   |

This script reads ``keccakf_with_padding.accepts`` (JSON-Lines, the same
fixture used by the go-corset ZkC bench) and emits the corresponding
``0x<hex>`` blob on stdout. The output is meant to be passed verbatim as
``IN_BYTES`` to the ``make compile`` / ``main.go`` pipeline.

Usage::

    python3 build_keccak_in_bytes.py \\
        ~/go/src/github/Consensys/go-corset/testdata/zkc/bench/keccakf_with_padding.accepts \\
        > in_bytes.hex

    # Drive the build with the resulting blob:
    cd ../..
    IN_BYTES="$(cat scripts/in_bytes.hex)" make compile TEST=keccak/keccak.rs
    make zkc-exec TEST=keccak/keccak.rs

The script is intentionally read-only on go-corset's side; if you change
the source ``.accepts`` file there, just rerun this script. The number
of vectors emitted must match ``N_VECTORS`` in ``keccak.rs`` (currently
10000) — the script prints the count to stderr so a mismatch is easy to
catch.
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

MESSAGE_TOTAL_BITS = 5440
MESSAGE_TOTAL_BYTES = MESSAGE_TOTAL_BITS // 8       # 680
LENGTH_FIELD_BYTES = 8
RESULT_BYTES = 32
VECTOR_BYTES = MESSAGE_TOTAL_BYTES + LENGTH_FIELD_BYTES + RESULT_BYTES  # 720


def encode_vector(obj: dict, lineno: int) -> bytes:
    msg_hex = obj["message"][2:]
    if len(msg_hex) != MESSAGE_TOTAL_BYTES * 2:
        raise ValueError(
            f"line {lineno}: message hex is {len(msg_hex)} chars, "
            f"expected {MESSAGE_TOTAL_BYTES * 2}"
        )
    res_hex = obj["result"][2:]
    if len(res_hex) != RESULT_BYTES * 2:
        raise ValueError(
            f"line {lineno}: result hex is {len(res_hex)} chars, expected {RESULT_BYTES * 2}"
        )

    msg_len_bits = int(obj["message_length"], 16)
    if msg_len_bits > MESSAGE_TOTAL_BITS:
        raise ValueError(
            f"line {lineno}: message_length={msg_len_bits} exceeds u{MESSAGE_TOTAL_BITS}"
        )

    return (
        bytes.fromhex(msg_hex)
        + msg_len_bits.to_bytes(LENGTH_FIELD_BYTES, "little")
        + bytes.fromhex(res_hex)
    )


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print(f"usage: {argv[0]} <keccakf_with_padding.accepts>", file=sys.stderr)
        return 1

    src = Path(argv[1])
    if not src.exists():
        print(f"error: not found: {src}", file=sys.stderr)
        return 1

    chunks: list[bytes] = []
    for lineno, raw in enumerate(src.read_text().splitlines(), start=1):
        line = raw.strip()
        if not line:
            continue
        chunks.append(encode_vector(json.loads(line), lineno))

    blob = b"".join(chunks)
    n = len(chunks)
    expected_bytes = n * VECTOR_BYTES
    if len(blob) != expected_bytes:
        print(
            f"error: assembled blob is {len(blob)} bytes, expected {expected_bytes} "
            f"({n} * {VECTOR_BYTES})",
            file=sys.stderr,
        )
        return 1

    sys.stdout.write("0x" + blob.hex() + "\n")
    print(
        f"emitted {n} vectors -> {len(blob)} bytes "
        f"({len(blob) * 2 + 2} hex chars on stdout). "
        f"Make sure N_VECTORS in keccak.rs == {n}.",
        file=sys.stderr,
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
