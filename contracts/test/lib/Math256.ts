import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { MaxUint256 } from "ethers";
import { TestMath256 } from "../../../typechain-types";
import { deployFromFactory } from "../common/deployment";

describe("Math256 Library", () => {
  let contract: TestMath256;

  async function deployTestMath256Fixture() {
    return deployFromFactory("TestMath256");
  }

  beforeEach(async () => {
    contract = (await loadFixture(deployTestMath256Fixture)) as TestMath256;
  });

  describe("min", () => {
    it("returns the smaller value", async () => {
      expect(await contract.min(1n, 2n)).to.equal(1n);
      expect(await contract.min(5n, 3n)).to.equal(3n);
      expect(await contract.min(4n, 4n)).to.equal(4n);
    });
  });

  describe("max", () => {
    it("returns the larger value", async () => {
      expect(await contract.max(1n, 2n)).to.equal(2n);
      expect(await contract.max(5n, 3n)).to.equal(5n);
      expect(await contract.max(4n, 4n)).to.equal(4n);
    });
  });

  describe("safeSub", () => {
    it("subtracts when the minuend is greater", async () => {
      expect(await contract.safeSub(10n, 3n)).to.equal(7n);
      expect(await contract.safeSub(MaxUint256, 1n)).to.equal(MaxUint256 - 1n);
    });

    it("saturates at zero otherwise", async () => {
      expect(await contract.safeSub(3n, 10n)).to.equal(0n);
      expect(await contract.safeSub(5n, 5n)).to.equal(0n);
      expect(await contract.safeSub(0n, 1n)).to.equal(0n);
    });
  });

  describe("nextPow2", () => {
    it("returns 1 for input 0", async () => {
      expect(await contract.nextPow2(0n)).to.equal(1n);
    });

    it("returns the same value for powers of 2", async () => {
      expect(await contract.nextPow2(1n)).to.equal(1n);
      expect(await contract.nextPow2(2n)).to.equal(2n);
      expect(await contract.nextPow2(4n)).to.equal(4n);
      expect(await contract.nextPow2(8n)).to.equal(8n);
      expect(await contract.nextPow2(16n)).to.equal(16n);
      expect(await contract.nextPow2(256n)).to.equal(256n);
      expect(await contract.nextPow2(2n ** 128n)).to.equal(2n ** 128n);
      expect(await contract.nextPow2(2n ** 255n)).to.equal(2n ** 255n);
    });

    it("returns the next power of 2 for values between powers", async () => {
      expect(await contract.nextPow2(3n)).to.equal(4n);
      expect(await contract.nextPow2(5n)).to.equal(8n);
      expect(await contract.nextPow2(7n)).to.equal(8n);
      expect(await contract.nextPow2(9n)).to.equal(16n);
      expect(await contract.nextPow2(15n)).to.equal(16n);
      expect(await contract.nextPow2(17n)).to.equal(32n);
      expect(await contract.nextPow2(100n)).to.equal(128n);
      expect(await contract.nextPow2(255n)).to.equal(256n);
      expect(await contract.nextPow2(2n ** 128n + 1n)).to.equal(2n ** 129n);
    });

    it("handles large values correctly", async () => {
      expect(await contract.nextPow2(2n ** 255n - 1n)).to.equal(2n ** 255n);
      expect(await contract.nextPow2(2n ** 254n + 1n)).to.equal(2n ** 255n);
      // MaxUint256 will overflow, but the function should still execute
      // The result will be 0 due to overflow (2^256 wraps to 0)
      expect(await contract.nextPow2(MaxUint256)).to.equal(0n);
    });
  });
});
