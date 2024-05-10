import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { Mimc } from "../typechain-types";
import mimcTestData from "./testData/mimc-test-data.json";
import { deployFromFactory } from "./utils/deployment";

describe("Mimc", () => {
  let mimc: Mimc;

  async function deployMIMCFixture() {
    return deployFromFactory("Mimc") as Promise<Mimc>;
  }

  beforeEach(async () => {
    mimc = await loadFixture(deployMIMCFixture);
  });

  describe("hash", () => {
    it("Should return mimc hash for each test case", async () => {
      for (const element of mimcTestData) {
        const msgs = ethers.concat(element.in);
        expect(await mimc.hash(msgs)).to.equal(element.out);
      }
    });
  });
});
