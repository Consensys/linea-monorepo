import { OnChainMessageStatus } from "../../types/enums";
import { formatMessageStatus } from "../formatMessageStatus";

describe("formatMessageStatus", () => {
  it("should return UNKNOWN when status is 0", () => {
    expect(formatMessageStatus(0n)).toBe(OnChainMessageStatus.UNKNOWN);
  });

  it("should return CLAIMABLE when status is 1", () => {
    expect(formatMessageStatus(1n)).toBe(OnChainMessageStatus.CLAIMABLE);
  });

  it("should return CLAIMED when status is 2", () => {
    expect(formatMessageStatus(2n)).toBe(OnChainMessageStatus.CLAIMED);
  });

  it("should return UNKNOWN for any other status", () => {
    expect(formatMessageStatus(3n)).toBe(OnChainMessageStatus.UNKNOWN);
    expect(formatMessageStatus(99n)).toBe(OnChainMessageStatus.UNKNOWN);
  });
});
