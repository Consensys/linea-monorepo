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

  describe("bitLength", () => {
    it("returns 0 for input 0", async () => {
      expect(await contract.bitLength(0n)).to.equal(0n);
    });

    it("returns correct bit length for powers of 2", async () => {
      expect(await contract.bitLength(1n)).to.equal(1n);
      expect(await contract.bitLength(2n)).to.equal(2n);
      expect(await contract.bitLength(4n)).to.equal(3n);
      expect(await contract.bitLength(8n)).to.equal(4n);
      expect(await contract.bitLength(16n)).to.equal(5n);
      expect(await contract.bitLength(32n)).to.equal(6n);
      expect(await contract.bitLength(64n)).to.equal(7n);
      expect(await contract.bitLength(128n)).to.equal(8n);
      expect(await contract.bitLength(256n)).to.equal(9n);
      expect(await contract.bitLength(2n ** 16n)).to.equal(17n);
      expect(await contract.bitLength(2n ** 32n)).to.equal(33n);
      expect(await contract.bitLength(2n ** 64n)).to.equal(65n);
      expect(await contract.bitLength(2n ** 128n)).to.equal(129n);
      expect(await contract.bitLength(2n ** 255n)).to.equal(256n);
    });

    it("returns correct bit length for values between powers of 2", async () => {
      expect(await contract.bitLength(3n)).to.equal(2n);
      expect(await contract.bitLength(5n)).to.equal(3n);
      expect(await contract.bitLength(7n)).to.equal(3n);
      expect(await contract.bitLength(9n)).to.equal(4n);
      expect(await contract.bitLength(15n)).to.equal(4n);
      expect(await contract.bitLength(17n)).to.equal(5n);
      expect(await contract.bitLength(31n)).to.equal(5n);
      expect(await contract.bitLength(33n)).to.equal(6n);
      expect(await contract.bitLength(100n)).to.equal(7n);
      expect(await contract.bitLength(255n)).to.equal(8n);
      expect(await contract.bitLength(256n)).to.equal(9n);
      expect(await contract.bitLength(257n)).to.equal(9n);
      expect(await contract.bitLength(2n ** 128n + 1n)).to.equal(129n);
      expect(await contract.bitLength(2n ** 255n - 1n)).to.equal(255n);
    });

    it("handles edge cases correctly", async () => {
      // MaxUint256 has all 256 bits set, so bit length is 256
      expect(await contract.bitLength(MaxUint256)).to.equal(256n);
      // 2^255 - 1 has bits 0-254 set, so bit length is 255
      expect(await contract.bitLength(2n ** 255n - 1n)).to.equal(255n);
      // 2^128 - 1 has bits 0-127 set, so bit length is 128
      expect(await contract.bitLength(2n ** 128n - 1n)).to.equal(128n);
      // 2^64 - 1 has bits 0-63 set, so bit length is 64
      expect(await contract.bitLength(2n ** 64n - 1n)).to.equal(64n);
      // 2^32 - 1 has bits 0-31 set, so bit length is 32
      expect(await contract.bitLength(2n ** 32n - 1n)).to.equal(32n);
    });

    it("returns correct bit length for various values across different bit ranges", async () => {
      // Small values
      expect(await contract.bitLength(1n)).to.equal(1n);
      expect(await contract.bitLength(127n)).to.equal(7n);
      expect(await contract.bitLength(128n)).to.equal(8n);
      expect(await contract.bitLength(255n)).to.equal(8n);

      // Medium values
      expect(await contract.bitLength(1000n)).to.equal(10n);
      expect(await contract.bitLength(10000n)).to.equal(14n);
      expect(await contract.bitLength(100000n)).to.equal(17n);
      expect(await contract.bitLength(1000000n)).to.equal(20n);

      // Large values
      expect(await contract.bitLength(2n ** 192n)).to.equal(193n);
      expect(await contract.bitLength(2n ** 192n + 1n)).to.equal(193n);
      expect(await contract.bitLength(2n ** 200n)).to.equal(201n);
      expect(await contract.bitLength(2n ** 254n)).to.equal(255n);
      expect(await contract.bitLength(2n ** 254n + 1n)).to.equal(255n);
    });

    it("verifies bit length matches mathematical definition", async () => {
      // For non-zero x, bitLength(x) = ceil(log2(x + 1))
      // This is equivalent to: position of highest set bit + 1
      const testCases = [
        1n,
        2n,
        3n,
        4n,
        5n,
        7n,
        8n,
        15n,
        16n,
        31n,
        32n,
        63n,
        64n,
        127n,
        128n,
        255n,
        256n,
        1000n,
        10000n,
        100000n,
        2n ** 16n - 1n,
        2n ** 16n,
        2n ** 16n + 1n,
        2n ** 32n - 1n,
        2n ** 32n,
        2n ** 32n + 1n,
        2n ** 64n - 1n,
        2n ** 64n,
        2n ** 64n + 1n,
        2n ** 128n - 1n,
        2n ** 128n,
        2n ** 128n + 1n,
        2n ** 255n - 1n,
        2n ** 255n,
        MaxUint256,
      ];

      for (const x of testCases) {
        const result = await contract.bitLength(x);
        if (x === 0n) {
          expect(result).to.equal(0n);
        } else {
          // Verify result is positive and reasonable
          expect(result).to.be.greaterThan(0n);
          expect(result).to.be.lessThanOrEqual(256n);

          // Verify: 2^(result-1) <= x < 2^result
          const lowerBound = 2n ** (result - 1n);
          const upperBound = 2n ** result;
          expect(x).to.be.greaterThanOrEqual(lowerBound);
          expect(x).to.be.lessThan(upperBound);
        }
      }
    });
  });
});
