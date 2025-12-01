import { mock, MockProxy } from "jest-mock-extended";
import type { IBlockchainClient, ILogger } from "@consensys/linea-shared-utils";
import type { PublicClient, TransactionReceipt, Address } from "viem";
import { DashboardABI } from "../../../core/abis/Dashboard.js";

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    getContract: jest.fn(),
    parseEventLogs: jest.fn(),
    encodeFunctionData: jest.fn(),
  };
});

import { encodeFunctionData, getContract, parseEventLogs } from "viem";

const mockedGetContract = getContract as jest.MockedFunction<typeof getContract>;
const mockedParseEventLogs = parseEventLogs as jest.MockedFunction<typeof parseEventLogs>;
const mockedEncodeFunctionData = encodeFunctionData as jest.MockedFunction<typeof encodeFunctionData>;

let DashboardContractClient: typeof import("../DashboardContractClient.js").DashboardContractClient;

beforeAll(async () => {
  ({ DashboardContractClient } = await import("../DashboardContractClient.js"));
});

describe("DashboardContractClient", () => {
  const contractAddress = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" as Address;

  let blockchainClient: MockProxy<IBlockchainClient<PublicClient, TransactionReceipt>>;
  let logger: MockProxy<ILogger>;
  let publicClient: PublicClient;
  const viemContractStub = {
    abi: DashboardABI,
    read: {
      obligations: jest.fn(),
      withdrawableValue: jest.fn(),
      totalValue: jest.fn(),
    },
  } as any;

  beforeEach(() => {
    jest.clearAllMocks();
    blockchainClient = mock<IBlockchainClient<PublicClient, TransactionReceipt>>();
    logger = mock<ILogger>();
    publicClient = {} as PublicClient;
    blockchainClient.getBlockchainClient.mockReturnValue(publicClient);
    mockedGetContract.mockReturnValue(viemContractStub);
    // Clear static state before each test
    (DashboardContractClient as any).blockchainClient = undefined;
    (DashboardContractClient as any).logger = undefined;
    (DashboardContractClient as any).clientCache.clear();
  });

  const createClient = () => {
    // Initialize logger if not already initialized (needed for constructor)
    if (!(DashboardContractClient as any).logger) {
      DashboardContractClient.initialize(blockchainClient, logger);
    }
    return new DashboardContractClient(contractAddress);
  };

  const buildReceipt = (logs: Array<{ address: string; data: string; topics: string[] }>): TransactionReceipt =>
    ({
      logs,
    }) as unknown as TransactionReceipt;

  it("initializes the viem contract and exposes it through getters", () => {
    const client = createClient();

    expect(mockedGetContract).toHaveBeenCalledWith({
      abi: DashboardABI,
      address: contractAddress,
      client: publicClient,
    });
    expect(client.getAddress()).toBe(contractAddress);
    expect(client.getContract()).toBe(viemContractStub);
  });

  it("returns node operator fees when FeeDisbursed event is present", () => {
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
        eventName: "FeeDisbursed",
        args: { fee: 123n },
        address: contractAddress,
      } as any,
    ]);

    const fee = client.getNodeOperatorFeesPaidFromTxReceipt(receipt);

    expect(fee).toBe(123n);
    expect(mockedParseEventLogs).toHaveBeenCalledWith({
      abi: viemContractStub.abi,
      eventName: "FeeDisbursed",
      logs: receipt.logs,
    });
  });

  it("ignores logs that fail to decode and returns zero when no FeeDisbursed event", () => {
    const client = createClient();
    const receipt = buildReceipt([
      {
        address: contractAddress.toUpperCase(),
        data: "0xdead",
        topics: [],
      },
    ]);

    mockedParseEventLogs.mockReturnValueOnce([]);

    const fee = client.getNodeOperatorFeesPaidFromTxReceipt(receipt);

    expect(fee).toBe(0n);
    expect(mockedParseEventLogs).toHaveBeenCalledTimes(1);
  });

  it("returns zero when logs belong to other contracts or events", () => {
    const client = createClient();
    const receipt = buildReceipt([
      {
        address: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
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
        eventName: "FeeDisbursed",
        args: { fee: 456n },
        address: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
      } as any,
    ]);

    const fee = client.getNodeOperatorFeesPaidFromTxReceipt(receipt);

    expect(fee).toBe(0n);
    expect(mockedParseEventLogs).toHaveBeenCalledTimes(1);
  });

  it("reads withdrawableValue via read and returns the result", async () => {
    const withdrawable = 1000n;
    viemContractStub.read.withdrawableValue.mockResolvedValueOnce(withdrawable);

    const client = createClient();
    const result = await client.withdrawableValue();

    expect(viemContractStub.read.withdrawableValue).toHaveBeenCalledTimes(1);
    expect(result).toBe(withdrawable);
  });

  it("reads totalValue via read and returns the result", async () => {
    const total = 5000n;
    viemContractStub.read.totalValue.mockResolvedValueOnce(total);

    const client = createClient();
    const result = await client.totalValue();

    expect(viemContractStub.read.totalValue).toHaveBeenCalledTimes(1);
    expect(result).toBe(total);
  });

  describe("initialize", () => {
    it("sets the static blockchainClient and logger", () => {
      DashboardContractClient.initialize(blockchainClient, logger);

      expect((DashboardContractClient as any).blockchainClient).toBe(blockchainClient);
      expect((DashboardContractClient as any).logger).toBe(logger);
    });
  });

  describe("getOrCreate", () => {
    it("throws error when blockchainClient is not initialized", () => {
      expect(() => {
        DashboardContractClient.getOrCreate(contractAddress);
      }).toThrow(
        "DashboardContractClient: blockchainClient must be initialized via DashboardContractClient.initialize() before use",
      );
    });

    it("throws error when logger is not initialized", () => {
      (DashboardContractClient as any).blockchainClient = blockchainClient;
      (DashboardContractClient as any).logger = undefined;

      expect(() => {
        DashboardContractClient.getOrCreate(contractAddress);
      }).toThrow(
        "DashboardContractClient: logger must be initialized via DashboardContractClient.initialize() before use",
      );
    });

    it("creates and caches a new client when not cached", () => {
      DashboardContractClient.initialize(blockchainClient, logger);

      const client = DashboardContractClient.getOrCreate(contractAddress);

      expect(client).toBeInstanceOf(DashboardContractClient);
      expect(client.getAddress()).toBe(contractAddress);
      expect(mockedGetContract).toHaveBeenCalledWith({
        abi: DashboardABI,
        address: contractAddress,
        client: publicClient,
      });
    });

    it("returns cached client when already exists", () => {
      DashboardContractClient.initialize(blockchainClient, logger);

      const client1 = DashboardContractClient.getOrCreate(contractAddress);
      mockedGetContract.mockClear();
      const client2 = DashboardContractClient.getOrCreate(contractAddress);

      expect(client1).toBe(client2);
      expect(mockedGetContract).not.toHaveBeenCalled();
    });

    it("creates separate clients for different addresses", () => {
      DashboardContractClient.initialize(blockchainClient, logger);
      const otherAddress = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" as Address;

      const client1 = DashboardContractClient.getOrCreate(contractAddress);
      const client2 = DashboardContractClient.getOrCreate(otherAddress);

      expect(client1).not.toBe(client2);
      expect(client1.getAddress()).toBe(contractAddress);
      expect(client2.getAddress()).toBe(otherAddress);
    });
  });

  describe("constructor", () => {
    it("throws error when blockchainClient is not initialized", () => {
      expect(() => {
        new DashboardContractClient(contractAddress);
      }).toThrow(
        "DashboardContractClient: blockchainClient must be initialized via DashboardContractClient.initialize() before use",
      );
    });

    it("uses static blockchainClient when initialized", () => {
      DashboardContractClient.initialize(blockchainClient, logger);

      const client = new DashboardContractClient(contractAddress);

      expect(client).toBeInstanceOf(DashboardContractClient);
      expect(mockedGetContract).toHaveBeenCalledWith({
        abi: DashboardABI,
        address: contractAddress,
        client: publicClient,
      });
    });

    it("throws error when logger is not initialized", () => {
      (DashboardContractClient as any).blockchainClient = blockchainClient;
      (DashboardContractClient as any).logger = undefined;

      expect(() => {
        new DashboardContractClient(contractAddress);
      }).toThrow(
        "DashboardContractClient: logger must be initialized via DashboardContractClient.initialize() before use",
      );
    });
  });

  describe("resumeBeaconChainDeposits", () => {
    it("sends transaction to resume beacon chain deposits", async () => {
      const txReceipt = { transactionHash: "0xresume" } as unknown as TransactionReceipt;
      const calldata = "0xcalldata" as `0x${string}`;
      mockedEncodeFunctionData.mockReturnValue(calldata);
      blockchainClient.sendSignedTransaction.mockResolvedValueOnce(txReceipt);

      const client = createClient();
      const result = await client.resumeBeaconChainDeposits();

      expect(mockedEncodeFunctionData).toHaveBeenCalledWith({
        abi: DashboardABI,
        functionName: "resumeBeaconChainDeposits",
        args: [],
      });
      expect(blockchainClient.sendSignedTransaction).toHaveBeenCalledWith(contractAddress, calldata);
      expect(logger.info).toHaveBeenCalledWith(`resumeBeaconChainDeposits started, dashboard=${contractAddress}`);
      expect(logger.info).toHaveBeenCalledWith(
        `resumeBeaconChainDeposits succeeded, dashboard=${contractAddress}, txHash=${txReceipt.transactionHash}`,
      );
      expect(result).toBe(txReceipt);
    });
  });

  describe("resumeBeaconChainDepositsIfSufficientBalance", () => {
    it("resumes deposits when withdrawableValue meets threshold", async () => {
      const minBalanceWei = 1000n;
      const withdrawableValue = 1500n;
      const txReceipt = { transactionHash: "0xresume" } as unknown as TransactionReceipt;
      const calldata = "0xcalldata" as `0x${string}`;

      viemContractStub.read.withdrawableValue.mockResolvedValueOnce(withdrawableValue);
      mockedEncodeFunctionData.mockReturnValue(calldata);
      blockchainClient.sendSignedTransaction.mockResolvedValueOnce(txReceipt);

      const client = createClient();
      const result = await client.resumeBeaconChainDepositsIfSufficientBalance(minBalanceWei);

      expect(viemContractStub.read.withdrawableValue).toHaveBeenCalledTimes(1);
      expect(logger.info).toHaveBeenCalledWith(
        `resumeBeaconChainDepositsIfSufficientBalance - withdrawableValue=${withdrawableValue} meets threshold ${minBalanceWei}, resuming beacon chain deposits`,
      );
      expect(mockedEncodeFunctionData).toHaveBeenCalledWith({
        abi: DashboardABI,
        functionName: "resumeBeaconChainDeposits",
        args: [],
      });
      expect(blockchainClient.sendSignedTransaction).toHaveBeenCalledWith(contractAddress, calldata);
      expect(result).toBe(txReceipt);
    });

    it("resumes deposits when withdrawableValue equals threshold", async () => {
      const minBalanceWei = 1000n;
      const withdrawableValue = 1000n;
      const txReceipt = { transactionHash: "0xresume" } as unknown as TransactionReceipt;
      const calldata = "0xcalldata" as `0x${string}`;

      viemContractStub.read.withdrawableValue.mockResolvedValueOnce(withdrawableValue);
      mockedEncodeFunctionData.mockReturnValue(calldata);
      blockchainClient.sendSignedTransaction.mockResolvedValueOnce(txReceipt);

      const client = createClient();
      const result = await client.resumeBeaconChainDepositsIfSufficientBalance(minBalanceWei);

      expect(viemContractStub.read.withdrawableValue).toHaveBeenCalledTimes(1);
      expect(logger.info).toHaveBeenCalledWith(
        `resumeBeaconChainDepositsIfSufficientBalance - withdrawableValue=${withdrawableValue} meets threshold ${minBalanceWei}, resuming beacon chain deposits`,
      );
      expect(result).toBe(txReceipt);
    });

    it("returns undefined when withdrawableValue is below threshold", async () => {
      const minBalanceWei = 1000n;
      const withdrawableValue = 500n;

      viemContractStub.read.withdrawableValue.mockResolvedValueOnce(withdrawableValue);

      const client = createClient();
      const result = await client.resumeBeaconChainDepositsIfSufficientBalance(minBalanceWei);

      expect(viemContractStub.read.withdrawableValue).toHaveBeenCalledTimes(1);
      expect(logger.info).toHaveBeenCalledWith(
        `resumeBeaconChainDepositsIfSufficientBalance - skipping resume as withdrawableValue=${withdrawableValue} is below the minimum balance threshold of ${minBalanceWei}`,
      );
      expect(mockedEncodeFunctionData).not.toHaveBeenCalled();
      expect(blockchainClient.sendSignedTransaction).not.toHaveBeenCalled();
      expect(result).toBeUndefined();
    });
  });
});

