import { describe, it, expect } from "@jest/globals";
import { BigNumber } from "ethers";
import { OnChainMessageStatus } from "../enum";
import { formatMessageStatus } from "../helpers";

describe("Helpers", () => {
  describe("formatMessageStatus", () => {
    it.each([
      { input: BigNumber.from(0), result: OnChainMessageStatus.UNKNOWN },
      { input: BigNumber.from(1), result: OnChainMessageStatus.CLAIMABLE },
      { input: BigNumber.from(2), result: OnChainMessageStatus.CLAIMED },
    ])("should return $result if status === $input", ({ input, result }) => {
      expect(formatMessageStatus(input)).toStrictEqual(result);
    });
  });
});
