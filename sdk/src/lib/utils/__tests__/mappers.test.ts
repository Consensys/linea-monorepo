import { describe, it, expect } from "@jest/globals";
import { BigNumber } from "ethers";
import { LogDescription } from "ethers/lib/utils";
import { MessageSentEvent } from "../../../typechain/ZkEvmV2";
import {
  TEST_ADDRESS_1,
  TEST_ADDRESS_2,
  TEST_CONTRACT_ADDRESS_1,
  TEST_MESSAGE_HASH,
  TEST_TRANSACTION_HASH,
} from "../testHelpers/constants";
import { mapMessageSentEventOrLogToMessage } from "../mappers";

describe("Mappers", () => {
  describe("mapMessageSentEventOrLogToMessage", () => {
    it("Should return Message object when we passed a MessageSentEvent object", () => {
      const input: MessageSentEvent = {
        address: TEST_CONTRACT_ADDRESS_1,
        data: "0x",
        topics: [],
        transactionHash: TEST_TRANSACTION_HASH,
        event: "MessageSent",
        eventSignature: "MessageSent(address,address,uint256,uint256,uint256,bytes,bytes32)",
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        args: {
          _from: TEST_ADDRESS_1,
          _to: TEST_ADDRESS_2,
          _fee: BigNumber.from(1),
          _value: BigNumber.from(2),
          _nonce: BigNumber.from(3),
          _calldata: "0x",
          _messageHash: TEST_MESSAGE_HASH,
        },
      };

      expect(mapMessageSentEventOrLogToMessage(input)).toStrictEqual({
        messageSender: TEST_ADDRESS_1,
        destination: TEST_ADDRESS_2,
        fee: BigNumber.from(1),
        value: BigNumber.from(2),
        messageNonce: BigNumber.from(3),
        calldata: "0x",
        messageHash: TEST_MESSAGE_HASH,
      });
    });

    it("Should return Message object when we passed a LogDescription object", () => {
      const input: LogDescription = {
        name: "MessageSent",
        signature: "MessageSent(address,address,uint256,uint256,uint256,bytes,bytes32)",
        topic: "",
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        args: {
          _from: TEST_ADDRESS_1,
          _to: TEST_ADDRESS_2,
          _fee: BigNumber.from(1),
          _value: BigNumber.from(2),
          _nonce: BigNumber.from(3),
          _calldata: "0x",
          _messageHash: TEST_MESSAGE_HASH,
        },
      };

      expect(mapMessageSentEventOrLogToMessage(input)).toStrictEqual({
        messageSender: TEST_ADDRESS_1,
        destination: TEST_ADDRESS_2,
        fee: BigNumber.from(1),
        value: BigNumber.from(2),
        messageNonce: BigNumber.from(3),
        calldata: "0x",
        messageHash: TEST_MESSAGE_HASH,
      });
    });
  });
});
