import { gweiToWei, weiToGwei, weiToGweiNumber, get0x02WithdrawalCredentials } from "../blockchain";
import { WEI_PER_GWEI } from "../../core/constants/blockchain";

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
