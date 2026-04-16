import { extractUncompressedPublicKeyFromDer } from "../ec-publickey";

describe("extractUncompressedPublicKeyFromDer", () => {
  // 64-byte XY coordinates (all 0xab)
  const PUBLIC_KEY_XY = new Uint8Array(64).fill(0xab);

  // 65-byte uncompressed secp256k1 public key: 0x04 || XY
  const UNCOMPRESSED_KEY = new Uint8Array([0x04, ...PUBLIC_KEY_XY]);

  // Standard DER header for secp256k1 SubjectPublicKeyInfo (23 bytes)
  // 30 56 — outer SEQUENCE (86 bytes)
  //   30 10 — algorithm SEQUENCE (16 bytes)
  //     06 07 2a8648ce3d0201 — OID 1.2.840.10045.2.1 (ecPublicKey)
  //     06 05 2b81040a       — OID 1.3.132.0.10 (secp256k1)
  //   03 42 00               — BIT STRING (66 bytes, 0 unused bits)
  const DER_HEADER = new Uint8Array([
    0x30, 0x56, 0x30, 0x10, 0x06, 0x07, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x02, 0x01, 0x06, 0x05, 0x2b, 0x81, 0x04, 0x00,
    0x0a, 0x03, 0x42, 0x00,
  ]);

  const DER_PUBLIC_KEY = new Uint8Array([...DER_HEADER, ...UNCOMPRESSED_KEY]);

  it("should extract 65-byte uncompressed key from DER-encoded SubjectPublicKeyInfo", () => {
    const result = extractUncompressedPublicKeyFromDer(DER_PUBLIC_KEY);

    expect(result).toBe(`0x04${"ab".repeat(64)}`);
  });

  it("should throw when ASN.1 structure is invalid", () => {
    const garbage = new Uint8Array(88).fill(0xff);

    expect(() => extractUncompressedPublicKeyFromDer(garbage)).toThrow(
      "Failed to parse DER-encoded SubjectPublicKeyInfo",
    );
  });

  it("should throw when key does not start with 0x04 prefix (compressed key)", () => {
    // Replace the 0x04 byte with 0x02 (compressed key prefix)
    const badDer = new Uint8Array(DER_PUBLIC_KEY);
    badDer[DER_HEADER.length] = 0x02;

    expect(() => extractUncompressedPublicKeyFromDer(badDer)).toThrow(
      "Expected 65-byte uncompressed secp256k1 public key (0x04 prefix)",
    );
  });

  it("should throw when key length is wrong", () => {
    // Build a SubjectPublicKeyInfo with only 33 bytes in the BIT STRING (compressed key)
    const compressedKey = new Uint8Array([0x02, ...new Uint8Array(32).fill(0xab)]);
    const shortBitString = new Uint8Array([0x03, 0x22, 0x00, ...compressedKey]); // BIT STRING (34 bytes, 0 unused)
    const shortDer = new Uint8Array([
      0x30,
      16 + shortBitString.length,
      // algorithm SEQUENCE (unchanged)
      0x30,
      0x10,
      0x06,
      0x07,
      0x2a,
      0x86,
      0x48,
      0xce,
      0x3d,
      0x02,
      0x01,
      0x06,
      0x05,
      0x2b,
      0x81,
      0x04,
      0x00,
      0x0a,
      ...shortBitString,
    ]);

    expect(() => extractUncompressedPublicKeyFromDer(shortDer)).toThrow(
      "Expected 65-byte uncompressed secp256k1 public key (0x04 prefix)",
    );
  });
});
