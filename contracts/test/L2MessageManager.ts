import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestL2MessageManager } from "../typechain-types";
import {
  DEFAULT_ADMIN_ROLE,
  INBOX_STATUS_CLAIMED,
  INBOX_STATUS_RECEIVED,
  L1_L2_MESSAGE_SETTER_ROLE,
} from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";
import { generateKeccak256Hash, generateNKeccak256Hashes } from "./utils/helpers";

describe("L2MessageManager", () => {
  let l2MessageManager: TestL2MessageManager;
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let admin: SignerWithAddress;
  let pauser: SignerWithAddress;
  let l1l2MessageSetter: SignerWithAddress;
  let notAuthorizedAccount: SignerWithAddress;

  async function deployL2MessageManagerFixture() {
    return deployUpgradableFromFactory("TestL2MessageManager", [
      pauser.address,
      l1l2MessageSetter.address,
    ]) as Promise<TestL2MessageManager>;
  }

  beforeEach(async () => {
    [admin, pauser, l1l2MessageSetter, notAuthorizedAccount] = await ethers.getSigners();
    l2MessageManager = await loadFixture(deployL2MessageManagerFixture);
  });

  describe("Initialization checks", () => {
    it("Deployer has default admin role", async () => {
      expect(await l2MessageManager.hasRole(DEFAULT_ADMIN_ROLE, admin.address)).to.be.true;
    });

    it("It should fail when not initializing", async () => {
      await expect(l2MessageManager.tryInitialize(admin.address)).to.be.revertedWith(
        "Initializable: contract is not initializing",
      );
    });
  });

  describe("Add L1->L2 message hashes in 'inboxL1L2MessageStatus'", () => {
    it("Should revert if the caller does not have the role 'L1_L2_MESSAGE_SETTER_ROLE'", async () => {
      const messageHashes = generateNKeccak256Hashes("message", 2);

      await expect(
        l2MessageManager.connect(notAuthorizedAccount).addL1L2MessageHashes(messageHashes),
      ).to.be.revertedWith(
        `AccessControl: account ${notAuthorizedAccount.address.toLowerCase()} is missing role ${L1_L2_MESSAGE_SETTER_ROLE}`,
      );
    });

    it("Should revert if message hashes array length is higher than one hundred", async () => {
      const messageHashes = generateNKeccak256Hashes("message", 101);

      await expect(
        l2MessageManager.connect(l1l2MessageSetter).addL1L2MessageHashes(messageHashes),
      ).to.be.revertedWithCustomError(l2MessageManager, "MessageHashesListLengthHigherThanOneHundred");
    });

    it("Should succeed if message hashes array param is correct", async () => {
      const messageHashes = generateNKeccak256Hashes("message", 100);
      await l2MessageManager.connect(l1l2MessageSetter).addL1L2MessageHashes(messageHashes);
      for (const messageHash of messageHashes) {
        expect(await l2MessageManager.connect(l1l2MessageSetter).inboxL1L2MessageStatus(messageHash)).to.equal(
          INBOX_STATUS_RECEIVED,
        );
      }
    });

    it("Should succeed if duplicates hashes exist in the array", async () => {
      const messageHashes = [
        ethers.constants.HashZero,
        ethers.constants.HashZero,
        generateKeccak256Hash("message1"),
        generateKeccak256Hash("message1"),
      ];
      await l2MessageManager.connect(l1l2MessageSetter).addL1L2MessageHashes(messageHashes);
      for (const messageHash of messageHashes) {
        expect(await l2MessageManager.connect(l1l2MessageSetter).inboxL1L2MessageStatus(messageHash)).to.equal(
          INBOX_STATUS_RECEIVED,
        );
      }
    });

    it("Should emit an event 'L1L2MessageHashesAddedToInbox' when succeed", async () => {
      const messageHashes = generateNKeccak256Hashes("message", 50);

      await expect(l2MessageManager.connect(l1l2MessageSetter).addL1L2MessageHashes(messageHashes))
        .to.emit(l2MessageManager, "L1L2MessageHashesAddedToInbox")
        .withArgs(messageHashes);
    });
  });

  describe("Update L1->L2 message status to 'claimed' in 'inboxL1L2MessageStatus'", () => {
    it("Should revert if the message hash has not the status 'received' in 'inboxL1L2MessageStatus' mapping", async () => {
      const messageHash = generateKeccak256Hash("message");
      await expect(l2MessageManager.updateL1L2MessageStatusToClaimed(messageHash)).to.be.revertedWithCustomError(
        l2MessageManager,
        "MessageDoesNotExistOrHasAlreadyBeenClaimed",
      );
    });

    it("Should succeed if message hash has the status 'received' in 'inboxL1L2MessageStatus' mapping", async () => {
      const messageHash = generateKeccak256Hash("message1");
      const messageHashes = generateNKeccak256Hashes("message", 50);

      await l2MessageManager.connect(l1l2MessageSetter).addL1L2MessageHashes(messageHashes);
      await l2MessageManager.updateL1L2MessageStatusToClaimed(messageHash);

      expect(await l2MessageManager.inboxL1L2MessageStatus(messageHash)).to.equal(INBOX_STATUS_CLAIMED);
    });
  });
});
