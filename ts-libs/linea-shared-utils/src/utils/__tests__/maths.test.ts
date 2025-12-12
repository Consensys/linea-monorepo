import { safeSub, min, absDiff } from "../maths";

describe("safeSub", () => {
  it("returns the difference when the minuend is greater than the subtrahend", () => {
    expect(safeSub(10n, 3n)).toBe(7n);
  });

  it("returns zero when values are equal", () => {
    expect(safeSub(5n, 5n)).toBe(0n);
  });

  it("returns zero when the minuend is smaller than the subtrahend", () => {
    expect(safeSub(2n, 8n)).toBe(0n);
  });
});

describe("min", () => {
  it("returns the first value when it is smaller", () => {
    expect(min(1n, 4n)).toBe(1n);
  });

  it("returns the second value when it is smaller", () => {
    expect(min(9n, 3n)).toBe(3n);
  });

  it("returns either value when they are equal", () => {
    expect(min(6n, 6n)).toBe(6n);
  });
});

describe("absDiff", () => {
  it("returns the difference when the first value is greater", () => {
    expect(absDiff(10n, 3n)).toBe(7n);
  });

  it("returns the difference when the second value is greater", () => {
    expect(absDiff(3n, 10n)).toBe(7n);
  });

  it("returns zero when values are equal", () => {
    expect(absDiff(5n, 5n)).toBe(0n);
  });

  it("returns zero when both values are zero", () => {
    expect(absDiff(0n, 0n)).toBe(0n);
  });

  it("handles large bigint values", () => {
    expect(absDiff(1000000000000000000n, 500000000000000000n)).toBe(500000000000000000n);
  });
});
