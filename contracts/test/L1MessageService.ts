import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { TestL1MessageService, TestReceivingContract } from "../typechain-types";
import {
  DEFAULT_ADMIN_ROLE,
  EMPTY_CALLDATA,
  GENERAL_PAUSE_TYPE,
  INBOX_STATUS_RECEIVED,
  INBOX_STATUS_UNKNOWN,
  INITIAL_WITHDRAW_LIMIT,
  L1_L2_PAUSE_TYPE,
  L2_L1_PAUSE_TYPE,
  LOW_NO_REFUND_MESSAGE_FEE,
  MESSAGE_FEE,
  MESSAGE_VALUE_1ETH,
  ONE_DAY_IN_SECONDS,
  OUTBOX_STATUS_SENT,
  PAUSE_MANAGER_ROLE,
  RATE_LIMIT_SETTER_ROLE,
} from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";
import { encodeSendMessage, generateKeccak256Hash } from "./utils/helpers";

describe("L1MessageService", () => {
  let l1MessageService: TestL1MessageService;
  let admin: SignerWithAddress;
  let pauser: SignerWithAddress;
  let limitSetter: SignerWithAddress;
  let notAuthorizedAccount: SignerWithAddress;
  let postmanAddress: SignerWithAddress;

  async function deployTestL1MessageServiceFixture(): Promise<TestL1MessageService> {
    return deployUpgradableFromFactory("TestL1MessageService", [
      limitSetter.address,
      pauser.address,
      ONE_DAY_IN_SECONDS,
      INITIAL_WITHDRAW_LIMIT,
    ]) as Promise<TestL1MessageService>;
  }

  beforeEach(async () => {
    [admin, pauser, limitSetter, notAuthorizedAccount, postmanAddress] = await ethers.getSigners();
    l1MessageService = await loadFixture(deployTestL1MessageServiceFixture);
  });

  describe("Initialisation tests", () => {
    it("Deployer has default admin role", async () => {
      expect(await l1MessageService.hasRole(DEFAULT_ADMIN_ROLE, admin.address)).to.be.true;
    });

    it("limitSetter has RATE_LIMIT_SETTER_ROLE", async () => {
      expect(await l1MessageService.hasRole(RATE_LIMIT_SETTER_ROLE, limitSetter.address)).to.be.true;
    });

    it("pauser has PAUSE_MANAGER_ROLE", async () => {
      expect(await l1MessageService.hasRole(PAUSE_MANAGER_ROLE, pauser.address)).to.be.true;
    });

    it("Should set rate limit and period", async () => {
      expect(await l1MessageService.periodInSeconds()).to.be.equal(ONE_DAY_IN_SECONDS);
      expect(await l1MessageService.limitInWei()).to.be.equal(INITIAL_WITHDRAW_LIMIT);
    });

    it("It should fail when not initializing", async () => {
      await expect(
        l1MessageService.tryInitialize(limitSetter.address, pauser.address, ONE_DAY_IN_SECONDS, INITIAL_WITHDRAW_LIMIT),
      ).to.be.revertedWith("Initializable: contract is not initializing");
    });

    it("Should initialise nextMessageNumber", async () => {
      expect(await l1MessageService.nextMessageNumber()).to.be.equal(1);
    });

    it("Should fail to deploy missing amount", async () => {
      await expect(
        deployUpgradableFromFactory("TestL1MessageService", [
          limitSetter.address,
          pauser.address,
          ONE_DAY_IN_SECONDS,
          0,
        ]),
      ).to.revertedWithCustomError(l1MessageService, "LimitIsZero");
    });

    it("Should fail to deploy missing limit period", async () => {
      await expect(
        deployUpgradableFromFactory("TestL1MessageService", [
          limitSetter.address,
          pauser.address,
          0,
          INITIAL_WITHDRAW_LIMIT,
        ]),
      ).to.revertedWithCustomError(l1MessageService, "PeriodIsZero");
    });

    it("Should fail with empty limiter address", async () => {
      await expect(
        deployUpgradableFromFactory("TestL1MessageService", [
          ethers.constants.AddressZero,
          pauser.address,
          ONE_DAY_IN_SECONDS,
          INITIAL_WITHDRAW_LIMIT,
        ]),
      ).to.revertedWithCustomError(l1MessageService, "ZeroAddressNotAllowed");
    });

    it("Should fail with empty pauser address", async () => {
      await expect(
        deployUpgradableFromFactory("TestL1MessageService", [
          limitSetter.address,
          ethers.constants.AddressZero,
          ONE_DAY_IN_SECONDS,
          INITIAL_WITHDRAW_LIMIT,
        ]),
      ).to.revertedWithCustomError(l1MessageService, "ZeroAddressNotAllowed");
    });
  });

  describe("Send messages", () => {
    it("Should fail when the fee is higher than the amount sent", async () => {
      await expect(
        l1MessageService.connect(admin).canSendMessage(notAuthorizedAccount.address, MESSAGE_FEE, "0x", {
          value: MESSAGE_FEE.sub(1),
        }),
      ).to.be.revertedWithCustomError(l1MessageService, "ValueSentTooLow");
    });

    it("Should fail when the to address is address 0", async () => {
      await expect(
        l1MessageService.connect(admin).canSendMessage(ethers.constants.AddressZero, MESSAGE_FEE, "0x", {
          value: MESSAGE_FEE,
        }),
      ).to.be.revertedWithCustomError(l1MessageService, "ZeroAddressNotAllowed");
    });

    it("Should send an ether only message with fees emitting the MessageSent event", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );
      const messageHash = ethers.utils.keccak256(expectedBytes);

      await expect(
        l1MessageService.connect(admin).canSendMessage(notAuthorizedAccount.address, MESSAGE_FEE, "0x", {
          value: MESSAGE_FEE.add(MESSAGE_VALUE_1ETH),
        }),
      )
        .to.emit(l1MessageService, "MessageSent")
        .withArgs(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          1,
          "0x",
          messageHash,
        );

      const messageStatus = await l1MessageService.outboxL1L2MessageStatus(messageHash);
      expect(messageStatus).to.be.equal(OUTBOX_STATUS_SENT);
    });

    it("Should send max limit ether only message with no fee emitting the MessageSent event", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        BigNumber.from(0),
        INITIAL_WITHDRAW_LIMIT,
        BigNumber.from(1),
        "0x",
      );
      const messageHash = ethers.utils.keccak256(expectedBytes);

      await expect(
        l1MessageService
          .connect(admin)
          .canSendMessage(notAuthorizedAccount.address, 0, "0x", { value: INITIAL_WITHDRAW_LIMIT }),
      )
        .to.emit(l1MessageService, "MessageSent")
        .withArgs(
          l1MessageService.address,
          notAuthorizedAccount.address,
          0,
          INITIAL_WITHDRAW_LIMIT,
          1,
          "0x",
          messageHash,
        );

      const messageStatus = await l1MessageService.outboxL1L2MessageStatus(messageHash);
      expect(messageStatus).to.be.equal(OUTBOX_STATUS_SENT);
    });

    // this is testing to allow even if claim is blocked
    it("Should send a message even when L2 to L1 communication is paused", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );
      const messageHash = ethers.utils.keccak256(expectedBytes);

      await l1MessageService.connect(pauser).pauseByType(L2_L1_PAUSE_TYPE);

      await expect(
        l1MessageService.connect(admin).canSendMessage(notAuthorizedAccount.address, MESSAGE_FEE, "0x", {
          value: MESSAGE_FEE.add(MESSAGE_VALUE_1ETH),
        }),
      )
        .to.emit(l1MessageService, "MessageSent")
        .withArgs(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          1,
          "0x",
          messageHash,
        );

      const messageStatus = await l1MessageService.outboxL1L2MessageStatus(messageHash);
      expect(messageStatus).to.be.equal(OUTBOX_STATUS_SENT);
    });
  });

  describe("Claiming messages", () => {
    it("Should fail when the message hash does not exist", async () => {
      await expect(
        l1MessageService.claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x",
          1,
        ),
      ).to.be.revertedWithCustomError(l1MessageService, "MessageDoesNotExistOrHasAlreadyBeenClaimed");
    });

    it("Should execute the claim message and send fees to recipient, left over fee to destination", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );

      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      await l1MessageService.claimMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        postmanAddress.address,
        "0x",
        1,
      );
    });

    it("Should claim message and send the fees when L1 to L2 communication is paused", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );

      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      // this is for sending only and should not affect claim
      await l1MessageService.connect(pauser).pauseByType(L1_L2_PAUSE_TYPE);

      await l1MessageService.claimMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        postmanAddress.address,
        "0x",
        1,
      );
    });

    it("Should execute the claim message and emit the MessageClaimed event", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );

      const messageHash = ethers.utils.keccak256(expectedBytes);
      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(messageHash);

      await expect(
        l1MessageService.claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          postmanAddress.address,
          "0x",
          1,
        ),
      )
        .to.emit(l1MessageService, "MessageClaimed")
        .withArgs(messageHash);
    });

    it("Should fail when the message hash has been claimed", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );

      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      await l1MessageService.claimMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        postmanAddress.address,
        "0x",
        1,
      );
      await expect(
        l1MessageService.claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          postmanAddress.address,
          "0x",
          1,
        ),
      ).to.be.revertedWithCustomError(l1MessageService, "MessageDoesNotExistOrHasAlreadyBeenClaimed");
    });

    it("Should execute the claim message and send the fees to msg.sender, left over fee to destination", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );

      const expectedSecondBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(2),
        "0x",
      );

      const destinationBalance = await notAuthorizedAccount.getBalance();

      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedSecondBytes));

      await l1MessageService
        .connect(admin)
        .claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x",
          1,
        );

      const adminBalance = await admin.getBalance();

      await l1MessageService
        .connect(admin)
        .claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x",
          2,
        );

      expect(await notAuthorizedAccount.getBalance()).to.be.greaterThan(
        destinationBalance.add(MESSAGE_VALUE_1ETH).add(MESSAGE_VALUE_1ETH),
      );
      expect(await admin.getBalance()).to.be.greaterThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.utils.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and send the fees to msg.sender and NOT refund the destination", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        LOW_NO_REFUND_MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );
      const destinationBalance = await notAuthorizedAccount.getBalance();

      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      await l1MessageService
        .connect(admin)
        .claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          LOW_NO_REFUND_MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x",
          1,
          { gasPrice: 1000000000 },
        );

      expect(await notAuthorizedAccount.getBalance()).to.be.equal(destinationBalance.add(MESSAGE_VALUE_1ETH));

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.utils.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and send fees to recipient contract and no refund sent", async () => {
      const factory = await ethers.getContractFactory("TestReceivingContract");
      const testContract = (await factory.deploy()) as TestReceivingContract;

      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        testContract.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );

      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      const adminBalance = await admin.getBalance();
      await l1MessageService
        .connect(admin)
        .claimMessage(
          l1MessageService.address,
          testContract.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x",
          1,
        );

      expect(await ethers.provider.getBalance(testContract.address)).to.be.equal(MESSAGE_VALUE_1ETH);
      expect(await admin.getBalance()).to.be.greaterThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.utils.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and send fees to EOA with calldata and no refund sent", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x123456789a",
      );

      const startingBalance = await notAuthorizedAccount.getBalance();
      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      const adminBalance = await admin.getBalance();
      await l1MessageService
        .connect(admin)
        .claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x123456789a",
          1,
        );

      expect(await notAuthorizedAccount.getBalance()).to.be.equal(startingBalance.add(MESSAGE_VALUE_1ETH));
      expect(await admin.getBalance()).to.be.greaterThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.utils.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and no fees to EOA with calldata and no refund sent", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        BigNumber.from(0),
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x12",
      );

      const startingBalance = await notAuthorizedAccount.getBalance();
      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      const adminBalance = await admin.getBalance();
      await l1MessageService
        .connect(admin)
        .claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          BigNumber.from(0),
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x12",
          1,
        );

      expect(await notAuthorizedAccount.getBalance()).to.be.equal(startingBalance.add(MESSAGE_VALUE_1ETH));
      expect(await admin.getBalance()).to.be.lessThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.utils.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and no fees to EOA with empty calldata and no refund sent", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        BigNumber.from(0),
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        EMPTY_CALLDATA,
      );

      const startingBalance = await notAuthorizedAccount.getBalance();
      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      const adminBalance = await admin.getBalance();
      await l1MessageService
        .connect(admin)
        .claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          BigNumber.from(0),
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          EMPTY_CALLDATA,
          1,
        );

      expect(await notAuthorizedAccount.getBalance()).to.be.equal(startingBalance.add(MESSAGE_VALUE_1ETH));
      expect(await admin.getBalance()).to.be.lessThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.utils.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message when there are no fees", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        notAuthorizedAccount.address,
        BigNumber.from(0),
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );
      const destinationBalance = await notAuthorizedAccount.getBalance();

      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      const adminBalance = await admin.getBalance();
      await l1MessageService
        .connect(admin)
        .claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          0,
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x",
          1,
        );

      expect(await notAuthorizedAccount.getBalance()).to.be.equal(destinationBalance.add(MESSAGE_VALUE_1ETH));
      expect(await admin.getBalance()).to.be.lessThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.utils.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should provide the correct origin sender", async () => {
      const sendCalldata = generateKeccak256Hash("setSender()").substring(0, 10);

      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        l1MessageService.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        sendCalldata,
      );
      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      const storedSenderBeforeSending = await l1MessageService.originalSender();
      expect(storedSenderBeforeSending).to.be.equal(ethers.constants.AddressZero);

      await expect(
        l1MessageService
          .connect(admin)
          .claimMessage(
            l1MessageService.address,
            l1MessageService.address,
            MESSAGE_FEE,
            MESSAGE_VALUE_1ETH,
            ethers.constants.AddressZero,
            sendCalldata,
            1,
          ),
      ).to.not.be.reverted;

      const newSender = await l1MessageService.originalSender();

      expect(newSender).to.be.equal(l1MessageService.address);
    });

    it("Should allow sending post claiming a message", async () => {
      const sendCalldata = generateKeccak256Hash("sendNewMessage()").substring(0, 10);

      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        l1MessageService.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        sendCalldata,
      );
      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      await expect(
        l1MessageService
          .connect(admin)
          .claimMessage(
            l1MessageService.address,
            l1MessageService.address,
            MESSAGE_FEE,
            MESSAGE_VALUE_1ETH,
            ethers.constants.AddressZero,
            sendCalldata,
            1,
          ),
      ).to.not.be.reverted;
    });

    it("Should fail on reentry when sending to recipient", async () => {
      const callSignature = generateKeccak256Hash("doReentry()").substring(0, 10);

      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        l1MessageService.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        callSignature,
      );
      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      await expect(
        l1MessageService
          .connect(admin)
          .claimMessage(
            l1MessageService.address,
            l1MessageService.address,
            MESSAGE_FEE,
            MESSAGE_VALUE_1ETH,
            ethers.constants.AddressZero,
            callSignature,
            BigNumber.from(1),
          ),
      ).to.be.revertedWith("ReentrancyGuard: reentrant call");
    });

    it("Should fail when the destination errors", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        l1MessageService.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );

      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));
      await expect(
        l1MessageService
          .connect(admin)
          .claimMessage(
            l1MessageService.address,
            l1MessageService.address,
            MESSAGE_FEE,
            MESSAGE_VALUE_1ETH,
            ethers.constants.AddressZero,
            "0x",
            1,
          ),
      )
        .to.be.revertedWithCustomError(l1MessageService, "MessageSendingFailed")
        .withArgs(l1MessageService.address);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.utils.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_RECEIVED,
      );
    });

    it("Should fail when the fee recipient fails errors", async () => {
      const expectedBytes = await encodeSendMessage(
        l1MessageService.address,
        admin.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        BigNumber.from(1),
        "0x",
      );

      await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
      await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

      await expect(
        l1MessageService
          .connect(admin)
          .claimMessage(
            l1MessageService.address,
            admin.address,
            MESSAGE_FEE,
            MESSAGE_VALUE_1ETH,
            l1MessageService.address,
            "0x",
            1,
          ),
      )
        .to.be.revertedWithCustomError(l1MessageService, "FeePaymentFailed")
        .withArgs(l1MessageService.address);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.utils.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_RECEIVED,
      );
    });

    it("Should revert with send over max limit amount only", async () => {
      await setHash(BigNumber.from(0), INITIAL_WITHDRAW_LIMIT.add(1));

      await expect(
        l1MessageService.claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          0,
          INITIAL_WITHDRAW_LIMIT.add(1),
          postmanAddress.address,
          "0x",
          1,
        ),
      ).to.revertedWithCustomError(l1MessageService, "RateLimitExceeded");
    });

    it("Should revert with send over max limit amount and fees", async () => {
      await setHash(BigNumber.from(1), INITIAL_WITHDRAW_LIMIT.add(1));

      await expect(
        l1MessageService.claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          1,
          INITIAL_WITHDRAW_LIMIT.add(1),
          postmanAddress.address,
          "0x",
          1,
        ),
      ).to.revertedWithCustomError(l1MessageService, "RateLimitExceeded");
    });

    it("Should revert with send over max limit amount and fees - multi tx", async () => {
      await setHashAndClaimMessage(MESSAGE_FEE, MESSAGE_VALUE_1ETH);

      await setHash(BigNumber.from(1), INITIAL_WITHDRAW_LIMIT.sub(MESSAGE_VALUE_1ETH).sub(MESSAGE_FEE).add(1));

      // limit - (fee+amount from success tx) + 1 = 1 over limit
      await expect(
        l1MessageService.claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          1,
          INITIAL_WITHDRAW_LIMIT.sub(MESSAGE_VALUE_1ETH).sub(MESSAGE_FEE).add(1),
          postmanAddress.address,
          "0x",
          1,
        ),
      ).to.revertedWithCustomError(l1MessageService, "RateLimitExceeded");
    });
  });

  describe("Resetting limits", () => {
    it("Should reset limits as limitSetter", async () => {
      let usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);

      await setHashAndClaimMessage(MESSAGE_FEE, MESSAGE_VALUE_1ETH);

      usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(MESSAGE_FEE.add(MESSAGE_VALUE_1ETH));

      await l1MessageService.connect(limitSetter).resetAmountUsedInPeriod();
      usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);
    });

    it("Should fail reset limits as non-RATE_LIMIT_SETTER_ROLE", async () => {
      let usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);

      await setHashAndClaimMessage(MESSAGE_FEE, MESSAGE_VALUE_1ETH);

      usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(MESSAGE_FEE.add(MESSAGE_VALUE_1ETH));

      await expect(l1MessageService.connect(admin).resetAmountUsedInPeriod()).to.be.revertedWith(
        "AccessControl: account " + admin.address.toLowerCase() + " is missing role " + RATE_LIMIT_SETTER_ROLE,
      );

      usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(MESSAGE_FEE.add(MESSAGE_VALUE_1ETH));
    });
  });

  describe("Pausing contracts", () => {
    it("Should fail general pausing as non-pauser", async () => {
      expect(await l1MessageService.pauseTypeStatuses(GENERAL_PAUSE_TYPE)).to.be.false;

      await expect(l1MessageService.connect(admin).pauseByType(GENERAL_PAUSE_TYPE)).to.be.revertedWith(
        "AccessControl: account " + admin.address.toLowerCase() + " is missing role " + PAUSE_MANAGER_ROLE,
      );

      expect(await l1MessageService.pauseTypeStatuses(GENERAL_PAUSE_TYPE)).to.be.false;
    });

    it("Should pause generally as pause manager", async () => {
      expect(await l1MessageService.pauseTypeStatuses(GENERAL_PAUSE_TYPE)).to.be.false;

      await l1MessageService.connect(pauser).pauseByType(GENERAL_PAUSE_TYPE);

      expect(await l1MessageService.pauseTypeStatuses(GENERAL_PAUSE_TYPE)).to.be.true;
    });

    it("Should fail when to claim the contract is generally paused", async () => {
      await l1MessageService.connect(pauser).pauseByType(GENERAL_PAUSE_TYPE);

      await expect(
        l1MessageService.claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x",
          1,
        ),
      )
        .to.be.revertedWithCustomError(l1MessageService, "IsPaused")
        .withArgs(GENERAL_PAUSE_TYPE);
    });

    it("Should fail to claim when the L2 to L1 communication is paused", async () => {
      await l1MessageService.connect(pauser).pauseByType(L2_L1_PAUSE_TYPE);

      await expect(
        l1MessageService.claimMessage(
          l1MessageService.address,
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ethers.constants.AddressZero,
          "0x",
          1,
        ),
      )
        .to.be.revertedWithCustomError(l1MessageService, "IsPaused")
        .withArgs(L2_L1_PAUSE_TYPE);
    });

    it("Should fail to send if the contract is generally paused", async () => {
      await l1MessageService.connect(pauser).pauseByType(GENERAL_PAUSE_TYPE);

      await expect(
        l1MessageService
          .connect(admin)
          .canSendMessage(notAuthorizedAccount.address, 0, "0x", { value: INITIAL_WITHDRAW_LIMIT }),
      )
        .to.be.revertedWithCustomError(l1MessageService, "IsPaused")
        .withArgs(GENERAL_PAUSE_TYPE);

      const usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);
    });

    it("Should fail to send if L1 to L2 communication is paused", async () => {
      await l1MessageService.connect(pauser).pauseByType(L1_L2_PAUSE_TYPE);

      await expect(
        l1MessageService
          .connect(admin)
          .canSendMessage(notAuthorizedAccount.address, 0, "0x", { value: INITIAL_WITHDRAW_LIMIT }),
      )
        .to.be.revertedWithCustomError(l1MessageService, "IsPaused")
        .withArgs(L1_L2_PAUSE_TYPE);

      const usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);
    });
  });

  async function setHash(fee: BigNumber, value: BigNumber) {
    const expectedBytes = await encodeSendMessage(
      l1MessageService.address,
      notAuthorizedAccount.address,
      fee,
      value,
      BigNumber.from(1),
      "0x",
    );

    await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
    await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));
  }

  async function setHashAndClaimMessage(fee: BigNumber, value: BigNumber) {
    const expectedBytes = await encodeSendMessage(
      l1MessageService.address,
      notAuthorizedAccount.address,
      fee,
      value,
      BigNumber.from(1),
      "0x",
    );

    await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT });
    await l1MessageService.addL2L1MessageHash(ethers.utils.keccak256(expectedBytes));

    await l1MessageService.claimMessage(
      l1MessageService.address,
      notAuthorizedAccount.address,
      fee,
      value,
      postmanAddress.address,
      "0x",
      1,
    );
  }
});
