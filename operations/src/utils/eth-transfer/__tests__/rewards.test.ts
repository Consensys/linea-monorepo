import { describe, it, expect } from "@jest/globals";
import { ethers } from "ethers";
import { calculateRewards } from "../rewards";

describe("Utils", () => {
  describe("calculateRewards", () => {
    it("should return 0 when the balance is less than 1 ETH", () => {
      const balance = ethers.parseEther("0.5");
      expect(calculateRewards(balance)).toStrictEqual(0n);
    });

    it("should return Math.floor(balance - 1 ETH) when balance > Number.MAX_SAFE_INTEGER (2^53 - 1)", () => {
      const balance = ethers.parseEther("9999999999999999999999999999999");
      expect(calculateRewards(balance)).toStrictEqual(ethers.parseEther("9999999999999999999999999999998"));
    });

    it("should return Math.floor(balance - 1 ETH) when balance < Number.MAX_SAFE_INTEGER (2^53 - 1)", () => {
      const balance = ethers.parseEther("101.55");
      expect(calculateRewards(balance)).toStrictEqual(ethers.parseEther("100"));
    });
  });
});
