import { deposit, computeMessageHash, computeMessageStorageSlot, createStateOverride } from "./deposit";
import { Client, Transport, Chain, Account, Address, BaseError, zeroAddress, Hex } from "viem";
import { readContract, sendTransaction, estimateFeesPerGas, estimateContractGas, multicall } from "viem/actions";
import { getContractsAddressesByChainId } from "@consensys/linea-sdk-core";
import { linea, mainnet } from "viem/chains";
import { TEST_ADDRESS_2, TEST_TRANSACTION_HASH } from "../../tests/constants";

jest.mock("viem/actions", () => ({
  readContract: jest.fn(),
  sendTransaction: jest.fn(),
  estimateFeesPerGas: jest.fn(),
  estimateContractGas: jest.fn(),
  multicall: jest.fn(),
}));

type MockClient = Client<Transport, Chain, Account>;

describe("deposit", () => {
  const mockClient = (chainId?: number, account?: Account): MockClient =>
    ({
      chain: chainId ? { id: chainId } : undefined,
      account,
    }) as unknown as MockClient;

  const mockL2Client = (chainId?: number, account?: Account): MockClient =>
    ({
      chain: chainId ? { id: chainId } : undefined,
      account,
    }) as unknown as MockClient;

  const mockAccount: Account = {
    address: "0x1111111111111111111111111111111111111111",
    type: "json-rpc",
  } as Account;

  const l1ChainId = mainnet.id;
  const l2ChainId = linea.id;
  const l1MessageServiceAddress = getContractsAddressesByChainId(mainnet.id).messageService;
  const l1TokenBridgeAddress = getContractsAddressesByChainId(mainnet.id).tokenBridge;
  const to = TEST_ADDRESS_2;
  const token = "0x7777777777777777777777777777777777777777" as Address;
  const amount = 123n;
  const data = "0x" as Hex;
  const fee = 10n;
  const nextMessageNumber = 42n;
  const l2ClaimingTxGasLimit = 1000n;

  beforeEach(() => {
    jest.clearAllMocks();
    (readContract as jest.Mock<ReturnType<typeof readContract>>).mockResolvedValue(nextMessageNumber);
    (sendTransaction as jest.Mock<ReturnType<typeof sendTransaction>>).mockResolvedValue(TEST_TRANSACTION_HASH);
    (estimateFeesPerGas as jest.Mock<ReturnType<typeof estimateFeesPerGas>>).mockResolvedValue({
      maxFeePerGas: 100_000n,
      maxPriorityFeePerGas: 99_000n,
    });
    (estimateContractGas as jest.Mock<ReturnType<typeof estimateContractGas>>).mockResolvedValue(l2ClaimingTxGasLimit);
    (multicall as jest.Mock<ReturnType<typeof multicall>>).mockResolvedValue(["TokenName", "TKN", 18, zeroAddress]);
  });

  it("throws if no account is provided", async () => {
    const client = mockClient(l1ChainId, undefined);
    await expect(deposit(client, { l2Client: mockL2Client(l2ChainId), token, to, amount })).rejects.toThrow(BaseError);
  });

  it("throws if no L1 or L2 chain id is found", async () => {
    const client = mockClient(undefined, mockAccount);
    const l2Client = mockL2Client(undefined, mockAccount);
    await expect(deposit(client, { l2Client, token, to, amount, account: mockAccount })).rejects.toThrow(BaseError);
  });

  it("sends ETH deposit transaction when token is zeroAddress", async () => {
    const client = mockClient(l1ChainId, mockAccount);
    const l2Client = mockL2Client(l2ChainId, mockAccount);
    const result = await deposit(client, {
      l2Client,
      token: zeroAddress,
      to,
      amount,
      data,
      account: mockAccount,
      fee,
    });
    expect(sendTransaction).toHaveBeenCalledWith(
      client,
      expect.objectContaining({
        to: l1MessageServiceAddress,
        value: amount + fee,
        account: mockAccount,
        data: expect.any(String),
      }),
    );
    expect(result).toBe(TEST_TRANSACTION_HASH);
  });

  it("sends ERC20 deposit transaction when token is not zeroAddress", async () => {
    const client = mockClient(l1ChainId, mockAccount);
    const l2Client = mockL2Client(l2ChainId, mockAccount);
    const result = await deposit(client, {
      l2Client,
      token,
      to,
      amount,
      data,
      account: mockAccount,
      fee,
    });
    expect(sendTransaction).toHaveBeenCalledWith(
      client,
      expect.objectContaining({
        to: l1TokenBridgeAddress,
        value: fee,
        account: mockAccount,
        data: expect.any(String),
      }),
    );
    expect(result).toBe(TEST_TRANSACTION_HASH);
  });

  it("estimates fee if not provided (ETH)", async () => {
    const client = mockClient(l1ChainId, mockAccount);
    const l2Client = mockL2Client(l2ChainId, mockAccount);
    await deposit(client, {
      l2Client,
      token: zeroAddress,
      to,
      amount,
      data,
      account: mockAccount,
    });
    expect(estimateFeesPerGas).toHaveBeenCalledWith(l2Client, expect.any(Object));
    expect(readContract).toHaveBeenCalledWith(client, expect.objectContaining({ functionName: "nextMessageNumber" }));
    expect(estimateContractGas).toHaveBeenCalled();
  });

  it("estimates fee if not provided (ERC20)", async () => {
    const client = mockClient(l1ChainId, mockAccount);
    const l2Client = mockL2Client(l2ChainId, mockAccount);
    await deposit(client, {
      l2Client,
      token,
      to,
      amount,
      data,
      account: mockAccount,
    });
    expect(estimateFeesPerGas).toHaveBeenCalledWith(l2Client, expect.any(Object));
    expect(readContract).toHaveBeenCalledWith(client, expect.objectContaining({ functionName: "nextMessageNumber" }));
    expect(estimateContractGas).toHaveBeenCalled();
    expect(multicall).toHaveBeenCalled();
  });
});

describe("deposit utility functions", () => {
  it("computeMessageHash returns correct hash", () => {
    const from = "0x1111111111111111111111111111111111111111" as Address;
    const to = "0x2222222222222222222222222222222222222222" as Address;
    const fee = 1n;
    const value = 2n;
    const nonce = 3n;
    const calldata = "0x" as Hex;

    const hash = computeMessageHash(from, to, fee, value, nonce, calldata);
    expect(hash).toBe("0x1ebd3a6c6d29012c12e2e6cc8c9cc3346ccd756b4b997e2c435b1a8b4c7c00e7");
  });

  it("computeMessageStorageSlot returns correct slot", () => {
    const messageHash = "0x" + "a".repeat(64);
    const slot = computeMessageStorageSlot(messageHash as Hex);
    expect(slot).toBe("0xdaa0649102c4e0b3eba6c3d3634428c0e6bf46eb30a919f60194ee4cbaca2db3");
  });

  it("createStateOverride returns correct structure", () => {
    const address = "0x1111111111111111111111111111111111111111" as Address;
    const slot = ("0x" + "b".repeat(64)) as Hex;
    const override = createStateOverride(address, slot);
    expect(override).toStrictEqual([
      {
        address,
        stateDiff: [
          {
            slot,
            value: "0x0000000000000000000000000000000000000000000000000000000000000001" as Hex,
          },
        ],
      },
    ]);
  });
});
