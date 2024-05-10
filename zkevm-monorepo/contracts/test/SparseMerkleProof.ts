import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { Mimc, SparseMerkleProof } from "../typechain-types";
import merkleProofTestData from "./testData/merkle-proof-data.json";
import { deployFromFactory } from "./utils/deployment";

describe("SparseMerkleProof", () => {
  let sparseMerkleProof: SparseMerkleProof;

  async function deploySparseMerkleProofFixture() {
    const mimc = (await deployFromFactory("Mimc")) as Mimc;
    const factory = await ethers.getContractFactory("SparseMerkleProof", {
      libraries: { Mimc: await mimc.getAddress() },
    });
    const sparseMerkleProof = await factory.deploy();
    await sparseMerkleProof.waitForDeployment();
    return sparseMerkleProof;
  }

  beforeEach(async () => {
    sparseMerkleProof = await loadFixture(deploySparseMerkleProofFixture);
  });

  describe("verifyProof", () => {
    describe("account proof", () => {
      it("Should return false when the account proof is not correct", async () => {
        const {
          accountProof: {
            proof: { proofRelatedNodes },
          },
        } = merkleProofTestData;

        const stateRoot = "0x0e080582960965e3c180b1457b16da48041e720af628ae6c1725d13bd98ba9f0";
        const leafIndex = 200;
        const result = await sparseMerkleProof.verifyProof(proofRelatedNodes, leafIndex, stateRoot);

        expect(result).to.be.false;
      });

      it("Should return true when the account proof is correct", async () => {
        const {
          accountProof: {
            proof: { proofRelatedNodes },
            leafIndex,
          },
        } = merkleProofTestData;
        const stateRoot = "0x0e080582960965e3c180b1457b16da48041e720af628ae6c1725d13bd98ba9f0";

        const result = await sparseMerkleProof.verifyProof(proofRelatedNodes, leafIndex, stateRoot);

        expect(result).to.be.true;
      });
    });

    describe("storage proof", () => {
      it("Should return false when the storage proof is not correct", async () => {
        const {
          storageProofs: [
            {
              proof: { proofRelatedNodes },
            },
          ],
        } = merkleProofTestData;

        const stateRoot = "0x0d2a66d5598b4fc5482c311f22d2dc657579b5452ab4b3e60fb1a9e9dbbfc99e";
        const leafIndex = 200;
        const result = await sparseMerkleProof.verifyProof(proofRelatedNodes, leafIndex, stateRoot);

        expect(result).to.be.false;
      });

      it("Should return true when the storage proof is correct", async () => {
        const { storageProofs } = merkleProofTestData;
        const stateRoot = "0x0d2a66d5598b4fc5482c311f22d2dc657579b5452ab4b3e60fb1a9e9dbbfc99e";

        for (const {
          proof: { proofRelatedNodes },
          leafIndex,
        } of storageProofs) {
          const result = await sparseMerkleProof.verifyProof(proofRelatedNodes, leafIndex, stateRoot);
          expect(result).to.be.true;
        }
      });
    });
  });

  describe("hashAccountValue", () => {
    it("Should return value hash", async () => {
      const {
        accountProof: {
          proof: { value },
        },
      } = merkleProofTestData;

      const hVal = await sparseMerkleProof.hashAccountValue(value);
      expect(hVal).to.be.equal("0x05d9557beb35be64f9f0be17af76dd4f19d5016b4108ce8a552458dcf8ec6d4b");
    });
  });

  describe("hashStorageValue", () => {
    it("Should return value hash", async () => {
      const {
        storageProofs: [
          {
            proof: { value },
          },
        ],
      } = merkleProofTestData;
      const hVal = await sparseMerkleProof.hashStorageValue(value);
      expect(hVal).to.be.equal("0x0cbfc04518a70ed18917be26dbd0efdf4c1f9f3def6d5de2b0bbd5c82b5e9c2d");
    });
  });

  describe("getLeaf", () => {
    describe("account leaf", () => {
      it("Should revert when leaf bytes length < 128", async () => {
        const {
          accountProof: {
            proof: { proofRelatedNodes },
          },
        } = merkleProofTestData;

        const wrongLeaftValue = `0x${proofRelatedNodes[proofRelatedNodes.length - 1].slice(4)}`;

        await expect(sparseMerkleProof.getLeaf(wrongLeaftValue))
          .to.revertedWithCustomError(sparseMerkleProof, "WrongBytesLength")
          .withArgs(128, ethers.dataLength(wrongLeaftValue));
      });

      it("Should return parsed leaf", async () => {
        const {
          accountProof: {
            proof: { proofRelatedNodes },
          },
        } = merkleProofTestData;

        const leaf = await sparseMerkleProof.getLeaf(proofRelatedNodes[proofRelatedNodes.length - 1]);

        expect(leaf.prev).to.be.equal(17750);
        expect(leaf.next).to.be.equal(13571);
        expect(leaf.hKey).to.be.equal("0x04f760ecc308f2f05e90be61b302e78e046595681e1a17c054ee417ffe5ac310");
        expect(leaf.hValue).to.be.equal("0x05d9557beb35be64f9f0be17af76dd4f19d5016b4108ce8a552458dcf8ec6d4b");
      });
    });

    describe("storage leaf", () => {
      it("Should revert when leaf bytes length < 128", async () => {
        const {
          storageProofs: [
            {
              proof: { proofRelatedNodes },
            },
          ],
        } = merkleProofTestData;

        const wrongLeaftValue = `0x${proofRelatedNodes[proofRelatedNodes.length - 1].slice(4)}`;

        await expect(sparseMerkleProof.getLeaf(wrongLeaftValue))
          .to.revertedWithCustomError(sparseMerkleProof, "WrongBytesLength")
          .withArgs(128, ethers.dataLength(wrongLeaftValue));
      });

      it("Should return parsed leaf", async () => {
        const {
          storageProofs: [
            {
              proof: { proofRelatedNodes },
            },
          ],
        } = merkleProofTestData;

        const leaf = await sparseMerkleProof.getLeaf(proofRelatedNodes[proofRelatedNodes.length - 1]);

        expect(leaf.prev).to.be.equal(2);
        expect(leaf.next).to.be.equal(5);
        expect(leaf.hKey).to.be.equal("0x01d265eebbf22fa41219519f676c05b01abaa5aca7abdb186422044c2971c80a");
        expect(leaf.hValue).to.be.equal("0x0cbfc04518a70ed18917be26dbd0efdf4c1f9f3def6d5de2b0bbd5c82b5e9c2d");
      });
    });
  });

  describe("getAccount", () => {
    it("Should revert when account bytes length < 192", async () => {
      const {
        accountProof: {
          proof: { value },
        },
      } = merkleProofTestData;

      const wrongAccountValue = `0x${value.slice(4)}`;

      await expect(sparseMerkleProof.getAccount(wrongAccountValue))
        .to.revertedWithCustomError(sparseMerkleProof, "WrongBytesLength")
        .withArgs(192, ethers.dataLength(wrongAccountValue));
    });

    it("Should return parsed leaf", async () => {
      const {
        accountProof: {
          proof: { value },
        },
      } = merkleProofTestData;

      const account = await sparseMerkleProof.getAccount(value);

      expect(account.nonce).to.be.equal(1);
      expect(account.balance).to.be.equal(0);
      expect(account.storageRoot).to.be.equal("0x0d2a66d5598b4fc5482c311f22d2dc657579b5452ab4b3e60fb1a9e9dbbfc99e");
      expect(account.mimcCodeHash).to.be.equal("0x00c24dd0468f02fbece668291f3c3eb20e06d1baec856f28430555967f2bf280");
      expect(account.keccakCodeHash).to.be.equal("0xd798c662debc23e8199fbf0b0a3a95649f2defe90af458d7f62c03881f916b3f");
      expect(account.codeSize).to.be.equal(12336);
    });
  });

  describe("mimcHash", () => {
    it("Should return mimc hash", async () => {
      const {
        accountProof: {
          proof: { value },
        },
      } = merkleProofTestData;

      const hashedValue = await sparseMerkleProof.mimcHash(value);

      expect(hashedValue).to.be.equal("0x0b99084c8c6234d0765eca2bc776c654f8b62ffc0aa3d0990b69e2aef431e732");
    });
  });
});
