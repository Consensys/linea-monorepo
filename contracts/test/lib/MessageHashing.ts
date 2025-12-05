import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestMessageHashing } from "../../typechain-types";
import { deployFromFactory } from "../common/deployment";

describe("MessageHashing Library", () => {
  let contract: TestMessageHashing;

  async function deployTestMessageHashingFixture() {
    return deployFromFactory("TestMessageHashing");
  }

  beforeEach(async () => {
    contract = (await loadFixture(deployTestMessageHashingFixture)) as TestMessageHashing;
  });

  describe("hashMessageWithEmptyCalldata", () => {
    it("Should return the same hash as hashMessage with empty calldata", async () => {
      const [from, to] = await ethers.getSigners();
      const fee = ethers.parseEther("0.001");
      const valueSent = ethers.parseEther("1");
      const messageNumber = 12345n;
      const emptyCalldata = "0x";

      const hashWithEmptyCalldata = await contract.hashMessageWithEmptyCalldata(
        from.address,
        to.address,
        fee,
        valueSent,
        messageNumber
      );

      const hashWithMessage = await contract.hashMessage(
        from.address,
        to.address,
        fee,
        valueSent,
        messageNumber,
        emptyCalldata
      );

      expect(hashWithEmptyCalldata).to.equal(hashWithMessage);
    });

    it("Should return the same hash as hashMessage with empty calldata for different parameters", async () => {
      const [from, to] = await ethers.getSigners();
      const fee = 1000n;
      const valueSent = 5000n;
      const messageNumber = 999999n;
      const emptyCalldata = "0x";

      const hashWithEmptyCalldata = await contract.hashMessageWithEmptyCalldata(
        from.address,
        to.address,
        fee,
        valueSent,
        messageNumber
      );

      const hashWithMessage = await contract.hashMessage(
        from.address,
        to.address,
        fee,
        valueSent,
        messageNumber,
        emptyCalldata
      );

      expect(hashWithEmptyCalldata).to.equal(hashWithMessage);
    });

    it("Should return the same hash as hashMessage with empty calldata for zero values", async () => {
      const [from, to] = await ethers.getSigners();
      const fee = 0n;
      const valueSent = 0n;
      const messageNumber = 0n;
      const emptyCalldata = "0x";

      const hashWithEmptyCalldata = await contract.hashMessageWithEmptyCalldata(
        from.address,
        to.address,
        fee,
        valueSent,
        messageNumber
      );

      const hashWithMessage = await contract.hashMessage(
        from.address,
        to.address,
        fee,
        valueSent,
        messageNumber,
        emptyCalldata
      );

      expect(hashWithEmptyCalldata).to.equal(hashWithMessage);
    });

    it("Should return the same hash as hashMessage with empty calldata for maximum values", async () => {
      const [from, to] = await ethers.getSigners();
      const fee = ethers.MaxUint256;
      const valueSent = ethers.MaxUint256;
      const messageNumber = ethers.MaxUint256;
      const emptyCalldata = "0x";

      const hashWithEmptyCalldata = await contract.hashMessageWithEmptyCalldata(
        from.address,
        to.address,
        fee,
        valueSent,
        messageNumber
      );

      const hashWithMessage = await contract.hashMessage(
        from.address,
        to.address,
        fee,
        valueSent,
        messageNumber,
        emptyCalldata
      );

      expect(hashWithEmptyCalldata).to.equal(hashWithMessage);
    });
  });
});

