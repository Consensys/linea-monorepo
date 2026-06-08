import { expect } from "chai";
import { network as hardhatNetwork } from "hardhat";

import { encodeSendMessage } from "../../../../common/helpers/encoding";
import {
  ADDRESS_ZERO,
  DEFAULT_ADMIN_ROLE,
  EMPTY_CALLDATA,
  GENERAL_PAUSE_TYPE,
  INBOX_STATUS_RECEIVED,
  INBOX_STATUS_UNKNOWN,
  INITIALIZED_ERROR_MESSAGE,
  INITIAL_WITHDRAW_LIMIT,
  L1_L2_PAUSE_TYPE,
  L2_L1_PAUSE_TYPE,
  LOW_NO_REFUND_MESSAGE_FEE,
  MESSAGE_FEE,
  MESSAGE_VALUE_1ETH,
  ONE_DAY_IN_SECONDS,
  PAUSE_ALL_ROLE,
  UNPAUSE_ALL_ROLE,
  RATE_LIMIT_SETTER_ROLE,
  USED_RATE_LIMIT_RESETTER_ROLE,
  PAUSE_L2_L1_ROLE,
  PAUSE_L1_L2_ROLE,
  pauseTypeRoles,
  unpauseTypeRoles,
} from "../../common/constants";
import { deployFromFactory, deployUpgradableFromFactory } from "../../common/deployment";
import {
  buildAccessErrorMessage,
  calculateRollingHash,
  expectEvent,
  expectHasRole,
  expectRevertWithCustomError,
  expectRevertWithReason,
  expectRevertWhenPaused,
  expectPaused,
  expectNotPaused,
  generateKeccak256Hash,
  validateRollingHashStorage,
  validateRollingHashNotZero,
  expectRollingHashUpdatedEvent,
  INITIAL_ROLLING_HASH,
} from "../../common/helpers";

import type {
  TestClaimingCaller,
  TestL1MessageService,
  TestL1MessageServiceMerkleProof,
  TestL1RevertContract,
  TestReceivingContract,
} from "../../../../typechain-types";
import type { HardhatEthersSigner as SignerWithAddress } from "@nomicfoundation/hardhat-ethers/types";

import { clearSnapshots, loadFixture } from "#hardhat-network-helpers";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;

const SHORT_DYNAMIC_MERKLE_PROOF_DEPTH = 8;

type ClaimProofParams = {
  data: string;
  fee: bigint;
  from: string;
  messageNumber: bigint;
  to: string;
  value: bigint;
};

describe("L1MessageService", () => {
  let l1MessageService: TestL1MessageService;
  let l1MessageServiceMerkleProof: TestL1MessageServiceMerkleProof;
  let l1TestRevert: TestL1RevertContract;
  let admin: SignerWithAddress;
  let pauser: SignerWithAddress;
  let limitSetter: SignerWithAddress;
  let notAuthorizedAccount: SignerWithAddress;
  let postmanAddress: SignerWithAddress;
  let l2Sender: SignerWithAddress;

  async function deployTestL1MessageServiceFixture(): Promise<TestL1MessageService> {
    return deployUpgradableFromFactory(
      "TestL1MessageService",
      [ONE_DAY_IN_SECONDS, INITIAL_WITHDRAW_LIMIT, pauseTypeRoles, unpauseTypeRoles],
      { unsafeAllow: ["incorrect-initializer-order"] },
    ) as unknown as Promise<TestL1MessageService>;
  }

  async function deployL1MessageServiceMerkleFixture(): Promise<TestL1MessageServiceMerkleProof> {
    return deployUpgradableFromFactory(
      "TestL1MessageServiceMerkleProof",
      [ONE_DAY_IN_SECONDS, INITIAL_WITHDRAW_LIMIT, pauseTypeRoles, unpauseTypeRoles],
      { unsafeAllow: ["incorrect-initializer-order"] },
    ) as unknown as Promise<TestL1MessageServiceMerkleProof>;
  }

  async function deployL1TestRevertFixture(): Promise<TestL1RevertContract> {
    return deployUpgradableFromFactory("TestL1RevertContract", []) as unknown as Promise<TestL1RevertContract>;
  }

  function efficientKeccak(left: string, right: string): string {
    return ethers.keccak256(ethers.concat([left, right]));
  }

  function buildSingleLeafProof(params: ClaimProofParams, depth: number) {
    const proof: string[] = [];
    let emptyNodeHash = ethers.ZeroHash;

    for (let height = 0; height < depth; height++) {
      proof.push(emptyNodeHash);
      emptyNodeHash = efficientKeccak(emptyNodeHash, emptyNodeHash);
    }

    const messageLeafHash = ethers.keccak256(
      encodeSendMessage(params.from, params.to, params.fee, params.value, params.messageNumber, params.data),
    );

    return {
      proof,
      merkleRoot: proof.reduce((node, sibling) => efficientKeccak(node, sibling), messageLeafHash),
      index: 0,
    };
  }

  function buildDefaultClaimProof() {
    return buildSingleLeafProof(
      {
        data: EMPTY_CALLDATA,
        fee: MESSAGE_FEE,
        from: admin.address,
        messageNumber: 1n,
        to: admin.address,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
      },
      32,
    );
  }

  function buildInvalidClaimProof() {
    const claimParams = buildDefaultClaimProof();
    const proof = [...claimParams.proof];
    proof[1] = ethers.ZeroHash;

    return { ...claimParams, proof };
  }

  before(async () => {
    await clearSnapshots();
    [admin, pauser, limitSetter, notAuthorizedAccount, postmanAddress, l2Sender] = await ethers.getSigners();
  });

  beforeEach(async () => {
    l1MessageService = await loadFixture(deployTestL1MessageServiceFixture);
    l1MessageServiceMerkleProof = await loadFixture(deployL1MessageServiceMerkleFixture);
    l1TestRevert = await loadFixture(deployL1TestRevertFixture);

    await l1MessageService.grantRole(PAUSE_L1_L2_ROLE, pauser.address);
    await l1MessageService.grantRole(PAUSE_L2_L1_ROLE, pauser.address);
    await l1MessageService.grantRole(PAUSE_ALL_ROLE, pauser.address);
    await l1MessageService.grantRole(UNPAUSE_ALL_ROLE, pauser.address);
    await l1MessageService.grantRole(RATE_LIMIT_SETTER_ROLE, limitSetter.address);
    await l1MessageService.grantRole(USED_RATE_LIMIT_RESETTER_ROLE, limitSetter.address);

    await l1MessageServiceMerkleProof.grantRole(PAUSE_ALL_ROLE, pauser.address);

    await l1MessageService.addFunds({ value: INITIAL_WITHDRAW_LIMIT * 2n });
    await l1MessageServiceMerkleProof.addFunds({ value: INITIAL_WITHDRAW_LIMIT * 2n });
  });

  describe("Initialisation tests", () => {
    it("Deployer has default admin role", async () => {
      await expectHasRole(l1MessageService, DEFAULT_ADMIN_ROLE, admin);
    });

    it("limitSetter has RATE_LIMIT_SETTER_ROLE", async () => {
      await expectHasRole(l1MessageService, RATE_LIMIT_SETTER_ROLE, limitSetter);
    });

    it("limitSetter has USED_RATE_LIMIT_RESETTER_ROLE", async () => {
      await expectHasRole(l1MessageService, USED_RATE_LIMIT_RESETTER_ROLE, limitSetter);
    });

    it("pauser has PAUSE_ALL_ROLE", async () => {
      await expectHasRole(l1MessageService, PAUSE_ALL_ROLE, pauser);
    });

    it("pauser has UNPAUSE_ALL_ROLE", async () => {
      await expectHasRole(l1MessageService, UNPAUSE_ALL_ROLE, pauser);
    });

    it("Should set rate limit and period", async () => {
      expect(await l1MessageService.periodInSeconds()).to.be.equal(ONE_DAY_IN_SECONDS);
      expect(await l1MessageService.limitInWei()).to.be.equal(INITIAL_WITHDRAW_LIMIT);
    });

    it("It should fail when not initializing", async () => {
      await expectRevertWithReason(
        l1MessageService.tryInitialize(ONE_DAY_IN_SECONDS, INITIAL_WITHDRAW_LIMIT, pauseTypeRoles, unpauseTypeRoles),
        INITIALIZED_ERROR_MESSAGE,
      );
    });

    it("Should initialize nextMessageNumber", async () => {
      expect(await l1MessageService.nextMessageNumber()).to.be.equal(1);
    });

    it("Should fail to deploy missing amount", async () => {
      await expectRevertWithCustomError(
        l1MessageService,
        deployUpgradableFromFactory("TestL1MessageService", [ONE_DAY_IN_SECONDS, 0, pauseTypeRoles, unpauseTypeRoles], {
          unsafeAllow: ["incorrect-initializer-order"],
        }),
        "LimitIsZero",
      );
    });

    it("Should fail to deploy missing limit period", async () => {
      await expectRevertWithCustomError(
        l1MessageService,
        deployUpgradableFromFactory(
          "TestL1MessageService",
          [0, INITIAL_WITHDRAW_LIMIT, pauseTypeRoles, unpauseTypeRoles],
          { unsafeAllow: ["incorrect-initializer-order"] },
        ),
        "PeriodIsZero",
      );
    });
  });

  describe("Send messages", () => {
    it("Should fail when the fee is higher than the amount sent", async () => {
      const sendMessageCall = l1MessageService
        .connect(admin)
        .canSendMessage(notAuthorizedAccount.address, MESSAGE_FEE, EMPTY_CALLDATA, { value: MESSAGE_FEE - 1n });

      await expectRevertWithCustomError(l1MessageService, sendMessageCall, "ValueSentTooLow");
    });

    it("Should fail when the to address is address 0", async () => {
      const sendMessageCall = l1MessageService
        .connect(admin)
        .canSendMessage(ADDRESS_ZERO, MESSAGE_FEE, EMPTY_CALLDATA, { value: MESSAGE_FEE });

      await expectRevertWithCustomError(l1MessageService, sendMessageCall, "ZeroAddressNotAllowed");
    });

    it("Should send an ether only message with fees emitting the MessageSent event", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      const messageHash = ethers.keccak256(expectedBytes);

      const sendMessageCall = l1MessageService
        .connect(admin)
        .canSendMessage(notAuthorizedAccount.address, MESSAGE_FEE, EMPTY_CALLDATA, {
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        });
      const eventArgs = [
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1,
        EMPTY_CALLDATA,
        messageHash,
      ];
      await expectEvent(l1MessageService, sendMessageCall, "MessageSent", eventArgs);
    });

    it("Should send max limit ether only message with no fee emitting the MessageSent event", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        0n,
        INITIAL_WITHDRAW_LIMIT,
        1n,
        EMPTY_CALLDATA,
      );
      const messageHash = ethers.keccak256(expectedBytes);

      const sendMessageCall = l1MessageService
        .connect(admin)
        .canSendMessage(notAuthorizedAccount.address, 0, EMPTY_CALLDATA, { value: INITIAL_WITHDRAW_LIMIT });
      const eventArgs = [
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        0,
        INITIAL_WITHDRAW_LIMIT,
        1,
        EMPTY_CALLDATA,
        messageHash,
      ];
      await expectEvent(l1MessageService, sendMessageCall, "MessageSent", eventArgs);
    });

    // this is testing to allow even if claim is blocked
    it("Should send a message even when L2 to L1 communication is paused", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      const messageHash = ethers.keccak256(expectedBytes);

      await l1MessageService.connect(pauser).pauseByType(L2_L1_PAUSE_TYPE);

      const sendMessageCall = l1MessageService
        .connect(admin)
        .canSendMessage(notAuthorizedAccount.address, MESSAGE_FEE, EMPTY_CALLDATA, {
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        });
      const eventArgs = [
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1,
        EMPTY_CALLDATA,
        messageHash,
      ];
      await expectEvent(l1MessageService, sendMessageCall, "MessageSent", eventArgs);
    });

    it("Should update the rolling hash when sending a message post migration", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      const messageHash = ethers.keccak256(expectedBytes);
      const rollingHash = calculateRollingHash(INITIAL_ROLLING_HASH, messageHash);

      const sendMessageCall = l1MessageService
        .connect(admin)
        .canSendMessage(notAuthorizedAccount.address, MESSAGE_FEE, EMPTY_CALLDATA, {
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        });
      const messageSentEventArgs = [
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1,
        EMPTY_CALLDATA,
        messageHash,
      ];

      await expectEvent(l1MessageService, sendMessageCall, "MessageSent", messageSentEventArgs);
      await expectRollingHashUpdatedEvent({
        contract: l1MessageService,
        updateCall: sendMessageCall,
        messageNumber: 1n,
        expectedRollingHash: rollingHash,
        messageHash,
      });

      await validateRollingHashStorage({
        contract: l1MessageService,
        messageNumber: 1n,
        expectedHash: rollingHash,
      });
      await validateRollingHashNotZero({ contract: l1MessageService, messageNumber: 1n });
    });

    it("Should use the previous existing rolling hash when sending a message post migration", async () => {
      let expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      let messageHash = ethers.keccak256(expectedBytes);
      let rollingHash = calculateRollingHash(INITIAL_ROLLING_HASH, messageHash);

      await l1MessageService.connect(admin).canSendMessage(notAuthorizedAccount.address, MESSAGE_FEE, EMPTY_CALLDATA, {
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
      });

      expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        2n,
        EMPTY_CALLDATA,
      );

      messageHash = ethers.keccak256(expectedBytes);

      rollingHash = calculateRollingHash(rollingHash, messageHash);

      const sendMessageCall = l1MessageService
        .connect(admin)
        .canSendMessage(notAuthorizedAccount.address, MESSAGE_FEE, EMPTY_CALLDATA, {
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        });
      const messageSentEventArgs = [
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        2,
        EMPTY_CALLDATA,
        messageHash,
      ];

      await expectEvent(l1MessageService, sendMessageCall, "MessageSent", messageSentEventArgs);
      await expectRollingHashUpdatedEvent({
        contract: l1MessageService,
        updateCall: sendMessageCall,
        messageNumber: 2n,
        expectedRollingHash: rollingHash,
        messageHash,
      });

      await validateRollingHashStorage({
        contract: l1MessageService,
        messageNumber: 2n,
        expectedHash: rollingHash,
      });
      await validateRollingHashNotZero({ contract: l1MessageService, messageNumber: 2n });
    });
  });

  describe("Claiming messages", () => {
    it("Should fail when the message hash does not exist", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      const messageHash = ethers.keccak256(expectedBytes);

      const claimMessageCall = l1MessageService.claimMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        ADDRESS_ZERO,
        EMPTY_CALLDATA,
        1,
      );

      await expectRevertWithCustomError(
        l1MessageService,
        claimMessageCall,
        "MessageDoesNotExistOrHasAlreadyBeenClaimed",
        [messageHash],
      );
    });

    it("Should execute the claim message and send fees to recipient, left over fee to destination", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      await expect(
        l1MessageService.claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          postmanAddress.address,
          EMPTY_CALLDATA,
          1,
        ),
      ).to.not.revert(ethers);
    });

    it("Should claim message and send the fees when L1 to L2 communication is paused", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      // this is for sending only and should not affect claim
      await l1MessageService.connect(pauser).pauseByType(L1_L2_PAUSE_TYPE);

      await expect(
        l1MessageService.claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          postmanAddress.address,
          EMPTY_CALLDATA,
          1,
        ),
      ).to.not.revert(ethers);
    });

    it("Should execute the claim message and emit the MessageClaimed event", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      const messageHash = ethers.keccak256(expectedBytes);

      await l1MessageService.addL2L1MessageHash(messageHash);

      const claimMessageCall = l1MessageService.claimMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        postmanAddress.address,
        EMPTY_CALLDATA,
        1,
      );

      await expectEvent(l1MessageService, claimMessageCall, "MessageClaimed", [messageHash]);
    });

    it("Should fail when the message hash has been claimed", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      const messageHash = ethers.keccak256(expectedBytes);

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      await l1MessageService.claimMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        postmanAddress.address,
        EMPTY_CALLDATA,
        1,
      );

      const claimMessageCall = l1MessageService.claimMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        postmanAddress.address,
        EMPTY_CALLDATA,
        1,
      );

      await expectRevertWithCustomError(
        l1MessageService,
        claimMessageCall,
        "MessageDoesNotExistOrHasAlreadyBeenClaimed",
        [messageHash],
      );
    });

    it("Should execute the claim message and send the fees to msg.sender, left over fee to destination", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      const expectedSecondBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        2n,
        EMPTY_CALLDATA,
      );

      const destinationBalance = await ethers.provider.getBalance(notAuthorizedAccount.address);

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));
      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedSecondBytes));

      await l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ADDRESS_ZERO,
          EMPTY_CALLDATA,
          1,
        );

      const adminBalance = await ethers.provider.getBalance(admin.address);

      await l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ADDRESS_ZERO,
          EMPTY_CALLDATA,
          2,
        );

      expect(await ethers.provider.getBalance(notAuthorizedAccount.address)).to.be.greaterThan(
        destinationBalance + MESSAGE_VALUE_1ETH + MESSAGE_VALUE_1ETH,
      );
      expect(await ethers.provider.getBalance(admin.address)).to.be.greaterThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and send the fees to msg.sender and NOT refund the destination", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        LOW_NO_REFUND_MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );
      const destinationBalance = await ethers.provider.getBalance(notAuthorizedAccount.address);

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      await l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          LOW_NO_REFUND_MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ADDRESS_ZERO,
          EMPTY_CALLDATA,
          1,
          { gasPrice: 1000000000 },
        );

      expect(await ethers.provider.getBalance(notAuthorizedAccount.address)).to.be.equal(
        destinationBalance + MESSAGE_VALUE_1ETH,
      );

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and send fees to recipient contract and no refund sent", async () => {
      const factory = await ethers.getContractFactory("TestReceivingContract");
      const testContract = (await factory.deploy()) as TestReceivingContract;

      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        await testContract.getAddress(),
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      const adminBalance = await ethers.provider.getBalance(admin.address);
      await l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          await testContract.getAddress(),
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ADDRESS_ZERO,
          EMPTY_CALLDATA,
          1,
        );

      expect(await ethers.provider.getBalance(await testContract.getAddress())).to.be.equal(MESSAGE_VALUE_1ETH);
      expect(await ethers.provider.getBalance(admin.address)).to.be.greaterThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and send fees to EOA with calldata and no refund sent", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        "0x123456789a",
      );

      const startingBalance = await ethers.provider.getBalance(notAuthorizedAccount.address);

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      const adminBalance = await ethers.provider.getBalance(admin.address);
      await l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ADDRESS_ZERO,
          "0x123456789a",
          1,
        );

      expect(await ethers.provider.getBalance(notAuthorizedAccount.address)).to.be.equal(
        startingBalance + MESSAGE_VALUE_1ETH,
      );
      expect(await ethers.provider.getBalance(admin.address)).to.be.greaterThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and no fees to EOA with calldata and no refund sent", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        0n,
        MESSAGE_VALUE_1ETH,
        1n,
        "0x12",
      );

      const startingBalance = await ethers.provider.getBalance(notAuthorizedAccount.address);

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      const adminBalance = await ethers.provider.getBalance(admin.address);
      await l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          0n,
          MESSAGE_VALUE_1ETH,
          ADDRESS_ZERO,
          "0x12",
          1,
        );

      expect(await ethers.provider.getBalance(notAuthorizedAccount.address)).to.be.equal(
        startingBalance + MESSAGE_VALUE_1ETH,
      );
      expect(await ethers.provider.getBalance(admin.address)).to.be.lessThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message and no fees to EOA with empty calldata and no refund sent", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        0n,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      const startingBalance = await ethers.provider.getBalance(notAuthorizedAccount.address);

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      const adminBalance = await ethers.provider.getBalance(admin.address);
      await l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          0n,
          MESSAGE_VALUE_1ETH,
          ADDRESS_ZERO,
          EMPTY_CALLDATA,
          1,
        );

      expect(await ethers.provider.getBalance(notAuthorizedAccount.address)).to.be.equal(
        startingBalance + MESSAGE_VALUE_1ETH,
      );
      expect(await ethers.provider.getBalance(admin.address)).to.be.lessThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should execute the claim message when there are no fees", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        0n,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );
      const destinationBalance = await ethers.provider.getBalance(notAuthorizedAccount.address);

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      const adminBalance = await ethers.provider.getBalance(admin.address);
      await l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          0,
          MESSAGE_VALUE_1ETH,
          ADDRESS_ZERO,
          EMPTY_CALLDATA,
          1,
        );

      expect(await ethers.provider.getBalance(notAuthorizedAccount.address)).to.be.equal(
        destinationBalance + MESSAGE_VALUE_1ETH,
      );
      expect(await ethers.provider.getBalance(admin.address)).to.be.lessThan(adminBalance);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_UNKNOWN,
      );
    });

    it("Should provide the correct origin sender", async () => {
      const claimingCaller = (await deployFromFactory(
        "TestClaimingCaller",
        l2Sender.address,
      )) as unknown as TestClaimingCaller;

      const expectedBytes = encodeSendMessage(
        l2Sender.address,
        await claimingCaller.getAddress(),
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      const storedSenderBeforeSending = await l1MessageService.originalSender();
      expect(storedSenderBeforeSending).to.be.equal(ADDRESS_ZERO);

      await expect(
        l1MessageService
          .connect(admin)
          .claimMessage(
            l2Sender.address,
            await claimingCaller.getAddress(),
            MESSAGE_FEE,
            MESSAGE_VALUE_1ETH,
            ADDRESS_ZERO,
            EMPTY_CALLDATA,
            1,
          ),
      ).to.not.revert(ethers);
    });

    it("Should allow sending post claiming a message", async () => {
      const sendCalldata = generateKeccak256Hash("sendNewMessage()").substring(0, 10);

      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        await l1MessageService.getAddress(),
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        sendCalldata,
      );

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      await expect(
        l1MessageService
          .connect(admin)
          .claimMessage(
            await l1MessageService.getAddress(),
            await l1MessageService.getAddress(),
            MESSAGE_FEE,
            MESSAGE_VALUE_1ETH,
            ADDRESS_ZERO,
            sendCalldata,
            1,
          ),
      ).to.not.revert(ethers);
    });

    it("Should fail on reentry when sending to recipient", async () => {
      const callSignature = generateKeccak256Hash("doReentry()").substring(0, 10);

      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        await l1MessageService.getAddress(),
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        callSignature,
      );

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      const claimMessageCall = l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          await l1MessageService.getAddress(),
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          ADDRESS_ZERO,
          callSignature,
          1n,
        );

      await expectRevertWithCustomError(l1MessageService, claimMessageCall, "ReentrantCall");
    });

    it("Should fail when the destination errors", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        await l1MessageService.getAddress(),
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));
      await expectRevertWithCustomError(
        l1MessageService,
        l1MessageService
          .connect(admin)
          .claimMessage(
            await l1MessageService.getAddress(),
            await l1MessageService.getAddress(),
            MESSAGE_FEE,
            MESSAGE_VALUE_1ETH,
            ADDRESS_ZERO,
            EMPTY_CALLDATA,
            1,
          ),
        "MessageSendingFailed",
        [await l1MessageService.getAddress()],
      );

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_RECEIVED,
      );
    });

    it("Should fail when the fee recipient fails errors", async () => {
      const expectedBytes = encodeSendMessage(
        await l1MessageService.getAddress(),
        admin.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        1n,
        EMPTY_CALLDATA,
      );

      await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

      const claimMessageCall = l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          admin.address,
          MESSAGE_FEE,
          MESSAGE_VALUE_1ETH,
          await l1MessageService.getAddress(),
          EMPTY_CALLDATA,
          1,
        );

      await expectRevertWithCustomError(l1MessageService, claimMessageCall, "FeePaymentFailed", [
        await l1MessageService.getAddress(),
      ]);

      expect(await l1MessageService.inboxL2L1MessageStatus(ethers.keccak256(expectedBytes))).to.be.equal(
        INBOX_STATUS_RECEIVED,
      );
    });

    it("Should revert with send over max limit amount only", async () => {
      await setHash(0n, INITIAL_WITHDRAW_LIMIT + 1n);

      const claimMessageCall = l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          0,
          INITIAL_WITHDRAW_LIMIT + 1n,
          postmanAddress.address,
          EMPTY_CALLDATA,
          1,
        );

      await expectRevertWithCustomError(l1MessageService, claimMessageCall, "RateLimitExceeded");
    });

    it("Should revert with send over max limit amount and fees", async () => {
      await setHash(1n, INITIAL_WITHDRAW_LIMIT + 1n);

      const claimMessageCall = l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          1,
          INITIAL_WITHDRAW_LIMIT + 1n,
          postmanAddress.address,
          EMPTY_CALLDATA,
          1,
        );

      await expectRevertWithCustomError(l1MessageService, claimMessageCall, "RateLimitExceeded");
    });

    it("Should revert with send over max limit amount and fees - multi tx", async () => {
      await setHashAndClaimMessage(MESSAGE_FEE, MESSAGE_VALUE_1ETH);

      await setHash(1n, INITIAL_WITHDRAW_LIMIT - MESSAGE_VALUE_1ETH - MESSAGE_FEE + 1n);

      const claimMessageCall = l1MessageService
        .connect(admin)
        .claimMessage(
          await l1MessageService.getAddress(),
          notAuthorizedAccount.address,
          1,
          INITIAL_WITHDRAW_LIMIT - MESSAGE_VALUE_1ETH - MESSAGE_FEE + 1n,
          postmanAddress.address,
          EMPTY_CALLDATA,
          1,
        );

      await expectRevertWithCustomError(l1MessageService, claimMessageCall, "RateLimitExceeded");
    });
  });

  describe("Claim Message with Proof", () => {
    it("Should be able to claim a message that was sent", async () => {
      const claimParams = buildDefaultClaimProof();

      await l1MessageServiceMerkleProof.addL2MerkleRoots([claimParams.merkleRoot], claimParams.proof.length);

      await expect(
        l1MessageServiceMerkleProof.claimMessageWithProof({
          proof: claimParams.proof,
          messageNumber: 1,
          leafIndex: claimParams.index,
          from: admin.address,
          to: admin.address,
          fee: MESSAGE_FEE,
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
          feeRecipient: ADDRESS_ZERO,
          merkleRoot: claimParams.merkleRoot,
          data: EMPTY_CALLDATA,
        }),
      ).to.not.revert(ethers);
    });

    it("Should be able to claim a message and emit a MessageClaimed event", async () => {
      const claimParams = buildDefaultClaimProof();

      await l1MessageServiceMerkleProof.addL2MerkleRoots([claimParams.merkleRoot], claimParams.proof.length);

      const messageLeafHash = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["address", "address", "uint256", "uint256", "uint256", "bytes"],
          [admin.address, admin.address, MESSAGE_FEE, MESSAGE_FEE + MESSAGE_VALUE_1ETH, "1", EMPTY_CALLDATA],
        ),
      );

      const claimMessageCall = l1MessageServiceMerkleProof.claimMessageWithProof({
        proof: claimParams.proof,
        messageNumber: 1,
        leafIndex: claimParams.index,
        from: admin.address,
        to: admin.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ADDRESS_ZERO,
        merkleRoot: claimParams.merkleRoot,
        data: EMPTY_CALLDATA,
      });

      await expectEvent(l1MessageServiceMerkleProof, claimMessageCall, "MessageClaimed", [messageLeafHash]);
    });

    it("Should fail to claim when the contract is generally paused", async () => {
      const claimParams = buildDefaultClaimProof();

      await l1MessageServiceMerkleProof.connect(pauser).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWhenPaused(
        l1MessageService,
        l1MessageServiceMerkleProof.claimMessageWithProof({
          proof: claimParams.proof,
          messageNumber: 1,
          leafIndex: claimParams.index,
          from: admin.address,
          to: admin.address,
          fee: MESSAGE_FEE,
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
          feeRecipient: ADDRESS_ZERO,
          merkleRoot: claimParams.merkleRoot,
          data: EMPTY_CALLDATA,
        }),
        GENERAL_PAUSE_TYPE,
      );
    });

    it("Should fail when the message has already been claimed", async () => {
      const claimParams = buildDefaultClaimProof();

      await l1MessageServiceMerkleProof.addL2MerkleRoots([claimParams.merkleRoot], claimParams.proof.length);

      await l1MessageServiceMerkleProof.claimMessageWithProof({
        proof: claimParams.proof,
        messageNumber: 1,
        leafIndex: claimParams.index,
        from: admin.address,
        to: admin.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ADDRESS_ZERO,
        merkleRoot: claimParams.merkleRoot,
        data: EMPTY_CALLDATA,
      });

      const claimMessageCall = l1MessageServiceMerkleProof.claimMessageWithProof({
        proof: claimParams.proof,
        messageNumber: 1,
        leafIndex: claimParams.index,
        from: admin.address,
        to: admin.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ADDRESS_ZERO,
        merkleRoot: claimParams.merkleRoot,
        data: EMPTY_CALLDATA,
      });

      await expectRevertWithCustomError(l1MessageServiceMerkleProof, claimMessageCall, "MessageAlreadyClaimed", [1]);
    });

    it("Should fail when l2 Merkle root does not exist on L1", async () => {
      const claimParams = buildDefaultClaimProof();

      const claimMessageCall = l1MessageServiceMerkleProof.claimMessageWithProof({
        proof: claimParams.proof,
        messageNumber: 1,
        leafIndex: claimParams.index,
        from: admin.address,
        to: admin.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ADDRESS_ZERO,
        merkleRoot: claimParams.merkleRoot,
        data: EMPTY_CALLDATA,
      });

      await expectRevertWithCustomError(l1MessageService, claimMessageCall, "L2MerkleRootDoesNotExist");
    });

    it("Should fail when l2 merkle proof is invalid", async () => {
      const claimParams = buildDefaultClaimProof();
      const invalidClaimParams = buildInvalidClaimProof();

      await l1MessageServiceMerkleProof.addL2MerkleRoots([claimParams.merkleRoot], claimParams.proof.length);

      const claimMessageCall = l1MessageServiceMerkleProof.claimMessageWithProof({
        proof: invalidClaimParams.proof,
        messageNumber: 1,
        leafIndex: claimParams.index,
        from: admin.address,
        to: admin.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ADDRESS_ZERO,
        merkleRoot: claimParams.merkleRoot,
        data: EMPTY_CALLDATA,
      });

      await expectRevertWithCustomError(l1MessageService, claimMessageCall, "InvalidMerkleProof");
    });

    it("Should fail claiming when the call transaction fails with receive()", async () => {
      const claimParams = buildSingleLeafProof(
        {
          data: "0xcd4aed30",
          fee: MESSAGE_FEE,
          from: admin.address,
          messageNumber: 1n,
          to: await l1TestRevert.getAddress(),
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        },
        32,
      );

      await l1MessageServiceMerkleProof.addL2MerkleRoots([claimParams.merkleRoot], claimParams.proof.length);

      await expect(
        l1MessageServiceMerkleProof.claimMessageWithProof({
          proof: claimParams.proof,
          messageNumber: 1,
          leafIndex: claimParams.index,
          from: admin.address,
          to: await l1TestRevert.getAddress(),
          fee: MESSAGE_FEE,
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
          feeRecipient: ADDRESS_ZERO,
          merkleRoot: claimParams.merkleRoot,
          data: "0xcd4aed30",
        }),
      ).to.revert(ethers);
    });

    it("Should fail claiming when the call transaction fails with empty fallback", async () => {
      const claimParams = buildSingleLeafProof(
        {
          data: "0xce398a64",
          fee: MESSAGE_FEE,
          from: admin.address,
          messageNumber: 1n,
          to: await l1TestRevert.getAddress(),
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        },
        SHORT_DYNAMIC_MERKLE_PROOF_DEPTH,
      );

      await l1MessageServiceMerkleProof.addL2MerkleRoots([claimParams.merkleRoot], claimParams.proof.length);

      const claimMessageCall = l1MessageServiceMerkleProof.claimMessageWithProof({
        proof: claimParams.proof,
        messageNumber: 1,
        leafIndex: claimParams.index,
        from: admin.address,
        to: await l1TestRevert.getAddress(),
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ADDRESS_ZERO,
        merkleRoot: claimParams.merkleRoot,
        data: "0xce398a64",
      });

      await expectRevertWithCustomError(l1MessageService, claimMessageCall, "MessageSendingFailed", [
        await l1TestRevert.getAddress(),
      ]);
    });

    it("Should fail on reentry", async () => {
      const nestedClaimMessageData = l1MessageServiceMerkleProof.interface.encodeFunctionData("doReentryWithParams", [
        {
          proof: [],
          messageNumber: 1,
          leafIndex: 0,
          from: admin.address,
          to: admin.address,
          fee: MESSAGE_FEE,
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
          feeRecipient: ADDRESS_ZERO,
          merkleRoot: ethers.ZeroHash,
          data: EMPTY_CALLDATA,
        },
      ]);

      const claimParams = buildSingleLeafProof(
        {
          data: nestedClaimMessageData,
          fee: MESSAGE_FEE,
          from: admin.address,
          messageNumber: 1n,
          to: await l1MessageServiceMerkleProof.getAddress(),
          value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        },
        SHORT_DYNAMIC_MERKLE_PROOF_DEPTH,
      );

      await l1MessageServiceMerkleProof.addL2MerkleRoots([claimParams.merkleRoot], claimParams.proof.length);

      const claimMessageWithProof = l1MessageServiceMerkleProof.claimMessageWithProof({
        proof: claimParams.proof,
        messageNumber: 1,
        leafIndex: claimParams.index,
        from: admin.address,
        to: await l1MessageServiceMerkleProof.getAddress(),
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ADDRESS_ZERO,
        merkleRoot: claimParams.merkleRoot,
        data: nestedClaimMessageData,
      });

      await expectRevertWithCustomError(l1MessageServiceMerkleProof, claimMessageWithProof, "ReentrantCall");
    });

    it("Should fail when the fee recipient fails errors", async () => {
      const claimParams = buildDefaultClaimProof();

      await l1MessageServiceMerkleProof.addL2MerkleRoots([claimParams.merkleRoot], claimParams.proof.length);

      const claimMessageCall = l1MessageServiceMerkleProof.claimMessageWithProof({
        proof: claimParams.proof,
        messageNumber: 1,
        leafIndex: claimParams.index,
        from: admin.address,
        to: admin.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: await l1MessageService.getAddress(),
        merkleRoot: claimParams.merkleRoot,
        data: EMPTY_CALLDATA,
      });

      await expectRevertWithCustomError(l1MessageServiceMerkleProof, claimMessageCall, "FeePaymentFailed", [
        await l1MessageService.getAddress(),
      ]);
    });

    it("Should fail when the merkle depth is different than the proof length", async () => {
      const claimParams = buildDefaultClaimProof();

      await l1MessageServiceMerkleProof.addL2MerkleRoots([claimParams.merkleRoot], claimParams.proof.length);

      const merkleDepth = await l1MessageServiceMerkleProof.l2MerkleRootsDepths(claimParams.merkleRoot);

      const claimMessageCall = l1MessageServiceMerkleProof.claimMessageWithProof({
        proof: claimParams.proof.slice(0, -1),
        messageNumber: 1,
        leafIndex: claimParams.index,
        from: admin.address,
        to: admin.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: await l1MessageService.getAddress(),
        merkleRoot: claimParams.merkleRoot,
        data: EMPTY_CALLDATA,
      });

      await expectRevertWithCustomError(
        l1MessageServiceMerkleProof,
        claimMessageCall,
        "ProofLengthDifferentThanMerkleDepth",
        [merkleDepth, claimParams.proof.slice(0, -1).length],
      );
    });
  });

  describe("Resetting limits", () => {
    it("Should reset limits as limitSetter", async () => {
      let usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);

      await setHashAndClaimMessage(MESSAGE_FEE, MESSAGE_VALUE_1ETH);

      usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(MESSAGE_FEE + MESSAGE_VALUE_1ETH);

      await l1MessageService.connect(limitSetter).resetAmountUsedInPeriod();
      usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);
    });

    it("Should fail reset limits as non-RATE_LIMIT_SETTER_ROLE", async () => {
      let usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);

      await setHashAndClaimMessage(MESSAGE_FEE, MESSAGE_VALUE_1ETH);

      usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(MESSAGE_FEE + MESSAGE_VALUE_1ETH);

      await expectRevertWithReason(
        l1MessageService.connect(admin).resetAmountUsedInPeriod(),
        buildAccessErrorMessage(admin, USED_RATE_LIMIT_RESETTER_ROLE),
      );

      usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(MESSAGE_FEE + MESSAGE_VALUE_1ETH);
    });
  });

  describe("Pausing contracts", () => {
    it("Should fail general pausing as non-pauser", async () => {
      await expectNotPaused(l1MessageService, GENERAL_PAUSE_TYPE);

      await expectRevertWithReason(
        l1MessageService.connect(admin).pauseByType(GENERAL_PAUSE_TYPE),
        buildAccessErrorMessage(admin, PAUSE_ALL_ROLE),
      );

      await expectNotPaused(l1MessageService, GENERAL_PAUSE_TYPE);
    });

    it("Should pause generally as pause manager", async () => {
      await expectNotPaused(l1MessageService, GENERAL_PAUSE_TYPE);

      await l1MessageService.connect(pauser).pauseByType(GENERAL_PAUSE_TYPE);

      await expectPaused(l1MessageService, GENERAL_PAUSE_TYPE);
    });

    it("Should fail when to claim the contract is generally paused", async () => {
      await l1MessageService.connect(pauser).pauseByType(GENERAL_PAUSE_TYPE);

      const claimMessageCall = l1MessageService.claimMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        ADDRESS_ZERO,
        EMPTY_CALLDATA,
        1,
      );

      await expectRevertWhenPaused(l1MessageService, claimMessageCall, GENERAL_PAUSE_TYPE);
    });

    it("Should fail to claim when the L2 to L1 communication is paused", async () => {
      await l1MessageService.connect(pauser).pauseByType(L2_L1_PAUSE_TYPE);

      const claimMessageCall = l1MessageService.claimMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        MESSAGE_FEE,
        MESSAGE_VALUE_1ETH,
        ADDRESS_ZERO,
        EMPTY_CALLDATA,
        1,
      );

      await expectRevertWhenPaused(l1MessageService, claimMessageCall, L2_L1_PAUSE_TYPE);
    });

    it("Should fail to send if the contract is generally paused", async () => {
      await l1MessageService.connect(pauser).pauseByType(GENERAL_PAUSE_TYPE);

      const claimMessageCall = l1MessageService
        .connect(admin)
        .canSendMessage(notAuthorizedAccount.address, 0, EMPTY_CALLDATA, { value: INITIAL_WITHDRAW_LIMIT });

      await expectRevertWhenPaused(l1MessageService, claimMessageCall, GENERAL_PAUSE_TYPE);

      const usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);
    });

    it("Should fail to send if L1 to L2 communication is paused", async () => {
      await l1MessageService.connect(pauser).pauseByType(L1_L2_PAUSE_TYPE);

      const claimMessageCall = l1MessageService
        .connect(admin)
        .canSendMessage(notAuthorizedAccount.address, 0, EMPTY_CALLDATA, { value: INITIAL_WITHDRAW_LIMIT });

      await expectRevertWhenPaused(l1MessageService, claimMessageCall, L1_L2_PAUSE_TYPE);

      const usedAmount = await l1MessageService.currentPeriodAmountInWei();
      expect(usedAmount).to.be.equal(0);
    });
  });

  async function setHash(fee: bigint, value: bigint) {
    const expectedBytes = encodeSendMessage(
      await l1MessageService.getAddress(),
      notAuthorizedAccount.address,
      fee,
      value,
      1n,
      EMPTY_CALLDATA,
    );

    await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));
  }

  async function setHashAndClaimMessage(fee: bigint, value: bigint) {
    const expectedBytes = encodeSendMessage(
      await l1MessageService.getAddress(),
      notAuthorizedAccount.address,
      fee,
      value,
      1n,
      EMPTY_CALLDATA,
    );

    await l1MessageService.addL2L1MessageHash(ethers.keccak256(expectedBytes));

    await expect(
      l1MessageService.claimMessage(
        await l1MessageService.getAddress(),
        notAuthorizedAccount.address,
        fee,
        value,
        postmanAddress.address,
        EMPTY_CALLDATA,
        1,
      ),
    ).to.not.revert(ethers);
  }
});
