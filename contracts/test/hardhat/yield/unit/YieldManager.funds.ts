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
  NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE,
  GENERAL_PAUSE_TYPE,
  NATIVE_YIELD_STAKING_PAUSE_TYPE,
  NATIVE_YIELD_REPORTING_PAUSE_TYPE,
  NATIVE_YIELD_UNSTAKING_PAUSE_TYPE,
  NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE,
} from "../../common/constants";
import { buildAccessErrorMessage, expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
import { setupSuccessfulYieldProviderWithdrawal } from "../helpers";

describe("Linea Rollup contract", () => {
  let yieldManager: TestYieldManager;

  let securityCouncil: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  let nativeYieldOperator: SignerWithAddress;
  let operationalSafe: SignerWithAddress;
  let l2YieldRecipient: SignerWithAddress;
  let mockLineaRollup: MockLineaRollup;

  const mockWithdrawalParams = ethers.hexlify(ethers.randomBytes(8));
  const mockWithdrawalParamsProof = ethers.hexlify(ethers.randomBytes(8));

  before(async () => {
    ({
      securityCouncil,
      operator: nonAuthorizedAccount,
      nativeYieldOperator,
      operationalSafe,
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
    it("Should revert when the caller is not the YIELD_PROVIDER_FUNDER_ROLE", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nonAuthorizedAccount).transferFundsToReserve(1n),
        "CallerMissingRole",
        [await yieldManager.RESERVE_OPERATOR_ROLE(), await yieldManager.YIELD_PROVIDER_FUNDER_ROLE()],
      );
    });

    it("Should revert when the caller is not the YIELD_PROVIDER_UNSTAKER_ROLE", async () => {
      const unstakerRole = await yieldManager.YIELD_PROVIDER_UNSTAKER_ROLE();
      await yieldManager.connect(securityCouncil).grantRole(unstakerRole, nonAuthorizedAccount.address);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nonAuthorizedAccount).transferFundsToReserve(1n),
        "CallerMissingRole",
        [await yieldManager.RESERVE_OPERATOR_ROLE(), await yieldManager.YIELD_PROVIDER_FUNDER_ROLE()],
      );
    });

    it("Should revert when the NATIVE_YIELD_RESERVE_FUNDING pause type is activated", async () => {
      await yieldManager.connect(operationalSafe).pauseByType(NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nativeYieldOperator).transferFundsToReserve(1n),
        "IsPaused",
        [NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE],
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
    it("Should revert when the caller is not the YIELD_PROVIDER_FUNDER_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expect(
        yieldManager.connect(nonAuthorizedAccount).fundYieldProvider(mockYieldProviderAddress, 1n),
      ).to.be.revertedWith(
        buildAccessErrorMessage(nonAuthorizedAccount, await yieldManager.YIELD_PROVIDER_FUNDER_ROLE()),
      );
    });

    it("Should revert when the NATIVE_YIELD_STAKING pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await yieldManager.connect(operationalSafe).pauseByType(NATIVE_YIELD_STAKING_PAUSE_TYPE);

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
      const lstPrincipalPayment = 10n;
      await yieldManager.setPayLSTPrincipalReturnVal(mockYieldProviderAddress, lstPrincipalPayment);

      await expect(
        yieldManager.connect(nativeYieldOperator).fundYieldProvider(mockYieldProviderAddress, transferAmount),
      )
        .to.emit(yieldManager, "YieldProviderFunded")
        .withArgs(mockYieldProviderAddress, transferAmount, lstPrincipalPayment, transferAmount - lstPrincipalPayment);

      const yieldProviderData = await yieldManager.getYieldProviderData(mockYieldProviderAddress);
      expect(yieldProviderData.userFunds).to.equal(transferAmount - lstPrincipalPayment);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(transferAmount - lstPrincipalPayment);
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

      await yieldManager.connect(operationalSafe).pauseByType(NATIVE_YIELD_REPORTING_PAUSE_TYPE);

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

    it("Should successfully report non-0 yield, update state and emit the expected event", async () => {
      // ARRANGE
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const reportedYield = ONE_ETHER;
      await yieldManager.connect(nativeYieldOperator).setReportYieldReturnVal(mockYieldProviderAddress, reportedYield);

      // ACT + ASSERT
      await expect(
        yieldManager.connect(nativeYieldOperator).reportYield(mockYieldProviderAddress, l2YieldRecipient.address),
      )
        .to.emit(yieldManager, "NativeYieldReported")
        .withArgs(mockYieldProviderAddress, l2YieldRecipient.address, reportedYield);

      const providerData = await yieldManager.getYieldProviderData(mockYieldProviderAddress);
      expect(providerData.userFunds).to.equal(reportedYield);
      expect(providerData.yieldReportedCumulative).to.equal(reportedYield);
      expect(await yieldManager.userFunds(mockYieldProviderAddress)).to.equal(reportedYield);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(reportedYield);
    });
  });

  describe("permissioned unstake", () => {
    it("Should revert when the caller is not the RESERVE_OPERATOR_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nonAuthorizedAccount).unstake(mockYieldProviderAddress, mockWithdrawalParams),
        "CallerMissingRole",
        [await yieldManager.RESERVE_OPERATOR_ROLE(), await yieldManager.YIELD_PROVIDER_UNSTAKER_ROLE()],
      );
    });

    it("Should revert when the caller is not the YIELD_PROVIDER_UNSTAKER_ROLE", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(nonAuthorizedAccount).unstake(mockYieldProviderAddress, mockWithdrawalParams),
        "CallerMissingRole",
        [await yieldManager.RESERVE_OPERATOR_ROLE(), await yieldManager.YIELD_PROVIDER_UNSTAKER_ROLE()],
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
      await yieldManager.connect(operationalSafe).pauseByType(NATIVE_YIELD_UNSTAKING_PAUSE_TYPE);

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

    it("Should revert when the NATIVE_YIELD_PERMISSIONLESS_UNSTAKING pause type is activated", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      await yieldManager.connect(operationalSafe).pauseByType(NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE);

      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .unstakePermissionless(mockYieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof),
        "IsPaused",
        [NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_PAUSE_TYPE],
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
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const targetReserveAmount = await yieldManager.getTargetWithdrawalReserveAmount();
      await yieldManager
        .connect(nativeYieldOperator)
        .setWithdrawableValueReturnVal(mockYieldProviderAddress, targetReserveAmount);
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
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const call = yieldManager
        .connect(nativeYieldOperator)
        .delegatecallWithdrawFromYieldProvider(mockYieldProviderAddress, 1n);

      await expect(call).to.be.revertedWithPanic(0x11);
    });
    it("Should revert if receiveCaller not configured correctly", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const call = yieldManager
        .connect(nativeYieldOperator)
        .delegatecallWithdrawFromYieldProvider(mockYieldProviderAddress, 0n);
      await expectRevertWithCustomError(yieldManager, call, "UnexpectedReceiveCaller");
    });

    it("Delegatecalls successfully and makes correct state transitions", async () => {
      const { mockYieldProviderAddress, mockYieldProvider } = await addMockYieldProvider(yieldManager);
      const withdrawAmount = ONE_ETHER;
      await setupSuccessfulYieldProviderWithdrawal(
        yieldManager,
        mockYieldProvider,
        nativeYieldOperator,
        withdrawAmount,
      );

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

  // describe("withdraw with target deficit priority and lst liability principal reduction", () => {
  //   const amount = ONE_ETHER;

  //   const setupProviderBalances = async (provider: string, userFunds: bigint) => {
  //     await yieldManager.setYieldProviderUserFunds(provider, userFunds);
  //     await yieldManager.setUserFundsInYieldProvidersTotal(userFunds);
  //   };

  //   it("With 0 targetDeficit and 0 lstLiabilityPrincipal, should successfully withdraw the full _amount", async () => {
  //     const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
  //     await setupProviderBalances(mockYieldProviderAddress, amount);
  //     await yieldManager.setPayLSTPrincipalReturnVal(mockYieldProviderAddress, 0n);

  //     const callData = yieldManager.interface.encodeFunctionData(
  //       "withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction",
  //       [mockYieldProviderAddress, amount, 0n],
  //     );
  //     const staticResult = await ethers.provider.call({
  //       to: await yieldManager.getAddress(),
  //       data: callData,
  //       from: nativeYieldOperator.address,
  //     });
  //     const [withdrawAmount, lstPrincipalPaid] = yieldManager.interface.decodeFunctionResult(
  //       "withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction",
  //       staticResult,
  //     );

  //     expect(withdrawAmount).to.equal(amount);
  //     expect(lstPrincipalPaid).to.equal(0n);

  //     await nativeYieldOperator.sendTransaction({
  //       to: await yieldManager.getAddress(),
  //       data: callData,
  //     });

  //     expect(await yieldManager.getYieldProviderUserFunds(mockYieldProviderAddress)).to.equal(0n);
  //     expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(0n);
  //     expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
  //   });

  //   it("With targetDeficit > _amount and 0 lstLiabilityPrincipal, should successfully withdraw the full _amount and pause staking", async () => {
  //     const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
  //     await setupProviderBalances(mockYieldProviderAddress, amount);
  //     await yieldManager.setPayLSTPrincipalReturnVal(mockYieldProviderAddress, 0n);

  //     const targetDeficit = amount + 1n;

  //     const callData = yieldManager.interface.encodeFunctionData(
  //       "withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction",
  //       [mockYieldProviderAddress, amount, targetDeficit],
  //     );
  //     const staticResult = await ethers.provider.call({
  //       to: await yieldManager.getAddress(),
  //       data: callData,
  //       from: nativeYieldOperator.address,
  //     });
  //     const [withdrawAmount, lstPrincipalPaid] = yieldManager.interface.decodeFunctionResult(
  //       "withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction",
  //       staticResult,
  //     );

  //     expect(withdrawAmount).to.equal(amount);
  //     expect(lstPrincipalPaid).to.equal(0n);

  //     await nativeYieldOperator.sendTransaction({
  //       to: await yieldManager.getAddress(),
  //       data: callData,
  //     });

  //     expect(await yieldManager.getYieldProviderUserFunds(mockYieldProviderAddress)).to.equal(0n);
  //     expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(0n);
  //     expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
  //   });

  //   it("With targetDeficit > _amount and non-0 lstLiabilityPrincipal, should pay LSTPrincipal, withdraw reduced _amount and pause staking", async () => {
  //     const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
  //     await setupProviderBalances(mockYieldProviderAddress, amount);
  //     await yieldManager.setPayLSTPrincipalReturnVal(mockYieldProviderAddress, 0n);
  //     await yieldManager.setYieldProviderLstLiabilityPrincipal(mockYieldProviderAddress, 5n);

  //     const targetDeficit = amount + 1n;

  //     const callData = yieldManager.interface.encodeFunctionData(
  //       "withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction",
  //       [mockYieldProviderAddress, amount, targetDeficit],
  //     );
  //     const staticResult = await ethers.provider.call({
  //       to: await yieldManager.getAddress(),
  //       data: callData,
  //       from: nativeYieldOperator.address,
  //     });
  //     const [withdrawAmount, lstPrincipalPaid] = yieldManager.interface.decodeFunctionResult(
  //       "withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction",
  //       staticResult,
  //     );

  //     expect(withdrawAmount).to.equal(amount);
  //     expect(lstPrincipalPaid).to.equal(0n);

  //     await nativeYieldOperator.sendTransaction({
  //       to: await yieldManager.getAddress(),
  //       data: callData,
  //     });

  //     expect(await yieldManager.getYieldProviderUserFunds(mockYieldProviderAddress)).to.equal(0n);
  //     expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(0n);
  //     expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.true;
  //   });

  //   it("With targetDeficit < _amount and 0 lstLiabilityPrincipal, should successfully withdraw the full _amount", async () => {
  //     const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
  //     await setupProviderBalances(mockYieldProviderAddress, amount);
  //     await yieldManager.setPayLSTPrincipalReturnVal(mockYieldProviderAddress, 0n);

  //     const targetDeficit = amount / 4n;

  //     const callData = yieldManager.interface.encodeFunctionData(
  //       "withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction",
  //       [mockYieldProviderAddress, amount, targetDeficit],
  //     );
  //     const staticResult = await ethers.provider.call({
  //       to: await yieldManager.getAddress(),
  //       data: callData,
  //       from: nativeYieldOperator.address,
  //     });
  //     const [withdrawAmount, lstPrincipalPaid] = yieldManager.interface.decodeFunctionResult(
  //       "withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction",
  //       staticResult,
  //     );

  //     expect(withdrawAmount).to.equal(amount);
  //     expect(lstPrincipalPaid).to.equal(0n);

  //     await nativeYieldOperator.sendTransaction({
  //       to: await yieldManager.getAddress(),
  //       data: callData,
  //     });

  //     expect(await yieldManager.getYieldProviderUserFunds(mockYieldProviderAddress)).to.equal(0n);
  //     expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(0n);
  //     expect(await yieldManager.isStakingPaused(mockYieldProviderAddress)).to.be.false;
  //   });

  //   it("With targetDeficit < _amount and non-0 lstLiabilityPrincipal, should pay LSTPrincipal and withdraw reduced _amount", async () => {
  //     const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
  //     const lstPrincipalPayment = amount / 4n;
  //     await setupProviderBalances(mockYieldProviderAddress, amount + lstPrincipalPayment);
  //     await yieldManager.setPayLSTPrincipalReturnVal(mockYieldProviderAddress, lstPrincipalPayment);

  //     const targetDeficit = amount / 2n;

  //     const callData = yieldManager.interface.encodeFunctionData(
  //       "withdrawWithTargetDeficitPriorityAndLSTLiabilityPrincipalReduction",
  //       [mockYieldProviderAddress, amount, targetDeficit],
  //     );

  //     await expect(
  //       nativeYieldOperator.sendTransaction({
  //         to: await yieldManager.getAddress(),
  //         data: callData,
  //       }),
  //     ).to.be.revertedWithPanic(0x11);
  //   });
  // });
});
