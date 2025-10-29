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
});
