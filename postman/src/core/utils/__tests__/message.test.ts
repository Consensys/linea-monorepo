import { describe, it, expect } from "@jest/globals";
import { formatMessageStatus } from "../message";
import { OnChainMessageStatus } from "../../enums/MessageEnums";

describe("Message utils", () => {
  describe("formatMessageStatus", () => {
    it("should return 'UNKNOWN' when status != 1 and status != 2", () => {
      expect(formatMessageStatus(0n)).toStrictEqual(OnChainMessageStatus.UNKNOWN);
    });

    it("should return 'CLAIMABLE' when status = 1", () => {
      expect(formatMessageStatus(1n)).toStrictEqual(OnChainMessageStatus.CLAIMABLE);
    });

    it("should return 'CLAIMED' when status = 2", () => {
      expect(formatMessageStatus(2n)).toStrictEqual(OnChainMessageStatus.CLAIMED);
    });
  });
});
