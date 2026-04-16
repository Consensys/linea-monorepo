import { Address, Hex, recoverAddress } from "viem";

import { bigintToHex32, decodeDerSignature, recoverYParity } from "../ecdsa-signature";

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    recoverAddress: jest.fn(),
  };
});

const recoverAddressMock = recoverAddress as jest.MockedFunction<typeof recoverAddress>;

describe("decodeDerSignature", () => {
  // DER-encoded ECDSA signature: 30 44 02 20 <r: 32 bytes> 02 20 <s: 32 bytes>
  const R_BYTES = new Uint8Array(32).fill(0x01);
  const S_BYTES = new Uint8Array(32).fill(0x02);
  const DER_SIGNATURE = new Uint8Array([0x30, 0x44, 0x02, 0x20, ...R_BYTES, 0x02, 0x20, ...S_BYTES]);

  it("should parse DER-encoded ECDSA signature into r and s", () => {
    const { r, s } = decodeDerSignature(DER_SIGNATURE);

    expect(r).toBe(BigInt(`0x${"01".repeat(32)}`));
    expect(s).toBe(BigInt(`0x${"02".repeat(32)}`));
  });

  it("should handle DER integers with leading zero padding", () => {
    // r = 0x80 needs 0x00 padding → DER: 02 02 00 80
    // s = 0x7f no padding → DER: 02 01 7f
    const derSig = new Uint8Array([0x30, 0x07, 0x02, 0x02, 0x00, 0x80, 0x02, 0x01, 0x7f]);

    const { r, s } = decodeDerSignature(derSig);

    expect(r).toBe(0x80n);
    expect(s).toBe(0x7fn);
  });

  it("should normalize s to lower half of curve order (EIP-2)", () => {
    const highSBytes = new Uint8Array(32).fill(0xff);
    const highSDerSig = new Uint8Array([0x30, 0x44, 0x02, 0x20, ...R_BYTES, 0x02, 0x20, ...highSBytes]);

    const { s } = decodeDerSignature(highSDerSig);

    const originalS = BigInt(`0x${"ff".repeat(32)}`);
    expect(s).not.toBe(originalS);
    const SECP256K1_HALF_N = 0xfffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141n / 2n;
    expect(s <= SECP256K1_HALF_N).toBe(true);
  });

  it("should not change s when already in lower half", () => {
    const { s } = decodeDerSignature(DER_SIGNATURE);

    // 0x020202...02 is well below half-N
    expect(s).toBe(BigInt(`0x${"02".repeat(32)}`));
  });

  it("should throw when DER structure is invalid", () => {
    expect(() => decodeDerSignature(new Uint8Array([0xff, 0x00]))).toThrow(
      "Failed to parse DER-encoded ECDSA signature",
    );
  });
});

describe("recoverYParity", () => {
  const HASH: Hex = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef";
  const R: Hex = `0x${"01".repeat(32)}`;
  const S: Hex = `0x${"02".repeat(32)}`;
  const ADDRESS: Address = "0xD42E308FC964b71E18126dF469c21B0d7bcb86cC";

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("should return 0 when yParity=0 recovers the correct address", async () => {
    recoverAddressMock.mockResolvedValueOnce(ADDRESS);

    const result = await recoverYParity(HASH, R, S, ADDRESS);

    expect(result).toBe(0);
    expect(recoverAddressMock).toHaveBeenCalledTimes(1);
  });

  it("should return 1 when yParity=0 does not match but yParity=1 does", async () => {
    recoverAddressMock
      .mockResolvedValueOnce("0x0000000000000000000000000000000000000001" as Address)
      .mockResolvedValueOnce(ADDRESS);

    const result = await recoverYParity(HASH, R, S, ADDRESS);

    expect(result).toBe(1);
    expect(recoverAddressMock).toHaveBeenCalledTimes(2);
  });

  it("should throw when neither candidate recovers the address", async () => {
    recoverAddressMock.mockResolvedValue("0x0000000000000000000000000000000000000001" as Address);

    await expect(recoverYParity(HASH, R, S, ADDRESS)).rejects.toThrow(
      "Failed to determine signature recovery parameter (yParity)",
    );
  });

  it("should handle recoverAddress throwing for one candidate", async () => {
    recoverAddressMock.mockRejectedValueOnce(new Error("invalid")).mockResolvedValueOnce(ADDRESS);

    const result = await recoverYParity(HASH, R, S, ADDRESS);

    expect(result).toBe(1);
  });

  it("should perform case-insensitive address comparison", async () => {
    recoverAddressMock.mockResolvedValueOnce(ADDRESS.toLowerCase() as Address);

    const result = await recoverYParity(HASH, R, S, ADDRESS);

    expect(result).toBe(0);
  });
});

describe("bigintToHex32", () => {
  it("should convert a small value to zero-padded 32-byte hex", () => {
    expect(bigintToHex32(1n)).toBe(`0x${"0".repeat(63)}1`);
  });

  it("should convert a 32-byte value without extra padding", () => {
    const value = BigInt(`0x${"ff".repeat(32)}`);
    expect(bigintToHex32(value)).toBe(`0x${"ff".repeat(32)}`);
  });
});
