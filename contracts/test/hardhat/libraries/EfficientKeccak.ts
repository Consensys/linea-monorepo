import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { TestEfficientKeccak } from "../../../typechain-types";
import { deployFromFactory } from "../common/deployment";
import { generateKeccak256, generateRandomBytes } from "../common/helpers";

describe("EfficientKeccak Library", () => {
  let contract: TestEfficientKeccak;

  async function deployTestEfficientLeftRightKeccakFixture() {
    return deployFromFactory("TestEfficientKeccak");
  }
  beforeEach(async () => {
    contract = (await loadFixture(deployTestEfficientLeftRightKeccakFixture)) as TestEfficientKeccak;
  });

  describe("efficientKeccak", () => {
    it("Should return the correct keccak hash for left/right", async () => {
      const leftValue = generateRandomBytes(32);
      const rightValue = generateRandomBytes(32);
      const solidityKeccakHash = generateKeccak256(["bytes32", "bytes32"], [leftValue, rightValue]);
      expect(await contract.efficientKeccakLeftRight(leftValue, rightValue)).to.equal(solidityKeccakHash);
    });

    it("Should return the correct keccak hash for 5 values", async () => {
      const v1 = generateRandomBytes(32);
      const v2 = generateRandomBytes(32);
      const v3 = generateRandomBytes(32);
      const v4 = generateRandomBytes(32);
      const v5 = generateRandomBytes(32);
      const solidityKeccakHash = generateKeccak256(
        ["bytes32", "bytes32", "bytes32", "bytes32", "bytes32"],
        [v1, v2, v3, v4, v5],
      );
      expect(await contract.efficientKeccak(v1, v2, v3, v4, v5)).to.equal(solidityKeccakHash);
    });

    it("Should return the correct keccak hash for some uint values (soundness alert state)", async () => {
      const v1 = generateRandomBytes(32);
      const v2 = 123;
      const v3 = 456;
      const v4 = generateRandomBytes(32);
      const v5 = 789;

      const solidityKeccakHashWithUints = generateKeccak256(
        ["bytes32", "uint256", "uint256", "bytes32", "uint256"],
        [v1, v2, v3, v4, v5],
      );

      expect(await contract.efficientKeccakWithSomeUints(v1, v2, v3, v4, v5)).to.equal(solidityKeccakHashWithUints);
    });
  });
});
