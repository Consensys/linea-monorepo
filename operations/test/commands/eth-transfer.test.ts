import { beforeEach, describe, expect, it, jest } from "@jest/globals";
import { err, ok } from "neverthrow";

import { runOperationsCommand } from "../helpers/run-command";

type UnknownFn = (...args: unknown[]) => unknown;
type AsyncUnknownFn = (...args: unknown[]) => Promise<unknown>;

const mockBuildHttpsAgent = jest.fn<UnknownFn>(() => ({ agent: "https" }));
const mockCreatePublicClient = jest.fn<UnknownFn>();
const mockEstimateTransactionGas = jest.fn<AsyncUnknownFn>();
const mockGetWeb3SignerSignature = jest.fn<AsyncUnknownFn>();
const mockParseSignature = jest.fn<UnknownFn>(() => ({ r: "0x1", s: "0x2", v: 27n }));
const mockSendRawTransaction = jest.fn<AsyncUnknownFn>();
const mockSerializeTransaction = jest.fn<UnknownFn>(() => "0xsigned");
const { parseEther } = jest.requireActual<typeof import("viem")>("viem");

jest.mock("viem", () => {
  const actual = jest.requireActual<typeof import("viem")>("viem");
  return {
    ...actual,
    createPublicClient: mockCreatePublicClient,
    http: jest.fn((url: string, options?: unknown) => {
      void options;
      return { url };
    }),
    parseSignature: mockParseSignature,
    serializeTransaction: mockSerializeTransaction,
  };
});
jest.mock("../../dist/utils/common/https-agent.js", () => ({
  buildHttpsAgent: mockBuildHttpsAgent,
}));
jest.mock("../../dist/utils/common/signature.js", () => ({
  getWeb3SignerSignature: mockGetWeb3SignerSignature,
}));
jest.mock("../../dist/utils/common/transactions.js", () => ({
  estimateTransactionGas: mockEstimateTransactionGas,
  sendRawTransaction: mockSendRawTransaction,
}));

const senderAddress = "0x0000000000000000000000000000000000000001";
const destinationAddress = "0x0000000000000000000000000000000000000002";

const baseArgs = [
  "eth-transfer",
  "--senderAddress",
  senderAddress,
  "--destinationAddress",
  destinationAddress,
  "--threshold",
  "2",
  "--blockchainRpcUrl",
  "http://127.0.0.1:8545",
  "--web3SignerUrl",
  "http://127.0.0.1:8546",
  "--web3SignerPublicKey",
  "0x1234",
];

describe("eth-transfer", () => {
  const publicClient = {
    getBalance: jest.fn<() => Promise<bigint>>(),
    getTransactionCount: jest.fn<() => Promise<number>>(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
    publicClient.getBalance.mockResolvedValue(parseEther("11"));
    publicClient.getTransactionCount.mockResolvedValue(7);
    mockCreatePublicClient.mockReturnValue(publicClient);
    mockEstimateTransactionGas.mockResolvedValue(ok({ gasLimit: 21_000n, baseFeePerGas: 10n, priorityFeePerGas: 2n }));
    mockGetWeb3SignerSignature.mockResolvedValue(ok("0xsignature"));
    mockSendRawTransaction.mockResolvedValue(ok({ status: "success", transactionHash: "0xtx" }));
  });

  it("loads command help from the oclif registry", async () => {
    const { error, stdout, stderr } = await runOperationsCommand(["eth-transfer", "--help"]);

    expect(error).toBeUndefined();
    expect(stderr).toStrictEqual("");
    expect(stdout).toContain("USAGE");
    expect(stdout).toContain("$ operations eth-transfer");
  });

  it("returns parser errors before executing side effects", async () => {
    const { error } = await runOperationsCommand([
      "eth-transfer",
      "--senderAddress",
      "0x0000000000000000000000000000000000000001",
      "--destinationAddress",
      "0x0000000000000000000000000000000000000002",
      "--threshold",
      "1",
      "--blockchainRpcUrl",
      "http://127.0.0.1:8545",
      "--web3SignerUrl",
      "http://127.0.0.1:8546",
      "--web3SignerPublicKey",
      "0x1234",
    ]);

    expect(error?.message).toContain("Threshold must be higher than 1 ETH");
  });

  it("does nothing when the sender balance is under the threshold", async () => {
    publicClient.getBalance.mockResolvedValue(parseEther("1.5"));

    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("is less than threshold. No action needed.");
    expect(mockEstimateTransactionGas).not.toHaveBeenCalled();
    expect(mockGetWeb3SignerSignature).not.toHaveBeenCalled();
    expect(mockSendRawTransaction).not.toHaveBeenCalled();
  });

  it("signs but skips broadcasting in dry-run mode", async () => {
    const { error, stdout } = await runOperationsCommand([...baseArgs, "--dryRun"]);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Gas estimation: gasLimit=21000 baseFeePerGas=10 priorityFeePerGas=2");
    expect(stdout).toContain("Dry run enabled: Skipping transaction submission to blockchain.");
    expect(mockGetWeb3SignerSignature).toHaveBeenCalledTimes(1);
    expect(mockSendRawTransaction).not.toHaveBeenCalled();
  });

  it("passes a TLS agent to Web3 Signer when TLS flags are provided", async () => {
    const { error, stdout } = await runOperationsCommand([
      ...baseArgs,
      "--dryRun",
      "--tls",
      "--web3SignerKeystorePath",
      "/tmp/keystore.p12",
      "--web3SignerKeystorePassphrase",
      "secret",
      "--web3SignerTrustedStorePath",
      "/tmp/truststore.p12",
      "--web3SignerTrustedStorePassphrase",
      "secret",
    ]);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Using TLS for Web3 Signer communication.");
    expect(mockBuildHttpsAgent).toHaveBeenCalledWith("/tmp/keystore.p12", "secret", "/tmp/truststore.p12", "secret");
    expect(mockGetWeb3SignerSignature).toHaveBeenCalledWith(
      "http://127.0.0.1:8546",
      "0x1234",
      expect.objectContaining({ nonce: 7, value: parseEther("10") }),
      { agent: "https" },
    );
  });

  it("broadcasts a signed transaction", async () => {
    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Broadcasting submitInvoice transaction to the network...");
    expect(stdout).toContain("Transaction succeed: transactionHash=0xtx rewards=10 ETH");
    expect(mockSendRawTransaction).toHaveBeenCalledTimes(1);
  });

  it("returns gas estimation errors", async () => {
    mockEstimateTransactionGas.mockResolvedValue(err(new Error("rpc unavailable")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to estimate gas. message=rpc unavailable");
    expect(mockGetWeb3SignerSignature).not.toHaveBeenCalled();
  });

  it("returns Web3 Signer errors", async () => {
    mockGetWeb3SignerSignature.mockResolvedValue(err(new Error("signer refused")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to get signature from Web3 Signer. message=signer refused");
    expect(mockSendRawTransaction).not.toHaveBeenCalled();
  });

  it("returns broadcast errors", async () => {
    mockSendRawTransaction.mockResolvedValue(err(new Error("mempool rejected")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to send transaction. message=mempool rejected");
  });

  it("returns reverted transaction errors", async () => {
    mockSendRawTransaction.mockResolvedValue(ok({ status: "reverted", transactionHash: "0xreverted" }));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Transaction failed. transactionHash=0xreverted");
  });
});
