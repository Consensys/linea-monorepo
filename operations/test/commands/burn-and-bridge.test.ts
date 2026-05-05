import { beforeEach, describe, expect, it, jest } from "@jest/globals";
import { err, ok } from "neverthrow";

import { runOperationsCommand } from "../helpers/run-command";

type UnknownFn = (...args: unknown[]) => unknown;
type AsyncUnknownFn = (...args: unknown[]) => Promise<unknown>;

const mockComputeBurnAndBridgeCalldata = jest.fn<UnknownFn>(() => "0xburnandbridge");
const mockComputeSwapCalldata = jest.fn<UnknownFn>(() => "0xswap");
const mockCreatePublicClient = jest.fn<UnknownFn>();
const mockCreateWalletClient = jest.fn<UnknownFn>();
const mockEstimateTransactionGas = jest.fn<AsyncUnknownFn>();
const mockGetBalance = jest.fn<AsyncUnknownFn>();
const mockGetInvoiceArrears = jest.fn<AsyncUnknownFn>();
const mockGetMinimumFee = jest.fn<AsyncUnknownFn>();
const mockGetQuote = jest.fn<AsyncUnknownFn>();
const mockParseEventLogs = jest.fn<UnknownFn>(() => []);
const mockSendTransaction = jest.fn<AsyncUnknownFn>();
const { parseEther, parseUnits } = jest.requireActual<typeof import("viem")>("viem");

jest.mock("date-fns-tz", () => jest.requireActual("../mocks/date-fns-tz"));
jest.mock("viem", () => {
  const actual = jest.requireActual<typeof import("viem")>("viem");
  return {
    ...actual,
    createPublicClient: mockCreatePublicClient,
    createWalletClient: mockCreateWalletClient,
    http: jest.fn((url: string, options?: unknown) => {
      void options;
      return { url };
    }),
    parseEventLogs: mockParseEventLogs,
  };
});
jest.mock("viem/accounts", () => ({
  privateKeyToAccount: jest.fn((privateKey: string) => ({
    address: "0x0000000000000000000000000000000000000abc",
    privateKey,
  })),
  privateKeyToAddress: jest.fn(() => "0x0000000000000000000000000000000000000abc"),
}));
jest.mock("viem/actions", () => {
  const actual = jest.requireActual<typeof import("viem/actions")>("viem/actions");
  return {
    ...actual,
    getBalance: mockGetBalance,
  };
});
jest.mock("../../dist/utils/burn-and-bridge/contract.js", () => ({
  computeBurnAndBridgeCalldata: mockComputeBurnAndBridgeCalldata,
  computeSwapCalldata: mockComputeSwapCalldata,
  getInvoiceArrears: mockGetInvoiceArrears,
  getMinimumFee: mockGetMinimumFee,
  getQuote: mockGetQuote,
}));
jest.mock("../../dist/utils/common/transactions.js", () => ({
  estimateTransactionGas: mockEstimateTransactionGas,
  sendTransaction: mockSendTransaction,
}));

const vaultAddress = "0x0000000000000000000000000000000000000001";
const messageServiceAddress = "0x0000000000000000000000000000000000000002";
const quoteAddress = "0x0000000000000000000000000000000000000003";

const baseArgs = [
  "burn-and-bridge",
  "--signerPrivateKey",
  "0x1234",
  "--rollupRevenueVaultContractAddress",
  vaultAddress,
  "--l2MessageServiceContractAddress",
  messageServiceAddress,
  "--quoteContractAddress",
  quoteAddress,
  "--rpcUrl",
  "http://127.0.0.1:8545",
  "--swapAmountSlippageBps",
  "50",
  "--swapDeadlineInSeconds",
  "300",
  "--poolTickSpacing",
  "50",
];

describe("burn-and-bridge", () => {
  const publicClient = {
    getTransactionCount: jest.fn<() => Promise<number>>(),
  };

  const walletClient = {};

  beforeEach(() => {
    jest.clearAllMocks();
    publicClient.getTransactionCount.mockResolvedValue(9);
    mockCreatePublicClient.mockReturnValue(publicClient);
    mockCreateWalletClient.mockReturnValue(walletClient);
    mockGetInvoiceArrears.mockResolvedValue(ok(parseEther("1")));
    mockGetBalance.mockResolvedValue(parseEther("5"));
    mockGetMinimumFee.mockResolvedValue(ok(parseEther("1")));
    mockGetQuote.mockResolvedValue(ok([parseUnits("10", 18)]));
    mockEstimateTransactionGas.mockResolvedValue(ok({ gasLimit: 50_000n, baseFeePerGas: 10n, priorityFeePerGas: 2n }));
    mockSendTransaction.mockResolvedValue(ok({ status: "success", transactionHash: "0xtx", logs: [] }));
    mockParseEventLogs.mockReturnValue([
      {
        args: {
          ethBurnt: parseEther("0.8"),
          lineaTokensBridged: parseUnits("3", 18),
        },
      },
    ]);
  });

  it("loads command help from the oclif registry", async () => {
    const { error, stdout, stderr } = await runOperationsCommand(["burn-and-bridge", "--help"]);

    expect(error).toBeUndefined();
    expect(stderr).toStrictEqual("");
    expect(stdout).toContain("USAGE");
    expect(stdout).toContain("$ operations burn-and-bridge");
  });

  it("does nothing when the vault balance is below the minimum fee", async () => {
    mockGetBalance.mockResolvedValue(parseEther("1"));

    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Vault balance is less than or equal to minimum fee. No action needed.");
    expect(mockEstimateTransactionGas).not.toHaveBeenCalled();
    expect(mockSendTransaction).not.toHaveBeenCalled();
  });

  it("computes burn-and-bridge calldata and skips broadcasting in dry-run mode", async () => {
    const { error, stdout } = await runOperationsCommand([...baseArgs, "--dryRun"]);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Burn and bridge will be performed.");
    expect(stdout).toContain("Minimum LINEA out (after slippage): minLineaOut=9.95 LINEA slippageBps=50");
    expect(stdout).toContain("Dry run mode - transaction not submitted.");
    expect(mockGetQuote).toHaveBeenCalledTimes(1);
    expect(mockComputeSwapCalldata).toHaveBeenCalledTimes(1);
    expect(mockSendTransaction).not.toHaveBeenCalled();
  });

  it("broadcasts a burn-and-bridge transaction and logs the emitted event", async () => {
    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Broadcasting transaction...");
    expect(stdout).toContain("Burn and bridge transaction successfully processed.");
    expect(stdout).toContain("transactionHash=0xtx");
    expect(mockSendTransaction).toHaveBeenCalledWith(
      walletClient,
      expect.objectContaining({
        data: "0xburnandbridge",
        gas: 50_000n,
        nonce: 9,
        to: vaultAddress,
      }),
    );
  });

  it("pays arrears without burn-and-bridge when arrears exceed the vault balance", async () => {
    mockGetInvoiceArrears.mockResolvedValue(ok(parseEther("6")));
    mockGetBalance.mockResolvedValue(parseEther("5"));
    mockParseEventLogs.mockReturnValue([
      {
        args: {
          amount: parseEther("5"),
          remainingArrears: parseEther("1"),
        },
      },
    ]);

    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("All funds will be used to pay arrears. No burn and bridge will be performed.");
    expect(stdout).toContain("successfully processed without burning");
    expect(mockGetQuote).not.toHaveBeenCalled();
    expect(mockComputeSwapCalldata).not.toHaveBeenCalled();
  });

  it("warns when a successful arrears-only transaction misses the expected event", async () => {
    mockGetInvoiceArrears.mockResolvedValue(ok(parseEther("6")));
    mockParseEventLogs.mockReturnValue([]);

    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("did not emit ArrearsPaid event as expected");
  });

  it("returns quote errors", async () => {
    mockGetQuote.mockResolvedValue(err(new Error("quote reverted")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to get quote from quote contract. message=quote reverted");
    expect(mockEstimateTransactionGas).not.toHaveBeenCalled();
  });

  it("returns gas estimation errors", async () => {
    mockEstimateTransactionGas.mockResolvedValue(err(new Error("gas rpc failed")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to estimate gas for burn and bridge transaction. message=gas rpc failed");
    expect(mockSendTransaction).not.toHaveBeenCalled();
  });

  it("returns send transaction errors", async () => {
    mockSendTransaction.mockResolvedValue(err(new Error("wallet rejected")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to send transaction. message=wallet rejected");
  });

  it("returns reverted transaction errors", async () => {
    mockSendTransaction.mockResolvedValue(ok({ status: "reverted", transactionHash: "0xreverted", logs: [] }));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Burn and bridge failed. transactionHash=0xreverted");
  });
});
