import { GetPublicKeyCommand, SignCommand } from "@aws-sdk/client-kms";
import { Address, Hex, keccak256, recoverAddress, serializeSignature, serializeTransaction } from "viem";
import { publicKeyToAddress } from "viem/accounts";

import { createLoggerMock } from "../../__tests__/helpers/factories";
import { ILogger } from "../../logging/ILogger";
import { AwsKmsSignerClientAdapter } from "../AwsKmsSignerClientAdapter";

const mockSend = jest.fn();

jest.mock("@aws-sdk/client-kms", () => ({
  KMSClient: jest.fn().mockImplementation(() => ({ send: mockSend })),
  CreateKeyCommand: jest.fn().mockImplementation((input: unknown) => ({ __input: input })),
  GetPublicKeyCommand: jest.fn().mockImplementation((input: unknown) => ({ __input: input })),
  SignCommand: jest.fn().mockImplementation((input: unknown) => ({ __input: input })),
}));

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    serializeTransaction: jest.fn(),
    keccak256: jest.fn(),
    recoverAddress: jest.fn(),
    serializeSignature: jest.fn(),
  };
});

jest.mock("viem/accounts", () => ({
  publicKeyToAddress: jest.fn(),
}));

const serializeTransactionMock = serializeTransaction as jest.MockedFunction<typeof serializeTransaction>;
const keccak256Mock = keccak256 as jest.MockedFunction<typeof keccak256>;
const recoverAddressMock = recoverAddress as jest.MockedFunction<typeof recoverAddress>;
const serializeSignatureMock = serializeSignature as jest.MockedFunction<typeof serializeSignature>;
const publicKeyToAddressMock = publicKeyToAddress as jest.MockedFunction<typeof publicKeyToAddress>;
const GetPublicKeyCommandMock = GetPublicKeyCommand as unknown as jest.Mock;
const SignCommandMock = SignCommand as unknown as jest.Mock;

describe("AwsKmsSignerClientAdapter", () => {
  const KMS_KEY_ID = "arn:aws:kms:us-east-1:123456789012:key/test-key-id";
  const EXPECTED_ADDRESS: Address = "0xD42E308FC964b71E18126dF469c21B0d7bcb86cC";
  const SERIALIZED_TX = "0x02serialized";
  const TX_HASH: Hex = "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef";
  const SIGNATURE_HEX: Hex = "0xsignature";

  // Full 88-byte DER-encoded SubjectPublicKeyInfo with 65-byte uncompressed key (0x04 || 0xab*64)
  const DER_PUBLIC_KEY = new Uint8Array([
    0x30,
    0x56,
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
    0x03,
    0x42,
    0x00,
    0x04,
    ...new Uint8Array(64).fill(0xab),
  ]);

  // DER-encoded ECDSA signature: 30 44 02 20 <r: 32 bytes of 0x01> 02 20 <s: 32 bytes of 0x02>
  const R_BYTES = new Uint8Array(32).fill(0x01);
  const S_BYTES = new Uint8Array(32).fill(0x02);
  const DER_SIGNATURE = new Uint8Array([0x30, 0x44, 0x02, 0x20, ...R_BYTES, 0x02, 0x20, ...S_BYTES]);

  const EXPECTED_R_HEX = `0x${"01".repeat(32)}` as Hex;
  const EXPECTED_S_HEX = `0x${"02".repeat(32)}` as Hex;

  const SAMPLE_TRANSACTION = {
    type: "eip1559",
    chainId: 1337,
    nonce: 0,
    gas: BigInt(21_000),
    maxFeePerGas: BigInt(1_000_000_000),
    maxPriorityFeePerGas: BigInt(100_000_000),
    to: "0x0000000000000000000000000000000000000000",
    value: BigInt(0),
    data: "0x",
  };

  let logger: jest.Mocked<ILogger>;

  beforeEach(() => {
    jest.clearAllMocks();
    mockSend.mockReset();
    logger = createLoggerMock({ name: "aws-kms-signer" });

    publicKeyToAddressMock.mockReturnValue(EXPECTED_ADDRESS);
    serializeTransactionMock.mockReturnValue(SERIALIZED_TX);
    keccak256Mock.mockReturnValue(TX_HASH);
    recoverAddressMock.mockResolvedValue(EXPECTED_ADDRESS);
    serializeSignatureMock.mockReturnValue(SIGNATURE_HEX);
  });

  const createAdapter = () => new AwsKmsSignerClientAdapter(logger, KMS_KEY_ID, { region: "us-east-1" });

  const initAdapter = async () => {
    mockSend.mockResolvedValueOnce({ PublicKey: DER_PUBLIC_KEY });
    const adapter = createAdapter();
    await adapter.init();
    return adapter;
  };

  describe("initialization", () => {
    it("should fetch public key from KMS and derive Ethereum address", async () => {
      mockSend.mockResolvedValueOnce({ PublicKey: DER_PUBLIC_KEY });
      const adapter = createAdapter();

      await adapter.init();

      expect(GetPublicKeyCommandMock).toHaveBeenCalledWith({ KeyId: KMS_KEY_ID });
      expect(publicKeyToAddressMock).toHaveBeenCalledWith(`0x04${"ab".repeat(64)}`);
      expect(adapter.getAddress()).toBe(EXPECTED_ADDRESS);
    });

    it("should throw when KMS returns empty public key", async () => {
      mockSend.mockResolvedValueOnce({ PublicKey: undefined });
      const adapter = createAdapter();

      await expect(adapter.init()).rejects.toThrow("AWS KMS returned empty public key");
    });
  });

  describe("create factory", () => {
    it("should create and initialize adapter in one step", async () => {
      mockSend.mockResolvedValueOnce({ PublicKey: DER_PUBLIC_KEY });

      const adapter = await AwsKmsSignerClientAdapter.create(logger, KMS_KEY_ID, { region: "us-east-1" });

      expect(adapter.getAddress()).toBe(EXPECTED_ADDRESS);
    });
  });

  describe("sign", () => {
    it("should sign transaction via AWS KMS and return serialized signature", async () => {
      const adapter = await initAdapter();
      mockSend.mockResolvedValueOnce({ Signature: DER_SIGNATURE });

      const signature = await adapter.sign(SAMPLE_TRANSACTION as any);

      expect(serializeTransactionMock).toHaveBeenCalledWith(SAMPLE_TRANSACTION);
      expect(keccak256Mock).toHaveBeenCalledWith(SERIALIZED_TX);
      expect(SignCommandMock).toHaveBeenCalledWith({
        KeyId: KMS_KEY_ID,
        Message: expect.any(Uint8Array),
        MessageType: "DIGEST",
        SigningAlgorithm: "ECDSA_SHA_256",
      });
      expect(recoverAddressMock).toHaveBeenCalledWith({
        hash: TX_HASH,
        signature: { r: EXPECTED_R_HEX, s: EXPECTED_S_HEX, yParity: 0 },
      });
      expect(serializeSignatureMock).toHaveBeenCalledWith({
        r: EXPECTED_R_HEX,
        s: EXPECTED_S_HEX,
        yParity: 0,
      });
      expect(signature).toBe(SIGNATURE_HEX);
    });

    it("should try yParity=1 when yParity=0 does not recover the correct address", async () => {
      const adapter = await initAdapter();
      mockSend.mockResolvedValueOnce({ Signature: DER_SIGNATURE });
      recoverAddressMock
        .mockResolvedValueOnce("0x0000000000000000000000000000000000000001" as Address)
        .mockResolvedValueOnce(EXPECTED_ADDRESS);

      const signature = await adapter.sign(SAMPLE_TRANSACTION as any);

      expect(recoverAddressMock).toHaveBeenCalledTimes(2);
      expect(serializeSignatureMock).toHaveBeenCalledWith({
        r: EXPECTED_R_HEX,
        s: EXPECTED_S_HEX,
        yParity: 1,
      });
      expect(signature).toBe(SIGNATURE_HEX);
    });

    it("should normalize s to lower half of curve order (EIP-2)", async () => {
      const adapter = await initAdapter();
      const highSBytes = new Uint8Array(32).fill(0xff);
      const highSDerSig = new Uint8Array([0x30, 0x44, 0x02, 0x20, ...R_BYTES, 0x02, 0x20, ...highSBytes]);
      mockSend.mockResolvedValueOnce({ Signature: highSDerSig });

      await adapter.sign(SAMPLE_TRANSACTION as any);

      const signCall = serializeSignatureMock.mock.calls[0][0];
      const originalSHex = `0x${"ff".repeat(32)}`;
      expect(signCall.s).not.toBe(originalSHex);
    });

    it("should throw when KMS returns empty signature", async () => {
      const adapter = await initAdapter();
      mockSend.mockResolvedValueOnce({ Signature: undefined });

      await expect(adapter.sign(SAMPLE_TRANSACTION as any)).rejects.toThrow("AWS KMS returned empty signature");
    });

    it("should throw when yParity cannot be determined", async () => {
      const adapter = await initAdapter();
      mockSend.mockResolvedValueOnce({ Signature: DER_SIGNATURE });
      recoverAddressMock.mockResolvedValue("0x0000000000000000000000000000000000000001" as Address);

      await expect(adapter.sign(SAMPLE_TRANSACTION as any)).rejects.toThrow(
        "Failed to determine signature recovery parameter (yParity)",
      );
    });

    it("should throw when not initialized", async () => {
      const adapter = createAdapter();

      await expect(adapter.sign(SAMPLE_TRANSACTION as any)).rejects.toThrow(
        "AwsKmsSignerClientAdapter not initialized. Call init() first.",
      );
    });
  });

  describe("getAddress", () => {
    it("should return the derived address after initialization", async () => {
      const adapter = await initAdapter();

      expect(adapter.getAddress()).toBe(EXPECTED_ADDRESS);
    });

    it("should throw when not initialized", () => {
      const adapter = createAdapter();

      expect(() => adapter.getAddress()).toThrow("AwsKmsSignerClientAdapter not initialized. Call init() first.");
    });
  });
});
