import { mock, MockProxy } from "jest-mock-extended";
import type { IBlockchainClient, ILogger } from "@consensys/linea-shared-utils";
import type { PublicClient, TransactionReceipt, Address } from "viem";
import { STETHABI } from "../../../core/abis/STETH.js";

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    getContract: jest.fn(),
  };
});

import { getContract } from "viem";

const mockedGetContract = getContract as jest.MockedFunction<typeof getContract>;

let STETHContractClient: typeof import("../STETHContractClient.js").STETHContractClient;

beforeAll(async () => {
  ({ STETHContractClient } = await import("../STETHContractClient.js"));
});

describe("STETHContractClient", () => {
  const contractAddress = "0x1111111111111111111111111111111111111111" as Address;

  let blockchainClient: MockProxy<IBlockchainClient<PublicClient, TransactionReceipt>>;
  let logger: MockProxy<ILogger>;
  let publicClient: PublicClient;
  const viemContractStub = {
    abi: STETHABI,
    read: { getPooledEthBySharesRoundUp: jest.fn() },
  } as any;

  beforeEach(() => {
    jest.clearAllMocks();
    blockchainClient = mock<IBlockchainClient<PublicClient, TransactionReceipt>>();
    logger = mock<ILogger>();
    publicClient = {} as PublicClient;
    blockchainClient.getBlockchainClient.mockReturnValue(publicClient);
    mockedGetContract.mockReturnValue(viemContractStub);
  });

  const createClient = () => new STETHContractClient(blockchainClient, contractAddress, logger);

  it("initializes the viem contract and exposes it through getters", () => {
    const client = createClient();

    expect(mockedGetContract).toHaveBeenCalledWith({
      abi: STETHABI,
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

  describe("getPooledEthBySharesRoundUp", () => {
    it("returns pooled ETH amount when call succeeds", async () => {
      const client = createClient();
      const sharesAmount = 1000000000000000000n; // 1 share
      const expectedValue = 1000000000000000000n; // 1 ETH
      viemContractStub.read.getPooledEthBySharesRoundUp.mockResolvedValueOnce(expectedValue);

      const result = await client.getPooledEthBySharesRoundUp(sharesAmount);

      expect(result).toBe(expectedValue);
      expect(viemContractStub.read.getPooledEthBySharesRoundUp).toHaveBeenCalledWith([sharesAmount]);
    });

    it("returns zero when value is null", async () => {
      const client = createClient();
      const sharesAmount = 1000000000000000000n;
      viemContractStub.read.getPooledEthBySharesRoundUp.mockResolvedValueOnce(null);

      const result = await client.getPooledEthBySharesRoundUp(sharesAmount);

      expect(result).toBe(0n);
      expect(viemContractStub.read.getPooledEthBySharesRoundUp).toHaveBeenCalledTimes(1);
    });

    it("returns undefined and logs error when call fails", async () => {
      const client = createClient();
      const sharesAmount = 1000000000000000000n;
      const error = new Error("Contract call failed");
      viemContractStub.read.getPooledEthBySharesRoundUp.mockRejectedValueOnce(error);

      const result = await client.getPooledEthBySharesRoundUp(sharesAmount);

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith(`getPooledEthBySharesRoundUp failed, error=${error}`);
      expect(viemContractStub.read.getPooledEthBySharesRoundUp).toHaveBeenCalledTimes(1);
    });
  });
});

