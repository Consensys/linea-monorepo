import { mock, MockProxy } from "jest-mock-extended";
import type { IBlockchainClient } from "@consensys/linea-shared-utils";
import type { PublicClient, TransactionReceipt, Address } from "viem";
import { VaultHubABI } from "../../../core/abis/VaultHub.js";

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    getContract: jest.fn(),
    decodeEventLog: jest.fn(),
  };
});

import { getContract, decodeEventLog } from "viem";

const mockedGetContract = getContract as jest.MockedFunction<typeof getContract>;
const mockedDecodeEventLog = decodeEventLog as jest.MockedFunction<typeof decodeEventLog>;

let VaultHubContractClient: typeof import("../VaultHubContractClient.js").VaultHubContractClient;

beforeAll(async () => {
  ({ VaultHubContractClient } = await import("../VaultHubContractClient.js"));
});

describe("VaultHubContractClient", () => {
  const contractAddress = "0x1111111111111111111111111111111111111111" as Address;

  let blockchainClient: MockProxy<IBlockchainClient<PublicClient, TransactionReceipt>>;
  let publicClient: PublicClient;
  const viemContractStub = { abi: VaultHubABI } as any;

  beforeEach(() => {
    jest.clearAllMocks();
    blockchainClient = mock<IBlockchainClient<PublicClient, TransactionReceipt>>();
    publicClient = {} as PublicClient;
    blockchainClient.getBlockchainClient.mockReturnValue(publicClient);
    mockedGetContract.mockReturnValue(viemContractStub);
  });

  const createClient = () => new VaultHubContractClient(blockchainClient, contractAddress);

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

    mockedDecodeEventLog.mockImplementation(({ data }) => {
      if (data === "0xdata") {
        return {
          eventName: "VaultRebalanced",
          args: { etherWithdrawn: 123n },
        } as any;
      }
      return { eventName: "Other", args: {} } as any;
    });

    const amount = client.getLiabilityPaymentFromTxReceipt(receipt);

    expect(amount).toBe(123n);
    expect(mockedDecodeEventLog).toHaveBeenCalledWith({
      abi: viemContractStub.abi,
      data: "0xdata",
      topics: ["0xtopic"],
    });
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

    mockedDecodeEventLog.mockImplementation(() => {
      throw new Error("bad log");
    });

    const amount = client.getLiabilityPaymentFromTxReceipt(receipt);

    expect(amount).toBe(0n);
    expect(mockedDecodeEventLog).toHaveBeenCalledTimes(1);
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

    mockedDecodeEventLog.mockReturnValueOnce({
      eventName: "LidoFeesSettled",
      args: { transferred: 456n },
    } as any);

    const amount = client.getLidoFeePaymentFromTxReceipt(receipt);

    expect(amount).toBe(456n);
    expect(mockedDecodeEventLog).toHaveBeenCalledWith({
      abi: viemContractStub.abi,
      data: "0xfeed",
      topics: ["0x01"],
    });
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

    mockedDecodeEventLog.mockReturnValueOnce({
      eventName: "OtherEvent",
      args: {},
    } as any);

    const amount = client.getLidoFeePaymentFromTxReceipt(receipt);

    expect(amount).toBe(0n);
    expect(mockedDecodeEventLog).toHaveBeenCalledTimes(1);
  });
});
