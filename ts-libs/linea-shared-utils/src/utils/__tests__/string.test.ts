import { bigintReplacer } from "../string";

describe("bigintReplacer", () => {
  it("serializes bigint values as decimal strings", () => {
    const json = JSON.stringify({ amount: 1234567890123456789n }, bigintReplacer);
    expect(json).toBe('{"amount":"1234567890123456789"}');
  });

  it("leaves non-bigint values unchanged", () => {
    expect(bigintReplacer("numeric", 42)).toBe(42);
    expect(bigintReplacer("text", "linea")).toBe("linea");
    const obj = { nested: true };
    expect(bigintReplacer("object", obj)).toBe(obj);
  });

  it("handles bigint entries inside nested structures", () => {
    const payload = {
      data: [{ value: 99n }, 1n, "plain"],
    };

    const json = JSON.stringify(payload, bigintReplacer);
    expect(json).toBe('{"data":[{"value":"99"},"1","plain"]}');
  });
});
