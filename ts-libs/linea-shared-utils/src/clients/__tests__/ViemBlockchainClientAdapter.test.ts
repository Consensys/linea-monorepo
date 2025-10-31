import {
  Address,
  Chain,
  ContractFunctionRevertedError,
  Hex,
  PublicClient,
  TimeoutError,
  createPublicClient,
  http,
  parseSignature,
  serializeTransaction,
  withTimeout,
} from "viem";
import { sendRawTransaction, waitForTransactionReceipt } from "viem/actions";
import { ViemBlockchainClientAdapter } from "../ViemBlockchainClientAdapter";
import { ILogger } from "../../logging/ILogger";
import { IContractSignerClient } from "../../core/client/IContractSignerClient";

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    http: jest.fn(() => "mock-transport"),
    createPublicClient: jest.fn(),
    withTimeout: jest.fn((fn: any) => fn({ signal: null })),
    serializeTransaction: jest.fn(),
    parseSignature: jest.fn(),
  };
});

jest.mock("viem/actions", () => ({
  sendRawTransaction: jest.fn(),
  waitForTransactionReceipt: jest.fn(),
}));

const mockedHttp = http as jest.MockedFunction<typeof http>;
const mockedCreatePublicClient = createPublicClient as jest.MockedFunction<typeof createPublicClient>;
const mockedWithTimeout = withTimeout as unknown as jest.MockedFunction<typeof withTimeout>;
const mockedSerializeTransaction = serializeTransaction as jest.MockedFunction<typeof serializeTransaction>;
const mockedParseSignature = parseSignature as jest.MockedFunction<typeof parseSignature>;
const mockedSendRawTransaction = sendRawTransaction as jest.MockedFunction<typeof sendRawTransaction>;
const mockedWaitForTransactionReceipt = waitForTransactionReceipt as jest.MockedFunction<
  typeof waitForTransactionReceipt
>;

const createLogger = (): jest.Mocked<ILogger> =>
  ({
    name: "viem-blockchain-client",
    info: jest.fn(),
    error: jest.fn(),
    warn: jest.fn(),
    debug: jest.fn(),
    warnOrError: jest.fn(),
  }) as jest.Mocked<ILogger>;

const createContractSignerClient = (): jest.Mocked<IContractSignerClient> =>
  ({
    getAddress: jest.fn().mockReturnValue("0xSIGNER"),
    sign: jest.fn(),
  }) as unknown as jest.Mocked<IContractSignerClient>;

const createPublicClientMock = () =>
  ({
    getChainId: jest.fn(),
    getBalance: jest.fn(),
    estimateFeesPerGas: jest.fn(),
    getTransactionCount: jest.fn(),
    estimateGas: jest.fn(),
  }) as unknown as jest.Mocked<PublicClient>;

describe("ViemBlockchainClientAdapter", () => {
  const rpcUrl = "https://rpc.local";
  const chain = { id: 11155111 } as Chain;
  const contractAddress = "0xCONTRACT" as Address;
  const calldata = "0xabcdef" as Hex;

  let logger: jest.Mocked<ILogger>;
  let contractSignerClient: jest.Mocked<IContractSignerClient>;
  let publicClientMock: jest.Mocked<PublicClient>;
  let adapter: ViemBlockchainClientAdapter;

  beforeEach(() => {
    jest.clearAllMocks();
    logger = createLogger();
    contractSignerClient = createContractSignerClient();
    publicClientMock = createPublicClientMock();
    mockedHttp.mockReturnValue("mock-transport" as any);
    mockedCreatePublicClient.mockReturnValue(publicClientMock as unknown as PublicClient);
    mockedWithTimeout.mockImplementation((fn: any, _opts?: any) => fn({ signal: null }));
    mockedParseSignature.mockReturnValue({ r: "0x1", s: "0x2", yParity: 1 } as any);
    mockedSerializeTransaction.mockReturnValue("0x02serialized");
    mockedSendRawTransaction.mockResolvedValue("0xHASH");
    mockedWaitForTransactionReceipt.mockResolvedValue({ transactionHash: "0xHASH", status: "success" } as any);

    adapter = new ViemBlockchainClientAdapter(logger, rpcUrl, chain, contractSignerClient, 3, 1000n, 300_000);
  });

  it("logs request and response bodies in viem transport hooks", async () => {
    expect(mockedHttp).toHaveBeenCalled();
    const transportConfig = mockedHttp.mock.calls[0]?.[1] as {
      onFetchRequest: (request: any) => Promise<void>;
      onFetchResponse: (response: any) => Promise<void>;
    };
    expect(transportConfig).toBeDefined();

    const requestBody = JSON.stringify({ foo: "bar" });
    const requestClone = { text: jest.fn().mockResolvedValue(requestBody) };
    const request = {
      method: "POST",
      url: "https://rpc.local",
      clone: jest.fn().mockReturnValue(requestClone),
    };

    await transportConfig.onFetchRequest(request);

    expect(logger.debug).toHaveBeenCalledWith("onFetchRequest", {
      method: "POST",
      url: "https://rpc.local",
      body: requestBody,
    });

    const failingRequest = {
      clone: jest.fn().mockReturnValue({
        text: jest.fn().mockRejectedValue(new Error("request-read-fail")),
      }),
    };

    await transportConfig.onFetchRequest(failingRequest);

    expect(logger.warn).toHaveBeenCalledWith("Failed to read request body", {
      err: expect.any(Error),
    });

    const responseBody = JSON.stringify({ ok: true });
    const responseClone = { text: jest.fn().mockResolvedValue(responseBody) };
    const response = {
      status: 200,
      statusText: "OK",
      clone: jest.fn().mockReturnValue(responseClone),
    };

    await transportConfig.onFetchResponse(response);

    expect(logger.debug).toHaveBeenCalledWith("onFetchResponse", {
      status: 200,
      statusText: "OK",
      body: responseBody,
    });

    const responseError = {
      clone: jest.fn().mockReturnValue({
        text: jest.fn().mockRejectedValue(new Error("read-fail")),
      }),
    };

    await transportConfig.onFetchResponse(responseError);

    expect(logger.warn).toHaveBeenCalledWith("Failed to read response body", {
      err: expect.any(Error),
    });
  });

  it("throws if sendTransactionsMaxRetries is less than 1", () => {
    expect(() => new ViemBlockchainClientAdapter(logger, rpcUrl, chain, contractSignerClient, 0, 1000n, 1_000)).toThrow(
      "sendTransactionsMaxRetries must be at least 1",
    );
  });

  it("exposes the underlying public client", () => {
    expect(adapter.getBlockchainClient()).toBe(publicClientMock);
  });

  it("delegates to public client for getChainId, getBalance, estimateGasFees", async () => {
    publicClientMock.getChainId.mockResolvedValue(5);
    publicClientMock.getBalance.mockResolvedValue(123n);
    publicClientMock.estimateFeesPerGas.mockResolvedValue({
      maxFeePerGas: 40n,
      maxPriorityFeePerGas: 3n,
    });

    await expect(adapter.getChainId()).resolves.toBe(5);
    await expect(adapter.getBalance("0xabc" as Address)).resolves.toBe(123n);
    await expect(adapter.estimateGasFees()).resolves.toEqual({
      maxFeePerGas: 40n,
      maxPriorityFeePerGas: 3n,
    });

    expect(publicClientMock.getChainId).toHaveBeenCalledTimes(1);
    expect(publicClientMock.getBalance).toHaveBeenCalledWith({ address: "0xabc" });
    expect(publicClientMock.estimateFeesPerGas).toHaveBeenCalledTimes(1);
  });

  it("uses default constructor parameters and default tx value", async () => {
    const defaultsAdapter = new ViemBlockchainClientAdapter(logger, rpcUrl, chain, contractSignerClient);

    const timeoutError = new TimeoutError({
      body: { message: "timeout" },
      url: "local:default",
    });

    publicClientMock.getTransactionCount.mockResolvedValue(4);
    publicClientMock.estimateGas.mockResolvedValue(200n);
    publicClientMock.getChainId.mockResolvedValue(chain.id);
    publicClientMock.estimateFeesPerGas
      .mockResolvedValueOnce({ maxFeePerGas: 10n, maxPriorityFeePerGas: 3n })
      .mockResolvedValueOnce({ maxFeePerGas: 8n, maxPriorityFeePerGas: 2n });
    contractSignerClient.sign.mockResolvedValue("0xSIGNATURE");

    mockedWithTimeout
      .mockImplementationOnce(async (fn: any, _opts?: any) => {
        await fn({ signal: null });
        throw timeoutError;
      })
      .mockImplementationOnce(async (fn: any, _opts?: any) => fn({ signal: null }));

    const receipt = await defaultsAdapter.sendSignedTransaction(contractAddress, calldata);

    expect(receipt).toEqual({ transactionHash: "0xHASH", status: "success" });
    expect(logger.warn).toHaveBeenCalledWith(
      "sendSignedTransaction retry attempt failed attempt=1 sendTransactionsMaxRetries=3",
      { error: timeoutError },
    );
    expect(publicClientMock.estimateFeesPerGas).toHaveBeenCalledTimes(2);
    expect(contractSignerClient.sign).toHaveBeenCalledTimes(2);
    expect(contractSignerClient.sign).toHaveBeenNthCalledWith(1, {
      to: contractAddress,
      type: "eip1559",
      data: calldata,
      chainId: chain.id,
      gas: 200n,
      maxFeePerGas: 10n,
      maxPriorityFeePerGas: 3n,
      nonce: 4,
      value: 0n,
    });
    expect(contractSignerClient.sign).toHaveBeenNthCalledWith(2, {
      to: contractAddress,
      type: "eip1559",
      data: calldata,
      chainId: chain.id,
      gas: 220n,
      maxFeePerGas: 11n,
      maxPriorityFeePerGas: 3n,
      nonce: 4,
      value: 0n,
    });
  });

  it("successfully sends a signed transaction on the first attempt", async () => {
    publicClientMock.getTransactionCount.mockResolvedValue(7);
    publicClientMock.estimateGas.mockResolvedValue(21_000n);
    publicClientMock.estimateFeesPerGas.mockResolvedValue({
      maxFeePerGas: 100n,
      maxPriorityFeePerGas: 2n,
    });
    publicClientMock.getChainId.mockResolvedValue(chain.id);
    contractSignerClient.sign.mockResolvedValue("0xSIGNATURE");

    const receipt = await adapter.sendSignedTransaction(contractAddress, calldata, 10n);

    expect(receipt).toEqual({ transactionHash: "0xHASH", status: "success" });
    expect(publicClientMock.estimateFeesPerGas).toHaveBeenCalledTimes(1);
    expect(contractSignerClient.sign).toHaveBeenCalledWith({
      to: contractAddress,
      type: "eip1559",
      data: calldata,
      chainId: chain.id,
      gas: 21_000n,
      maxFeePerGas: 100n,
      maxPriorityFeePerGas: 2n,
      nonce: 7,
      value: 10n,
    });
    expect(mockedParseSignature).toHaveBeenCalledWith("0xSIGNATURE");
    expect(mockedSerializeTransaction).toHaveBeenCalledWith(
      {
        to: contractAddress,
        type: "eip1559",
        data: calldata,
        chainId: chain.id,
        gas: 21_000n,
        maxFeePerGas: 100n,
        maxPriorityFeePerGas: 2n,
        nonce: 7,
        value: 10n,
      },
      { r: "0x1", s: "0x2", yParity: 1 },
    );
    expect(mockedSendRawTransaction).toHaveBeenCalledWith(publicClientMock, {
      serializedTransaction: "0x02serialized",
    });
    expect(mockedWaitForTransactionReceipt).toHaveBeenCalledWith(publicClientMock, { hash: "0xHASH" });
  });

  it("retries on timeout and applies gas bump multipliers", async () => {
    adapter = new ViemBlockchainClientAdapter(logger, rpcUrl, chain, contractSignerClient, 3, 1_000n, 300_000);

    const timeoutError = new TimeoutError({
      body: { message: "timeout" },
      url: "local:test",
    });

    publicClientMock.getTransactionCount.mockResolvedValue(5);
    publicClientMock.estimateGas.mockResolvedValue(200n);
    publicClientMock.getChainId.mockResolvedValue(chain.id);
    publicClientMock.estimateFeesPerGas
      .mockResolvedValueOnce({ maxFeePerGas: 10n, maxPriorityFeePerGas: 3n })
      .mockResolvedValueOnce({ maxFeePerGas: 8n, maxPriorityFeePerGas: 2n });
    contractSignerClient.getAddress.mockReturnValue("0xSIGNER");
    contractSignerClient.sign.mockResolvedValue("0xSIGNATURE");

    mockedWithTimeout
      .mockImplementationOnce(async (fn: any, _opts?: any) => {
        await fn({ signal: null });
        throw timeoutError;
      })
      .mockImplementationOnce(async (fn: any, _opts?: any) => fn({ signal: null }));

    const receipt = await adapter.sendSignedTransaction(contractAddress, calldata, 0n);

    expect(receipt).toEqual({ transactionHash: "0xHASH", status: "success" });
    expect(publicClientMock.estimateFeesPerGas).toHaveBeenCalledTimes(2);
    expect(logger.warn).toHaveBeenCalledWith(
      "sendSignedTransaction retry attempt failed attempt=1 sendTransactionsMaxRetries=3",
      { error: timeoutError },
    );
    expect(contractSignerClient.sign).toHaveBeenCalledTimes(2);
    expect(contractSignerClient.sign).toHaveBeenNthCalledWith(1, {
      to: contractAddress,
      type: "eip1559",
      data: calldata,
      chainId: chain.id,
      gas: 200n,
      maxFeePerGas: 10n,
      maxPriorityFeePerGas: 3n,
      nonce: 5,
      value: 0n,
    });
    expect(contractSignerClient.sign).toHaveBeenNthCalledWith(2, {
      to: contractAddress,
      type: "eip1559",
      data: calldata,
      chainId: chain.id,
      gas: 220n,
      maxFeePerGas: 11n,
      maxPriorityFeePerGas: 3n,
      nonce: 5,
      value: 0n,
    });
  });

  it("rethrows non-timeout errors without retrying", async () => {
    publicClientMock.getTransactionCount.mockResolvedValue(1);
    publicClientMock.estimateGas.mockResolvedValue(100n);
    publicClientMock.getChainId.mockResolvedValue(chain.id);
    publicClientMock.estimateFeesPerGas.mockResolvedValue({
      maxFeePerGas: 9n,
      maxPriorityFeePerGas: 1n,
    });
    contractSignerClient.sign.mockResolvedValue("0xSIGNATURE");

    const fatalError = new Error("boom");
    mockedWithTimeout.mockImplementationOnce(async (fn: any, _opts?: any) => {
      await fn({ signal: null });
      throw fatalError;
    });

    await expect(adapter.sendSignedTransaction(contractAddress, calldata, 0n)).rejects.toThrow(fatalError);
    expect(publicClientMock.estimateFeesPerGas).toHaveBeenCalledTimes(1);
    expect(logger.error).toHaveBeenCalledWith("sendSignedTransaction error", { error: fatalError });
    expect(mockedWithTimeout).toHaveBeenCalledTimes(1);
  });

  it("logs and rethrows ContractFunctionRevertedError without retrying", async () => {
    publicClientMock.getTransactionCount.mockResolvedValue(1);
    publicClientMock.estimateGas.mockResolvedValue(50n);
    publicClientMock.getChainId.mockResolvedValue(chain.id);
    publicClientMock.estimateFeesPerGas.mockResolvedValue({
      maxFeePerGas: 5n,
      maxPriorityFeePerGas: 1n,
    });
    contractSignerClient.sign.mockResolvedValue("0xSIGNATURE");

    const revertError = new ContractFunctionRevertedError({
      abi: [] as any,
      functionName: "test",
      message: "execution reverted",
    });
    (revertError as any).data = { errorName: "RevertReason" };

    mockedWithTimeout.mockImplementationOnce(async (fn: any, _opts?: any) => {
      await fn({ signal: null });
      throw revertError;
    });

    await expect(adapter.sendSignedTransaction(contractAddress, calldata, 0n)).rejects.toBe(revertError);

    expect(logger.error).toHaveBeenNthCalledWith(1, "❌ sendSignedTransaction contract call reverted:", {
      shortMessage: revertError.shortMessage,
    });
    expect(logger.error).toHaveBeenNthCalledWith(2, "Reason:", { reason: "RevertReason" });
    expect(mockedWithTimeout).toHaveBeenCalledTimes(1);
  });

  it("falls back to ContractFunctionRevertedError revert message when error name missing", async () => {
    publicClientMock.getTransactionCount.mockResolvedValue(1);
    publicClientMock.estimateGas.mockResolvedValue(50n);
    publicClientMock.getChainId.mockResolvedValue(chain.id);
    publicClientMock.estimateFeesPerGas.mockResolvedValue({
      maxFeePerGas: 5n,
      maxPriorityFeePerGas: 1n,
    });
    contractSignerClient.sign.mockResolvedValue("0xSIGNATURE");

    const revertError = new ContractFunctionRevertedError({
      abi: [] as any,
      functionName: "test",
      message: "execution reverted",
    });
    (revertError as any).data = undefined;

    mockedWithTimeout.mockImplementationOnce(async (fn: any, _opts?: any) => {
      await fn({ signal: null });
      throw revertError;
    });

    await expect(adapter.sendSignedTransaction(contractAddress, calldata, 0n)).rejects.toBe(revertError);
    expect(logger.error).toHaveBeenNthCalledWith(1, "❌ sendSignedTransaction contract call reverted:", {
      shortMessage: revertError.shortMessage,
    });
    expect(logger.error).toHaveBeenNthCalledWith(2, "Reason:", { reason: revertError.message });
    expect(mockedWithTimeout).toHaveBeenCalledTimes(1);
  });

  it("throws after exhausting timeout retries", async () => {
    adapter = new ViemBlockchainClientAdapter(logger, rpcUrl, chain, contractSignerClient, 2, 1_000n, 1_000);

    const timeoutError = new TimeoutError({
      body: { message: "timeout" },
      url: "local:test",
    });

    publicClientMock.getTransactionCount.mockResolvedValue(1);
    publicClientMock.estimateGas.mockResolvedValue(100n);
    publicClientMock.getChainId.mockResolvedValue(chain.id);
    publicClientMock.estimateFeesPerGas.mockResolvedValue({
      maxFeePerGas: 9n,
      maxPriorityFeePerGas: 1n,
    });

    mockedWithTimeout.mockImplementation(async (fn: any, _opts?: any) => {
      await fn({ signal: null });
      throw timeoutError;
    });

    await expect(adapter.sendSignedTransaction(contractAddress, calldata, 0n)).rejects.toBe(timeoutError);

    expect(logger.warn).toHaveBeenCalledTimes(1);
    expect(logger.error).toHaveBeenCalledWith(
      "sendSignedTransaction retry attempts exhausted sendTransactionsMaxRetries=2",
      { error: timeoutError },
    );
    expect(mockedWithTimeout).toHaveBeenCalledTimes(2);
    expect(contractSignerClient.sign).toHaveBeenCalledTimes(2);
  });
});
