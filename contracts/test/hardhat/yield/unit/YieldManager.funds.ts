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
  NATIVE_YIELD_RESERVE_FUNDING_PAUSE_TYPE,
  GENERAL_PAUSE_TYPE,
  NATIVE_YIELD_STAKING_PAUSE_TYPE,
  ONE_ETHER,
} from "../../common/constants";
import { buildAccessErrorMessage, expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";

describe("Linea Rollup contract", () => {
  let yieldManager: TestYieldManager;

  let securityCouncil: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  let nativeYieldOperator: SignerWithAddress;
  let operationalSafe: SignerWithAddress;
  let mockLineaRollup: MockLineaRollup;

  before(async () => {
    ({
      securityCouncil,
      operator: nonAuthorizedAccount,
      nativeYieldOperator,
      operationalSafe,
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

    it("Should successfully send ETH to the YieldProvider, update state and emit the expected event", async () => {
      const { mockYieldProviderAddress } = await addMockYieldProvider(yieldManager);
      const l1MessageService = await mockLineaRollup.getAddress();
      const yieldManagerAddress = await yieldManager.getAddress();

      // Arrange funds so that not in withdrawal reserve deficit
      const minimumReserveAmount = await yieldManager.getMinimumWithdrawalReserveAmount();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(minimumReserveAmount)]);
      const transferAmount = ONE_ETHER;
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(transferAmount)]);

      // Act
      await expect(
        yieldManager.connect(nativeYieldOperator).fundYieldProvider(mockYieldProviderAddress, transferAmount),
      )
        .to.emit(yieldManager, "YieldProviderFunded")
        .withArgs(mockYieldProviderAddress, transferAmount, 0n, transferAmount);

      // Assert
      const yieldProviderData = await yieldManager.getYieldProviderData(mockYieldProviderAddress);
      expect(yieldProviderData.userFunds).to.equal(transferAmount);

      expect(await yieldManager.userFundsInYieldProvidersTotal()).to.equal(transferAmount);
    });
  });
});
