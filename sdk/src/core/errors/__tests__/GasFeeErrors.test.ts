import { describe, it } from "@jest/globals";
import { FeeEstimationError, GasEstimationError } from "../GasFeeErrors";
import { serialize } from "../../utils/serialize";
import { Direction, MessageStatus } from "../../enums/MessageEnums";
import { MessageProps } from "../../entities/Message";
import { ZERO_ADDRESS, ZERO_HASH } from "../../constants";

describe("BaseError", () => {
  describe("FeeEstimationError", () => {
    it("Should log error message", () => {
      expect(serialize(new FeeEstimationError("An error message."))).toStrictEqual(
        serialize({
          name: "FeeEstimationError",
          message: "An error message.",
        }),
      );
    });
  });

  describe("GasEstimationError", () => {
    it("Should log error message", () => {
      const rejectedMessage: MessageProps = {
        messageHash: ZERO_HASH,
        messageSender: ZERO_ADDRESS,
        destination: ZERO_ADDRESS,
        fee: 0n,
        value: 0n,
        messageNonce: 0n,
        calldata: "0x",
        contractAddress: ZERO_ADDRESS,
        sentBlockNumber: 0,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.SENT,
        claimNumberOfRetry: 0,
      };

      const estimationError = new Error("estimation error");

      expect(serialize(new GasEstimationError(estimationError.message, rejectedMessage))).toStrictEqual(
        serialize({
          name: "GasEstimationError",
          message: "estimation error",
          rejectedMessage: {
            messageHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
            messageSender: "0x0000000000000000000000000000000000000000",
            destination: "0x0000000000000000000000000000000000000000",
            fee: 0n,
            value: 0n,
            messageNonce: 0n,
            calldata: "0x",
            contractAddress: "0x0000000000000000000000000000000000000000",
            sentBlockNumber: 0,
            direction: "L1_TO_L2",
            status: "SENT",
            claimNumberOfRetry: 0,
          },
        }),
      );
    });
  });
});
