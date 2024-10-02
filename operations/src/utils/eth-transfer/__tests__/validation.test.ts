import { describe, it, expect } from "@jest/globals";
import { validateETHThreshold } from "../validation.js";

describe("Validation utils", () => {
  describe("validateETHThreshold", () => {
    it("should throw an error when the input threshold is less than 1 ETH", () => {
      const invalidThreshold = "0.5";
      expect(() => validateETHThreshold(invalidThreshold)).toThrow("Threshold must be higher than 1 ETH");
    });

    it("should return the input when it is higher than 1 ETH", () => {
      const threshold = "2";
      expect(validateETHThreshold(threshold)).toStrictEqual(threshold);
    });
  });
});
