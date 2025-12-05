import { mock, MockProxy } from "jest-mock-extended";
import type { ILogger, IBlockchainClient } from "@consensys/linea-shared-utils";
import type { PublicClient, TransactionReceipt, Address, Hex } from "viem";
import { LineaRollupYieldExtensionABI } from "../../../core/abis/LineaRollupYieldExtension.js";

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    getContract: jest.fn(),
    encodeFunctionData: jest.fn(),
  };
});

import { getContract, encodeFunctionData } from "viem";

const mockedGetContract = getContract as jest.MockedFunction<typeof getContract>;
const mockedEncodeFunctionData = encodeFunctionData as jest.MockedFunction<typeof encodeFunctionData>;

let LineaRollupYieldExtensionContractClient: typeof import("../LineaRollupYieldExtensionContractClient.js").LineaRollupYieldExtensionContractClient;

beforeAll(async () => {
  ({ LineaRollupYieldExtensionContractClient } = await import("../LineaRollupYieldExtensionContractClient.js"));
});

describe("LineaRollupYieldExtensionContractClient", () => {
  const contractAddress = "0x1111111111111111111111111111111111111111" as Address;

  let logger: MockProxy<ILogger>;
  let blockchainClient: MockProxy<IBlockchainClient<PublicClient, TransactionReceipt>>;
  let publicClient: PublicClient;
  const contractStub = { abi: LineaRollupYieldExtensionABI } as any;

  beforeEach(() => {
    jest.clearAllMocks();
    logger = mock<ILogger>();
    blockchainClient = mock<IBlockchainClient<PublicClient, TransactionReceipt>>();
    publicClient = {} as PublicClient;
    blockchainClient.getBlockchainClient.mockReturnValue(publicClient);
    mockedGetContract.mockReturnValue(contractStub);
  });

  const createClient = () => new LineaRollupYieldExtensionContractClient(logger, blockchainClient, contractAddress);

  it("initializes viem contract with provided address and client", () => {
    const client = createClient();

    expect(mockedGetContract).toHaveBeenCalledWith({
      abi: LineaRollupYieldExtensionABI,
      address: contractAddress,
      client: publicClient,
    });
    expect(client.getContract()).toBe(contractStub);
  });

  it("exposes the configured contract address", () => {
    const client = createClient();

    expect(client.getAddress()).toBe(contractAddress);
  });

  it("gets the contract balance", async () => {
    const balance = 1_000_000_000_000_000_000n; // 1 ETH
    blockchainClient.getBalance.mockResolvedValueOnce(balance);

    const client = createClient();
    await expect(client.getBalance()).resolves.toBe(balance);

    expect(blockchainClient.getBalance).toHaveBeenCalledWith(contractAddress);
  });

  it("encodes calldata and relays transferFundsForNativeYield to the blockchain client", async () => {
    const client = createClient();
    const amount = 123n;
    const calldata = "0xdeadbeef" as Hex;
    const txReceipt = { transactionHash: "0xhash" } as unknown as TransactionReceipt;

    mockedEncodeFunctionData.mockReturnValue(calldata);
    blockchainClient.sendSignedTransaction.mockResolvedValue(txReceipt);

    const receipt = await client.transferFundsForNativeYield(amount);

    expect(receipt).toBe(txReceipt);
    expect(logger.debug).toHaveBeenCalledWith("transferFundsForNativeYield started, amount=123");
    expect(mockedEncodeFunctionData).toHaveBeenCalledWith({
      abi: contractStub.abi,
      functionName: "transferFundsForNativeYield",
      args: [amount],
    });
    expect(blockchainClient.sendSignedTransaction).toHaveBeenCalledWith(contractAddress, calldata);
    expect(logger.info).toHaveBeenCalledWith("transferFundsForNativeYield succeeded, amount=123, txHash=0xhash");
  });
});
