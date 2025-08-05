import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { deployFromFactory } from "../common/deployment";
import { expectEvent, expectRevertWithCustomError } from "../common/helpers";
import { LineaSequencerUptimeFeed } from "../../../typechain-types";
import { ZeroAddress } from "ethers";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { expect } from "chai";

describe("LineaSequencerUptimeFeed", () => {
  let contract: LineaSequencerUptimeFeed;

  let admin: SignerWithAddress;
  let l2Sender: SignerWithAddress;
  let notAuthorized: SignerWithAddress;

  async function deployLineaSequencerUptimeFeedFixture() {
    const initialStatus = true;
    const adminAddress = admin.address;
    const l2SenderAddress = l2Sender.address;
    return deployFromFactory("LineaSequencerUptimeFeed", initialStatus, adminAddress, l2SenderAddress);
  }

  before(async () => {
    [admin, l2Sender, notAuthorized] = await ethers.getSigners();
  });

  beforeEach(async () => {
    contract = (await loadFixture(deployLineaSequencerUptimeFeedFixture)) as LineaSequencerUptimeFeed;
  });

  describe("constructor", () => {
    it("should revert when admin address is zero", async () => {
      const initialStatus = true;
      const adminAddress = ZeroAddress;
      await expectRevertWithCustomError(
        contract,
        deployFromFactory("LineaSequencerUptimeFeed", initialStatus, adminAddress, l2Sender.address),
        "ZeroAddressNotAllowed",
      );
    });

    it("should revert when l2Sender address is zero", async () => {
      const initialStatus = true;

      const l2SenderAddress = ZeroAddress;
      await expectRevertWithCustomError(
        contract,
        deployFromFactory("LineaSequencerUptimeFeed", initialStatus, admin.address, l2SenderAddress),
        "ZeroAddressNotAllowed",
      );
    });

    it("should set initial status correctly", async () => {
      const initialStatus = true;
      const contract = (await deployFromFactory(
        "LineaSequencerUptimeFeed",
        initialStatus,
        admin.address,
        l2Sender.address,
      )) as LineaSequencerUptimeFeed;
      const deployTx = contract.deploymentTransaction();
      const block = await ethers.provider.getBlock(deployTx!.blockNumber!);

      const latestAnswer = await contract.latestAnswer();
      expect(latestAnswer).to.equal(1n);
      expect(await contract.latestTimestamp()).to.be.equal(block!.timestamp);
      expect(await contract.latestRound()).to.be.equal(0n);
    });

    it("should set admin and l2Sender roles correctly", async () => {
      const initialStatus = true;
      const contract = (await deployFromFactory(
        "LineaSequencerUptimeFeed",
        initialStatus,
        admin.address,
        l2Sender.address,
      )) as LineaSequencerUptimeFeed;

      expect(await contract.hasRole(await contract.DEFAULT_ADMIN_ROLE(), admin.address)).to.be.true;
      expect(await contract.hasRole(await contract.L2_SENDER_ROLE(), l2Sender.address)).to.be.true;
    });
  });

  describe("typeAndVersion", () => {
    it("should return the correct type and version", async () => {
      const typeAndVersion = await contract.typeAndVersion();
      expect(typeAndVersion).to.equal("LineaSequencerUptimeFeed 1.0.0");
    });
  });

  describe("updateStatus", () => {
    it("should revert if caller does not have the L2_SENDER_ROLE", async () => {
      const newStatus = true;
      const timestamp = Date.now();

      await expectRevertWithCustomError(
        contract,
        contract.connect(notAuthorized).updateStatus(newStatus, timestamp),
        "InvalidSender",
      );
    });

    it("should ignore update if latest recorded timestamp is newer", async () => {
      const newStatus = true;
      const timestamp = Math.floor(Date.now() / 1000);

      await contract.connect(l2Sender).updateStatus(newStatus, timestamp);
      const startedAt = (await contract.latestRoundData()).startedAt;

      const oldTimestamp = startedAt - 1n;

      await expectEvent(contract, contract.connect(l2Sender).updateStatus(newStatus, oldTimestamp), "UpdateIgnored", [
        newStatus,
        startedAt,
        true,
        oldTimestamp,
      ]);

      const latestTimestamp = await contract.latestTimestamp();
      expect(latestTimestamp).to.equal(timestamp);
    });

    it("should update status and emit RoundUpdated event", async () => {
      const newStatus = true;
      const timestamp = Math.floor(Date.now() / 1000);
      await expectEvent(contract, contract.connect(l2Sender).updateStatus(newStatus, timestamp), "RoundUpdated", [
        newStatus ? 1n : 0n,
        timestamp,
      ]);
      const latestAnswer = await contract.latestAnswer();
      expect(latestAnswer).to.equal(newStatus ? 1n : 0n);
      expect(await contract.latestTimestamp()).to.be.equal(timestamp);
    });
  });
});
