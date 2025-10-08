import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

import { TestYieldManager } from "contracts/typechain-types";
import { deployYieldManagerForUnitTest } from "../helpers/deploy";
import { addMockYieldProvider } from "../helpers/mocks";
import {
  GENERAL_PAUSE_TYPE,
  NATIVE_YIELD_STAKING_PAUSE_TYPE,
  NATIVE_YIELD_UNSTAKING_PAUSE_TYPE,
  NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE,
  NATIVE_YIELD_PERMISSIONLESS_REBALANCE_PAUSE_TYPE,
  NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE,
  NATIVE_YIELD_REPORTING_PAUSE_TYPE,
  LST_WITHDRAWAL_PAUSE_TYPE,
} from "../../common/constants";
import { setWithdrawalReserveBalance, setWithdrawalReserveToMinimum } from "../helpers/setup";
import { buildAccessErrorMessage, expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";

describe("YieldManager contract - control operations", () => {
  let yieldManager: TestYieldManager;

  let securityCouncil: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  let nativeYieldOperator: SignerWithAddress;

  before(async () => {
    ({ securityCouncil, operator: nonAuthorizedAccount, nativeYieldOperator } = await loadFixture(getAccountsFixture));
  });

  describe("pausing", () => {
    it("Security council should be able to activate GENERAL_PAUSE_TYPE", async () => {
      await expect(yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE))
        .to.emit(yieldManager, "Paused")
        .withArgs(securityCouncil.address, GENERAL_PAUSE_TYPE);
      expect(await yieldManager.isPaused(GENERAL_PAUSE_TYPE)).to.be.true;
    });

    it("Security council should be able to activate NATIVE_YIELD_STAKING_PAUSE_TYPE", async () => {
      await expect(yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_STAKING_PAUSE_TYPE))
        .to.emit(yieldManager, "Paused")
        .withArgs(securityCouncil.address, NATIVE_YIELD_STAKING_PAUSE_TYPE);
      expect(await yieldManager.isPaused(NATIVE_YIELD_STAKING_PAUSE_TYPE)).to.be.true;
    });

    it("Security council should be able to activate NATIVE_YIELD_UNSTAKING_PAUSE_TYPE", async () => {
      await expect(yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_UNSTAKING_PAUSE_TYPE))
        .to.emit(yieldManager, "Paused")
        .withArgs(securityCouncil.address, NATIVE_YIELD_UNSTAKING_PAUSE_TYPE);
      expect(await yieldManager.isPaused(NATIVE_YIELD_UNSTAKING_PAUSE_TYPE)).to.be.true;
    });

    it("Security council should be able to activate NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE", async () => {
      await expect(yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE))
        .to.emit(yieldManager, "Paused")
        .withArgs(securityCouncil.address, NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE);
      expect(await yieldManager.isPaused(NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE)).to.be.true;
    });

    it("Security council should be able to activate NATIVE_YIELD_PERMISSIONLESS_REBALANCE_PAUSE_TYPE", async () => {
      await expect(yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_PERMISSIONLESS_REBALANCE_PAUSE_TYPE))
        .to.emit(yieldManager, "Paused")
        .withArgs(securityCouncil.address, NATIVE_YIELD_PERMISSIONLESS_REBALANCE_PAUSE_TYPE);
      expect(await yieldManager.isPaused(NATIVE_YIELD_PERMISSIONLESS_REBALANCE_PAUSE_TYPE)).to.be.true;
    });

    it("Security council should be able to activate NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE", async () => {
      await expect(yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE))
        .to.emit(yieldManager, "Paused")
        .withArgs(securityCouncil.address, NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE);
      expect(await yieldManager.isPaused(NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE)).to.be.true;
    });

    it("Security council should be able to activate NATIVE_YIELD_REPORTING_PAUSE_TYPE", async () => {
      await expect(yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_REPORTING_PAUSE_TYPE))
        .to.emit(yieldManager, "Paused")
        .withArgs(securityCouncil.address, NATIVE_YIELD_REPORTING_PAUSE_TYPE);
      expect(await yieldManager.isPaused(NATIVE_YIELD_REPORTING_PAUSE_TYPE)).to.be.true;
    });

    it("Security council should be able to activate LST_WITHDRAWAL_PAUSE_TYPE", async () => {
      await expect(yieldManager.connect(securityCouncil).pauseByType(LST_WITHDRAWAL_PAUSE_TYPE))
        .to.emit(yieldManager, "Paused")
        .withArgs(securityCouncil.address, LST_WITHDRAWAL_PAUSE_TYPE);
      expect(await yieldManager.isPaused(LST_WITHDRAWAL_PAUSE_TYPE)).to.be.true;
    });
  });

  describe("unpausing", () => {
    it("Security council should be able to unpause GENERAL_PAUSE_TYPE", async () => {
      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);
      await expect(yieldManager.connect(securityCouncil).unPauseByType(GENERAL_PAUSE_TYPE))
        .to.emit(yieldManager, "UnPaused")
        .withArgs(securityCouncil.address, GENERAL_PAUSE_TYPE);
      expect(await yieldManager.isPaused(GENERAL_PAUSE_TYPE)).to.be.false;
    });

    async function pauseThenUnpause(pauseType: bigint) {
      await yieldManager.connect(securityCouncil).pauseByType(pauseType);
      await expect(yieldManager.connect(securityCouncil).unPauseByType(pauseType))
        .to.emit(yieldManager, "UnPaused")
        .withArgs(securityCouncil.address, pauseType);
      expect(await yieldManager.isPaused(pauseType)).to.be.false;
    }

    it("Security council should be able to unpause NATIVE_YIELD_STAKING_PAUSE_TYPE", async () => {
      await pauseThenUnpause(NATIVE_YIELD_STAKING_PAUSE_TYPE);
    });

    it("Security council should be able to unpause NATIVE_YIELD_UNSTAKING_PAUSE_TYPE", async () => {
      await pauseThenUnpause(NATIVE_YIELD_UNSTAKING_PAUSE_TYPE);
    });

    it("Security council should be able to unpause NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE", async () => {
      await pauseThenUnpause(NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE);
    });

    it("Security council should be able to unpause NATIVE_YIELD_PERMISSIONLESS_REBALANCE_PAUSE_TYPE", async () => {
      await pauseThenUnpause(NATIVE_YIELD_PERMISSIONLESS_REBALANCE_PAUSE_TYPE);
    });

    it("Security council should be able to unpause NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE", async () => {
      await pauseThenUnpause(NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE);
    });

    it("Security council should be able to unpause NATIVE_YIELD_REPORTING_PAUSE_TYPE", async () => {
      await pauseThenUnpause(NATIVE_YIELD_REPORTING_PAUSE_TYPE);
    });

    it("Security council should be able to unpause LST_WITHDRAWAL_PAUSE_TYPE", async () => {
      await pauseThenUnpause(LST_WITHDRAWAL_PAUSE_TYPE);
    });
  });

  beforeEach(async () => {
    ({ yieldManager } = await loadFixture(deployYieldManagerForUnitTest));
  });

  // pauseStaking() unit tests
  describe("Pausing staking", () => {
    it("Should revert when adding if the caller does not have the STAKING_PAUSER_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      const requiredRole = await yieldManager.STAKING_PAUSER_ROLE();

      await expect(
        yieldManager.connect(nonAuthorizedAccount).pauseStaking(mockYieldProviderAddress),
      ).to.be.revertedWith(buildAccessErrorMessage(nonAuthorizedAccount, requiredRole));
    });

    it("Should revert when pausing an unknown yield provider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).pauseStaking(ethers.Wallet.createRandom().address),
        "UnknownYieldProvider",
      );
    });

    it("Should revert if staking already paused", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(nativeYieldOperator).pauseStaking(mockYieldProviderAddress);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).pauseStaking(mockYieldProviderAddress),
        "StakingAlreadyPaused",
      );
    });

    it("Should successfully pause and emit the correct event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expect(yieldManager.connect(nativeYieldOperator).pauseStaking(mockYieldProviderAddress))
        .to.emit(yieldManager, "YieldProviderStakingPaused")
        .withArgs(mockYieldProviderAddress);

      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
    });
  });

  // pauseStakingIfNotAlready() unit tests
  describe("Pausing staking if not already", () => {
    it("If not currently paused, should pause", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;

      await yieldManager.connect(nativeYieldOperator).pauseStakingIfNotAlready(mockYieldProviderAddress);

      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
    });

    it("If already paused, no-op", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.setYieldProviderIsStakingPaused(mockYieldProviderAddress, true);

      await yieldManager.connect(nativeYieldOperator).pauseStakingIfNotAlready(mockYieldProviderAddress);

      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
    });
  });

  // unpauseStaking() unit tests
  describe("Unpausing staking", () => {
    it("Should revert when adding if the caller does not have the STAKING_UNPAUSER_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      const requiredRole = await yieldManager.STAKING_UNPAUSER_ROLE();

      await expect(
        yieldManager.connect(nonAuthorizedAccount).unpauseStaking(mockYieldProviderAddress),
      ).to.be.revertedWith(buildAccessErrorMessage(nonAuthorizedAccount, requiredRole));
    });

    it("Should revert when pausing an unknown yield provider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).unpauseStaking(ethers.Wallet.createRandom().address),
        "UnknownYieldProvider",
      );
    });

    it("Should revert if staking already unpaused", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).unpauseStaking(mockYieldProviderAddress),
        "StakingAlreadyUnpaused",
      );
    });

    it("Should revert if the withdrawal reserve is in deficit", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(nativeYieldOperator).pauseStaking(mockYieldProviderAddress);
      await setWithdrawalReserveBalance(yieldManager, 0n);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).unpauseStaking(mockYieldProviderAddress),
        "InsufficientWithdrawalReserve",
      );
    });

    it("Should revert if ossification has been initiated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(nativeYieldOperator).pauseStaking(mockYieldProviderAddress);
      await setWithdrawalReserveToMinimum(yieldManager);
      await yieldManager.setYieldProviderIsOssificationInitiated(mockYieldProviderAddress, true);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).unpauseStaking(mockYieldProviderAddress),
        "UnpauseStakingForbiddenDuringOssification",
      );
    });

    it("Should revert if ossification has completed", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(nativeYieldOperator).pauseStaking(mockYieldProviderAddress);
      await setWithdrawalReserveToMinimum(yieldManager);
      await yieldManager.setYieldProviderIsOssified(mockYieldProviderAddress, true);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).unpauseStaking(mockYieldProviderAddress),
        "UnpauseStakingForbiddenDuringOssification",
      );
    });

    it("Should revert if there is a current lst liability", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(nativeYieldOperator).pauseStaking(mockYieldProviderAddress);
      await setWithdrawalReserveToMinimum(yieldManager);
      await yieldManager.setYieldProviderLstLiabilityPrincipal(mockYieldProviderAddress, 1n);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).unpauseStaking(mockYieldProviderAddress),
        "UnpauseStakingForbiddenWithCurrentLSTLiability",
      );
    });

    it("Should successfully unpause and emit the correct event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(nativeYieldOperator).pauseStaking(mockYieldProviderAddress);
      await setWithdrawalReserveToMinimum(yieldManager);

      await expect(yieldManager.connect(nativeYieldOperator).unpauseStaking(mockYieldProviderAddress))
        .to.emit(yieldManager, "YieldProviderStakingUnpaused")
        .withArgs(mockYieldProviderAddress);

      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
    });
  });

  // initiateOssification() unit tests
  describe("Initiate ossification", () => {
    it("Should revert when adding if the caller does not have the OSSIFIER_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      const requiredRole = await yieldManager.OSSIFIER_ROLE();

      await expect(
        yieldManager.connect(nativeYieldOperator).initiateOssification(mockYieldProviderAddress),
      ).to.be.revertedWith(buildAccessErrorMessage(nativeYieldOperator, requiredRole));
    });

    it("Should revert when requesting for an unknown yield provider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(securityCouncil).initiateOssification(ethers.Wallet.createRandom().address),
        "UnknownYieldProvider",
      );
    });

    it("Should revert if ossification has completed", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.setYieldProviderIsOssified(mockYieldProviderAddress, true);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(securityCouncil).initiateOssification(mockYieldProviderAddress),
        "AlreadyOssified",
      );
    });

    it("Should successfully initiate ossification, pause staking and emit the correct event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expect(yieldManager.connect(securityCouncil).initiateOssification(mockYieldProviderAddress))
        .to.emit(yieldManager, "YieldProviderOssificationInitiated")
        .withArgs(mockYieldProviderAddress);

      expect(await yieldManager.isOssificationInitiated(mockYieldProviderAddress)).to.be.true;
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
    });
  });

  // undoInitiateOssification() unit tests
  describe("Undo initiate ossification", () => {
    it("Should revert when adding if the caller does not have the OSSIFIER_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      const requiredRole = await yieldManager.OSSIFIER_ROLE();

      await expect(
        yieldManager.connect(nativeYieldOperator).undoInitiateOssification(mockYieldProviderAddress),
      ).to.be.revertedWith(buildAccessErrorMessage(nativeYieldOperator, requiredRole));
    });

    it("Should revert when requesting for an unknown yield provider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(securityCouncil).undoInitiateOssification(ethers.Wallet.createRandom().address),
        "UnknownYieldProvider",
      );
    });

    it("Should revert if ossification not initiated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(securityCouncil).undoInitiateOssification(mockYieldProviderAddress),
        "OssificationNotInitiated",
      );
    });

    it("Should revert if ossification completed", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.setYieldProviderIsOssificationInitiated(mockYieldProviderAddress, true);
      await yieldManager.setYieldProviderIsOssified(mockYieldProviderAddress, true);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(securityCouncil).undoInitiateOssification(mockYieldProviderAddress),
        "AlreadyOssified",
      );
    });

    it("Should successfully revert previous ossification initiation, unpause staking and emit the correct event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).initiateOssification(mockYieldProviderAddress);

      await expect(yieldManager.connect(securityCouncil).undoInitiateOssification(mockYieldProviderAddress))
        .to.emit(yieldManager, "YieldProviderOssificationReverted")
        .withArgs(mockYieldProviderAddress);

      expect(await yieldManager.isOssificationInitiated(mockYieldProviderAddress)).to.be.false;
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
    });
  });

  // processPendingOssification() unit tests
  describe("Process pending ossification", () => {
    it("Should revert when adding if the caller does not have the OSSIFIER_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      const requiredRole = await yieldManager.OSSIFIER_ROLE();

      await expect(
        yieldManager.connect(nativeYieldOperator).processPendingOssification(mockYieldProviderAddress),
      ).to.be.revertedWith(buildAccessErrorMessage(nativeYieldOperator, requiredRole));
    });

    it("Should revert when requesting for an unknown yield provider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(securityCouncil).processPendingOssification(ethers.Wallet.createRandom().address),
        "UnknownYieldProvider",
      );
    });

    it("Should revert if ossification not initiated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(securityCouncil).processPendingOssification(mockYieldProviderAddress),
        "OssificationNotInitiated",
      );
    });

    it("Should revert if ossification completed", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.setYieldProviderIsOssificationInitiated(mockYieldProviderAddress, true);
      await yieldManager.setYieldProviderIsOssified(mockYieldProviderAddress, true);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(securityCouncil).processPendingOssification(mockYieldProviderAddress),
        "AlreadyOssified",
      );
    });

    it("If YieldProvider does not return completed ossification, should not complete ossification and emit the correct event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).initiateOssification(mockYieldProviderAddress);
      await yieldManager
        .connect(securityCouncil)
        .setProcessPendingOssificationReturnVal(mockYieldProviderAddress, false);

      const isOssificationComplete = await yieldManager
        .connect(securityCouncil)
        .processPendingOssification.staticCall(mockYieldProviderAddress);
      expect(isOssificationComplete).to.be.false;

      await expect(yieldManager.connect(securityCouncil).processPendingOssification(mockYieldProviderAddress))
        .to.emit(yieldManager, "YieldProviderOssificationProcessed")
        .withArgs(mockYieldProviderAddress, false);

      expect(await yieldManager.isOssificationInitiated(mockYieldProviderAddress)).to.be.true;
      expect(await yieldManager.isOssified(mockYieldProviderAddress)).to.be.false;
    });

    it("If YieldProvider returns completed ossification, should complete ossification and emit the correct event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).initiateOssification(mockYieldProviderAddress);
      await yieldManager
        .connect(securityCouncil)
        .setProcessPendingOssificationReturnVal(mockYieldProviderAddress, true);

      const isOssificationComplete = await yieldManager
        .connect(securityCouncil)
        .processPendingOssification.staticCall(mockYieldProviderAddress);
      expect(isOssificationComplete).to.be.true;

      await expect(yieldManager.connect(securityCouncil).processPendingOssification(mockYieldProviderAddress))
        .to.emit(yieldManager, "YieldProviderOssificationProcessed")
        .withArgs(mockYieldProviderAddress, true);

      expect(await yieldManager.isOssified(mockYieldProviderAddress)).to.be.true;
    });
  });
});
