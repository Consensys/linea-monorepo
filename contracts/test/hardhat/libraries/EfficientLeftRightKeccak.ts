import { expect } from "chai";

import { deployFromFactory } from "../common/deployment";
import { generateKeccak256, generateRandomBytes } from "../common/helpers";

import type { TestEfficientLeftRightKeccak } from "../../../typechain-types";

import { loadFixture } from "#hardhat-network-helpers";

describe("EfficientLeftRightKeccak Library", () => {
  let contract: TestEfficientLeftRightKeccak;

  async function deployTestEfficientLeftRightKeccakFixture(): Promise<TestEfficientLeftRightKeccak> {
    return deployFromFactory<TestEfficientLeftRightKeccak>("TestEfficientLeftRightKeccak");
  }
  beforeEach(async () => {
    contract = await loadFixture(deployTestEfficientLeftRightKeccakFixture);
  });

  describe("efficientKeccak", () => {
    it("Should return the correct keccak hash", async () => {
      const leftValue = generateRandomBytes(32);
      const rightValue = generateRandomBytes(32);
      const solidityKeccakHash = generateKeccak256(["bytes32", "bytes32"], [leftValue, rightValue]);
      expect(await contract.efficientKeccak(leftValue, rightValue)).to.equal(solidityKeccakHash);
    });
  });
});
