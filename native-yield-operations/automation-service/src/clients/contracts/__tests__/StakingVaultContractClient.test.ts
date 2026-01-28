import { mock, MockProxy } from "jest-mock-extended";
import type { IBlockchainClient, ILogger } from "@consensys/linea-shared-utils";
import type { PublicClient, TransactionReceipt, Address } from "viem";
import { StakingVaultABI } from "../../../core/abis/StakingVault.js";

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    getContract: jest.fn(),
  };
});

import { getContract } from "viem";

const mockedGetContract = getContract as jest.MockedFunction<typeof getContract>;

let StakingVaultContractClient: typeof import("../StakingVaultContractClient.js").StakingVaultContractClient;

beforeAll(async () => {
  ({ StakingVaultContractClient } = await import("../StakingVaultContractClient.js"));
});

describe("StakingVaultContractClient", () => {
  const contractAddress = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" as Address;

  let blockchainClient: MockProxy<IBlockchainClient<PublicClient, TransactionReceipt>>;
  let logger: MockProxy<ILogger>;
  let publicClient: PublicClient;
  const viemContractStub = {
    abi: StakingVaultABI,
    read: {
      beaconChainDepositsPaused: jest.fn(),
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
    (StakingVaultContractClient as any).blockchainClient = undefined;
    (StakingVaultContractClient as any).logger = undefined;
    (StakingVaultContractClient as any).clientCache.clear();
  });

  const createClient = () => {
    // Initialize logger if not already initialized (needed for constructor)
    if (!(StakingVaultContractClient as any).logger) {
      StakingVaultContractClient.initialize(blockchainClient, logger);
    }
    return new StakingVaultContractClient(contractAddress);
  };

  it("initializes the viem contract and exposes it through getters", () => {
    const client = createClient();

    expect(mockedGetContract).toHaveBeenCalledWith({
      abi: StakingVaultABI,
      address: contractAddress,
      client: publicClient,
    });
    expect(client.getAddress()).toBe(contractAddress);
    expect(client.getContract()).toBe(viemContractStub);
  });

  it("gets the contract balance", async () => {
    const balance = 1_000_000_000_000_000_000n; // 1 ETH
    blockchainClient.getBalance.mockResolvedValueOnce(balance);

    const client = createClient();
    await expect(client.getBalance()).resolves.toBe(balance);

    expect(blockchainClient.getBalance).toHaveBeenCalledWith(contractAddress);
  });

  it("throws error when getBalance is called and blockchainClient is not initialized", async () => {
    const client = createClient();
    // Clear the static blockchainClient after client creation
    (StakingVaultContractClient as any).blockchainClient = undefined;

    await expect(client.getBalance()).rejects.toThrow(
      "StakingVaultContractClient: blockchainClient must be initialized via StakingVaultContractClient.initialize() before use",
    );
  });

  it("reads beaconChainDepositsPaused via read and returns the result", async () => {
    const paused = true;
    viemContractStub.read.beaconChainDepositsPaused.mockResolvedValueOnce(paused);

    const client = createClient();
    const result = await client.beaconChainDepositsPaused();

    expect(viemContractStub.read.beaconChainDepositsPaused).toHaveBeenCalledTimes(1);
    expect(result).toBe(paused);
  });

  it("reads beaconChainDepositsPaused when deposits are not paused", async () => {
    const paused = false;
    viemContractStub.read.beaconChainDepositsPaused.mockResolvedValueOnce(paused);

    const client = createClient();
    const result = await client.beaconChainDepositsPaused();

    expect(viemContractStub.read.beaconChainDepositsPaused).toHaveBeenCalledTimes(1);
    expect(result).toBe(paused);
  });

  describe("initialize", () => {
    it("sets the static blockchainClient and logger", () => {
      StakingVaultContractClient.initialize(blockchainClient, logger);

      expect((StakingVaultContractClient as any).blockchainClient).toBe(blockchainClient);
      expect((StakingVaultContractClient as any).logger).toBe(logger);
    });
  });

  describe("getOrCreate", () => {
    it("throws error when blockchainClient is not initialized", () => {
      expect(() => {
        StakingVaultContractClient.getOrCreate(contractAddress);
      }).toThrow(
        "StakingVaultContractClient: blockchainClient must be initialized via StakingVaultContractClient.initialize() before use",
      );
    });

    it("throws error when logger is not initialized", () => {
      (StakingVaultContractClient as any).blockchainClient = blockchainClient;
      (StakingVaultContractClient as any).logger = undefined;

      expect(() => {
        StakingVaultContractClient.getOrCreate(contractAddress);
      }).toThrow(
        "StakingVaultContractClient: logger must be initialized via StakingVaultContractClient.initialize() before use",
      );
    });

    it("creates and caches a new client when not cached", () => {
      StakingVaultContractClient.initialize(blockchainClient, logger);

      const client = StakingVaultContractClient.getOrCreate(contractAddress);

      expect(client).toBeInstanceOf(StakingVaultContractClient);
      expect(client.getAddress()).toBe(contractAddress);
      expect(mockedGetContract).toHaveBeenCalledWith({
        abi: StakingVaultABI,
        address: contractAddress,
        client: publicClient,
      });
    });

    it("returns cached client when already exists", () => {
      StakingVaultContractClient.initialize(blockchainClient, logger);

      const client1 = StakingVaultContractClient.getOrCreate(contractAddress);
      mockedGetContract.mockClear();
      const client2 = StakingVaultContractClient.getOrCreate(contractAddress);

      expect(client1).toBe(client2);
      expect(mockedGetContract).not.toHaveBeenCalled();
    });

    it("creates separate clients for different addresses", () => {
      StakingVaultContractClient.initialize(blockchainClient, logger);
      const otherAddress = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" as Address;

      const client1 = StakingVaultContractClient.getOrCreate(contractAddress);
      const client2 = StakingVaultContractClient.getOrCreate(otherAddress);

      expect(client1).not.toBe(client2);
      expect(client1.getAddress()).toBe(contractAddress);
      expect(client2.getAddress()).toBe(otherAddress);
    });
  });

  describe("constructor", () => {
    it("throws error when blockchainClient is not initialized", () => {
      expect(() => {
        new StakingVaultContractClient(contractAddress);
      }).toThrow(
        "StakingVaultContractClient: blockchainClient must be initialized via StakingVaultContractClient.initialize() before use",
      );
    });

    it("uses static blockchainClient when initialized", () => {
      StakingVaultContractClient.initialize(blockchainClient, logger);

      const client = new StakingVaultContractClient(contractAddress);

      expect(client).toBeInstanceOf(StakingVaultContractClient);
      expect(mockedGetContract).toHaveBeenCalledWith({
        abi: StakingVaultABI,
        address: contractAddress,
        client: publicClient,
      });
    });

    it("throws error when logger is not initialized", () => {
      (StakingVaultContractClient as any).blockchainClient = blockchainClient;
      (StakingVaultContractClient as any).logger = undefined;

      expect(() => {
        new StakingVaultContractClient(contractAddress);
      }).toThrow(
        "StakingVaultContractClient: logger must be initialized via StakingVaultContractClient.initialize() before use",
      );
    });
  });
});

