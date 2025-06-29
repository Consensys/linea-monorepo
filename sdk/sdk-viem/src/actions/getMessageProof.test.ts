import { getMessageProof } from "./getMessageProof";
import { Client, Transport, Chain, Account, Hex, BaseError } from "viem";
import { getMessageSentEvents } from "./getMessageSentEvents";
import { getContractEvents, getTransactionReceipt } from "viem/actions";
import { getContractsAddressesByChainId } from "@consensys/linea-sdk-core";
import {
  generateL2MerkleTreeAddedLog,
  generateL2MessagingBlockAnchoredLog,
  generateMessageSentLog,
  generateTransactionReceipt,
} from "../../tests/utils";
import { TEST_MERKLE_ROOT, TEST_MERKLE_ROOT_2, TEST_MESSAGE_HASH } from "../../tests/constants";

jest.mock("./getMessageSentEvents");
jest.mock("viem/actions", () => ({
  getContractEvents: jest.fn(),
  getTransactionReceipt: jest.fn(),
}));

type MockClient = Client<Transport, Chain, Account>;

describe("getMessageProof", () => {
  const mainnetId = 1;
  const lineaId = 59144;
  const messageHash: Hex = TEST_MESSAGE_HASH;
  const l2BlockNumber = 42n;
  const treeDepth = 5;
  const merkleRoot = TEST_MERKLE_ROOT;
  const proof = [
    "0x0000000000000000000000000000000000000000000000000000000000000000",
    "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
    "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
    "0x21ddb9a356815c3fac1026b6dec5df3124afbadb485c9ba5a3e3398a04b7ba85",
    "0xe58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a19344",
  ] as Hex[];
  const leafIndex = 0;

  const mockClient = (chainId?: number): MockClient =>
    ({ chain: chainId ? { id: chainId } : undefined }) as unknown as MockClient;

  const mockL2Client = (chainId?: number): MockClient =>
    ({ chain: chainId ? { id: chainId } : undefined }) as unknown as MockClient;

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("throws if l2Client.chain is not set", async () => {
    const client = mockClient(mainnetId);
    const l2Client = mockL2Client();
    await expect(getMessageProof(client, { l2Client, messageHash })).rejects.toThrow(BaseError);
  });

  it("throws if client.chain is not set", async () => {
    const client = mockClient();
    const l2Client = mockL2Client(lineaId);
    await expect(getMessageProof(client, { l2Client, messageHash })).rejects.toThrow(BaseError);
  });

  it("throws if no MessageSent event is found", async () => {
    const client = mockClient(mainnetId);
    const l2Client = mockL2Client(lineaId);
    (getMessageSentEvents as jest.Mock<ReturnType<typeof getMessageSentEvents>>).mockResolvedValue([]);
    await expect(getMessageProof(client, { l2Client, messageHash })).rejects.toThrow(
      `Message hash does not exist on L2. Message hash: ${messageHash}`,
    );
  });

  it("throws if no L2MessagingBlockAnchored event is found", async () => {
    const client = mockClient(mainnetId);
    const l2Client = mockL2Client(lineaId);
    const messageSentLog = generateMessageSentLog({ blockNumber: l2BlockNumber });
    (getMessageSentEvents as jest.Mock<ReturnType<typeof getMessageSentEvents>>).mockResolvedValue([
      {
        messageSender: messageSentLog.args._from!,
        destination: messageSentLog.args._to!,
        fee: messageSentLog.args._fee!,
        value: messageSentLog.args._value!,
        messageNonce: messageSentLog.args._nonce!,
        calldata: messageSentLog.args._calldata!,
        messageHash: messageSentLog.args._messageHash!,
        blockNumber: messageSentLog.blockNumber,
        logIndex: messageSentLog.logIndex,
        contractAddress: messageSentLog.address,
        transactionHash: messageSentLog.transactionHash,
      },
    ]);
    (getContractEvents as jest.Mock<ReturnType<typeof getContractEvents>>).mockResolvedValue([]);
    await expect(getMessageProof(client, { l2Client, messageHash })).rejects.toThrow(
      `L2 block number ${l2BlockNumber} has not been finalized on L1.`,
    );
  });

  it("throws if no MessageSent events in block range", async () => {
    const client = mockClient(mainnetId);
    const l2Client = mockL2Client(lineaId);
    const messageSentLog = generateMessageSentLog({ blockNumber: l2BlockNumber });
    (getMessageSentEvents as jest.Mock<ReturnType<typeof getMessageSentEvents>>)
      .mockResolvedValueOnce([
        {
          messageSender: messageSentLog.args._from!,
          destination: messageSentLog.args._to!,
          fee: messageSentLog.args._fee!,
          value: messageSentLog.args._value!,
          messageNonce: messageSentLog.args._nonce!,
          calldata: messageSentLog.args._calldata!,
          messageHash: messageSentLog.args._messageHash!,
          blockNumber: messageSentLog.blockNumber,
          logIndex: messageSentLog.logIndex,
          contractAddress: messageSentLog.address,
          transactionHash: messageSentLog.transactionHash,
        },
      ])
      .mockResolvedValueOnce([]); // for block range call
    (getContractEvents as jest.Mock<ReturnType<typeof getContractEvents>>).mockResolvedValue([
      generateL2MessagingBlockAnchoredLog(l2BlockNumber, {
        address: getContractsAddressesByChainId(mainnetId).messageService,
      }),
    ]);
    (getTransactionReceipt as jest.Mock).mockResolvedValue(
      generateTransactionReceipt({
        logs: [
          generateL2MerkleTreeAddedLog(merkleRoot, treeDepth, {
            address: getContractsAddressesByChainId(mainnetId).messageService,
          }),
          generateL2MessagingBlockAnchoredLog(l2BlockNumber, {
            address: getContractsAddressesByChainId(mainnetId).messageService,
          }),
        ],
      }),
    );

    await expect(getMessageProof(client, { l2Client, messageHash })).rejects.toThrow(
      `No MessageSent events found in this block range on L2.`,
    );
  });

  it("throws if merkle root does not match", async () => {
    const client = mockClient(mainnetId);
    const l2Client = mockL2Client(lineaId);
    const messageSentLog = generateMessageSentLog({ blockNumber: l2BlockNumber });
    (getMessageSentEvents as jest.Mock<ReturnType<typeof getMessageSentEvents>>).mockResolvedValue([
      {
        messageSender: messageSentLog.args._from!,
        destination: messageSentLog.args._to!,
        fee: messageSentLog.args._fee!,
        value: messageSentLog.args._value!,
        messageNonce: messageSentLog.args._nonce!,
        calldata: messageSentLog.args._calldata!,
        messageHash: messageSentLog.args._messageHash!,
        blockNumber: messageSentLog.blockNumber,
        logIndex: messageSentLog.logIndex,
        contractAddress: messageSentLog.address,
        transactionHash: messageSentLog.transactionHash,
      },
    ]);
    (getContractEvents as jest.Mock<ReturnType<typeof getContractEvents>>).mockResolvedValue([
      generateL2MessagingBlockAnchoredLog(l2BlockNumber, {
        address: getContractsAddressesByChainId(mainnetId).messageService,
      }),
    ]);
    (getTransactionReceipt as jest.Mock).mockResolvedValue(
      generateTransactionReceipt({
        logs: [
          generateL2MerkleTreeAddedLog(TEST_MERKLE_ROOT_2, treeDepth, {
            address: getContractsAddressesByChainId(mainnetId).messageService,
          }),
          generateL2MessagingBlockAnchoredLog(l2BlockNumber, {
            address: getContractsAddressesByChainId(mainnetId).messageService,
          }),
        ],
      }),
    );

    await expect(getMessageProof(client, { l2Client, messageHash })).rejects.toThrow("Merkle tree build failed.");
  });

  it("returns proof on success", async () => {
    const client = mockClient(mainnetId);
    const l2Client = mockL2Client(lineaId);
    const messageSentLog = generateMessageSentLog({ blockNumber: l2BlockNumber });
    (getMessageSentEvents as jest.Mock<ReturnType<typeof getMessageSentEvents>>).mockResolvedValue([
      {
        messageSender: messageSentLog.args._from!,
        destination: messageSentLog.args._to!,
        fee: messageSentLog.args._fee!,
        value: messageSentLog.args._value!,
        messageNonce: messageSentLog.args._nonce!,
        calldata: messageSentLog.args._calldata!,
        messageHash: messageSentLog.args._messageHash!,
        blockNumber: messageSentLog.blockNumber,
        logIndex: messageSentLog.logIndex,
        contractAddress: messageSentLog.address,
        transactionHash: messageSentLog.transactionHash,
      },
    ]);
    (getContractEvents as jest.Mock<ReturnType<typeof getContractEvents>>).mockResolvedValue([
      generateL2MessagingBlockAnchoredLog(l2BlockNumber, {
        address: getContractsAddressesByChainId(mainnetId).messageService,
      }),
    ]);
    (getTransactionReceipt as jest.Mock).mockResolvedValue(
      generateTransactionReceipt({
        logs: [
          generateL2MerkleTreeAddedLog(TEST_MERKLE_ROOT, treeDepth, {
            address: getContractsAddressesByChainId(mainnetId).messageService,
          }),
          generateL2MessagingBlockAnchoredLog(l2BlockNumber, {
            address: getContractsAddressesByChainId(mainnetId).messageService,
          }),
        ],
      }),
    );

    const result = await getMessageProof(client, { l2Client, messageHash });
    expect(result).toEqual({ proof, root: merkleRoot, leafIndex });
  });
});
