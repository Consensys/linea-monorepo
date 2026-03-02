import { networkHelpers } from "../common/connection.js";
const { loadFixture } = networkHelpers;

import { expect } from "chai";
import type { TestEfficientLeftRightKeccak } from "../../../typechain-types/index.js";
import { deployFromFactory } from "../common/deployment.js";
import { generateKeccak256, generateRandomBytes } from "../common/helpers/index.js";

describe("EfficientLeftRightKeccak Library", () => {
  let contract: TestEfficientLeftRightKeccak;

  async function deployTestEfficientLeftRightKeccakFixture() {
    return deployFromFactory("TestEfficientLeftRightKeccak");
  }
  beforeEach(async () => {
    contract = (await loadFixture(deployTestEfficientLeftRightKeccakFixture)) as TestEfficientLeftRightKeccak;
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
