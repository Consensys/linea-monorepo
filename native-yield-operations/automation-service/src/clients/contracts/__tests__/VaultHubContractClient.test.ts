import { mock, MockProxy } from "jest-mock-extended";
import type { IBlockchainClient, ILogger } from "@consensys/linea-shared-utils";
import type { PublicClient, TransactionReceipt, Address } from "viem";
import { VaultHubABI } from "../../../core/abis/VaultHub.js";

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    getContract: jest.fn(),
    parseEventLogs: jest.fn(),
  };
});

import { getContract, parseEventLogs } from "viem";

const mockedGetContract = getContract as jest.MockedFunction<typeof getContract>;
const mockedParseEventLogs = parseEventLogs as jest.MockedFunction<typeof parseEventLogs>;

let VaultHubContractClient: typeof import("../VaultHubContractClient.js").VaultHubContractClient;

beforeAll(async () => {
  ({ VaultHubContractClient } = await import("../VaultHubContractClient.js"));
});

describe("VaultHubContractClient", () => {
  const contractAddress = "0x1111111111111111111111111111111111111111" as Address;

  let blockchainClient: MockProxy<IBlockchainClient<PublicClient, TransactionReceipt>>;
  let logger: MockProxy<ILogger>;
  let publicClient: PublicClient;
  const viemContractStub = { abi: VaultHubABI, read: { settleableLidoFeesValue: jest.fn() } } as any;

  beforeEach(() => {
    jest.clearAllMocks();
    blockchainClient = mock<IBlockchainClient<PublicClient, TransactionReceipt>>();
    logger = mock<ILogger>();
    publicClient = {} as PublicClient;
    blockchainClient.getBlockchainClient.mockReturnValue(publicClient);
    mockedGetContract.mockReturnValue(viemContractStub);
  });

  const createClient = () => new VaultHubContractClient(blockchainClient, contractAddress, logger);

  const buildReceipt = (logs: Array<{ address: string; data: string; topics: string[] }>): TransactionReceipt =>
    ({
      logs,
    }) as unknown as TransactionReceipt;

  it("initializes the viem contract and exposes it through getters", () => {
    const client = createClient();

    expect(mockedGetContract).toHaveBeenCalledWith({
      abi: VaultHubABI,
      address: contractAddress,
      client: publicClient,
    });
    expect(client.getAddress()).toBe(contractAddress);
    expect(client.getContract()).toBe(viemContractStub);
  });

  it("returns liability payment when VaultRebalanced event is present", () => {
    const client = createClient();
    const receipt = buildReceipt([
      {
        address: "0x0000000000000000000000000000000000000000",
        data: "0x",
        topics: [],
      },
      {
        address: contractAddress,
        data: "0xdata",
        topics: ["0xtopic"],
      },
    ]);

    mockedParseEventLogs.mockReturnValueOnce([
      {
        eventName: "VaultRebalanced",
        args: { etherWithdrawn: 123n },
        address: contractAddress,
      } as any,
    ]);

    const amount = client.getLiabilityPaymentFromTxReceipt(receipt);

    expect(amount).toBe(123n);
    expect(mockedParseEventLogs).toHaveBeenCalledWith({
      abi: viemContractStub.abi,
      eventName: "VaultRebalanced",
      logs: receipt.logs,
    });
  });

  it("returns zero when VaultRebalanced event is present but etherWithdrawn is undefined", () => {
    const client = createClient();
    const receipt = buildReceipt([
      {
        address: contractAddress,
        data: "0xdata",
        topics: ["0xtopic"],
      },
    ]);

    mockedParseEventLogs.mockReturnValueOnce([
      {
        eventName: "VaultRebalanced",
        args: { etherWithdrawn: undefined },
        address: contractAddress,
      } as any,
    ]);

    const amount = client.getLiabilityPaymentFromTxReceipt(receipt);

    expect(amount).toBe(0n);
    expect(mockedParseEventLogs).toHaveBeenCalledTimes(1);
    expect(logger.warn).not.toHaveBeenCalled();
  });

  it("ignores logs that fail to decode and returns zero when no VaultRebalanced event", () => {
    const client = createClient();
    const receipt = buildReceipt([
      {
        address: contractAddress.toUpperCase(),
        data: "0xdead",
        topics: [],
      },
    ]);

    mockedParseEventLogs.mockReturnValueOnce([]);

    const amount = client.getLiabilityPaymentFromTxReceipt(receipt);

    expect(amount).toBe(0n);
    expect(mockedParseEventLogs).toHaveBeenCalledTimes(1);
    expect(logger.warn).toHaveBeenCalledWith(
      "getLiabilityPaymentFromTxReceipt - VaultRebalanced event not found in receipt",
    );
  });

  it("returns lido fee payment when LidoFeesSettled event is present", () => {
    const client = createClient();
    const receipt = buildReceipt([
      {
        address: contractAddress,
        data: "0xfeed",
        topics: ["0x01"],
      },
    ]);

    mockedParseEventLogs.mockReturnValueOnce([
      {
        eventName: "LidoFeesSettled",
        args: { transferred: 456n },
        address: contractAddress,
      } as any,
    ]);

    const amount = client.getLidoFeePaymentFromTxReceipt(receipt);

    expect(amount).toBe(456n);
    expect(mockedParseEventLogs).toHaveBeenCalledWith({
      abi: viemContractStub.abi,
      eventName: "LidoFeesSettled",
      logs: receipt.logs,
    });
  });

  it("returns zero when LidoFeesSettled event is present but transferred is undefined", () => {
    const client = createClient();
    const receipt = buildReceipt([
      {
        address: contractAddress,
        data: "0xfeed",
        topics: ["0x01"],
      },
    ]);

    mockedParseEventLogs.mockReturnValueOnce([
      {
        eventName: "LidoFeesSettled",
        args: { transferred: undefined },
        address: contractAddress,
      } as any,
    ]);

    const amount = client.getLidoFeePaymentFromTxReceipt(receipt);

    expect(amount).toBe(0n);
    expect(mockedParseEventLogs).toHaveBeenCalledTimes(1);
    expect(logger.warn).not.toHaveBeenCalled();
  });

  it("returns zero lido fee when logs belong to other contracts or events", () => {
    const client = createClient();
    const receipt = buildReceipt([
      {
        address: "0x2222222222222222222222222222222222222222",
        data: "0xaaa",
        topics: [],
      },
      {
        address: contractAddress,
        data: "0xbb",
        topics: [],
      },
    ]);

    mockedParseEventLogs.mockReturnValueOnce([
      {
        eventName: "LidoFeesSettled",
        args: { transferred: 456n },
        address: "0x2222222222222222222222222222222222222222",
      } as any,
    ]);

    const amount = client.getLidoFeePaymentFromTxReceipt(receipt);

    expect(amount).toBe(0n);
    expect(mockedParseEventLogs).toHaveBeenCalledTimes(1);
    expect(logger.warn).toHaveBeenCalledWith(
      "getLidoFeePaymentFromTxReceipt - LidoFeesSettled event not found in receipt",
    );
  });

  describe("settleableLidoFeesValue", () => {
    const vaultAddress = "0x2222222222222222222222222222222222222222" as Address;

    it("returns settleable Lido fees value when call succeeds", async () => {
      const client = createClient();
      const expectedValue = 1000000000000000000n;
      viemContractStub.read.settleableLidoFeesValue.mockResolvedValueOnce(expectedValue);

      const result = await client.settleableLidoFeesValue(vaultAddress);

      expect(result).toBe(expectedValue);
      expect(viemContractStub.read.settleableLidoFeesValue).toHaveBeenCalledWith([vaultAddress]);
    });

    it("returns zero when value is null", async () => {
      const client = createClient();
      viemContractStub.read.settleableLidoFeesValue.mockResolvedValueOnce(null);

      const result = await client.settleableLidoFeesValue(vaultAddress);

      expect(result).toBe(0n);
      expect(viemContractStub.read.settleableLidoFeesValue).toHaveBeenCalledTimes(1);
    });

    it("returns undefined and logs error when call fails", async () => {
      const client = createClient();
      const error = new Error("Contract call failed");
      viemContractStub.read.settleableLidoFeesValue.mockRejectedValueOnce(error);

      const result = await client.settleableLidoFeesValue(vaultAddress);

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith(`settleableLidoFeesValue failed, error=${error}`);
      expect(viemContractStub.read.settleableLidoFeesValue).toHaveBeenCalledTimes(1);
    });
  });
});
