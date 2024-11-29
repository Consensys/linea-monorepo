import { L2MessagingBlockAnchoredEvent, MessageSentEvent } from "../../clients/typechain/LineaRollup";
import { MessageClaimedEvent, ServiceVersionMigratedEvent } from "../../clients/typechain/L2MessageService";
import { L2_MESSAGING_BLOCK_ANCHORED_EVENT_SIGNATURE, MESSAGE_SENT_EVENT_SIGNATURE } from "../../core/constants";
import { L2MessagingBlockAnchored, MessageClaimed, MessageSent, ServiceVersionMigrated } from "../../core/types";

export const TEST_L1_SIGNER_PRIVATE_KEY = "0x0000000000000000000000000000000000000000000000000000000000000001";
export const TEST_L2_SIGNER_PRIVATE_KEY = "0x0000000000000000000000000000000000000000000000000000000000000002";

export const TEST_ADDRESS_1 = "0x0000000000000000000000000000000000000001";
export const TEST_ADDRESS_2 = "0x0000000000000000000000000000000000000002";

export const TEST_CONTRACT_ADDRESS_1 = "0x1000000000000000000000000000000000000000";
export const TEST_CONTRACT_ADDRESS_2 = "0x2000000000000000000000000000000000000000";

export const TEST_MESSAGE_HASH = "0x1010101010101010101010101010101010101010101010101010101010101010";
export const TEST_MESSAGE_HASH_2 = "0x1010101010101010101010101010101010101010101010101010101010101020";

export const TEST_TRANSACTION_HASH = "0x2020202020202020202020202020202020202020202020202020202020202020";
export const TEST_BLOCK_HASH = "0x1000000000000000000000000000000000000000000000000000000000000000";

export const TEST_MERKLE_ROOT = "0xfc3dfe7470d41465e77e7c929170578b14a066a2272c2469b60162c5282e05a6";
export const TEST_MERKLE_ROOT_2 = "0x7777777777777777777777777777777777777777777777777777777777777777";

export const TEST_RPC_URL = "http://localhost:8545";

export const testMessageSentEvent: MessageSent = {
  blockNumber: 51,
  logIndex: 1,
  contractAddress: TEST_CONTRACT_ADDRESS_1,
  transactionHash: TEST_TRANSACTION_HASH,
  messageSender: TEST_ADDRESS_1,
  destination: TEST_ADDRESS_2,
  fee: 0n,
  value: 0n,
  messageNonce: 1n,
  calldata: "0x",
  messageHash: TEST_MESSAGE_HASH,
};

export const testL2MessagingBlockAnchoredEvent: L2MessagingBlockAnchored = {
  blockNumber: 51,
  logIndex: 0,
  contractAddress: TEST_CONTRACT_ADDRESS_1,
  transactionHash: TEST_TRANSACTION_HASH,
  l2Block: 51n,
};

export const testMessageClaimedEvent: MessageClaimed = {
  blockNumber: 100_000,
  logIndex: 0,
  contractAddress: TEST_CONTRACT_ADDRESS_1,
  transactionHash: TEST_TRANSACTION_HASH,
  messageHash: TEST_MESSAGE_HASH,
};

export const testServiceVersionMigratedEvent: ServiceVersionMigrated = {
  blockNumber: 51,
  logIndex: 0,
  contractAddress: TEST_CONTRACT_ADDRESS_1,
  transactionHash: TEST_TRANSACTION_HASH,
  version: 2n,
};

export const testMessageSentEventLog: MessageSentEvent.Log = {
  blockNumber: 51,
  blockHash: TEST_BLOCK_HASH,
  transactionIndex: 0,
  removed: false,
  address: TEST_CONTRACT_ADDRESS_1,
  data: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000000",
  topics: [
    MESSAGE_SENT_EVENT_SIGNATURE,
    `0x000000000000000000000000${TEST_ADDRESS_1.slice(2)}`,
    `0x000000000000000000000000${TEST_ADDRESS_2.slice(2)}`,
    TEST_MESSAGE_HASH,
  ],
  transactionHash: TEST_TRANSACTION_HASH,
  index: 1,
  removeListener: jest.fn(),
  getBlock: jest.fn(),
  getTransaction: jest.fn(),
  getTransactionReceipt: jest.fn(),
  event: "MessageSent",
  eventSignature: "MessageSent(address,address,uint256,uint256,uint256,bytes,bytes32)",
  decode: jest.fn(),
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  args: {
    _from: TEST_ADDRESS_1,
    _to: TEST_ADDRESS_2,
    _fee: 0n,
    _value: 0n,
    _nonce: 1n,
    _calldata: "0x",
    _messageHash: TEST_MESSAGE_HASH,
  },
};

export const testMessageClaimedEventLog: MessageClaimedEvent.Log = {
  blockNumber: 100_000,
  blockHash: TEST_BLOCK_HASH,
  transactionIndex: 0,
  removed: false,
  address: TEST_CONTRACT_ADDRESS_1,
  data: "0x",
  topics: ["0xa4c827e719e911e8f19393ccdb85b5102f08f0910604d340ba38390b7ff2ab0e", TEST_MESSAGE_HASH],
  transactionHash: TEST_TRANSACTION_HASH,
  index: 0,
  removeListener: jest.fn(),
  getBlock: jest.fn(),
  getTransaction: jest.fn(),
  getTransactionReceipt: jest.fn(),
  event: "MessageClaimed",
  eventSignature: "MessageClaimed(bytes32)",
  decode: jest.fn(),
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  args: {
    _messageHash: TEST_MESSAGE_HASH,
  },
};

export const testServiceVersionMigratedEventLog: ServiceVersionMigratedEvent.Log = {
  blockNumber: 51,
  blockHash: TEST_BLOCK_HASH,
  transactionIndex: 0,
  removed: false,
  address: TEST_CONTRACT_ADDRESS_1,
  data: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000000",
  topics: [
    "0x6f4cd2683fd248db2513412cb2f25f767694689d6f083941f7437c8fe1f87964",
    `0x000000000000000000000000${TEST_ADDRESS_1.slice(2)}`,
    "0x0000000000000000000000000000000000000000000000000000000000000002",
  ],
  transactionHash: TEST_TRANSACTION_HASH,
  index: 0,
  removeListener: jest.fn(),
  getBlock: jest.fn(),
  getTransaction: jest.fn(),
  getTransactionReceipt: jest.fn(),
  event: "ServiceVersionMigrated",
  eventSignature: "ServiceVersionMigrated(uint256)",
  decode: jest.fn(),
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  args: {
    version: 2n,
  },
};

export const testL2MessagingBlockAnchoredEventLog: L2MessagingBlockAnchoredEvent.Log = {
  blockNumber: 51,
  blockHash: TEST_BLOCK_HASH,
  transactionIndex: 0,
  removed: false,
  address: TEST_CONTRACT_ADDRESS_1,
  data: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000000",
  topics: [
    L2_MESSAGING_BLOCK_ANCHORED_EVENT_SIGNATURE,
    "0x0000000000000000000000000000000000000000000000000000000000000033",
  ],
  transactionHash: TEST_TRANSACTION_HASH,
  index: 0,
  removeListener: jest.fn(),
  getBlock: jest.fn(),
  getTransaction: jest.fn(),
  getTransactionReceipt: jest.fn(),
  event: "L2MessagingBlockAnchored",
  eventSignature: "L2MessagingBlockAnchored(uint256)",
  decode: jest.fn(),
  // eslint-disable-next-line @typescript-eslint/ban-ts-comment
  // @ts-ignore
  args: {
    l2Block: 51n,
  },
};
