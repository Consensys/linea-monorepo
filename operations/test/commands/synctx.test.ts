import { beforeEach, describe, expect, it, jest } from "@jest/globals";

import { normalizeWhitespace, runOperationsCommand } from "../helpers/run-command";

type UnknownFn = (...args: unknown[]) => unknown;
type AsyncUnknownFn = (...args: unknown[]) => Promise<unknown>;

const mockBroadcastTransaction = jest.fn<AsyncUnknownFn>(async () => ({ hash: "0xbroadcast" }));
const mockHexlify = jest.fn((bytes: Uint8Array) => `0x${Buffer.from(bytes).toString("hex")}`);
const mockKeccak256 = jest.fn<UnknownFn>(() => "0xhash");
const mockRandomBytes = jest.fn((length: number) => Uint8Array.from({ length }, (_, index) => index + 1));
const mockResolveAddress = jest.fn(async (address: string) => address);
const mockSend = jest.fn<AsyncUnknownFn>(async () => ({}));
const mockTransaction = {
  from: jest.fn((transaction: { nonce?: number }) => ({ serialized: `0xserialized${transaction.nonce ?? ""}` })),
};
const mockJsonRpcProvider = class {
  send = mockSend;
  broadcastTransaction = mockBroadcastTransaction;

  constructor(public readonly url: string) {}
};

jest.mock("ethers", () => ({
  ethers: {
    JsonRpcProvider: mockJsonRpcProvider,
    Transaction: mockTransaction,
    hexlify: mockHexlify,
    keccak256: mockKeccak256,
    randomBytes: mockRandomBytes,
    resolveAddress: mockResolveAddress,
  },
}));

const tx = {
  hash: "0xhash",
  nonce: 1,
  gas: "21000",
  gasPrice: "1000000000",
  input: "0x",
  value: "1",
  type: 0,
  to: "0x0000000000000000000000000000000000000001",
};

describe("synctx", () => {
  beforeEach(() => {
    mockSend.mockReset();
    mockBroadcastTransaction.mockReset();
    mockKeccak256.mockReset();
    mockKeccak256.mockReturnValue("0xhash");
    mockBroadcastTransaction.mockResolvedValue({ hash: "0xbroadcast" });
  });

  it("loads command help from the oclif registry", async () => {
    const { error, stdout, stderr } = await runOperationsCommand(["synctx", "--help"]);

    expect(error).toBeUndefined();
    expect(stderr).toStrictEqual("");
    expect(stdout).toContain("USAGE");
    expect(stdout).toContain("$ operations synctx");
  });

  it("returns runtime validation errors", async () => {
    const { error } = await runOperationsCommand(["synctx", "--target", "http://127.0.0.1:8545"]);
    const message = normalizeWhitespace(error?.message ?? "");

    expect(message).toContain("Invalid flag values are supplied");
    expect(message).toContain("exclusive, and at least one needs to be specified");
  });

  it("returns runtime validation errors for invalid node targets", async () => {
    const { error } = await runOperationsCommand(["synctx", "--source", "localhost:8500", "--target", "not-a-url"]);

    expect(error?.message).toContain("Invalid nodes supplied to source and/or target; must be valid URLs");
  });

  it("stops when the source txpool has no pending transactions", async () => {
    mockSend
      .mockResolvedValueOnce("Geth/v1")
      .mockResolvedValueOnce("Geth/v1")
      .mockResolvedValueOnce({ pending: {}, queued: {} })
      .mockResolvedValueOnce({ pending: {}, queued: {} });

    const { error, stdout } = await runOperationsCommand([
      "synctx",
      "--source",
      "http://127.0.0.1:8500",
      "--target",
      "http://127.0.0.1:8501",
    ]);

    expect(error).toBeUndefined();
    expect(stdout).toContain("No pending transactions found on source node");
    expect(mockBroadcastTransaction).not.toHaveBeenCalled();
  });

  it("computes the pending tx delta and stops before broadcasting in dry-run mode", async () => {
    mockSend
      .mockResolvedValueOnce("Geth/v1")
      .mockResolvedValueOnce("Geth/v1")
      .mockResolvedValueOnce({ pending: { [tx.to]: { "1": tx } }, queued: {} })
      .mockResolvedValueOnce({ pending: {}, queued: {} });

    const { error, stdout } = await runOperationsCommand([
      "synctx",
      "--source",
      "8500",
      "--target",
      "8501",
      "--local",
      "--dry-run",
      "--concurrency",
      "2",
    ]);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Source geth node: http://localhost:8500");
    expect(stdout).toContain("Target geth node: http://localhost:8501");
    expect(stdout).toContain("Pending transactions to process: 1");
    expect(stdout).toContain("Total batches to process: 1");
    expect(mockBroadcastTransaction).not.toHaveBeenCalled();
  });

  it("broadcasts transactions from a file", async () => {
    mockSend.mockResolvedValueOnce("Geth/v1");
    mockKeccak256.mockReturnValue("0xfile");

    const { error, stdout } = await runOperationsCommand([
      "synctx",
      "--target",
      "http://127.0.0.1:8501",
      "--file",
      "test/fixtures/synctx-transactions.json",
      "--concurrency",
      "1",
    ]);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Skip checking source node type as txs file is supplied");
    expect(stdout).toContain("Pending transactions to process: 1");
    expect(stdout).toContain("Total count: 1 - Success: 1 - Errors: 0 - Total Success: 1 - Total Errors: 0");
    expect(mockBroadcastTransaction).toHaveBeenCalledWith("0xserialized1");
  });

  it("logs broadcast errors and continues processing the batch", async () => {
    mockSend
      .mockResolvedValueOnce("Geth/v1")
      .mockResolvedValueOnce("Geth/v1")
      .mockResolvedValueOnce({
        pending: {
          [tx.to]: {
            "1": tx,
            "2": { ...tx, hash: "0xhash2", nonce: 2 },
          },
        },
        queued: {},
      })
      .mockResolvedValueOnce({ pending: {}, queued: {} });
    mockKeccak256.mockReturnValueOnce("0xhash").mockReturnValueOnce("0xhash2");
    mockBroadcastTransaction.mockResolvedValueOnce({ hash: "0xbroadcast" }).mockRejectedValueOnce(new Error("boom"));

    const { error, stdout } = await runOperationsCommand([
      "synctx",
      "--source",
      "http://127.0.0.1:8500",
      "--target",
      "http://127.0.0.1:8501",
      "--concurrency",
      "2",
    ]);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Error broadcasting transaction: boom");
    expect(stdout).toContain("Total count: 2 - Success: 1 - Errors: 1 - Total Success: 1 - Total Errors: 1");
  });
});
