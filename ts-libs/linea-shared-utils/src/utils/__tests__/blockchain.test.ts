import { WEI_PER_GWEI, SLOTS_PER_EPOCH } from "../../core/constants/blockchain";
import { gweiToWei, weiToGwei, weiToGweiNumber, get0x02WithdrawalCredentials, slotToEpoch } from "../blockchain";

describe("weiToGwei", () => {
  it("returns zero when converting zero wei", () => {
    expect(weiToGwei(0n)).toBe(0n);
  });

  it("converts exact multiples of WEI_PER_GWEI to whole gwei", () => {
    expect(weiToGwei(5n * WEI_PER_GWEI)).toBe(5n);
  });

  it("floors fractional gwei values", () => {
    const fractionalWei = 1234n * WEI_PER_GWEI + 567n;
    expect(weiToGwei(fractionalWei)).toBe(1234n);
  });
});

describe("weiToGweiNumber", () => {
  it("returns a number representation of gwei", () => {
    expect(weiToGweiNumber(7n * WEI_PER_GWEI)).toBe(7);
  });

  it("handles large but safe values", () => {
    const gweiValue = 9_000_000n;
    expect(weiToGweiNumber(gweiValue * WEI_PER_GWEI)).toBe(Number(gweiValue));
  });
});

describe("gweiToWei", () => {
  it("converts gwei to wei", () => {
    expect(gweiToWei(42n)).toBe(42n * WEI_PER_GWEI);
  });

  it("is the inverse of weiToGwei for whole gwei amounts", () => {
    const gweiValue = 1_234_567n;
    const wei = gweiToWei(gweiValue);
    expect(weiToGwei(wei)).toBe(gweiValue);
  });
});

describe("get0x02WithdrawalCredentials", () => {
  it("converts a valid lowercase address to withdrawal credentials format", () => {
    const address = "0x2101af8b812b529fc303c976b6dd747618cfdadb";
    const expected = "0x0200000000000000000000002101af8b812b529fc303c976b6dd747618cfdadb";
    expect(get0x02WithdrawalCredentials(address)).toBe(expected);
  });

  it("converts a valid uppercase address to lowercase withdrawal credentials", () => {
    const address = "0x2101AF8B812B529FC303C976B6DD747618CFDADB";
    const expected = "0x0200000000000000000000002101af8b812b529fc303c976b6dd747618cfdadb";
    expect(get0x02WithdrawalCredentials(address)).toBe(expected);
  });

  it("converts a valid mixed-case address to lowercase withdrawal credentials", () => {
    const address = "0x2101Af8B812b529Fc303c976B6Dd747618cFdAdB";
    const expected = "0x0200000000000000000000002101af8b812b529fc303c976b6dd747618cfdadb";
    expect(get0x02WithdrawalCredentials(address)).toBe(expected);
  });

  it("handles zero address", () => {
    const address = "0x0000000000000000000000000000000000000000";
    const expected = "0x0200000000000000000000000000000000000000000000000000000000000000";
    expect(get0x02WithdrawalCredentials(address)).toBe(expected);
  });

  it("throws error for invalid address format - too short", () => {
    const invalidAddress = "0x2101af8b812b529fc303c976b6dd747618cfdad";
    expect(() => get0x02WithdrawalCredentials(invalidAddress)).toThrow("Invalid Ethereum address");
  });

  it("throws error for invalid address format - too long", () => {
    const invalidAddress = "0x2101af8b812b529fc303c976b6dd747618cfdadba";
    expect(() => get0x02WithdrawalCredentials(invalidAddress)).toThrow("Invalid Ethereum address");
  });

  it("throws error for invalid address format - missing 0x prefix", () => {
    const invalidAddress = "2101af8b812b529fc303c976b6dd747618cfdadb";
    expect(() => get0x02WithdrawalCredentials(invalidAddress)).toThrow("Invalid Ethereum address");
  });

  it("throws error for invalid address format - invalid hex characters", () => {
    const invalidAddress = "0x2101af8b812b529fc303c976b6dd747618cfdadg";
    expect(() => get0x02WithdrawalCredentials(invalidAddress)).toThrow("Invalid Ethereum address");
  });

  it("throws error for empty string", () => {
    expect(() => get0x02WithdrawalCredentials("")).toThrow("Invalid Ethereum address");
  });
});

describe("slotToEpoch", () => {
  it("returns zero when converting zero slot", () => {
    expect(slotToEpoch(0)).toBe(0);
  });

  it("converts exact multiples of SLOTS_PER_EPOCH to whole epochs", () => {
    expect(slotToEpoch(1 * SLOTS_PER_EPOCH)).toBe(1);
    expect(slotToEpoch(5 * SLOTS_PER_EPOCH)).toBe(5);
    expect(slotToEpoch(100 * SLOTS_PER_EPOCH)).toBe(100);
  });

  it("floors fractional epoch values", () => {
    // Slot 31 should be in epoch 0 (31 / 32 = 0.96875)
    expect(slotToEpoch(31)).toBe(0);
    // Slot 33 should be in epoch 1 (33 / 32 = 1.03125)
    expect(slotToEpoch(33)).toBe(1);
    // Slot 63 should be in epoch 1 (63 / 32 = 1.96875)
    expect(slotToEpoch(63)).toBe(1);
    // Slot 64 should be in epoch 2 (64 / 32 = 2.0)
    expect(slotToEpoch(64)).toBe(2);
  });

  it("handles slots at epoch boundaries", () => {
    // First slot of epoch 1
    expect(slotToEpoch(32)).toBe(1);
    // Last slot of epoch 0
    expect(slotToEpoch(31)).toBe(0);
    // First slot of epoch 2
    expect(slotToEpoch(64)).toBe(2);
    // Last slot of epoch 1
    expect(slotToEpoch(63)).toBe(1);
  });

  it("handles large slot numbers", () => {
    const largeEpoch = 1000;
    const slot = largeEpoch * SLOTS_PER_EPOCH;
    expect(slotToEpoch(slot)).toBe(largeEpoch);
    expect(slotToEpoch(slot + 15)).toBe(largeEpoch); // Should floor
    expect(slotToEpoch(slot + 31)).toBe(largeEpoch); // Last slot of epoch, should still floor
    expect(slotToEpoch(slot + 32)).toBe(largeEpoch + 1); // First slot of next epoch
  });
});
