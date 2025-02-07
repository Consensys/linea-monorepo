import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestL2MessageService, TestMessageServiceBase } from "../../../typechain-types";
import {
  INITIALIZED_ERROR_MESSAGE,
  INITIAL_WITHDRAW_LIMIT,
  L1_L2_MESSAGE_SETTER_ROLE,
  ONE_DAY_IN_SECONDS,
  pauseTypeRoles,
  unpauseTypeRoles,
} from "../common/constants";
import { deployUpgradableFromFactory } from "../common/deployment";
import { expectEvent, expectRevertWithCustomError, expectRevertWithReason } from "../common/helpers";
import { generateRoleAssignments } from "contracts/common/helpers";
import { L2_MESSAGE_SERVICE_ROLES } from "contracts/common/constants";

describe("MessageServiceBase", () => {
  let messageServiceBase: TestMessageServiceBase;
  let messageService: TestL2MessageService;
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let admin: SignerWithAddress;
  let remoteSender: SignerWithAddress;
  let securityCouncil: SignerWithAddress;
  let l1L2MessageSetter: SignerWithAddress;

  async function deployMessageServiceBaseFixture() {
    const roleAddresses = generateRoleAssignments(L2_MESSAGE_SERVICE_ROLES, securityCouncil.address, [
      { role: L1_L2_MESSAGE_SETTER_ROLE, addresses: [l1L2MessageSetter.address] },
    ]);

    const messageService = (await deployUpgradableFromFactory("TestL2MessageService", [
      ONE_DAY_IN_SECONDS,
      INITIAL_WITHDRAW_LIMIT,
      securityCouncil.address,
      roleAddresses,
      pauseTypeRoles,
      unpauseTypeRoles,
    ])) as unknown as TestL2MessageService;

    const messageServiceBase = (await deployUpgradableFromFactory("TestMessageServiceBase", [
      await messageService.getAddress(),
      remoteSender.address,
    ])) as unknown as TestMessageServiceBase;
    return { messageService, messageServiceBase };
  }

  beforeEach(async () => {
    [admin, remoteSender, securityCouncil, l1L2MessageSetter] = await ethers.getSigners();
    const contracts = await loadFixture(deployMessageServiceBaseFixture);
    messageService = contracts.messageService;
    messageServiceBase = contracts.messageServiceBase;
  });

  describe("Initialization checks", () => {
    it("Should revert if message service address is address(0)", async () => {
      await expectRevertWithCustomError(
        messageService,
        deployUpgradableFromFactory("TestMessageServiceBase", [ethers.ZeroAddress, remoteSender.address]),
        "ZeroAddressNotAllowed",
      );
    });

    it("It should fail when not initializing", async () => {
      await expectRevertWithReason(
        messageServiceBase.tryInitialize(await messageService.getAddress(), remoteSender.address),
        INITIALIZED_ERROR_MESSAGE,
      );
    });

    it("Should revert if remote sender address is address(0)", async () => {
      await expectRevertWithCustomError(
        messageServiceBase,
        deployUpgradableFromFactory("TestMessageServiceBase", [await messageService.getAddress(), ethers.ZeroAddress]),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should set the value of remoteSender variable in storage", async () => {
      expect(await messageServiceBase.remoteSender()).to.equal(remoteSender.address);
    });

    it("Should set the value of messageService variable in storage", async () => {
      expect(await messageServiceBase.messageService()).to.equal(await messageService.getAddress());
    });
  });

  describe("RemoteSenderSet event", () => {
    it("Should emit RemoteSenderSet event when testSetRemoteSender is called", async () => {
      const newRemoteSender = ethers.Wallet.createRandom().address;
      await expectEvent(
        messageServiceBase,
        messageServiceBase.testSetRemoteSender(newRemoteSender),
        "RemoteSenderSet",
        [newRemoteSender, admin.address],
      );
    });
  });

  describe("onlyMessagingService() modifier", () => {
    it("Should revert if msg.sender is not the message service address", async () => {
      await expectRevertWithCustomError(
        messageServiceBase,
        messageServiceBase.withOnlyMessagingService(),
        "CallerIsNotMessageService",
      );
    });

    it("Should succeed if msg.sender is the message service address", async () => {
      expect(await messageService.callMessageServiceBase(await messageServiceBase.getAddress())).to.not.be.reverted;
    });
  });

  describe("onlyAuthorizedRemoteSender() modifier", () => {
    it("Should revert if sender is not allowed", async () => {
      await expectRevertWithCustomError(
        messageServiceBase,
        messageServiceBase.withOnlyAuthorizedRemoteSender(),
        "SenderNotAuthorized",
      );
    });

    /**
     * TO DISCUSS
     * Suggest remove this test as it is coupled to implementation detail of `_messageSender` in storage.
     * It is replaced with a transient storage slot.
     * Regular storage slot can have a default non-zero value, transient storage slot cannot.
     * ...
     * Furthermore the original test is really testing for a specific magic value 'address(123456789)'
     * that we hardcoded as the default value for the storage slot `_messageSender`.
     * Arguably the original test is not testing for useful behaviour - it is not really testing the case stated in the 'it' statement
     * ...
     * What we care about is that onlyAuthorizedRemoteSender modifier is authorizing calls correctly.
     * It's a slightly tricky modifier to test because the variable it checks - `messageService.sender()` - is a valid value
     * only during an intermediate state in the middle of function flow, and not at start or end of the transaction.
     * ...
     * Proposed testup
     * - Construct a call A from `remoteSender` to `messageService`
     * - Call A will created a nested call from `messageService` to `messageServiceBase`, invoking onlyAuthorizedRemoteSender modifier
     * - We can enable this function flow with test fixture methods if needed
     */

    // it("Should succeed if original sender is allowed", async () => {
    //   const messageServiceBase = (await deployUpgradableFromFactory("TestMessageServiceBase", [
    //     await messageService.getAddress(),
    //     // Zero address fails here because we cannot initialize with zero address
    //     // Magic value for address(123456789) which was old L2MessageServiceV1.DEFAULT_SENDER_ADDRESS
    //     "0x00000000000000000000000000000000075BCd15",
    //   ])) as unknown as TestMessageServiceBase;
    //   await expect(messageServiceBase.withOnlyAuthorizedRemoteSender()).to.not.be.reverted;
    // });

    it("Should succeed if original sender is allowed", async () => {
      // Construct a call A from `remoteSender` to `messageService`
      // Call A will created a nested call from `messageService` to `messageServiceBase`, invoking onlyAuthorizedRemoteSender modifier
      const call = messageService.simulateClaimMessageWithoutChecks(
        remoteSender,
        messageServiceBase,
        0,
        "0xfcd38105", // keccak256("withOnlyAuthorizedRemoteSender()")
      );
      await expect(call).to.not.be.reverted;
    });
  });
});
