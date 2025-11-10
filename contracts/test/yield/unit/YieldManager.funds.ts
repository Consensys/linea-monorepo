// Unit tests on functions handling ETH transfer

import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

import { MockLineaRollup, TestYieldManager } from "contracts/typechain-types";
import { deployYieldManagerForUnitTest } from "../helpers/deploy";
import { addMockYieldProvider } from "../helpers/mocks";
import {
  ONE_THOUSAND_ETHER,
  ONE_ETHER,
  GENERAL_PAUSE_TYPE,
  NATIVE_YIELD_STAKING_PAUSE_TYPE,
  NATIVE_YIELD_REPORTING_PAUSE_TYPE,
  NATIVE_YIELD_UNSTAKING_PAUSE_TYPE,
  NATIVE_YIELD_PERMISSIONLESS_ACTIONS_PAUSE_TYPE,
} from "../../common/constants";
import { buildAccessErrorMessage, expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
import {
  decrementBalance,
  fundYieldProviderForWithdrawal,
  getBalance,
  incrementBalance,
  setBalance,
  setWithdrawalReserveToTarget,
} from "../helpers";

describe("YieldManager contract - ETH transfer operations", () => {
  let yieldManager: TestYieldManager;

  let securityCouncil: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  let nativeYieldOperator: SignerWithAddress;
  let l2YieldRecipient: SignerWithAddress;
  let mockLineaRollup: MockLineaRollup;

  const mockWithdrawalParams = ethers.hexlify(ethers.randomBytes(8));
  const mockWithdrawalParamsProof = ethers.hexlify(ethers.randomBytes(8));

  before(async () => {
    ({
      securityCouncil,
      operator: nonAuthorizedAccount,
      nativeYieldOperator,
      l2YieldRecipient,
    } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ yieldManager, mockLineaRollup } = await loadFixture(deployYieldManagerForUnitTest));
  });

  describe("receiving ETH from the L1MessageService", () => {
    it("Should revert when the caller is not the L1MessageService", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).receiveFundsFromReserve({ value: 1n }),
        "SenderNotL1MessageService",
      );
    });

    it("Should revert when the withdrawal reserve is below minimum", async () => {
      const l1MessageService = await mockLineaRollup.getAddress();
      const minimumEffectiveBalance = await yieldManager.getEffectiveMinimumWithdrawalReserve();
      const yieldManagerAddress = await yieldManager.getAddress();

      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(minimumEffectiveBalance)]);

      // Act - Use helper function on MockLineaRollup
      await expectRevertWithCustomError(
        yieldManager,
        mockLineaRollup.connect(nativeYieldOperator).callReceiveFundsFromReserve(yieldManagerAddress, 1n),
        "InsufficientWithdrawalReserve",
      );
    });

    it("Should successfully receive funds and emit the expected event", async () => {
      const l1MessageService = await mockLineaRollup.getAddress();
      const minimumEffectiveBalance = await yieldManager.getEffectiveMinimumWithdrawalReserve();
      const yieldManagerAddress = await yieldManager.getAddress();

      await ethers.provider.send("hardhat_setBalance", [
        l1MessageService,
        ethers.toBeHex(minimumEffectiveBalance + 1n),
      ]);

      await expect(mockLineaRollup.connect(nativeYieldOperator).callReceiveFundsFromReserve(yieldManagerAddress, 1n))
        .to.emit(yieldManager, "ReserveFundsReceived")
        .withArgs(1n);
    });
  });

  describe("sending ETH to the L1MessageService", () => {
    it("Should revert when the caller is not the YIELD_PROVIDER_UNSTAKER_ROLE", async () => {
      await expect(yieldManager.connect(nonAuthorizedAccount).transferFundsToReserve(1n)).to.be.revertedWith(
        buildAccessErrorMessage(nonAuthorizedAccount, await yieldManager.YIELD_PROVIDER_UNSTAKER_ROLE()),
      );
    });

    it("Should revert when the NATIVE_YIELD_UNSTAKING pause type is activated", async () => {
      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_UNSTAKING_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).transferFundsToReserve(1n),
        "IsPaused",
        [NATIVE_YIELD_UNSTAKING_PAUSE_TYPE],
      );
    });

    it("Should revert when the GENERAL pause type is activated", async () => {
      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).transferFundsToReserve(1n),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });

    it("Should successfully send ETH to the L1MessageService", async () => {
      const yieldManagerAddress = await yieldManager.getAddress();
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(ONE_THOUSAND_ETHER)]);
      const transferAmount = 1_000_000_000_000_000_000n;
      const l1MessageService = await mockLineaRollup.getAddress();

      const reserveBalanceBefore = await ethers.provider.getBalance(l1MessageService);
      const yieldManagerBalanceBefore = await ethers.provider.getBalance(yieldManagerAddress);

      await yieldManager.connect(nativeYieldOperator).transferFundsToReserve(transferAmount);

      const reserveBalanceAfter = await ethers.provider.getBalance(l1MessageService);
      const yieldManagerBalanceAfter = await ethers.provider.getBalance(yieldManagerAddress);

      expect(reserveBalanceAfter).to.equal(reserveBalanceBefore + transferAmount);
      expect(yieldManagerBalanceAfter).to.equal(yieldManagerBalanceBefore - transferAmount);
    });
  });

  describe("sending ETH to a YieldProvider", () => {
    it("Should revert when the caller is not the YIELD_PROVIDER_STAKING_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expect(
        yieldManager.connect(nonAuthorizedAccount).fundYieldProvider(mockYieldProviderAddress, 1n),
      ).to.be.revertedWith(
        buildAccessErrorMessage(nonAuthorizedAccount, await yieldManager.YIELD_PROVIDER_STAKING_ROLE()),
      );
    });

    it("Should revert when the NATIVE_YIELD_STAKING pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_STAKING_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).fundYieldProvider(mockYieldProviderAddress, 1n),
        "IsPaused",
        [NATIVE_YIELD_STAKING_PAUSE_TYPE],
      );
    });

    it("Should revert when the GENERAL pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).fundYieldProvider(mockYieldProviderAddress, 1n),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });

    it("Should revert when sending to an unknown YieldProvider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).fundYieldProvider(ethers.Wallet.createRandom().address, 1n),
        "UnknownYieldProvider",
      );
    });

    it("Should revert when the withdrawal reserve is in deficit", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).fundYieldProvider(mockYieldProviderAddress, 1n),
        "InsufficientWithdrawalReserve",
      );
    });
    it("With 0 LSTPrincipal payment, should successfully send ETH to the YieldProvider, update state and emit the expected event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const yieldManagerAddress = await yieldManager.getAddress();

      const minimumReserveAmount = await yieldManager.getMinimumWithdrawalReserveAmount();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(minimumReserveAmount)]);
      const transferAmount = 40n;
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(transferAmount)]);

      await expect(
        yieldManager.connect(nativeYieldOperator).fundYieldProvider(mockYieldProviderAddress, transferAmount),
      )
        .to.emit(yieldManager, "YieldProviderFunded")
        .withArgs(mockYieldProviderAddress, transferAmount, 0n, transferAmount);

      const yieldProviderData = await yieldManager.getYieldProviderData(mockYieldProviderAddress);
      expect(yieldProviderData.userFunds).to.equal(transferAmount);

      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(transferAmount);
    });

    it("With non-0 LSTPrincipal payment, should successfully send ETH to the YieldProvider, update state and emit the expected event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const yieldManagerAddress = await yieldManager.getAddress();

      const minimumReserveAmount = await yieldManager.getMinimumWithdrawalReserveAmount();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(minimumReserveAmount)]);
      const transferAmount = 40n;
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(transferAmount)]);
      // lstPrincipal can only be accrued when it is 'backed' by existing user funds
      const lstPrincipalPayment = 10n;
      await yieldManager.setYieldProviderUserFunds(mockYieldProviderAddress, lstPrincipalPayment);
      await yieldManager.setUserFundsInYieldProvidersTotal(lstPrincipalPayment);
      await yieldManager.setPayLSTPrincipalReturnVal(mockYieldProviderAddress, lstPrincipalPayment);

      await expect(
        yieldManager.connect(nativeYieldOperator).fundYieldProvider(mockYieldProviderAddress, transferAmount),
      )
        .to.emit(yieldManager, "YieldProviderFunded")
        .withArgs(mockYieldProviderAddress, transferAmount, lstPrincipalPayment, transferAmount - lstPrincipalPayment);

      expect(await yieldManager.userFunds(mockYieldProviderAddress)).to.equal(transferAmount);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(transferAmount);
    });
  });

  describe("reporting yield", () => {
    it("Should revert when the caller is not the YIELD_REPORTER_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expect(
        yieldManager.connect(nonAuthorizedAccount).reportYield(mockYieldProviderAddress, l2YieldRecipient.address),
      ).to.be.revertedWith(buildAccessErrorMessage(nonAuthorizedAccount, await yieldManager.YIELD_REPORTER_ROLE()));
    });

    it("Should revert when the NATIVE_YIELD_REPORTING pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_REPORTING_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).reportYield(mockYieldProviderAddress, l2YieldRecipient.address),
        "IsPaused",
        [NATIVE_YIELD_REPORTING_PAUSE_TYPE],
      );
    });

    it("Should revert when the GENERAL pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).reportYield(mockYieldProviderAddress, l2YieldRecipient.address),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });

    it("Should revert when reporting for an unknown YieldProvider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .reportYield(ethers.Wallet.createRandom().address, l2YieldRecipient.address),
        "UnknownYieldProvider",
      );
    });

    it("Should revert when distributing yield to an unknown L2YieldRecipient", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const unknownRecipient = ethers.Wallet.createRandom().address;

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).reportYield(mockYieldProviderAddress, unknownRecipient),
        "UnknownL2YieldRecipient",
      );
    });

    it("Should successfully report positive yield, update state and emit the expected event", async () => {
      // ARRANGE
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const reportedYield = ONE_ETHER;
      const outstandingNegativeYield = 0n;

      await yieldManager
        .connect(nativeYieldOperator)
        .setReportYieldReturnVal_NewReportedYield(mockYieldProviderAddress, reportedYield);
      await yieldManager
        .connect(nativeYieldOperator)
        .setReportYieldReturnVal_OutstandingNegativeYield(mockYieldProviderAddress, outstandingNegativeYield);

      // ACT + ASSERT
      await expect(
        yieldManager.connect(nativeYieldOperator).reportYield(mockYieldProviderAddress, l2YieldRecipient.address),
      )
        .to.emit(yieldManager, "NativeYieldReported")
        .withArgs(mockYieldProviderAddress, l2YieldRecipient.address, reportedYield, outstandingNegativeYield);

      const providerData = await yieldManager.getYieldProviderData(mockYieldProviderAddress);
      expect(providerData.userFunds).to.equal(reportedYield);
      expect(providerData.yieldReportedCumulative).to.equal(reportedYield);
      expect(await yieldManager.userFunds(mockYieldProviderAddress)).to.equal(reportedYield);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(reportedYield);
      expect(await yieldManager.getYieldProviderLastReportedNegativeYield(mockYieldProviderAddress)).to.equal(
        outstandingNegativeYield,
      );
    });

    it("Should successfully report negative yield, update state and emit the expected event", async () => {
      // ARRANGE
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const reportedYield = 0n;
      const outstandingNegativeYield = ONE_ETHER * 2n;

      await yieldManager
        .connect(nativeYieldOperator)
        .setReportYieldReturnVal_NewReportedYield(mockYieldProviderAddress, reportedYield);
      await yieldManager
        .connect(nativeYieldOperator)
        .setReportYieldReturnVal_OutstandingNegativeYield(mockYieldProviderAddress, outstandingNegativeYield);

      // ACT + ASSERT
      await expect(
        yieldManager.connect(nativeYieldOperator).reportYield(mockYieldProviderAddress, l2YieldRecipient.address),
      )
        .to.emit(yieldManager, "NativeYieldReported")
        .withArgs(mockYieldProviderAddress, l2YieldRecipient.address, reportedYield, outstandingNegativeYield);

      const providerData = await yieldManager.getYieldProviderData(mockYieldProviderAddress);
      expect(providerData.userFunds).to.equal(reportedYield);
      expect(providerData.yieldReportedCumulative).to.equal(reportedYield);
      expect(await yieldManager.userFunds(mockYieldProviderAddress)).to.equal(reportedYield);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(reportedYield);
      expect(await yieldManager.getYieldProviderLastReportedNegativeYield(mockYieldProviderAddress)).to.equal(
        outstandingNegativeYield,
      );
    });
  });

  describe("permissioned unstake", () => {
    it("Should revert when the caller is not the YIELD_PROVIDER_UNSTAKER_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expect(
        yieldManager.connect(nonAuthorizedAccount).unstake(mockYieldProviderAddress, mockWithdrawalParams),
      ).to.be.revertedWith(
        buildAccessErrorMessage(nonAuthorizedAccount, await yieldManager.YIELD_PROVIDER_UNSTAKER_ROLE()),
      );
    });

    it("Should revert when the GENERAL pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).unstake(mockYieldProviderAddress, mockWithdrawalParams),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });

    it("Should revert when the NATIVE_YIELD_UNSTAKING pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_UNSTAKING_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).unstake(mockYieldProviderAddress, mockWithdrawalParams),
        "IsPaused",
        [NATIVE_YIELD_UNSTAKING_PAUSE_TYPE],
      );
    });

    it("Should revert when unstaking from an unknown YieldProvider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).unstake(ethers.Wallet.createRandom().address, mockWithdrawalParams),
        "UnknownYieldProvider",
      );
    });

    it("Should successfully unstake from a YieldProvider", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expect(yieldManager.connect(nativeYieldOperator).unstake(mockYieldProviderAddress, mockWithdrawalParams)).to
        .not.be.reverted;
    });
  });

  describe("permissionless unstake", () => {
    it("Should revert when the GENERAL pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });

    it("Should revert when the NATIVE_YIELD_PERMISSIONLESS_ACTIONS pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_PERMISSIONLESS_ACTIONS_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof),
        "IsPaused",
        [NATIVE_YIELD_PERMISSIONLESS_ACTIONS_PAUSE_TYPE],
      );
    });

    it("Should revert when unstaking from an unknown YieldProvider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(ethers.Wallet.createRandom().address, mockWithdrawalParams, mockWithdrawalParamsProof),
        "UnknownYieldProvider",
      );
    });

    it("Should revert when the withdrawal reserve is not in deficit", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      const minimumReserveAmount = await yieldManager.getMinimumWithdrawalReserveAmount();
      await ethers.provider.send("hardhat_setBalance", [
        await mockLineaRollup.getAddress(),
        ethers.toBeHex(minimumReserveAmount),
      ]);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof),
        "WithdrawalReserveNotInDeficit",
      );
    });

    it("Should revert when there is sufficient withdrawable YieldProvider value to cover target deficit", async () => {
      // Arrange - Put targetDeficit on YieldProvider
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, targetReserveAmount);
      // Arrange - Ensure 0 balance on L1MessageService
      await ethers.provider.send("hardhat_setBalance", [await mockLineaRollup.getAddress(), ethers.toBeHex(0)]);

      // Act
      const unstakeAmount = 1n;
      await yieldManager
        .connect(nativeYieldOperator)
        .setUnstakePermissionlessReturnVal(mockYieldProviderAddress, unstakeAmount);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof),
        "PermissionlessUnstakeRequestPlusAvailableFundsExceedsTargetDeficit",
      );
    });

    it("Should revert when there is sufficient YieldManager balance to cover target deficit", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      await ethers.provider.send("hardhat_setBalance", [
        await yieldManager.getAddress(),
        ethers.toBeHex(targetReserveAmount),
      ]);
      const unstakeAmount = 1n;
      await yieldManager
        .connect(nativeYieldOperator)
        .setUnstakePermissionlessReturnVal(mockYieldProviderAddress, unstakeAmount);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof),
        "PermissionlessUnstakeRequestPlusAvailableFundsExceedsTargetDeficit",
      );
    });

    it("Should revert when the YieldProvider returns 0 unstake amount", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      await ethers.provider.send("hardhat_setBalance", [
        await yieldManager.getAddress(),
        ethers.toBeHex(targetReserveAmount),
      ]);
      const unstakeAmount = 0n;
      await yieldManager
        .connect(nativeYieldOperator)
        .setUnstakePermissionlessReturnVal(mockYieldProviderAddress, unstakeAmount);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof),
        "YieldProviderReturnedZeroUnstakeAmount",
      );
    });

    it("Should successfully submit the unstake request, change state and emit the expected event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const unstakeAmount = targetReserveAmount;

      await yieldManager
        .connect(nativeYieldOperator)
        .setUnstakePermissionlessReturnVal(mockYieldProviderAddress, unstakeAmount);

      await yieldManager
        .connect(nativeYieldOperator)
        .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof);

      expect(await yieldManager.pendingPermissionlessUnstake()).to.equal(unstakeAmount);
    });

    it("Should revert if unstake amount is larger than target deficit", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const unstakeAmount = targetReserveAmount + 1n;
      await yieldManager
        .connect(nativeYieldOperator)
        .setUnstakePermissionlessReturnVal(mockYieldProviderAddress, unstakeAmount);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof),
        "PermissionlessUnstakeRequestPlusAvailableFundsExceedsTargetDeficit",
      );
    });

    it("After submitting one unstake request that restores the reserve deficit, the next permissionless request reverts", async () => {
      // Arrange - First do unstake permissionless up to maximum capacity
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const unstakeAmount = targetReserveAmount;

      await yieldManager
        .connect(nativeYieldOperator)
        .setUnstakePermissionlessReturnVal(mockYieldProviderAddress, unstakeAmount);

      await yieldManager
        .connect(nativeYieldOperator)
        .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof);

      expect(await yieldManager.pendingPermissionlessUnstake()).to.equal(unstakeAmount);

      // Arrange - Then do unstake of 1
      const secondUnstakeAmount = 1n;
      await yieldManager
        .connect(nativeYieldOperator)
        .setUnstakePermissionlessReturnVal(mockYieldProviderAddress, secondUnstakeAmount);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof),
        "PermissionlessUnstakeRequestPlusAvailableFundsExceedsTargetDeficit",
      );
    });
  });

  describe("delegatecall helper for withdraw from yield provider helper", () => {
    it("Should revert if the YieldProvider does not have sufficient balance", async () => {
      const { mockYieldProviderAddress, mockWithdrawTarget } = await addMockYieldProvider(yieldManager);
      const call = yieldManager
        .connect(nativeYieldOperator)
        .delegatecallWithdrawFromYieldProvider(mockYieldProviderAddress, 1n);

      await expect(call).to.be.revertedWithCustomError(mockWithdrawTarget, "MockWithdrawFailed");
    });

    it("Delegatecalls successfully and makes correct state transitions", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawAmount);

      const beforeYieldProviderUserFunds = await yieldManager.userFunds(mockYieldProviderAddress);
      const beforeUserFundsInYieldProvidersTotal = await yieldManager.userFundsInYieldProvidersTotal();

      // Act
      await yieldManager
        .connect(nativeYieldOperator)
        .delegatecallWithdrawFromYieldProvider(mockYieldProviderAddress, withdrawAmount);

      // Assert
      const afterYieldProviderUserFunds = await yieldManager.userFunds(mockYieldProviderAddress);
      const afterUserFundsInYieldProvidersTotal = await yieldManager.userFundsInYieldProvidersTotal();

      expect(afterYieldProviderUserFunds).to.equal(beforeYieldProviderUserFunds - withdrawAmount);
      expect(afterUserFundsInYieldProvidersTotal).to.equal(beforeUserFundsInYieldProvidersTotal - withdrawAmount);
    });
  });

  describe("decrementPendingPermissionlessUnstake helper", () => {
    it("if pendingPermissionlessUnstake = 0, should be a no-op", async () => {
      expect(await yieldManager.getPendingPermissionlessUnstake()).to.equal(0n);

      await yieldManager.connect(nativeYieldOperator).decrementPendingPermissionlessUnstake(ONE_ETHER);

      expect(await yieldManager.getPendingPermissionlessUnstake()).to.equal(0n);
    });

    it("if pendingPermissionlessUnstake <= _amount, should reduce pendingPermissionlessUnstake to 0", async () => {
      await yieldManager.setPendingPermissionlessUnstake(ONE_ETHER);

      await yieldManager.connect(nativeYieldOperator).decrementPendingPermissionlessUnstake(ONE_THOUSAND_ETHER);

      expect(await yieldManager.getPendingPermissionlessUnstake()).to.equal(0n);
    });

    it("if pendingPermissionlessUnstake > _amount, should reduce pendingPermissionlessUnstake accordingly", async () => {
      const startingPending = ONE_THOUSAND_ETHER;
      await yieldManager.setPendingPermissionlessUnstake(startingPending);

      await yieldManager.connect(nativeYieldOperator).decrementPendingPermissionlessUnstake(ONE_ETHER);

      expect(await yieldManager.getPendingPermissionlessUnstake()).to.equal(startingPending - ONE_ETHER);
    });
  });

  describe("withdraw with target deficit priority and lst liability principal reduction", () => {
    it("With 0 targetDeficit and 0 lstLiabilityPrincipal paid, should successfully withdraw the full _amount", async () => {
      // Arrange
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawRequestAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawRequestAmount);

      // Act
      const [actualWithdrawAmount, lstPrincipalPaid] =
        await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction.staticCall(
          mockYieldProviderAddress,
          withdrawRequestAmount,
          0,
        );

      // Assert
      expect(actualWithdrawAmount).eq(withdrawRequestAmount);
      expect(lstPrincipalPaid).eq(0);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
    });
    it("With 0 targetDeficit and lstLiabilityPrincipal paid < withdrawRequestAmount, should pay whole lstLiabilityPrincipal, and withdraw remainder from YieldProvider", async () => {
      // Arrange
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawRequestAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawRequestAmount);

      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, withdrawRequestAmount / 2n);

      // Act
      const [actualWithdrawAmount, lstPrincipalPaid] =
        await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction.staticCall(
          mockYieldProviderAddress,
          withdrawRequestAmount,
          0,
        );

      // Assert
      expect(actualWithdrawAmount).eq(withdrawRequestAmount / 2n);
      expect(lstPrincipalPaid).eq(withdrawRequestAmount / 2n);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
    });
    it("With 0 targetDeficit and lstLiabilityPrincipal paid = withdrawRequestAmount, should withdraw nothing from YieldProvider", async () => {
      // Arrange
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawRequestAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawRequestAmount);

      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, withdrawRequestAmount);

      // Act
      const [actualWithdrawAmount, lstPrincipalPaid] =
        await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction.staticCall(
          mockYieldProviderAddress,
          withdrawRequestAmount,
          0,
        );

      // Assert
      expect(actualWithdrawAmount).eq(0n);
      expect(lstPrincipalPaid).eq(withdrawRequestAmount);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
    });
    it("With targetDeficit > _amount and 0 lstLiabilityPrincipal paid, should successfully withdraw the full _amount and pause staking", async () => {
      // Arrange
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawRequestAmount = ONE_ETHER;
      const targetDeficit = withdrawRequestAmount + 1n;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawRequestAmount);

      // Act
      const [actualWithdrawAmount, lstPrincipalPaid] =
        await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction.staticCall(
          mockYieldProviderAddress,
          withdrawRequestAmount,
          targetDeficit,
        );
      await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction(
        mockYieldProviderAddress,
        withdrawRequestAmount,
        targetDeficit,
      );

      // Assert
      expect(actualWithdrawAmount).eq(withdrawRequestAmount);
      expect(lstPrincipalPaid).eq(0);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
    });
    it("With targetDeficit > _amount and non-0 lstLiabilityPrincipal paid, should withdraw full _amount, pause staking and make no liability payment", async () => {
      // Arrange
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawRequestAmount = ONE_ETHER;
      const lstLiabilityPrincipalForPayment = ONE_ETHER / 2n;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawRequestAmount);
      await yieldManager.setPayLSTPrincipalReturnVal(mockYieldProviderAddress, lstLiabilityPrincipalForPayment);
      const targetDeficit = withdrawRequestAmount + 1n;

      // Act
      const [actualWithdrawAmount, lstPrincipalPaid] =
        await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction.staticCall(
          mockYieldProviderAddress,
          withdrawRequestAmount,
          targetDeficit,
        );
      await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction(
        mockYieldProviderAddress,
        withdrawRequestAmount,
        targetDeficit,
      );

      // Assert
      expect(actualWithdrawAmount).eq(withdrawRequestAmount);
      expect(lstPrincipalPaid).eq(0);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
    });
    it("With targetDeficit < _amount and 0 lstLiabilityPrincipal paid, should successfully withdraw the full _amount", async () => {
      // Arrange
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawRequestAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawRequestAmount);
      const targetDeficit = withdrawRequestAmount - 1n;

      // Act
      const [actualWithdrawAmount, lstPrincipalPaid] =
        await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction.staticCall(
          mockYieldProviderAddress,
          withdrawRequestAmount,
          targetDeficit,
        );
      await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction(
        mockYieldProviderAddress,
        withdrawRequestAmount,
        targetDeficit,
      );

      // Assert
      expect(actualWithdrawAmount).eq(withdrawRequestAmount);
      expect(lstPrincipalPaid).eq(0);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
    });
    it("With targetDeficit < _amount and non-0 lstLiabilityPrincipal paid, should withdraw _amount - excess", async () => {
      // Arrange
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawRequestAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawRequestAmount);
      const deficitExcess = ONE_ETHER / 10n;
      const targetDeficit = withdrawRequestAmount - deficitExcess;

      // We trust implementation to return < availableAmount
      await yieldManager.setPayLSTPrincipalReturnVal(mockYieldProviderAddress, deficitExcess);

      // Act
      const [actualWithdrawAmount, lstPrincipalPaid] =
        await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction.staticCall(
          mockYieldProviderAddress,
          withdrawRequestAmount,
          targetDeficit,
        );
      await yieldManager.withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction(
        mockYieldProviderAddress,
        withdrawRequestAmount,
        targetDeficit,
      );

      // Assert
      expect(actualWithdrawAmount).eq(withdrawRequestAmount - deficitExcess);
      expect(lstPrincipalPaid).eq(deficitExcess);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
    });
  });

  describe("withdraw from yield provider", () => {
    it("Should revert when the GENERAL pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(mockYieldProviderAddress, 1n),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });

    it("Should revert when the NATIVE_YIELD_UNSTAKING pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_UNSTAKING_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(mockYieldProviderAddress, 1n),
        "IsPaused",
        [NATIVE_YIELD_UNSTAKING_PAUSE_TYPE],
      );
    });

    it("Should revert when unstaking from an unknown YieldProvider", async () => {
      const unknownYieldProvider = ethers.Wallet.createRandom().address;

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(unknownYieldProvider, 1n),
        "UnknownYieldProvider",
      );
    });

    it("Should revert when the caller does not have YIELD_PROVIDER_UNSTAKER_ROLE role", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const unstakerRole = await yieldManager.YIELD_PROVIDER_UNSTAKER_ROLE();

      await expect(
        yieldManager.connect(nonAuthorizedAccount).withdrawFromYieldProvider(mockYieldProviderAddress, 1n),
      ).to.be.revertedWith(buildAccessErrorMessage(nonAuthorizedAccount, unstakerRole));
    });

    it("With 0 targetDeficit and 0 lstLiabilityPrincipal paid, should successfully withdraw the full _amount to the YieldManager", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawAmount);

      const l1MessageService = await mockLineaRollup.getAddress();
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const yieldManagerAddress = await yieldManager.getAddress();

      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(targetReserveAmount)]);
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(withdrawAmount)]);

      expect(await yieldManager.getTargetReserveDeficit()).to.equal(0n);

      await expect(
        yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(mockYieldProviderAddress, withdrawAmount),
      )
        .to.emit(yieldManager, "YieldProviderWithdrawal")
        .withArgs(mockYieldProviderAddress, withdrawAmount, withdrawAmount, 0n, 0n);

      expect(await yieldManager.userFunds(mockYieldProviderAddress)).to.equal(0n);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(0n);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
      expect(await ethers.provider.getBalance(l1MessageService)).to.equal(targetReserveAmount);
    });

    it("With targetDeficit > _amount and 0 lstLiabilityPrincipal paid, should withdraw the full _amount to the reserve", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawAmount);

      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const l1MessageService = await mockLineaRollup.getAddress();
      const yieldManagerAddress = await yieldManager.getAddress();
      const targetDeficit = withdrawAmount * 2n;
      const reserveBalanceBefore = targetReserveAmount - targetDeficit;

      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(reserveBalanceBefore)]);
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(withdrawAmount)]);

      expect(await yieldManager.getTargetReserveDeficit()).to.equal(targetDeficit);
      expect(targetDeficit).to.be.above(withdrawAmount);

      await expect(
        yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(mockYieldProviderAddress, withdrawAmount),
      )
        .to.emit(yieldManager, "YieldProviderWithdrawal")
        .withArgs(mockYieldProviderAddress, withdrawAmount, withdrawAmount, withdrawAmount, 0n);

      expect(await ethers.provider.getBalance(l1MessageService)).to.equal(reserveBalanceBefore + withdrawAmount);
      expect(await yieldManager.userFunds(mockYieldProviderAddress)).to.equal(0n);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(0n);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
    });

    it("With targetDeficit > _amount and non-0 lstLiabilityPrincipal paid, should withdraw the full _amount to the reserve, pause staking and make no liability payment", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawAmount);

      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const l1MessageService = await mockLineaRollup.getAddress();
      const yieldManagerAddress = await yieldManager.getAddress();
      const targetDeficit = withdrawAmount * 2n;
      const reserveBalanceBefore = targetReserveAmount - targetDeficit;

      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, withdrawAmount / 2n);

      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(reserveBalanceBefore)]);
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(withdrawAmount)]);

      expect(await yieldManager.getTargetReserveDeficit()).to.equal(targetDeficit);

      await expect(
        yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(mockYieldProviderAddress, withdrawAmount),
      )
        .to.emit(yieldManager, "YieldProviderWithdrawal")
        .withArgs(mockYieldProviderAddress, withdrawAmount, withdrawAmount, withdrawAmount, 0n);

      expect(await ethers.provider.getBalance(l1MessageService)).to.equal(reserveBalanceBefore + withdrawAmount);
      expect(await yieldManager.userFunds(mockYieldProviderAddress)).to.equal(0n);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(0n);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
    });

    it("With targetDeficit < _amount and 0 lstLiabilityPrincipal paid, should successfully withdraw the full _amount and send target deficit to the reserve", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawAmount);

      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const l1MessageService = await mockLineaRollup.getAddress();
      const yieldManagerAddress = await yieldManager.getAddress();
      const targetDeficit = withdrawAmount / 4n;
      const reserveBalanceBefore = targetReserveAmount - targetDeficit;

      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(reserveBalanceBefore)]);
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(withdrawAmount)]);

      expect(await yieldManager.getTargetReserveDeficit()).to.equal(targetDeficit);

      await expect(
        yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(mockYieldProviderAddress, withdrawAmount),
      )
        .to.emit(yieldManager, "YieldProviderWithdrawal")
        .withArgs(mockYieldProviderAddress, withdrawAmount, withdrawAmount, targetDeficit, 0n);

      expect(await ethers.provider.getBalance(l1MessageService)).to.equal(reserveBalanceBefore + targetDeficit);
      expect(await yieldManager.userFunds(mockYieldProviderAddress)).to.equal(0n);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(0n);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
    });

    it("With targetDeficit < _amount and non-0 lstLiabilityPrincipal paid, should withdraw reduced _amount and send target deficit to the reserve", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawAmount);

      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const l1MessageService = await mockLineaRollup.getAddress();
      const yieldManagerAddress = await yieldManager.getAddress();
      const targetDeficit = withdrawAmount / 4n;
      const lstPrincipalPayment = withdrawAmount / 5n;
      const reserveBalanceBefore = targetReserveAmount - targetDeficit;

      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, lstPrincipalPayment);

      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(reserveBalanceBefore)]);
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(withdrawAmount)]);

      expect(await yieldManager.getTargetReserveDeficit()).to.equal(targetDeficit);

      const expectedWithdrawnAmount = withdrawAmount - lstPrincipalPayment;

      await expect(
        yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(mockYieldProviderAddress, withdrawAmount),
      )
        .to.emit(yieldManager, "YieldProviderWithdrawal")
        .withArgs(
          mockYieldProviderAddress,
          withdrawAmount,
          expectedWithdrawnAmount,
          targetDeficit,
          lstPrincipalPayment,
        );

      expect(await ethers.provider.getBalance(l1MessageService)).to.equal(reserveBalanceBefore + targetDeficit);
      // LST principal payment is not counted as user funds decrement, but as negative yield in the next reportYield call.
      // We tolerate userFunds > withdrawableValue, but not the other way round.
      expect(await yieldManager.userFunds(mockYieldProviderAddress)).to.equal(lstPrincipalPayment);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(lstPrincipalPayment);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
    });
  });

  describe("adding to withdrawal reserve", () => {
    it("Should revert when the GENERAL pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).addToWithdrawalReserve(mockYieldProviderAddress, 1n),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });
    it("Should revert when the NATIVE_YIELD_UNSTAKING pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_UNSTAKING_PAUSE_TYPE);
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).addToWithdrawalReserve(mockYieldProviderAddress, 1n),
        "IsPaused",
        [NATIVE_YIELD_UNSTAKING_PAUSE_TYPE],
      );
    });
    it("Should revert when rebalancing from an unknown YieldProvider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).addToWithdrawalReserve(ethers.Wallet.createRandom().address, 1n),
        "UnknownYieldProvider",
      );
    });
    it("Should revert when the caller does not have YIELD_PROVIDER_UNSTAKER_ROLE role", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await expect(
        yieldManager.connect(nonAuthorizedAccount).addToWithdrawalReserve(mockYieldProviderAddress, 1n),
      ).to.be.revertedWith(
        buildAccessErrorMessage(nonAuthorizedAccount, await yieldManager.YIELD_PROVIDER_UNSTAKER_ROLE()),
      );
    });
    it("With YieldManager balance > _amount, will send _amount from YieldManager to reserve", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const rebalanceAmount = ONE_ETHER;
      const yieldManagerAddress = await yieldManager.getAddress();
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(rebalanceAmount * 2n)]);
      await expect(
        yieldManager.connect(nativeYieldOperator).addToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(mockYieldProviderAddress, rebalanceAmount, rebalanceAmount, rebalanceAmount, 0n, 0n);

      expect(await ethers.provider.getBalance(yieldManagerAddress)).to.equal(rebalanceAmount);
      expect(await ethers.provider.getBalance(await mockLineaRollup.getAddress())).to.equal(rebalanceAmount);
    });
    it("With YieldManager balance < _amount, 0 targetDeficit and 0 lstLiabilityPrincipal, should withdraw from YieldProvider to the reserve", async () => {
      const rebalanceAmount = ONE_ETHER * 2n;
      // Arrange - setup remainder of rebalanceAmount on YieldProvider
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, rebalanceAmount / 2n);
      // Arrange - setup insufficient YieldManager balance
      const yieldManagerAddress = await yieldManager.getAddress();
      await incrementBalance(yieldManagerAddress, rebalanceAmount / 2n);
      // Arrange - setup 0 target deficit
      const l1MessageService = await mockLineaRollup.getAddress();
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(targetReserveAmount)]);
      expect(await yieldManager.getTargetReserveDeficit()).to.equal(0n);

      // Act
      await expect(
        yieldManager.connect(nativeYieldOperator).addToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          rebalanceAmount,
          rebalanceAmount,
          rebalanceAmount / 2n,
          rebalanceAmount / 2n,
          0n,
        );

      expect(await ethers.provider.getBalance(l1MessageService)).to.equal(targetReserveAmount + rebalanceAmount);
      expect(await ethers.provider.getBalance(yieldManager)).to.equal(0);
      expect(await ethers.provider.getBalance(mockWithdrawTarget)).to.equal(0);
    });
    it("With YieldManager balance < _amount, 0 targetDeficit and non-0 lstLiabilityPrincipal paid, should partial withdraw from YieldProvider to the reserve", async () => {
      const rebalanceAmount = ONE_ETHER * 2n;
      // Arrange - setup remainder of rebalanceAmount on YieldProvider
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, rebalanceAmount / 2n);
      // Arrange - setup insufficient YieldManager balance
      const yieldManagerAddress = await yieldManager.getAddress();
      await incrementBalance(yieldManagerAddress, rebalanceAmount / 2n);
      // Arrange - setup 0 target deficit
      const l1MessageService = await mockLineaRollup.getAddress();
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(targetReserveAmount)]);
      expect(await yieldManager.getTargetReserveDeficit()).to.equal(0n);
      // Arrange setup non-0 lstLiabilityPrincipal paid
      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, rebalanceAmount / 4n);

      // Act
      const expectedToReserve = rebalanceAmount - rebalanceAmount / 4n;
      await expect(
        yieldManager.connect(nativeYieldOperator).addToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          rebalanceAmount,
          expectedToReserve,
          rebalanceAmount / 2n,
          rebalanceAmount / 4n,
          rebalanceAmount / 4n,
        );

      expect(await ethers.provider.getBalance(l1MessageService)).to.equal(targetReserveAmount + expectedToReserve);
      expect(await ethers.provider.getBalance(yieldManager)).to.equal(0);
      // Accept imperfection of mock, that mockWithdrawTarget balance is not 0 here/
    });
    it("With YieldManager balance < _amount, targetDeficit > _amount and 0 lstLiabilityPrincipal paid, should withdraw _amount from YieldProvider to the reserve", async () => {
      const rebalanceAmount = ONE_ETHER * 2n;
      // Arrange - setup half of rebalanceAmount on YieldProvider
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, rebalanceAmount / 2n);
      // Arrange - setup other half of rebalanceAmount on YieldManager
      const yieldManagerAddress = await yieldManager.getAddress();
      await incrementBalance(yieldManagerAddress, rebalanceAmount / 2n);
      // Arrange - setup targetDeficit > _amount
      const l1MessageService = await mockLineaRollup.getAddress();
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const targetDeficit = rebalanceAmount * 2n;
      const startingL1MessageServiceBalance = targetReserveAmount - targetDeficit;
      await ethers.provider.send("hardhat_setBalance", [
        l1MessageService,
        ethers.toBeHex(startingL1MessageServiceBalance),
      ]);
      expect(await yieldManager.getTargetReserveDeficit()).to.equal(targetDeficit);
      expect(targetDeficit).to.above(rebalanceAmount);

      // Act
      await expect(
        yieldManager.connect(nativeYieldOperator).addToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          rebalanceAmount,
          rebalanceAmount,
          rebalanceAmount / 2n,
          rebalanceAmount / 2n,
          0n,
        );

      expect(await ethers.provider.getBalance(l1MessageService)).to.equal(
        startingL1MessageServiceBalance + rebalanceAmount,
      );
      expect(await ethers.provider.getBalance(yieldManager)).to.equal(0);
      expect(await ethers.provider.getBalance(mockWithdrawTarget)).to.equal(0);
    });
    it("With YieldManager balance < _amount, targetDeficit > _amount and non-0 lstLiabilityPrincipal paid, should withdraw _amount from YieldProvider to the reserve", async () => {
      const rebalanceAmount = ONE_ETHER * 2n;
      // Arrange - setup half of rebalanceAmount on YieldProvider
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, rebalanceAmount / 2n);
      // Arrange - setup other half of rebalanceAmount on YieldManager
      const yieldManagerAddress = await yieldManager.getAddress();
      await incrementBalance(yieldManagerAddress, rebalanceAmount / 2n);
      // Arrange - setup targetDeficit > _amount
      const l1MessageService = await mockLineaRollup.getAddress();
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      const targetDeficit = rebalanceAmount * 2n;
      const startingL1MessageServiceBalance = targetReserveAmount - targetDeficit;
      await ethers.provider.send("hardhat_setBalance", [
        l1MessageService,
        ethers.toBeHex(startingL1MessageServiceBalance),
      ]);
      expect(await yieldManager.getTargetReserveDeficit()).to.equal(targetDeficit);
      expect(targetDeficit).to.above(rebalanceAmount);
      // Arrange setup non-0 lstLiabilityPrincipal paid
      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, rebalanceAmount / 4n);

      // Act
      await expect(
        yieldManager.connect(nativeYieldOperator).addToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          rebalanceAmount,
          rebalanceAmount,
          rebalanceAmount / 2n,
          rebalanceAmount / 2n,
          0n,
        );

      expect(await ethers.provider.getBalance(l1MessageService)).to.equal(
        startingL1MessageServiceBalance + rebalanceAmount,
      );
      expect(await ethers.provider.getBalance(yieldManager)).to.equal(0);
      expect(await ethers.provider.getBalance(mockWithdrawTarget)).to.equal(0);
    });
    // Ok to skip the following two cases because we won't explore new paths in addToWithdrawalReserve()
    // - With YieldManager balance > _amount, targetDeficit > _amount and 0 lstLiabilityPrincipal paid
    // - With YieldManager balance > _amount, targetDeficit > _amount and non-0 lstLiabilityPrincipal paid
  });

  describe("safely adding to withdrawal reserve", () => {
    it("Should revert when the GENERAL pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(mockYieldProviderAddress, 1n),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });
    it("Should revert when the NATIVE_YIELD_UNSTAKING pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_UNSTAKING_PAUSE_TYPE);
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(mockYieldProviderAddress, 1n),
        "IsPaused",
        [NATIVE_YIELD_UNSTAKING_PAUSE_TYPE],
      );
    });
    it("Should revert when rebalancing from an unknown YieldProvider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(ethers.Wallet.createRandom().address, 1n),
        "UnknownYieldProvider",
      );
    });
    it("Should revert when the caller does not have YIELD_PROVIDER_UNSTAKER_ROLE role", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await expect(
        yieldManager.connect(nonAuthorizedAccount).safeAddToWithdrawalReserve(mockYieldProviderAddress, 1n),
      ).to.be.revertedWith(
        buildAccessErrorMessage(nonAuthorizedAccount, await yieldManager.YIELD_PROVIDER_UNSTAKER_ROLE()),
      );
    });
    it("With YieldManager balance > _amount, maxSafeRebalanceAmount > _amount, will send _amount from YieldManager to reserve", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      // Arrange - Setup YieldManager balance
      const yieldManagerBalance = ONE_ETHER * 4n;
      const yieldManagerAddress = await yieldManager.getAddress();
      await setBalance(yieldManagerAddress, yieldManagerBalance);

      // Act
      const rebalanceAmount = ONE_ETHER * 2n;
      await expect(
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(mockYieldProviderAddress, rebalanceAmount, rebalanceAmount, rebalanceAmount, 0n, 0n);

      expect(await ethers.provider.getBalance(yieldManagerAddress)).to.equal(yieldManagerBalance - rebalanceAmount);
      expect(await ethers.provider.getBalance(await mockLineaRollup.getAddress())).to.equal(rebalanceAmount);
    });
    it("With YieldManager balance < _amount, maxSafeRebalanceAmount < _amount, will send maxSafeRebalanceAmount from YieldManager to reserve", async () => {
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      // Arrange - Setup withdrawable amount
      const withdrawableAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawableAmount);
      await yieldManager.setWithdrawableValueReturnVal(mockYieldProviderAddress, withdrawableAmount);
      // Arrange - Setup YieldManager balance
      const yieldManagerBalance = ONE_ETHER * 2n;
      const yieldManagerAddress = await yieldManager.getAddress();
      await setBalance(yieldManagerAddress, yieldManagerBalance);
      // Arrange - Get before figures
      const l1MessageServiceBalanceBefore = await getBalance(mockLineaRollup);

      // Act
      const rebalanceAmount = ONE_ETHER * 4n;
      // Assert
      const maxSafeRebalanceAmount = yieldManagerBalance + withdrawableAmount;
      await expect(
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          maxSafeRebalanceAmount,
          maxSafeRebalanceAmount,
          yieldManagerBalance,
          withdrawableAmount,
          0n,
        );

      expect(await ethers.provider.getBalance(yieldManagerAddress)).to.equal(0);
      expect(await ethers.provider.getBalance(mockWithdrawTarget)).to.equal(0);
      expect(await ethers.provider.getBalance(await mockLineaRollup.getAddress())).to.equal(
        l1MessageServiceBalanceBefore + maxSafeRebalanceAmount,
      );
    });
    it("With YieldManager balance < _amount, maxSafeRebalanceAmount < _amount, 0 targetDeficit and non-0 lstLiabilityPrincipal paid, should partial withdraw from YieldProvider to the reserve", async () => {
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      // Arrange - Setup withdrawable amount
      const withdrawableAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawableAmount);
      await yieldManager.setWithdrawableValueReturnVal(mockYieldProviderAddress, withdrawableAmount);
      // Arrange - Setup YieldManager balance
      const yieldManagerBalance = ONE_ETHER * 2n;
      const yieldManagerAddress = await yieldManager.getAddress();
      await setBalance(yieldManagerAddress, yieldManagerBalance);
      // Arrange - setup 0 target deficit
      await setWithdrawalReserveToTarget(yieldManager);
      expect(await yieldManager.getTargetReserveDeficit()).to.equal(0n);
      // Arrange setup non-0 lstLiabilityPrincipal paid
      const lstLiabilityPrincipal = ONE_ETHER;
      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, lstLiabilityPrincipal);
      // Arrange - Get before figures
      const l1MessageServiceBalanceBefore = await getBalance(mockLineaRollup);

      // Act
      const rebalanceAmount = ONE_ETHER * 4n;
      // Assert
      const maxSafeRebalanceAmount = yieldManagerBalance + withdrawableAmount;
      await expect(
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          maxSafeRebalanceAmount,
          maxSafeRebalanceAmount - lstLiabilityPrincipal,
          yieldManagerBalance,
          0n,
          lstLiabilityPrincipal,
        );
      expect(await ethers.provider.getBalance(yieldManagerAddress)).to.equal(0);
      expect(await ethers.provider.getBalance(mockWithdrawTarget)).to.equal(withdrawableAmount);
      expect(await ethers.provider.getBalance(await mockLineaRollup.getAddress())).to.equal(
        l1MessageServiceBalanceBefore + maxSafeRebalanceAmount - lstLiabilityPrincipal,
      );
    });
    it("With YieldManager balance < _amount, maxSafeRebalanceAmount < _amount, targetDeficit < yieldManagerBalance, targetDeficit < yieldProviderBalance & and non-0 lstLiabilityPrincipal paid, should partial withdraw maxSafeRebalanceAmount from YieldProvider to the reserve", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      // Arrange - Setup withdrawable amount
      const withdrawableAmount = ONE_ETHER * 2n;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawableAmount);
      await yieldManager.setWithdrawableValueReturnVal(mockYieldProviderAddress, withdrawableAmount);
      // Arrange - Setup YieldManager balance
      const yieldManagerBalance = ONE_ETHER;
      const yieldManagerAddress = await yieldManager.getAddress();
      await setBalance(yieldManagerAddress, yieldManagerBalance);
      // Arrange - targetDeficit < maxSafeRebalanceAmount
      await setWithdrawalReserveToTarget(yieldManager);
      const maxSafeRebalanceAmount = yieldManagerBalance + withdrawableAmount;
      const targetDeficit = maxSafeRebalanceAmount - ONE_ETHER * 2n;
      await decrementBalance(await mockLineaRollup.getAddress(), targetDeficit);
      expect(await yieldManager.getTargetReserveDeficit()).to.eq(targetDeficit);
      // Arrange setup non-0 lstLiabilityPrincipal paid
      const lstLiabilityPrincipal = ONE_ETHER;
      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, lstLiabilityPrincipal);
      // Arrange - Get before figures
      const l1MessageServiceBalanceBefore = await getBalance(mockLineaRollup);

      // Act
      const rebalanceAmount = ONE_ETHER * 4n;
      // Assert
      await expect(
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          maxSafeRebalanceAmount,
          maxSafeRebalanceAmount - lstLiabilityPrincipal,
          yieldManagerBalance,
          withdrawableAmount - lstLiabilityPrincipal,
          lstLiabilityPrincipal,
        );
      expect(await ethers.provider.getBalance(yieldManagerAddress)).to.equal(0);
      expect(await ethers.provider.getBalance(await mockLineaRollup.getAddress())).to.equal(
        l1MessageServiceBalanceBefore + maxSafeRebalanceAmount - lstLiabilityPrincipal,
      );
    });
    it("With YieldManager balance < _amount, maxSafeRebalanceAmount < _amount, targetDeficit >= yieldManagerBalance & and non-0 lstLiabilityPrincipal paid, should withdraw maxSafeRebalanceAmount from YieldProvider to the reserve", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      // Arrange - Setup withdrawable amount
      const withdrawableAmount = ONE_ETHER * 2n;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawableAmount);
      await yieldManager.setWithdrawableValueReturnVal(mockYieldProviderAddress, withdrawableAmount);
      // Arrange - Setup YieldManager balance
      const yieldManagerBalance = ONE_ETHER;
      const yieldManagerAddress = await yieldManager.getAddress();
      await setBalance(yieldManagerAddress, yieldManagerBalance);
      // Arrange - targetDeficit < maxSafeRebalanceAmount
      await setWithdrawalReserveToTarget(yieldManager);
      const maxSafeRebalanceAmount = yieldManagerBalance + withdrawableAmount;
      const targetDeficit = maxSafeRebalanceAmount - ONE_ETHER;
      await decrementBalance(await mockLineaRollup.getAddress(), targetDeficit);
      expect(await yieldManager.getTargetReserveDeficit()).to.eq(targetDeficit);
      // Arrange setup non-0 lstLiabilityPrincipal paid
      const lstLiabilityPrincipal = ONE_ETHER;
      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, lstLiabilityPrincipal);
      // Arrange - Get before figures
      const l1MessageServiceBalanceBefore = await getBalance(mockLineaRollup);

      // Act
      const rebalanceAmount = ONE_ETHER * 4n;
      // Assert
      await expect(
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          maxSafeRebalanceAmount,
          maxSafeRebalanceAmount,
          yieldManagerBalance,
          withdrawableAmount,
          0n,
        );
      expect(await ethers.provider.getBalance(yieldManagerAddress)).to.equal(0);
      expect(await ethers.provider.getBalance(await mockLineaRollup.getAddress())).to.equal(
        l1MessageServiceBalanceBefore + maxSafeRebalanceAmount,
      );
    });

    it("With YieldManager balance < _amount, maxSafeRebalanceAmount < _amount, targetDeficit > _amount and 0 lstLiabilityPrincipal paid, should withdraw _amount from YieldProvider to the reserve", async () => {
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      // Arrange - Setup withdrawable amount
      const withdrawableAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawableAmount);
      await yieldManager.setWithdrawableValueReturnVal(mockYieldProviderAddress, withdrawableAmount);
      // Arrange - Setup YieldManager balance
      const yieldManagerBalance = ONE_ETHER * 2n;
      const yieldManagerAddress = await yieldManager.getAddress();
      await setBalance(yieldManagerAddress, yieldManagerBalance);
      // Arrange - targetDeficit > _amount
      const rebalanceAmount = ONE_ETHER * 4n;
      await setWithdrawalReserveToTarget(yieldManager);
      await decrementBalance(await mockLineaRollup.getAddress(), rebalanceAmount + 1n);
      expect(await yieldManager.getTargetReserveDeficit()).to.be.above(rebalanceAmount);
      // Arrange setup 0 lstLiabilityPrincipal paid
      // Arrange - Get before figures
      const l1MessageServiceBalanceBefore = await getBalance(mockLineaRollup);

      // Act
      // Assert
      const maxSafeRebalanceAmount = yieldManagerBalance + withdrawableAmount;
      await expect(
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          maxSafeRebalanceAmount,
          maxSafeRebalanceAmount,
          yieldManagerBalance,
          withdrawableAmount,
          0n,
        );
      expect(await ethers.provider.getBalance(yieldManagerAddress)).to.equal(0);
      expect(await ethers.provider.getBalance(mockWithdrawTarget)).to.equal(0);
      expect(await ethers.provider.getBalance(await mockLineaRollup.getAddress())).to.equal(
        l1MessageServiceBalanceBefore + maxSafeRebalanceAmount,
      );
    });
    it("With YieldManager balance < _amount, maxSafeRebalanceAmount < _amount, targetDeficit > _amount and non-0 lstLiabilityPrincipal paid, should withdraw _amount from YieldProvider to the reserve", async () => {
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      // Arrange - Setup withdrawable amount
      const withdrawableAmount = ONE_ETHER;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawableAmount);
      await yieldManager.setWithdrawableValueReturnVal(mockYieldProviderAddress, withdrawableAmount);
      // Arrange - Setup YieldManager balance
      const yieldManagerBalance = ONE_ETHER * 2n;
      const yieldManagerAddress = await yieldManager.getAddress();
      await setBalance(yieldManagerAddress, yieldManagerBalance);
      // Arrange - targetDeficit > _amount
      const rebalanceAmount = ONE_ETHER * 4n;
      await setWithdrawalReserveToTarget(yieldManager);
      await decrementBalance(await mockLineaRollup.getAddress(), rebalanceAmount + 1n);
      expect(await yieldManager.getTargetReserveDeficit()).to.be.above(rebalanceAmount);
      // Arrange setup non-0 lstLiabilityPrincipal paid
      const lstLiabilityPrincipal = ONE_ETHER;
      await yieldManager
        .connect(nativeYieldOperator)
        .setPayLSTPrincipalReturnVal(mockYieldProviderAddress, lstLiabilityPrincipal);
      // Arrange - Get before figures
      const l1MessageServiceBalanceBefore = await getBalance(mockLineaRollup);

      // Act
      // Assert
      const maxSafeRebalanceAmount = yieldManagerBalance + withdrawableAmount;
      await expect(
        yieldManager.connect(nativeYieldOperator).safeAddToWithdrawalReserve(mockYieldProviderAddress, rebalanceAmount),
      )
        .to.emit(yieldManager, "WithdrawalReserveAugmented")
        .withArgs(
          mockYieldProviderAddress,
          maxSafeRebalanceAmount,
          maxSafeRebalanceAmount,
          yieldManagerBalance,
          withdrawableAmount,
          0n,
        );
      expect(await ethers.provider.getBalance(yieldManagerAddress)).to.equal(0);
      expect(await ethers.provider.getBalance(mockWithdrawTarget)).to.equal(0);
      expect(await ethers.provider.getBalance(await mockLineaRollup.getAddress())).to.equal(
        l1MessageServiceBalanceBefore + maxSafeRebalanceAmount,
      );
    });
  });

  describe("permissionlessly replenishing the withdrawal reserve", () => {
    it("Should revert when the GENERAL pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).replenishWithdrawalReserve(mockYieldProviderAddress),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });

    it("Should revert when the NATIVE_YIELD_PERMISSIONLESS_ACTIONS pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_PERMISSIONLESS_ACTIONS_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).replenishWithdrawalReserve(mockYieldProviderAddress),
        "IsPaused",
        [NATIVE_YIELD_PERMISSIONLESS_ACTIONS_PAUSE_TYPE],
      );
    });

    it("Should revert when rebalancing from an unknown YieldProvider", async () => {
      const unknownYieldProvider = ethers.Wallet.createRandom().address;

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).replenishWithdrawalReserve(unknownYieldProvider),
        "UnknownYieldProvider",
      );
    });

    it("Should revert when there is no reserve deficit", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const minimumReserve = await yieldManager.getEffectiveMinimumWithdrawalReserve();

      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(minimumReserve)]);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).replenishWithdrawalReserve(mockYieldProviderAddress),
        "WithdrawalReserveNotInDeficit",
      );
    });

    it("Should revert if there is 0 available balance on YieldManager and YieldProvider", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const yieldManagerAddress = await yieldManager.getAddress();

      // Arrange balances
      await setBalance(l1MessageService, 0n);
      await setBalance(yieldManagerAddress, 0n);
      await setBalance(mockYieldProviderAddress, 0n);

      // Act
      const call = yieldManager.connect(nativeYieldOperator).replenishWithdrawalReserve(mockYieldProviderAddress);

      // Assert
      expectRevertWithCustomError(yieldManager, call, "NoAvailableFundsToReplenishWithdrawalReserve");
    });

    it("If YieldManager balance > targetDeficit, settle targetDeficit from YieldManager balance", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const yieldManagerAddress = await yieldManager.getAddress();

      // Arrange - L1MessageService balance = 0
      const beforeL1MessageServiceBalance = 0n;
      await ethers.provider.send("hardhat_setBalance", [
        l1MessageService,
        ethers.toBeHex(beforeL1MessageServiceBalance),
      ]);
      // Arrange - YieldManager balance
      const targetDeficit = await yieldManager.getTargetReserveDeficit();
      const beforeYieldManagerBalance = targetDeficit + ONE_ETHER;
      await ethers.provider.send("hardhat_setBalance", [
        yieldManagerAddress,
        ethers.toBeHex(beforeYieldManagerBalance),
      ]);

      await expect(yieldManager.connect(nativeYieldOperator).replenishWithdrawalReserve(mockYieldProviderAddress))
        .to.emit(yieldManager, "WithdrawalReserveReplenished")
        .withArgs(mockYieldProviderAddress, targetDeficit, targetDeficit, targetDeficit, 0n);
      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
      expect(await ethers.provider.getBalance(l1MessageService)).eq(beforeL1MessageServiceBalance + targetDeficit);
      expect(await ethers.provider.getBalance(yieldManagerAddress)).eq(beforeYieldManagerBalance - targetDeficit);
    });

    it("If YieldManager balance < targetDeficit, settle remainder by withdrawing available value from YieldProvider to reserve", async () => {
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      // Arrange - put half of target reserve amount on YieldProvider
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      await fundYieldProviderForWithdrawal(
        yieldManager,
        mockYieldProvider,
        nativeYieldOperator,
        targetReserveAmount / 2n,
      );
      const mockWithdrawTargetAddress = await mockWithdrawTarget.getAddress();
      const beforeMockWithdrawTargetBalance = await ethers.provider.getBalance(mockWithdrawTargetAddress);
      // Arrange - put other half on YieldManager
      const yieldManagerAddress = await yieldManager.getAddress();
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(targetReserveAmount / 2n)]);
      const beforeYieldManagerBalance = await ethers.provider.getBalance(yieldManagerAddress);
      // Arrange - 0 balance on L1MessageService
      const l1MessageService = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(0)]);
      expect(await yieldManager.getTargetReserveDeficit()).eq(targetReserveAmount);
      const beforeL1MessageServiceBalance = await ethers.provider.getBalance(l1MessageService);

      // Act
      await expect(yieldManager.connect(nativeYieldOperator).replenishWithdrawalReserve(mockYieldProviderAddress))
        .to.emit(yieldManager, "WithdrawalReserveReplenished")
        .withArgs(
          mockYieldProviderAddress,
          targetReserveAmount,
          targetReserveAmount,
          targetReserveAmount / 2n,
          targetReserveAmount / 2n,
        );

      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
      expect(await ethers.provider.getBalance(l1MessageService)).eq(
        beforeL1MessageServiceBalance + targetReserveAmount,
      );
      expect(await ethers.provider.getBalance(yieldManagerAddress)).eq(
        beforeYieldManagerBalance - targetReserveAmount / 2n,
      );
      expect(await ethers.provider.getBalance(mockWithdrawTargetAddress)).eq(
        beforeMockWithdrawTargetBalance - targetReserveAmount / 2n,
      );
    });

    it("If YieldManager + YieldProvider balance > targetDeficit, should only rebalance required to restore targetDeficit", async () => {
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      // Arrange - put target reserve amount on YieldProvider
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, targetReserveAmount);
      const mockWithdrawTargetAddress = await mockWithdrawTarget.getAddress();
      const beforeMockWithdrawTargetBalance = await ethers.provider.getBalance(mockWithdrawTargetAddress);
      // Arrange - put half of target reserve amount on YieldManager
      const yieldManagerAddress = await yieldManager.getAddress();
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(targetReserveAmount / 2n)]);
      const beforeYieldManagerBalance = await ethers.provider.getBalance(yieldManagerAddress);
      // Arrange - 0 balance on L1MessageService
      const l1MessageService = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(0)]);
      expect(await yieldManager.getTargetReserveDeficit()).eq(targetReserveAmount);
      const beforeL1MessageServiceBalance = await ethers.provider.getBalance(l1MessageService);

      // Act
      await expect(yieldManager.connect(nativeYieldOperator).replenishWithdrawalReserve(mockYieldProviderAddress))
        .to.emit(yieldManager, "WithdrawalReserveReplenished")
        .withArgs(
          mockYieldProviderAddress,
          targetReserveAmount,
          targetReserveAmount,
          targetReserveAmount / 2n,
          targetReserveAmount / 2n,
        );

      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
      expect(await ethers.provider.getBalance(l1MessageService)).eq(
        beforeL1MessageServiceBalance + targetReserveAmount,
      );
      expect(await ethers.provider.getBalance(yieldManagerAddress)).eq(
        beforeYieldManagerBalance - targetReserveAmount / 2n,
      );
      expect(await ethers.provider.getBalance(mockWithdrawTargetAddress)).eq(
        beforeMockWithdrawTargetBalance - targetReserveAmount / 2n,
      );
    });

    it("If there is remaining target deficit, should pause staking", async () => {
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      // Arrange - put half target reserve amount on YieldProvider
      const { mockYieldProviderAddress, mockYieldProvider, mockWithdrawTarget } =
        await addMockYieldProvider(yieldManager);
      await fundYieldProviderForWithdrawal(
        yieldManager,
        mockYieldProvider,
        nativeYieldOperator,
        targetReserveAmount / 2n,
      );
      const mockWithdrawTargetAddress = await mockWithdrawTarget.getAddress();
      const beforeMockWithdrawTargetBalance = await ethers.provider.getBalance(mockWithdrawTargetAddress);
      // Arrange - put quarter of target reserve amount on YieldManager
      const yieldManagerAddress = await yieldManager.getAddress();
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(targetReserveAmount / 4n)]);
      const beforeYieldManagerBalance = await ethers.provider.getBalance(yieldManagerAddress);
      // Arrange - 0 balance on L1MessageService
      const l1MessageService = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(0)]);
      expect(await yieldManager.getTargetReserveDeficit()).eq(targetReserveAmount);
      const beforeL1MessageServiceBalance = await ethers.provider.getBalance(l1MessageService);

      // Act
      await expect(yieldManager.connect(nativeYieldOperator).replenishWithdrawalReserve(mockYieldProviderAddress))
        .to.emit(yieldManager, "WithdrawalReserveReplenished")
        .withArgs(
          mockYieldProviderAddress,
          targetReserveAmount,
          (targetReserveAmount * 3n) / 4n,
          targetReserveAmount / 4n,
          targetReserveAmount / 2n,
        );

      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
      expect(await ethers.provider.getBalance(l1MessageService)).eq(
        beforeL1MessageServiceBalance + (targetReserveAmount * 3n) / 4n,
      );
      expect(await ethers.provider.getBalance(yieldManagerAddress)).eq(
        beforeYieldManagerBalance - targetReserveAmount / 4n,
      );
      expect(await ethers.provider.getBalance(mockWithdrawTargetAddress)).eq(
        beforeMockWithdrawTargetBalance - targetReserveAmount / 2n,
      );
    });
  });

  describe("withdraw LST", () => {
    it("Should revert when the GENERAL pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).withdrawLST(mockYieldProviderAddress, 0n, ethers.ZeroAddress),
        "IsPaused",
        [GENERAL_PAUSE_TYPE],
      );
    });

    it("Should revert when the NATIVE_YIELD_PERMISSIONLESS_ACTIONS pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(securityCouncil).pauseByType(NATIVE_YIELD_PERMISSIONLESS_ACTIONS_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).withdrawLST(mockYieldProviderAddress, 0n, ethers.ZeroAddress),
        "IsPaused",
        [NATIVE_YIELD_PERMISSIONLESS_ACTIONS_PAUSE_TYPE],
      );
    });

    it("Should revert when choosing an unknown YieldProvider", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .withdrawLST(ethers.Wallet.createRandom().address, 0n, ethers.ZeroAddress),
        "UnknownYieldProvider",
      );
    });

    it("Should revert if the sender is not the L1MessageService", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).withdrawLST(mockYieldProviderAddress, 0n, ethers.ZeroAddress),
        "SenderNotL1MessageService",
      );
    });

    it("Should revert if L1MessageService does not have withdraw LST flag toggled", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();

      await mockLineaRollup.setWithdrawLSTAllowed(false);
      // Arrange - set gas funds for L1MessageService to be signer
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(ONE_ETHER)]);
      const l1Signer = await ethers.getImpersonatedSigner(l1MessageService);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(l1Signer).withdrawLST(mockYieldProviderAddress, 0n, ethers.ZeroAddress),
        "LSTWithdrawalNotAllowed",
      );

      await mockLineaRollup.setWithdrawLSTAllowed(true);
      await ethers.provider.send("hardhat_stopImpersonatingAccount", [l1MessageService]);
    });

    it("Should revert if LST withdraw amount > userFunds for yield provider", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const withdrawAmount = ONE_ETHER;

      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawAmount / 2n);

      // Arrange - set gas funds for L1MessageService to be signer
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(ONE_ETHER)]);
      const l1Signer = await ethers.getImpersonatedSigner(l1MessageService);
      await mockLineaRollup.setWithdrawLSTAllowed(true);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(l1Signer).withdrawLST(mockYieldProviderAddress, withdrawAmount, ethers.ZeroAddress),
        "LSTWithdrawalExceedsYieldProviderFunds",
      );

      await ethers.provider.send("hardhat_stopImpersonatingAccount", [l1MessageService]);
    });

    it("Should revert if lstPrincipalAmount + LST withdraw amount > userFunds for yield provider", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const fundAmount = ONE_ETHER * 10n;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, fundAmount);
      // Arrange - Setup lstPrincipalAmount
      await yieldManager.setYieldProviderLstLiabilityPrincipal(mockYieldProviderAddress, fundAmount - ONE_ETHER + 1n);
      // Arrange - set gas funds for L1MessageService to be signer
      const withdrawAmount = ONE_ETHER;
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(withdrawAmount)]);
      const l1Signer = await ethers.getImpersonatedSigner(l1MessageService);
      await mockLineaRollup.setWithdrawLSTAllowed(true);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(l1Signer).withdrawLST(mockYieldProviderAddress, withdrawAmount, ethers.ZeroAddress),
        "LSTWithdrawalExceedsYieldProviderFunds",
      );

      await ethers.provider.send("hardhat_stopImpersonatingAccount", [l1MessageService]);
    });

    it("Should revert if lstPrincipalAmount + LST withdraw amount + lastReportedNegativeYield > userFunds for yield provider", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const fundAmount = ONE_ETHER * 10n;
      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, fundAmount);
      // Arrange - Setup lstPrincipalAmount
      const withdrawAmount = ONE_ETHER;
      const negativeYield = 1n;
      await yieldManager.setYieldProviderLstLiabilityPrincipal(
        mockYieldProviderAddress,
        fundAmount - withdrawAmount - negativeYield + 1n,
      );
      await yieldManager.setYieldProviderLastReportedNegativeYield(mockYieldProviderAddress, negativeYield);
      // Arrange - set gas funds for L1MessageService to be signer
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(withdrawAmount)]);
      const l1Signer = await ethers.getImpersonatedSigner(l1MessageService);
      await mockLineaRollup.setWithdrawLSTAllowed(true);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(l1Signer).withdrawLST(mockYieldProviderAddress, withdrawAmount, ethers.ZeroAddress),
        "LSTWithdrawalExceedsYieldProviderFunds",
      );

      await ethers.provider.send("hardhat_stopImpersonatingAccount", [l1MessageService]);
    });

    it("Should successfully withdraw LST, pause staking and emit the expected event", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const withdrawAmount = ONE_ETHER;
      const recipient = ethers.Wallet.createRandom().address;

      await fundYieldProviderForWithdrawal(yieldManager, mockYieldProvider, nativeYieldOperator, withdrawAmount);

      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(ONE_ETHER)]);
      const l1Signer = await ethers.getImpersonatedSigner(l1MessageService);
      await mockLineaRollup.setWithdrawLSTAllowed(true);

      // Act
      await expect(yieldManager.connect(l1Signer).withdrawLST(mockYieldProviderAddress, withdrawAmount, recipient))
        .to.emit(yieldManager, "LSTMinted")
        .withArgs(mockYieldProviderAddress, recipient, withdrawAmount);

      expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;

      await ethers.provider.send("hardhat_stopImpersonatingAccount", [l1MessageService]);
    });
  });
});
