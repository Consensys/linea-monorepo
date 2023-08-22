import { describe, it, expect } from "@jest/globals";
import { isEmptyBytes, subtractSeconds } from "../helpers";

describe("Helpers", () => {
  describe("subtractSeconds", () => {
    it("should substract X seconds to Date", async () => {
      const inputDate = new Date("2023-08-01T08:30:30Z");
      const secondsToSubstract = 300; // 5 minutes
      const newDate = subtractSeconds(inputDate, secondsToSubstract);
      expect(newDate).toStrictEqual(new Date("2023-08-01T08:25:30Z"));
    });
  });

  describe("isEmptyBytes", () => {
    it.each([
      { input: "", result: true },
      { input: "0x", result: true },
      { input: "0", result: true },
      { input: "0x01", result: false },
    ])("should return $result if value === $input", ({ input, result }) => {
      expect(isEmptyBytes(input)).toBe(result);
    });
  });
});
