import { BigNumber } from "ethers";
import { L1NetworkConfig, L2NetworkConfig } from "../../postman/utils/types";
import { MessageClaimedEvent, MessageSentEvent } from "../../../typechain/ZkEvmV2";
import { MESSAGE_SENT_EVENT_SIGNATURE } from "../constants";

export const TEST_L1_SIGNER_PRIVATE_KEY = "0x0000000000000000000000000000000000000000000000000000000000000001";
export const TEST_L2_SIGNER_PRIVATE_KEY = "0x0000000000000000000000000000000000000000000000000000000000000002";

export const TEST_ADDRESS_1 = "0x0000000000000000000000000000000000000001";
export const TEST_ADDRESS_2 = "0x0000000000000000000000000000000000000002";

export const TEST_CONTRACT_ADDRESS_1 = "0x1000000000000000000000000000000000000000";
export const TEST_CONTRACT_ADDRESS_2 = "0x2000000000000000000000000000000000000000";

export const TEST_MESSAGE_HASH = "0x1010101010101010101010101010101010101010101010101010101010101010";
export const TEST_TRANSACTION_HASH = "0x2020202020202020202020202020202020202020202020202020202020202020";
export const TEST_BLOCK_HASH = "0x1000000000000000000000000000000000000000000000000000000000000000";

export const TEST_RPC_URL = "http://localhost:8545";

export const testMessageSentEvent: MessageSentEvent = {
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
  logIndex: 1,
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
    _fee: BigNumber.from(0),
    _value: BigNumber.from(0),
    _nonce: BigNumber.from(1),
    _calldata: "0x",
    _messageHash: TEST_MESSAGE_HASH,
  },
};

export const testMessageClaimedEvent: MessageClaimedEvent = {
  blockNumber: 100_000,
  blockHash: TEST_BLOCK_HASH,
  transactionIndex: 0,
  removed: false,
  address: TEST_CONTRACT_ADDRESS_1,
  data: "0x",
  topics: ["0xa4c827e719e911e8f19393ccdb85b5102f08f0910604d340ba38390b7ff2ab0e", TEST_MESSAGE_HASH],
  transactionHash: TEST_TRANSACTION_HASH,
  logIndex: 0,
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

export const testL1NetworkConfig: L1NetworkConfig = {
  claiming: {
    signerPrivateKey: TEST_L1_SIGNER_PRIVATE_KEY,
    messageSubmissionTimeout: 300_000,
    maxFeePerGas: 100_000_000,
    gasEstimationPercentile: 15,
  },
  listener: {
    pollingInterval: 4000,
    maxFetchMessagesFromDb: 3,
  },
  rpcUrl: "http://localhost:8445",
  messageServiceContractAddress: TEST_CONTRACT_ADDRESS_1,
};

export const testL2NetworkConfig: L2NetworkConfig = {
  claiming: {
    signerPrivateKey: TEST_L2_SIGNER_PRIVATE_KEY,
    messageSubmissionTimeout: 300_000,
    maxFeePerGas: 100_000_000,
    gasEstimationPercentile: 15,
  },
  listener: {
    pollingInterval: 3000,
    maxFetchMessagesFromDb: 3,
  },
  rpcUrl: "http://localhost:8545",
  messageServiceContractAddress: TEST_CONTRACT_ADDRESS_2,
};
