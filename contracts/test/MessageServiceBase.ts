import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestL2MessageService, TestMessageServiceBase } from "../typechain-types";
import { INITIAL_WITHDRAW_LIMIT } from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";

describe("MessageServiceBase", () => {
  let messageServiceBase: TestMessageServiceBase;
  let messageService: TestL2MessageService;
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let admin: SignerWithAddress;
  let remoteSender: SignerWithAddress;
  let securityCouncil: SignerWithAddress;
  let l1L2MessageSetter: SignerWithAddress;

  async function deployMessageServiceBaseFixture() {
    const messageService = (await deployUpgradableFromFactory("TestL2MessageService", [
      securityCouncil.address,
      l1L2MessageSetter.address,
      86400,
      INITIAL_WITHDRAW_LIMIT,
    ])) as TestL2MessageService;

    const messageServiceBase = (await deployUpgradableFromFactory("TestMessageServiceBase", [
      messageService.address,
      remoteSender.address,
    ])) as TestMessageServiceBase;
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
      await expect(
        deployUpgradableFromFactory("TestMessageServiceBase", [ethers.constants.AddressZero, remoteSender.address]),
      ).to.be.revertedWithCustomError(messageServiceBase, "ZeroAddressNotAllowed");
    });

    it("It should fail when not initializing", async () => {
      await expect(messageServiceBase.tryInitialize(messageService.address, remoteSender.address)).to.be.revertedWith(
        "Initializable: contract is not initializing",
      );
    });

    it("Should revert if remote sender address is address(0)", async () => {
      await expect(
        deployUpgradableFromFactory("TestMessageServiceBase", [messageService.address, ethers.constants.AddressZero]),
      ).to.be.revertedWithCustomError(messageServiceBase, "ZeroAddressNotAllowed");
    });

    it("Should set the value of remoteSender variable in storage", async () => {
      expect(await messageServiceBase.remoteSender()).to.equal(remoteSender.address);
    });

    it("Should set the value of messageService variable in storage", async () => {
      expect(await messageServiceBase.messageService()).to.equal(messageService.address);
    });
  });

  describe("onlyMessagingService() modifier", () => {
    it("Should revert if msg.sender is not the message service address", async () => {
      await expect(messageServiceBase.withOnlyMessagingService()).to.be.revertedWithCustomError(
        messageServiceBase,
        "CallerIsNotMessageService",
      );
    });

    it("Should succeed if msg.sender is the message service address", async () => {
      expect(await messageService.callMessageServiceBase(messageServiceBase.address)).to.not.be.reverted;
    });
  });

  describe("onlyAuthorizedRemoteSender() modifier", () => {
    it("Should revert if sender is not allowed", async () => {
      await expect(messageServiceBase.withOnlyAuthorizedRemoteSender()).to.be.revertedWithCustomError(
        messageServiceBase,
        "SenderNotAuthorized",
      );
    });

    it("Should succeed if original sender is allowed", async () => {
      const messageServiceBase = (await deployUpgradableFromFactory("TestMessageServiceBase", [
        messageService.address,
        "0x00000000000000000000000000000000075BCd15",
      ])) as TestMessageServiceBase;
      await expect(messageServiceBase.withOnlyAuthorizedRemoteSender()).to.not.be.reverted;
    });
  });
});
