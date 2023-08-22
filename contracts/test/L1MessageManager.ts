import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { TestL1MessageManager } from "../typechain-types";
import {
  INBOX_STATUS_RECEIVED,
  INBOX_STATUS_UNKNOWN,
  OUTBOX_STATUS_RECEIVED,
  OUTBOX_STATUS_SENT,
} from "./utils/constants";
import { deployFromFactory } from "./utils/deployment";
import { generateKeccak256Hash, generateNKeccak256Hashes } from "./utils/helpers";

describe("L1MessageManager", () => {
  let l1MessageManager: TestL1MessageManager;

  async function deployTestL1MessageManagerFixture(): Promise<TestL1MessageManager> {
    return deployFromFactory("TestL1MessageManager") as Promise<TestL1MessageManager>;
  }

  beforeEach(async () => {
    l1MessageManager = await loadFixture(deployTestL1MessageManagerFixture);
  });

  describe("Add L2->L1 message hash in 'inboxL2L1MessageStatus' mapping", () => {
    it("Should revert if the message hash already exists in 'inboxL2L1MessageStatus' mapping", async () => {
      const messageHash = generateKeccak256Hash("message1");
      await l1MessageManager.addL2L1MessageHash(messageHash);

      await expect(l1MessageManager.addL2L1MessageHash(messageHash))
        .to.be.revertedWithCustomError(l1MessageManager, "MessageAlreadyReceived")
        .withArgs(messageHash);
    });

    it("Should succeed if message hash does not exist in 'inboxL2L1MessageStatus' mapping", async () => {
      const messageHash = generateKeccak256Hash("message1");
      await l1MessageManager.addL2L1MessageHash(messageHash);

      expect(await l1MessageManager.inboxL2L1MessageStatus(messageHash)).to.equal(INBOX_STATUS_RECEIVED);
    });

    it("Should emit an event 'L2L1MessageHashAddedToInbox' when succeed", async () => {
      const messageHash = generateKeccak256Hash("message1");

      await expect(l1MessageManager.addL2L1MessageHash(messageHash))
        .to.emit(l1MessageManager, "L2L1MessageHashAddedToInbox")
        .withArgs(messageHash);
    });
  });

  describe("Update L2->L1 message hash status in 'inboxL2L1MessageStatus' mapping to 'claimed'", () => {
    it("Should revert if the message hash has not the status 'received' in 'inboxL2L1MessageStatus' mapping", async () => {
      const messageHash = generateKeccak256Hash("message1");

      await l1MessageManager.addL2L1MessageHash(messageHash);
      await l1MessageManager.updateL2L1MessageStatusToClaimed(messageHash);

      await expect(l1MessageManager.updateL2L1MessageStatusToClaimed(messageHash)).to.be.revertedWithCustomError(
        l1MessageManager,
        "MessageDoesNotExistOrHasAlreadyBeenClaimed",
      );
    });

    it("Should succeed if message hash has the status 'received' in 'inboxL2L1MessageStatus' mapping", async () => {
      const messageHash = generateKeccak256Hash("message1");

      await l1MessageManager.addL2L1MessageHash(messageHash);
      await l1MessageManager.updateL2L1MessageStatusToClaimed(messageHash);

      expect(await l1MessageManager.inboxL2L1MessageStatus(messageHash)).to.equal(INBOX_STATUS_UNKNOWN);
    });
  });

  describe("Add L1->L2 message hash in 'outboxL1L2MessageStatus' mapping", () => {
    it("Should succeed if the message hash does not exists in 'outboxL1L2MessageStatus' mapping", async () => {
      const messageHash = generateKeccak256Hash("message1");
      await l1MessageManager.addL1L2MessageHash(messageHash);
      expect(await l1MessageManager.outboxL1L2MessageStatus(messageHash)).to.equal(OUTBOX_STATUS_SENT);
    });
  });

  describe("Update L1->L2 message hashes in 'outboxL1L2MessageStatus' mapping to 'received'", () => {
    it("Should revert if one of the message hashes has the status 'unknown'", async () => {
      const messageHashes = generateNKeccak256Hashes("message", 2);

      await expect(l1MessageManager.updateL1L2MessageStatusToReceived(messageHashes))
        .to.be.revertedWithCustomError(l1MessageManager, "L1L2MessageNotSent")
        .withArgs(messageHashes[0]);
    });

    it("Should succeed if the message hash exists in 'outboxL1L2MessageStatus' mapping and has the status 'sent'", async () => {
      const messageHashes = generateNKeccak256Hashes("message", 100);
      for (const messageHash of messageHashes) {
        await l1MessageManager.addL1L2MessageHash(messageHash);
      }

      await l1MessageManager.updateL1L2MessageStatusToReceived(messageHashes);
      for (const messageHash of messageHashes) {
        expect(await l1MessageManager.outboxL1L2MessageStatus(messageHash)).to.equal(OUTBOX_STATUS_RECEIVED);
      }
    });

    it("Should emit an event 'L1L2MessagesReceivedOnL2' when succeed", async () => {
      const messageHashes = generateNKeccak256Hashes("message", 2);
      for (const messageHash of messageHashes) {
        await l1MessageManager.addL1L2MessageHash(messageHash);
      }

      await expect(l1MessageManager.updateL1L2MessageStatusToReceived(messageHashes))
        .to.emit(l1MessageManager, "L1L2MessagesReceivedOnL2")
        .withArgs(messageHashes);
    });
  });
});
