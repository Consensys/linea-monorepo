import { describe, it, expect } from "@jest/globals";
import { mapMessageEntityToMessage, mapMessageToMessageEntity } from "../mappers";
import { generateMessageEntity, generateMessageFromDb } from "../../../utils/testHelpers/helpers";
import {
  TEST_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_2,
  TEST_MESSAGE_HASH,
  TEST_TRANSACTION_HASH,
} from "../../../utils/testHelpers/constants";
import { Direction, MessageStatus } from "../enums";

describe("Mappers", () => {
  describe("mapMessageToMessageEntity", () => {
    it("Should return MessageEntity", () => {
      const messageInDb = generateMessageFromDb();

      expect(mapMessageToMessageEntity(messageInDb)).toStrictEqual({
        calldata: "0x",
        claimNumberOfRetry: 0,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        destination: TEST_CONTRACT_ADDRESS_1,
        direction: Direction.L1_TO_L2,
        fee: "10",
        id: 1,
        messageContractAddress: TEST_CONTRACT_ADDRESS_2,
        messageHash: TEST_MESSAGE_HASH,
        messageNonce: 1,
        messageSender: TEST_ADDRESS_1,
        sentBlockNumber: 100000,
        status: MessageStatus.SENT,
        value: "2",
      });
    });
  });

  describe("mapMessageEntityToMessage", () => {
    it("Should return MessageEntity", () => {
      const messageEntity = generateMessageEntity();

      expect(mapMessageEntityToMessage(messageEntity)).toStrictEqual({
        calldata: "0x",
        claimNumberOfRetry: 0,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        destination: TEST_CONTRACT_ADDRESS_1,
        direction: Direction.L1_TO_L2,
        fee: "10",
        messageContractAddress: TEST_CONTRACT_ADDRESS_2,
        messageHash: TEST_MESSAGE_HASH,
        messageNonce: 1,
        messageSender: TEST_ADDRESS_1,
        sentBlockNumber: 100000,
        status: MessageStatus.SENT,
        value: "2",
        claimGasEstimationThreshold: 1.0,
        claimLastRetriedAt: undefined,
        claimTxCreationDate: undefined,
        claimTxGasLimit: 60_000,
        claimTxHash: TEST_TRANSACTION_HASH,
        claimTxMaxFeePerGas: 100_000_000n,
        claimTxMaxPriorityFeePerGas: 50_000_000n,
        claimTxNonce: 1,
      });
    });
  });
});
