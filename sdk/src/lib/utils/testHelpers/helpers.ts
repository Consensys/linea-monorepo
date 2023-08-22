import { TransactionReceipt, TransactionResponse } from "@ethersproject/providers";
import { BigNumber, ethers } from "ethers";
import { Direction, MessageStatus } from "../../postman/utils/enums";
import { MessageInDb } from "../../postman/utils/types";
import {
  TEST_ADDRESS_1,
  TEST_ADDRESS_2,
  TEST_BLOCK_HASH,
  TEST_CONTRACT_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_2,
  TEST_MESSAGE_HASH,
  TEST_TRANSACTION_HASH,
} from "./constants";
import { MessageEntity } from "../../postman/entity/Message.entity";
import { Message } from "../types";
import { MESSAGE_SENT_EVENT_SIGNATURE } from "../constants";

const mocks = new Map();

export const mockProperty = <T extends object, K extends keyof T>(object: T, property: K, value: T[K]) => {
  const descriptor = Object.getOwnPropertyDescriptor(object, property);
  const mocksForThisObject = mocks.get(object) || {};
  mocksForThisObject[property] = descriptor;
  mocks.set(object, mocksForThisObject);
  Object.defineProperty(object, property, { get: () => value });
};

export const undoMockProperty = <T extends object, K extends keyof T>(object: T, property: K) => {
  Object.defineProperty(object, property, mocks.get(object)[property]);
};

export const generateMessageFromDb = (overrides?: Partial<MessageInDb>): MessageInDb => ({
  id: 1,
  messageSender: TEST_ADDRESS_1,
  destination: TEST_CONTRACT_ADDRESS_1,
  fee: "10",
  value: "2",
  messageNonce: 1,
  calldata: "0x",
  messageHash: TEST_MESSAGE_HASH,
  messageContractAddress: TEST_CONTRACT_ADDRESS_2,
  sentBlockNumber: 100_000,
  direction: Direction.L1_TO_L2,
  status: MessageStatus.SENT,
  claimNumberOfRetry: 0,
  createdAt: new Date("2023-08-04"),
  updatedAt: new Date("2023-08-04"),
  ...overrides,
});

export const generateMessageEntity = (overrides?: Partial<MessageEntity>): MessageEntity => ({
  id: 1,
  messageSender: TEST_ADDRESS_1,
  destination: TEST_CONTRACT_ADDRESS_1,
  fee: "10",
  value: "2",
  messageNonce: 1,
  calldata: "0x",
  messageHash: TEST_MESSAGE_HASH,
  messageContractAddress: TEST_CONTRACT_ADDRESS_2,
  sentBlockNumber: 100_000,
  direction: Direction.L1_TO_L2,
  status: MessageStatus.SENT,
  claimNumberOfRetry: 0,
  createdAt: new Date("2023-08-04"),
  updatedAt: new Date("2023-08-04"),
  claimGasEstimationThreshold: 1.0,
  claimLastRetriedAt: undefined,
  claimTxCreationDate: undefined,
  claimTxGasLimit: 60_000,
  claimTxHash: TEST_TRANSACTION_HASH,
  claimTxMaxFeePerGas: 100_000_000n,
  claimTxMaxPriorityFeePerGas: 50_000_000n,
  claimTxNonce: 1,
  ...overrides,
});

export const generateTransactionReceipt = (overrides?: Partial<TransactionReceipt>): TransactionReceipt => ({
  transactionHash: TEST_TRANSACTION_HASH,
  blockHash: TEST_BLOCK_HASH,
  to: TEST_CONTRACT_ADDRESS_1,
  from: TEST_ADDRESS_1,
  contractAddress: TEST_CONTRACT_ADDRESS_1,
  transactionIndex: 0,
  gasUsed: BigNumber.from(70_000),
  logsBloom: "",
  logs: [
    {
      transactionIndex: 0,
      blockNumber: 100_000,
      removed: false,
      transactionHash: TEST_TRANSACTION_HASH,
      address: TEST_CONTRACT_ADDRESS_1,
      topics: [
        MESSAGE_SENT_EVENT_SIGNATURE,
        `0x000000000000000000000000${TEST_ADDRESS_1.slice(2)}`,
        `0x000000000000000000000000${TEST_ADDRESS_1.slice(2)}`,
        TEST_MESSAGE_HASH,
      ],
      data: "0x00000000000000000000000000000000000000000000000000038d7ea4c68000000000000000000000000000000000000000000000000000015fb7f9b8c3800000000000000000000000000000000000000000000000000000000000000003d700000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000000",
      logIndex: 0,
      blockHash: TEST_BLOCK_HASH,
    },
  ],
  blockNumber: 100_000,
  confirmations: 1,
  cumulativeGasUsed: BigNumber.from(75_000),
  effectiveGasPrice: BigNumber.from(30_000),
  byzantium: true,
  type: 2,
  status: 1,
  ...overrides,
});

export const generateMessage = (
  overrides?: Partial<Message & { feeRecipient?: string }>,
): Message & { feeRecipient?: string } => ({
  messageSender: TEST_ADDRESS_1,
  destination: TEST_ADDRESS_2,
  fee: BigNumber.from(100_000_000),
  value: ethers.utils.parseEther("1"),
  calldata: "0x",
  messageNonce: BigNumber.from(1),
  messageHash: TEST_MESSAGE_HASH,
  feeRecipient: TEST_ADDRESS_2,
  ...overrides,
});

export const generateTransactionResponse = (overrides?: Partial<TransactionResponse>): TransactionResponse => ({
  hash: TEST_TRANSACTION_HASH,
  type: 0,
  accessList: undefined,
  blockHash: TEST_BLOCK_HASH,
  blockNumber: 1077297,
  confirmations: 212238,
  from: TEST_ADDRESS_1,
  gasPrice: BigNumber.from("3492211493612"),
  gasLimit: BigNumber.from("74959"),
  to: TEST_CONTRACT_ADDRESS_2,
  value: BigNumber.from("4313771350571206145"),
  nonce: 40,
  data: `0x9f3ce55a000000000000000000000000${TEST_ADDRESS_1.slice(
    2,
  )}000000000000000000000000000000000000000000000000002386f26fc1000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000`,
  chainId: 59140,
  wait: jest.fn(),
  ...overrides,
});
