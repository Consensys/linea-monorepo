import { afterEach, beforeEach, describe, expect, it, jest } from "@jest/globals";
import { err, ok } from "neverthrow";

import { runOperationsCommand } from "../helpers/run-command";

type UnknownFn = (...args: unknown[]) => unknown;
type AsyncUnknownFn = (...args: unknown[]) => Promise<unknown>;

const mockBuildHttpsAgent = jest.fn<UnknownFn>(() => ({ agent: "https" }));
const mockComputeSubmitInvoiceCalldata = jest.fn<UnknownFn>(() => "0xsubmitinvoice");
const mockCreateAwsCostExplorerClient = jest.fn<UnknownFn>(() => ({ aws: "client" }));
const mockCreatePublicClient = jest.fn<UnknownFn>();
const mockEstimateTransactionGas = jest.fn<AsyncUnknownFn>();
const mockFetchEthereumPrice = jest.fn<AsyncUnknownFn>();
const mockFlattenResultsByTime = jest.fn(
  (
    results: Array<{
      Estimated?: boolean;
      TimePeriod?: { Start?: string };
      Total?: Record<string, { Amount?: string }>;
    }>,
    metric: string,
  ) =>
    results.map((result) => ({
      amount: result.Total?.[metric]?.Amount,
      date: result.TimePeriod?.Start,
      estimated: result.Estimated,
    })),
);
const mockGenerateQueryParameters = jest.fn((params: unknown) => params);
const mockGetBlock = jest.fn<AsyncUnknownFn>();
const mockGetDailyAwsCosts = jest.fn<AsyncUnknownFn>();
const mockGetDuneClient = jest.fn<UnknownFn>(() => ({ dune: "client" }));
const mockGetLastInvoiceDate = jest.fn<AsyncUnknownFn>();
const mockGetWeb3SignerSignature = jest.fn<AsyncUnknownFn>();
const mockParseEventLogs = jest.fn<UnknownFn>(() => []);
const mockParseSignature = jest.fn<UnknownFn>(() => ({ r: "0x1", s: "0x2", v: 27n }));
const mockRunDuneQuery = jest.fn<AsyncUnknownFn>();
const mockSendRawTransaction = jest.fn<AsyncUnknownFn>();
const mockSerializeTransaction = jest.fn<UnknownFn>(() => "0xsigned");
const { parseEther } = jest.requireActual<typeof import("viem")>("viem");

jest.mock("date-fns-tz", () => jest.requireActual("../mocks/date-fns-tz"));
jest.mock("viem", () => {
  const actual = jest.requireActual<typeof import("viem")>("viem");
  return {
    ...actual,
    createPublicClient: mockCreatePublicClient,
    http: jest.fn((url: string, options?: unknown) => {
      void options;
      return { url };
    }),
    parseEventLogs: mockParseEventLogs,
    parseSignature: mockParseSignature,
    serializeTransaction: mockSerializeTransaction,
  };
});
jest.mock("viem/actions", () => {
  const actual = jest.requireActual<typeof import("viem/actions")>("viem/actions");
  return {
    ...actual,
    getBlock: mockGetBlock,
  };
});
jest.mock("../../dist/utils/common/aws.js", () => ({
  createAwsCostExplorerClient: mockCreateAwsCostExplorerClient,
  flattenResultsByTime: mockFlattenResultsByTime,
  getDailyAwsCosts: mockGetDailyAwsCosts,
}));
jest.mock("../../dist/utils/common/coingecko.js", () => ({
  fetchEthereumPrice: mockFetchEthereumPrice,
}));
jest.mock("../../dist/utils/common/dune.js", () => ({
  generateQueryParameters: mockGenerateQueryParameters,
  getDuneClient: mockGetDuneClient,
  runDuneQuery: mockRunDuneQuery,
}));
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
jest.mock("../../dist/utils/submit-invoice/contract.js", () => ({
  computeSubmitInvoiceCalldata: mockComputeSubmitInvoiceCalldata,
  getLastInvoiceDate: mockGetLastInvoiceDate,
}));

const senderAddress = "0x0000000000000000000000000000000000000001";
const contractAddress = "0x0000000000000000000000000000000000000002";
const awsFilters = JSON.stringify({ Granularity: "DAILY", Metrics: ["AmortizedCost"], GroupBy: [] });

const baseArgs = [
  "submit-invoice",
  "--senderAddress",
  senderAddress,
  "--contractAddress",
  contractAddress,
  "--periodDays",
  "2",
  "--reportingLagDays",
  "1",
  "--rpcUrl",
  "http://127.0.0.1:8545",
  "--web3SignerUrl",
  "http://127.0.0.1:8546",
  "--web3SignerPublicKey",
  "0x1234",
  "--duneApiKey",
  "dune-key",
  "--duneQueryId",
  "42",
  "--awsCostsApiFilters",
  awsFilters,
  "--coingeckoApiBaseUrl",
  "https://api.coingecko.test",
  "--coingeckoApiKey",
  "coingecko-key",
];

const awsCostResponse = (amount: string, estimated = false) =>
  ok({
    ResultsByTime: [
      {
        Estimated: estimated,
        TimePeriod: { Start: "2024-01-02" },
        Total: { AmortizedCost: { Amount: amount } },
      },
    ],
  });

describe("submit-invoice", () => {
  const publicClient = {
    getTransactionCount: jest.fn<() => Promise<number>>(),
  };

  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers().setSystemTime(new Date("2024-01-20T12:00:00.000Z"));
    publicClient.getTransactionCount.mockResolvedValue(4);
    mockCreatePublicClient.mockReturnValue(publicClient);
    mockGetLastInvoiceDate.mockResolvedValue(ok(1_704_067_200n));
    mockGetDailyAwsCosts.mockResolvedValueOnce(awsCostResponse("12")).mockResolvedValueOnce(awsCostResponse("100"));
    mockRunDuneQuery.mockResolvedValue(ok({ result: { rows: [{ total_costs_per_day: 0.5 }] } }));
    mockFetchEthereumPrice.mockResolvedValue(ok({ ethereum: { usd: 2_000 } }));
    mockEstimateTransactionGas.mockResolvedValue(ok({ gasLimit: 80_000n, baseFeePerGas: 10n, priorityFeePerGas: 2n }));
    mockGetWeb3SignerSignature.mockResolvedValue(ok("0xsignature"));
    mockSendRawTransaction.mockResolvedValue(
      ok({ status: "success", transactionHash: "0xtx", blockNumber: 10n, logs: [] }),
    );
    mockGetBlock.mockResolvedValue({ timestamp: 1_705_755_610n });
    mockParseEventLogs.mockReturnValue([
      {
        eventName: "InvoiceProcessed",
        args: {
          receiver: senderAddress,
          startTimestamp: 1_704_067_201n,
          endTimestamp: 1_704_240_000n,
          amountPaid: parseEther("0.55"),
          amountRequested: parseEther("0.55"),
        },
      },
    ]);
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it("loads command help from the oclif registry", async () => {
    const { error, stdout, stderr } = await runOperationsCommand(["submit-invoice", "--help"]);

    expect(error).toBeUndefined();
    expect(stderr).toStrictEqual("");
    expect(stdout).toContain("USAGE");
    expect(stdout).toContain("$ operations submit-invoice");
  });

  it("does nothing when no invoice period is ready", async () => {
    mockGetLastInvoiceDate.mockResolvedValue(ok(1_705_579_200n));

    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("No invoice to process at this time.");
    expect(mockGetDailyAwsCosts).not.toHaveBeenCalled();
    expect(mockEstimateTransactionGas).not.toHaveBeenCalled();
  });

  it("computes invoice costs and skips broadcasting in dry-run mode", async () => {
    const { error, stdout } = await runOperationsCommand([...baseArgs, "--dryRun"]);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Invoice period to process:");
    expect(stdout).toContain("Total AWS costs costsInUsd=100");
    expect(stdout).toContain("Total on-chain costs costsInEth=0.5");
    expect(stdout).toContain("Total costs to invoice: costsInEth=0.55 etherPriceInUsd=2000");
    expect(stdout).toContain("Dry run mode - transaction not submitted.");
    expect(mockGetWeb3SignerSignature).toHaveBeenCalledTimes(1);
    expect(mockSendRawTransaction).not.toHaveBeenCalled();
  });

  it("broadcasts a submitted invoice and logs the processed event", async () => {
    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("Broadcasting submitInvoice transaction to the network...");
    expect(stdout).toContain("Invoice successfully submitted: transactionHash=0xtx");
    expect(mockSendRawTransaction).toHaveBeenCalledTimes(1);
    expect(mockGetBlock).toHaveBeenCalledWith(publicClient, { blockNumber: 10n });
  });

  it("stops when AWS costs are still estimated", async () => {
    mockGetDailyAwsCosts.mockReset();
    mockGetDailyAwsCosts
      .mockResolvedValueOnce(awsCostResponse("12"))
      .mockResolvedValueOnce(awsCostResponse("100", true));

    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("AWS costs are undefined, likely due to data still being estimated.");
    expect(mockRunDuneQuery).not.toHaveBeenCalled();
  });

  it("returns AWS filter validation errors", async () => {
    const { error } = await runOperationsCommand([
      ...baseArgs.slice(0, baseArgs.indexOf("--awsCostsApiFilters") + 1),
      JSON.stringify({ Granularity: "DAILY", Metrics: [], GroupBy: [] }),
      ...baseArgs.slice(baseArgs.indexOf("--coingeckoApiBaseUrl")),
    ]);

    expect(error?.message).toContain("AWS Costs API Filters must specify one metric.");
  });

  it("returns AWS fetch errors", async () => {
    mockGetDailyAwsCosts.mockReset();
    mockGetDailyAwsCosts.mockResolvedValue(err(new Error("aws unavailable")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to fetch AWS costs historical data. message=aws unavailable");
  });

  it("returns Dune errors when the query has no rows", async () => {
    mockRunDuneQuery.mockResolvedValue(ok({ result: { rows: [] } }));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("No Dune query result returned for the specified period.");
  });

  it("returns CoinGecko missing-price errors", async () => {
    mockFetchEthereumPrice.mockResolvedValue(ok({ ethereum: {} }));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Ethereum price data is missing in the CoinGecko response.");
  });

  it("returns CoinGecko zero-price errors", async () => {
    mockFetchEthereumPrice.mockResolvedValue(ok({ ethereum: { usd: 0 } }));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Ethereum price data is missing in the CoinGecko response.");
  });

  it("does nothing when computed costs are zero", async () => {
    mockGetDailyAwsCosts.mockReset();
    mockGetDailyAwsCosts.mockResolvedValueOnce(awsCostResponse("0")).mockResolvedValueOnce(awsCostResponse("0"));
    mockRunDuneQuery.mockResolvedValue(ok({ result: { rows: [{ total_costs_per_day: 0 }] } }));

    const { error, stdout } = await runOperationsCommand(baseArgs);

    expect(error).toBeUndefined();
    expect(stdout).toContain("No costs to process at this time.");
    expect(mockEstimateTransactionGas).not.toHaveBeenCalled();
  });

  it("returns gas estimation errors", async () => {
    mockEstimateTransactionGas.mockResolvedValue(err(new Error("gas failed")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to estimate gas for submitInvoice transaction. message=gas failed");
    expect(mockGetWeb3SignerSignature).not.toHaveBeenCalled();
  });

  it("returns Web3 Signer errors", async () => {
    mockGetWeb3SignerSignature.mockResolvedValue(err(new Error("signer failed")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to get signature from Web3 Signer. message=signer failed");
    expect(mockSendRawTransaction).not.toHaveBeenCalled();
  });

  it("returns send transaction errors", async () => {
    mockSendRawTransaction.mockResolvedValue(err(new Error("send failed")));

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Failed to send transaction. message=send failed");
  });

  it("returns reverted invoice transaction errors", async () => {
    mockSendRawTransaction.mockResolvedValue(
      ok({ status: "reverted", transactionHash: "0xreverted", blockNumber: 10n, logs: [] }),
    );

    const { error } = await runOperationsCommand(baseArgs);

    expect(error?.message).toContain("Invoice submission failed. transactionHash=0xreverted");
  });
});
