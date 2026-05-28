# How to

To trigger the test with inputs one must construct a valid input. A valid input is structured like so

`0x <input_len> <result> <inputs>`

where
- `<input_len>` is a 4 byte, little endian integer indicating the length in bytes of `<inputs>`,
- `<inputs>` is an input byte slice, composed of 3-byte little-endian words
- `<result>` is the **full state** of the Poseidon hash at the end of the computation

## Example input string

input_len ≡ `30 00 00 00` ≡ 48 in little-endian
input     ≡ `000000 010000 020000 030000 040000 050000 060000 070000 080000 090000 0a0000 0b0000 0c0000 0d0000 0e0000 0f0000` (where e.g. `050000 ≡ 05 00 00` is 5  in little endian)
result    ≡ `67a1d2223fcc1149c0c9f819a25f722c2f603a5eb4c61453ffce495a3a716b546314f2172196381206c3610774d65476010c66277a0575741ec3e152fb3d97610000000100000200000300000400000500000600000700000800000900000a00000b00000c00000d00000e00000f0000`

full input string ≡ `0x3000000067a1d2223fcc1149c0c9f819a25f722c2f603a5eb4c61453ffce495a3a716b546314f2172196381206c3610774d65476010c66277a0575741ec3e152fb3d97610000000100000200000300000400000500000600000700000800000900000a00000b00000c00000d00000e00000f0000`

## Example command

```bash
zkc-test poseidon1/poseidon1_with_in_bytes.rs IN_BYTES="0x3000000067a1d2223fcc1149c0c9f819a25f722c2f603a5eb4c61453ffce495a3a716b546314f2172196381206c3610774d65476010c66277a0575741ec3e152fb3d97610000000100000200000300000400000500000600000700000800000900000a00000b00000c00000d00000e00000f0000"
zkc-test poseidon1/poseidon1_with_embedded.rs
````
