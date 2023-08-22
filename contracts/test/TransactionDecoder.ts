import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestTransactionDecoder } from "../typechain-types";
import { getRLPEncodeTransactions } from "./utils/helpers";

describe("TransactionDecoder", () => {
  let contract: TestTransactionDecoder;
  const {
    eip1559Transaction,
    eip1559TransactionHashes,
    legacyTransaction,
    legacyTransactionHashes,
    eip2930Transaction,
    eip2930TransactionHashes,
  } = getRLPEncodeTransactions("test-transactions.json");

  async function deployTestTransactionDecoderFixture() {
    const factory = await ethers.getContractFactory("TestTransactionDecoder");
    contract = await factory.deploy();
    return contract;
  }

  beforeEach(async () => {
    contract = await loadFixture(deployTestTransactionDecoderFixture);
  });

  describe("Decoding transactions", () => {
    it("Should decode the hashes from an EIP1559 transaction", async () => {
      const hashes = await contract.decodeTransactionAndHashes(eip1559Transaction);
      expect(hashes).to.deep.equal(eip1559TransactionHashes);
    });

    it("Should decode the hashes from the legacy transaction", async () => {
      const hashes = await contract.decodeTransactionAndHashes(legacyTransaction);
      expect(hashes).to.deep.equal(legacyTransactionHashes);
    });

    it("Should decode the hashes from and EIP2930 transaction", async () => {
      const hashes = await contract.decodeTransactionAndHashes(eip2930Transaction);
      expect(hashes).to.deep.equal(eip2930TransactionHashes);
    });

    it("Should revert with too short data", async () => {
      await expect(contract.decodeTransactionAndHashes("0x")).to.be.reverted;
    });

    it("Should revert with unknown tx type", async () => {
      await expect(contract.decodeTransactionAndHashes("0x03")).to.be.reverted;
    });
  });
});
