import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
import {
  deployAndAddSingleLidoStVaultYieldProvider,
  fundLidoStVaultYieldProvider,
  getWithdrawLSTCall,
} from "../helpers";
import {
  MockVaultHub,
  MockSTETH,
  MockLineaRollup,
  TestYieldManager,
  MockDashboard,
  MockStakingVault,
  TestLidoStVaultYieldProvider,
} from "contracts/typechain-types";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { ONE_ETHER, REPORT_YIELD_OPERATION_TYPE, ZERO_VALUE } from "../../common/constants";
import { ethers } from "hardhat";

describe("LidoStVaultYieldProvider contract - yield operations", () => {
  let yieldProvider: TestLidoStVaultYieldProvider;
  let nativeYieldOperator: SignerWithAddress;
  let securityCouncil: SignerWithAddress;
  let l2YieldRecipient: SignerWithAddress;
  let mockVaultHub: MockVaultHub;
  let mockSTETH: MockSTETH;
  let mockLineaRollup: MockLineaRollup;
  let yieldManager: TestYieldManager;
  let mockDashboard: MockDashboard;
  let mockStakingVault: MockStakingVault;

  let yieldManagerAddress: string;
  let mockStakingVaultAddress: string;
  let yieldProviderAddress: string;
  let l2YieldRecipientAddress: string;

  before(async () => {
    ({ nativeYieldOperator, securityCouncil, l2YieldRecipient } = await loadFixture(getAccountsFixture));
    l2YieldRecipientAddress = await l2YieldRecipient.getAddress();
  });

  beforeEach(async () => {
    ({
      yieldProvider,
      yieldProviderAddress,
      mockDashboard,
      mockStakingVault,
      yieldManager,
      mockVaultHub,
      mockSTETH,
      mockLineaRollup,
    } = await loadFixture(deployAndAddSingleLidoStVaultYieldProvider));

    yieldManagerAddress = await yieldManager.getAddress();
    mockStakingVaultAddress = await mockStakingVault.getAddress();
  });

  describe("syncExternalLiabilitySettlement", () => {
    it("If ETH value of Lido liabilityShares >= YieldManager liabilityPrincipal, no-op", async () => {
      // Arrange
      const liabilityShares = ONE_ETHER;
      const ethValueOfLidoLiabilityShares = ONE_ETHER * 2n;
      const liabilityPrincipalBefore = ONE_ETHER;
      await mockSTETH.connect(securityCouncil).setPooledEthBySharesRoundUpReturn(ethValueOfLidoLiabilityShares);
      await yieldManager
        .connect(securityCouncil)
        .setYieldProviderLstLiabilityPrincipal(yieldProviderAddress, liabilityPrincipalBefore);

      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProvider);
      await yieldManager.setYieldProviderLstLiabilityPrincipal(yieldProvider, liabilityPrincipalBefore);

      // Act
      const [lstLiabilityPrincipalSynced, isLstLiabilityPrincipalChanged] = await yieldManager
        .connect(securityCouncil)
        .syncExternalLiabilitySettlement.staticCall(yieldProviderAddress, liabilityShares, liabilityPrincipalBefore);
      await yieldManager
        .connect(securityCouncil)
        .syncExternalLiabilitySettlement(yieldProviderAddress, liabilityShares, liabilityPrincipalBefore);

      // Assert
      expect(isLstLiabilityPrincipalChanged).eq(false);
      expect(lstLiabilityPrincipalSynced).eq(liabilityPrincipalBefore);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(userFundsInYieldProvidersTotalBefore);
      expect(await yieldManager.userFunds(yieldProvider)).eq(userFundsBefore);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);
    });
    it("If ETH value of Lido liabilityShares < YieldManager liabilityPrincipal, will decrement liabilityPrincipal to sync", async () => {
      // Arrange liabilityShares and ETH value
      const liabilityShares = ZERO_VALUE;
      const ethValueOfLidoLiabilityShares = ZERO_VALUE;
      await mockSTETH.connect(securityCouncil).setPooledEthBySharesRoundUpReturn(ethValueOfLidoLiabilityShares);
      // Arrange - Set up userFunds and lstLiabilityPrincipals
      const liabilityPrincipalBefore = ONE_ETHER * 2n;
      const call = getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        liabilityPrincipalBefore,
      );
      await call;

      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProvider);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);

      // Act
      const [lstLiabilityPrincipalSynced, isLstLiabilityPrincipalChanged] = await yieldManager
        .connect(securityCouncil)
        .syncExternalLiabilitySettlement.staticCall(yieldProviderAddress, liabilityShares, liabilityPrincipalBefore);
      await yieldManager
        .connect(securityCouncil)
        .syncExternalLiabilitySettlement(yieldProviderAddress, liabilityShares, liabilityPrincipalBefore);

      // Assert
      expect(isLstLiabilityPrincipalChanged).eq(true);
      expect(lstLiabilityPrincipalSynced).eq(ethValueOfLidoLiabilityShares);
      const lstLiabilityPrincipalDecrement = liabilityPrincipalBefore - BigInt(ethValueOfLidoLiabilityShares);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(
        userFundsInYieldProvidersTotalBefore - lstLiabilityPrincipalDecrement,
      );
      expect(await yieldManager.userFunds(yieldProvider)).eq(userFundsBefore - lstLiabilityPrincipalDecrement);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(
        liabilityPrincipalBefore - lstLiabilityPrincipalDecrement,
      );
    });
  });

  describe("payMaximumPossibleLSTLiability", () => {
    it("If ossified, should be a no-op", async () => {
      // Arrange - setup lst liability
      const liabilityPrincipalBefore = ONE_ETHER;
      await getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        liabilityPrincipalBefore,
      );
      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProvider);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);
      // Arrange - setup ossified. Note with real Lido contracts 'processPendingOssification' will not succeed with an LST liability
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProvider);
      await yieldManager.connect(securityCouncil).processPendingOssification(yieldProvider);
      // Act
      const liabilityPaidETH = await yieldManager
        .connect(securityCouncil)
        .payMaximumPossibleLSTLiability.staticCall(yieldProvider);
      await yieldManager.connect(securityCouncil).payMaximumPossibleLSTLiability(yieldProvider);
      // Assert
      expect(liabilityPaidETH).eq(0);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(userFundsInYieldProvidersTotalBefore);
      expect(await yieldManager.userFunds(yieldProvider)).eq(userFundsBefore);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);
    });
    it("If no Lido liabilities, no-op", async () => {
      // Arrange - setup lst liability principal
      const liabilityPrincipalBefore = ONE_ETHER;
      await getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        liabilityPrincipalBefore,
      );
      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProvider);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);
      // Arrange - setup 0 Lido liability
      const liabilityShares = ZERO_VALUE;
      await mockDashboard.connect(securityCouncil).setLiabilitySharesReturn(liabilityShares);
      // Act
      const liabilityPaidETH = await yieldManager
        .connect(securityCouncil)
        .payMaximumPossibleLSTLiability.staticCall(yieldProvider);
      await yieldManager.connect(securityCouncil).payMaximumPossibleLSTLiability(yieldProvider);
      // Assert
      expect(liabilityPaidETH).eq(0);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(userFundsInYieldProvidersTotalBefore);
      expect(await yieldManager.userFunds(yieldProvider)).eq(userFundsBefore);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);
    });
    it("If no Vault balance, no-op", async () => {
      // Arrange - setup lst liability principal
      const liabilityPrincipalBefore = ONE_ETHER;
      await getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        liabilityPrincipalBefore,
      );
      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProvider);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);
      // Arrange - setup Lido liability
      const liabilityShares = ONE_ETHER;
      await mockDashboard.connect(securityCouncil).setLiabilitySharesReturn(liabilityShares);
      // Arrange - setup Vault balance (counted in shares)
      await mockSTETH.connect(securityCouncil).setSharesByPooledEthReturn(ZERO_VALUE);
      // Act
      const liabilityPaidETH = await yieldManager
        .connect(securityCouncil)
        .payMaximumPossibleLSTLiability.staticCall(yieldProvider);
      await yieldManager.connect(securityCouncil).payMaximumPossibleLSTLiability(yieldProvider);
      // Assert
      expect(liabilityPaidETH).eq(0);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(userFundsInYieldProvidersTotalBefore);
      expect(await yieldManager.userFunds(yieldProvider)).eq(userFundsBefore);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);
    });
    it("If VAULT_BALANCE >0 and LIDO_LIABILITY_SHARE >0, rebalance with lower of the two (VAULT_BALANCE lower)", async () => {
      // Arrange - setup lst liability principal
      const liabilityPrincipalBefore = ONE_ETHER;
      await getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        liabilityPrincipalBefore,
      );
      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProvider);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);
      // Arrange - setup Lido liability
      const liabilityShares = ONE_ETHER * 2n;
      await mockDashboard.connect(securityCouncil).setLiabilitySharesReturn(liabilityShares);
      // Arrange - setup Vault balance (counted in shares)
      await mockDashboard.setRebalanceVaultWithSharesWithdrawingFromVault(true);
      await mockSTETH.connect(securityCouncil).setSharesByPooledEthReturn(ONE_ETHER);
      // Arrange - setup post-rebalance Lido LST liability
      const ethValueOfLidoLiabilitySharesAfterRebalance = ONE_ETHER;
      await mockSTETH
        .connect(securityCouncil)
        .setPooledEthBySharesRoundUpReturn(ethValueOfLidoLiabilitySharesAfterRebalance);

      // Act
      const liabilityPaidETH = await yieldManager
        .connect(securityCouncil)
        .payMaximumPossibleLSTLiability.staticCall(yieldProvider);
      await yieldManager.connect(securityCouncil).payMaximumPossibleLSTLiability(yieldProvider);

      // Assert
      expect(liabilityPaidETH).eq(ONE_ETHER);
      const syncExternalLiabilitySettlementDifference =
        liabilityPrincipalBefore - ethValueOfLidoLiabilitySharesAfterRebalance;
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(
        userFundsInYieldProvidersTotalBefore - syncExternalLiabilitySettlementDifference,
      );
      expect(await yieldManager.userFunds(yieldProvider)).eq(
        userFundsBefore - syncExternalLiabilitySettlementDifference,
      );
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(
        liabilityPrincipalBefore - syncExternalLiabilitySettlementDifference,
      );
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(vaultBalanceBefore - liabilityPaidETH);
    });
    it("If VAULT_BALANCE >0 and LIDO_LIABILITY_SHARE >0, rebalance with lower of the two (LIDO_LIABILITY_SHARE lower)", async () => {
      // Arrange - setup lst liability principal
      const liabilityPrincipalBefore = (ONE_ETHER * 3n) / 2n;
      await getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        liabilityPrincipalBefore,
      );
      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProvider);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);
      // Arrange - setup Lido liability
      const liabilityShares = ONE_ETHER;
      await mockDashboard.connect(securityCouncil).setLiabilitySharesReturn(liabilityShares);
      // Arrange - setup Vault balance (counted in shares)
      await mockDashboard.setRebalanceVaultWithSharesWithdrawingFromVault(true);
      await mockSTETH.connect(securityCouncil).setSharesByPooledEthReturn(ONE_ETHER * 2n);
      // Arrange - setup post-rebalance Lido LST liability
      const ethValueOfLidoLiabilitySharesAfterRebalance = ONE_ETHER;
      await mockSTETH
        .connect(securityCouncil)
        .setPooledEthBySharesRoundUpReturn(ethValueOfLidoLiabilitySharesAfterRebalance);

      // Act
      const liabilityPaidETH = await yieldManager
        .connect(securityCouncil)
        .payMaximumPossibleLSTLiability.staticCall(yieldProvider);
      await yieldManager.connect(securityCouncil).payMaximumPossibleLSTLiability(yieldProvider);

      // Assert
      expect(liabilityPaidETH).eq(liabilityShares);
      const syncExternalLiabilitySettlementDifference =
        liabilityPrincipalBefore - ethValueOfLidoLiabilitySharesAfterRebalance;
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(
        userFundsInYieldProvidersTotalBefore - syncExternalLiabilitySettlementDifference,
      );
      expect(await yieldManager.userFunds(yieldProvider)).eq(
        userFundsBefore - syncExternalLiabilitySettlementDifference,
      );
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(
        liabilityPrincipalBefore - syncExternalLiabilitySettlementDifference,
      );
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(vaultBalanceBefore - liabilityPaidETH);
    });
  });

  describe("payLSTPrincipal", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      await expect(yieldProvider.payLSTPrincipal(yieldProviderAddress, ONE_ETHER)).to.be.revertedWithCustomError(
        yieldProvider,
        "ContextIsNotYieldManager",
      );
    });

    it("Should return 0 if ossification pending", async () => {
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      const lstPrincipalPaid = await yieldManager
        .connect(securityCouncil)
        .payLSTPrincipalExternal.staticCall(yieldProviderAddress, ONE_ETHER);
      await yieldManager.connect(securityCouncil).payLSTPrincipalExternal(yieldProviderAddress, ONE_ETHER);

      expect(lstPrincipalPaid).eq(0);
    });

    it("Should return 0 if ossified", async () => {
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await yieldManager.connect(securityCouncil).processPendingOssification(yieldProviderAddress);
      const lstPrincipalPaid = await yieldManager
        .connect(securityCouncil)
        .payLSTPrincipalExternal.staticCall(yieldProviderAddress, ONE_ETHER);
      await yieldManager.connect(securityCouncil).payLSTPrincipalExternal(yieldProviderAddress, ONE_ETHER);

      expect(lstPrincipalPaid).eq(0);
    });

    it("If no lst liability principal, be no-op", async () => {
      // Arrange
      const lstLiabilityBefore = 0n;
      await yieldManager.connect(securityCouncil).setYieldProviderLstLiabilityPrincipal(yieldProviderAddress, 0);

      // Act
      const lstLiabilityPaid = await yieldManager
        .connect(securityCouncil)
        .payLSTPrincipalExternal.staticCall(yieldProviderAddress, ONE_ETHER);
      await yieldManager.connect(securityCouncil).payLSTPrincipalExternal(yieldProviderAddress, ONE_ETHER);

      // Arrange
      expect(lstLiabilityPaid).eq(0);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldManagerAddress)).eq(lstLiabilityBefore);
    });

    it("If no available funds, be no-op", async () => {
      // Arrange
      const lstLiabilityBefore = ONE_ETHER;
      await yieldManager
        .connect(securityCouncil)
        .setYieldProviderLstLiabilityPrincipal(yieldProviderAddress, lstLiabilityBefore);
      await mockSTETH.connect(securityCouncil).setPooledEthBySharesRoundUpReturn(lstLiabilityBefore);

      // Act
      const amountAvailable = ZERO_VALUE;
      const lstLiabilityPaid = await yieldManager
        .connect(securityCouncil)
        .payLSTPrincipalExternal.staticCall(yieldProviderAddress, amountAvailable);
      await yieldManager.connect(securityCouncil).payLSTPrincipalExternal(yieldProviderAddress, amountAvailable);

      // Arrange
      expect(lstLiabilityPaid).eq(0);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress)).eq(lstLiabilityBefore);
    });

    it("If LIABILITY_PRINCIPAL >0 and AVAILABLE_FUNDS >0, rebalance with lower of the two (LIABILITY_PRINCIPAL lower)", async () => {
      // Arrange
      const lstLiabilityBefore = ONE_ETHER;
      await yieldManager
        .connect(securityCouncil)
        .setYieldProviderLstLiabilityPrincipal(yieldProviderAddress, lstLiabilityBefore);
      await mockSTETH.connect(securityCouncil).setPooledEthBySharesRoundUpReturn(lstLiabilityBefore);

      // Act
      const amountAvailable = ONE_ETHER * 2n;
      const lstLiabilityPaid = await yieldManager
        .connect(securityCouncil)
        .payLSTPrincipalExternal.staticCall(yieldProviderAddress, amountAvailable);
      await yieldManager.connect(securityCouncil).payLSTPrincipalExternal(yieldProviderAddress, amountAvailable);

      // Arrange
      expect(lstLiabilityPaid).eq(ONE_ETHER);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress)).eq(
        lstLiabilityBefore - ONE_ETHER,
      );
    });

    it("If LIABILITY_PRINCIPAL >0 and AVAILABLE_FUNDS >0, rebalance with lower of the two (AVAILABLE_FUNDS lower)", async () => {
      // Arrange
      const lstLiabilityBefore = ONE_ETHER * 2n;
      await yieldManager
        .connect(securityCouncil)
        .setYieldProviderLstLiabilityPrincipal(yieldProviderAddress, lstLiabilityBefore);
      await mockSTETH.connect(securityCouncil).setPooledEthBySharesRoundUpReturn(lstLiabilityBefore);

      // Act
      const amountAvailable = ONE_ETHER;
      const lstLiabilityPaid = await yieldManager
        .connect(securityCouncil)
        .payLSTPrincipalExternal.staticCall(yieldProviderAddress, amountAvailable);
      await yieldManager.connect(securityCouncil).payLSTPrincipalExternal(yieldProviderAddress, amountAvailable);

      // Arrange
      expect(lstLiabilityPaid).eq(ONE_ETHER);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress)).eq(
        lstLiabilityBefore - ONE_ETHER,
      );
    });

    it("If external liability settlement occurred, should succeed", async () => {
      // Arrange - Set up userFunds + lstLiabilityPrincipal
      const liabilityPrincipalBefore = ONE_ETHER * 2n;
      await getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        liabilityPrincipalBefore,
      );
      // Arrange - set up liabilityShares < liabilityPrincipal
      const ethValueOfLidoLiabilityShare = (ONE_ETHER * 3n) / 2n;
      await mockSTETH.connect(securityCouncil).setPooledEthBySharesRoundUpReturn(ethValueOfLidoLiabilityShare);
      // Arrange - Get before figures
      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProvider);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(liabilityPrincipalBefore);

      // Act
      const amountAvailable = ONE_ETHER;
      const lstLiabilityPaid = await yieldManager
        .connect(securityCouncil)
        .payLSTPrincipalExternal.staticCall(yieldProviderAddress, amountAvailable);
      await yieldManager.connect(securityCouncil).payLSTPrincipalExternal(yieldProviderAddress, amountAvailable);

      // Arrange
      expect(lstLiabilityPaid).eq(ONE_ETHER);
      const expectedExternalLiabilitySettlement = liabilityPrincipalBefore - ethValueOfLidoLiabilityShare;
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(
        userFundsInYieldProvidersTotalBefore - expectedExternalLiabilitySettlement,
      );
      expect(await yieldManager.userFunds(yieldProvider)).eq(userFundsBefore - expectedExternalLiabilitySettlement);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(
        liabilityPrincipalBefore - expectedExternalLiabilitySettlement - lstLiabilityPaid,
      );
    });
  });

  describe("payObligations", () => {
    it("If 0 available yield, no-op", async () => {
      const amountAvailable = ZERO_VALUE;
      const obligationsPaid = await yieldManager
        .connect(securityCouncil)
        .payObligations.staticCall(yieldProviderAddress, amountAvailable);
      await yieldManager.connect(securityCouncil).payObligations(yieldProviderAddress, amountAvailable);
      expect(obligationsPaid).eq(0);
    });
    it("If obligationsPaid <= availableYield, succeed", async () => {
      // Arrange - Set up Vault balance
      const vaultBalance = ONE_ETHER * 2n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, vaultBalance);
      // Arrange - Obligations paid
      const expectedObligationsPaid = ONE_ETHER;
      await mockVaultHub.setIsSettleVaultObligationsWithdrawingFromVault(true);
      await mockVaultHub.setSettleVaultObligationAmount(expectedObligationsPaid);
      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);

      // Act
      const availableYield = ONE_ETHER * 2n;
      const obligationsPaid = await yieldManager
        .connect(securityCouncil)
        .payObligations.staticCall(yieldProviderAddress, availableYield);

      await yieldManager.connect(securityCouncil).payObligations(yieldProviderAddress, availableYield);
      // Assert
      expect(expectedObligationsPaid).eq(obligationsPaid);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(vaultBalanceBefore - obligationsPaid);
    });
    it("If obligationsPaid > availableYield, succeed", async () => {
      // Arrange - Set up Vault balance
      const vaultBalance = ONE_ETHER * 2n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, vaultBalance);
      // Arrange - Obligations paid
      const expectedObligationsPaid = ONE_ETHER * 2n;
      await mockVaultHub.setIsSettleVaultObligationsWithdrawingFromVault(true);
      await mockVaultHub.setSettleVaultObligationAmount(expectedObligationsPaid);
      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);

      // Act
      const availableYield = ONE_ETHER;
      const obligationsPaid = await yieldManager
        .connect(securityCouncil)
        .payObligations.staticCall(yieldProviderAddress, availableYield);

      await yieldManager.connect(securityCouncil).payObligations(yieldProviderAddress, availableYield);
      // Assert
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(vaultBalanceBefore - obligationsPaid);
      expect(expectedObligationsPaid).eq(obligationsPaid);
    });
  });

  describe("payNodeOperatorFees", () => {
    it("If 0 available yield, no-op", async () => {
      const amountAvailable = ZERO_VALUE;
      const nodeOperatorFeesPaid = await yieldManager
        .connect(securityCouncil)
        .payNodeOperatorFees.staticCall(yieldProviderAddress, amountAvailable);
      await yieldManager.connect(securityCouncil).payNodeOperatorFees(yieldProviderAddress, amountAvailable);
      expect(nodeOperatorFeesPaid).eq(0);
    });
    it("If vault balance < current fees, no-op", async () => {
      // Arrange - Set up Vault balance
      const vaultBalance = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, vaultBalance);
      // Arrange - Set up current fees
      const operatorFees = ONE_ETHER * 2n;
      await mockDashboard.setNodeOperatorDisbursableFeeReturn(operatorFees);
      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);

      // Act
      const amountAvailable = ONE_ETHER;
      const nodeOperatorFeesPaid = await yieldManager
        .connect(securityCouncil)
        .payNodeOperatorFees.staticCall(yieldProviderAddress, amountAvailable);
      await yieldManager.connect(securityCouncil).payNodeOperatorFees(yieldProviderAddress, amountAvailable);

      // Assert
      expect(nodeOperatorFeesPaid).eq(0);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(vaultBalanceBefore);
    });
    it("If vault balance > current fees, and feesPaid <= availableYield, succeed", async () => {
      // Arrange - Set up Vault balance
      const vaultBalance = ONE_ETHER * 2n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, vaultBalance);
      // Arrange - Set up current fees
      const operatorFees = ONE_ETHER;
      await mockDashboard.setNodeOperatorDisbursableFeeReturn(operatorFees);
      await mockDashboard.setIsDisburseNodeOperatorFeeWithdrawingFromVault(true);
      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);

      // Act
      const amountAvailable = (ONE_ETHER * 3n) / 2n;
      const nodeOperatorFeesPaid = await yieldManager
        .connect(securityCouncil)
        .payNodeOperatorFees.staticCall(yieldProviderAddress, amountAvailable);
      await yieldManager.connect(securityCouncil).payNodeOperatorFees(yieldProviderAddress, amountAvailable);

      // Assert
      expect(nodeOperatorFeesPaid).eq(operatorFees);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(vaultBalanceBefore - operatorFees);
    });
    it("If vault balance > current fees, and feesPaid > availableYield, succeed", async () => {
      // Arrange - Set up Vault balance
      const vaultBalance = ONE_ETHER * 2n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, vaultBalance);
      // Arrange - Set up current fees
      const operatorFees = ONE_ETHER;
      await mockDashboard.setNodeOperatorDisbursableFeeReturn(operatorFees);
      await mockDashboard.setIsDisburseNodeOperatorFeeWithdrawingFromVault(true);
      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);

      // Act
      const amountAvailable = 1n;
      const nodeOperatorFeesPaid = await yieldManager
        .connect(securityCouncil)
        .payNodeOperatorFees.staticCall(yieldProviderAddress, amountAvailable);
      await yieldManager.connect(securityCouncil).payNodeOperatorFees(yieldProviderAddress, amountAvailable);

      // Assert
      expect(nodeOperatorFeesPaid).eq(operatorFees);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(vaultBalanceBefore - operatorFees);
    });
  });

  describe("handlePositiveYieldAccounting", () => {
    it("If 0 LST liability, obligations or node operator fees, should report full _availableAmount", async () => {
      // Act
      const availableYield = ONE_ETHER;
      const newReportedYield = await yieldManager
        .connect(securityCouncil)
        .handlePostiveYieldAccounting.staticCall(yieldProviderAddress, availableYield);
      await yieldManager.connect(securityCouncil).handlePostiveYieldAccounting(yieldProviderAddress, availableYield);

      // Assert
      expect(newReportedYield).eq(ONE_ETHER);
    });
    it("If LST liability payment > availableYield, should succeed", async () => {
      // Arrange - setup lst liability principal
      const liabilityPrincipalBefore = ONE_ETHER;
      await getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        liabilityPrincipalBefore,
      );
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);
      // Arrange - setup Lido liability
      const liabilityShares = ONE_ETHER * 2n;
      await mockDashboard.connect(securityCouncil).setLiabilitySharesReturn(liabilityShares);
      // Arrange - setup Vault balance (counted in shares)
      await mockDashboard.setRebalanceVaultWithSharesWithdrawingFromVault(true);
      await mockSTETH.connect(securityCouncil).setSharesByPooledEthReturn(ONE_ETHER);
      // Arrange - setup post-rebalance Lido LST liability
      const ethValueOfLidoLiabilitySharesAfterRebalance = ONE_ETHER;
      await mockSTETH
        .connect(securityCouncil)
        .setPooledEthBySharesRoundUpReturn(ethValueOfLidoLiabilitySharesAfterRebalance);

      // Act
      const availableYield = ONE_ETHER / 2n;
      const newReportedYield = await yieldManager
        .connect(securityCouncil)
        .handlePostiveYieldAccounting.staticCall(yieldProviderAddress, availableYield);
      await yieldManager.connect(securityCouncil).handlePostiveYieldAccounting(yieldProviderAddress, availableYield);

      // Assert
      expect(newReportedYield).eq(0);
      const expectedLiabilityPaidEth = ONE_ETHER;
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(
        vaultBalanceBefore - expectedLiabilityPaidEth,
      );
    });
    it("Should decrement reported yield by liabilities paid", async () => {
      // Arrange - setup lst liability principal
      const liabilityPrincipalBefore = ONE_ETHER;
      await getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        liabilityPrincipalBefore,
      );
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);
      // Arrange - setup Lido liability
      const liabilityShares = ONE_ETHER * 2n;
      await mockDashboard.connect(securityCouncil).setLiabilitySharesReturn(liabilityShares);
      // Arrange - setup Vault balance (counted in shares)
      await mockDashboard.setRebalanceVaultWithSharesWithdrawingFromVault(true);
      await mockSTETH.connect(securityCouncil).setSharesByPooledEthReturn(ONE_ETHER);
      // Arrange - setup post-rebalance Lido LST liability
      const ethValueOfLidoLiabilitySharesAfterRebalance = ONE_ETHER;
      await mockSTETH
        .connect(securityCouncil)
        .setPooledEthBySharesRoundUpReturn(ethValueOfLidoLiabilitySharesAfterRebalance);

      // Act
      const availableYield = ONE_ETHER * 2n;
      const newReportedYield = await yieldManager
        .connect(securityCouncil)
        .handlePostiveYieldAccounting.staticCall(yieldProviderAddress, availableYield);
      await yieldManager.connect(securityCouncil).handlePostiveYieldAccounting(yieldProviderAddress, availableYield);

      // Assert
      expect(newReportedYield).eq(availableYield - ethValueOfLidoLiabilitySharesAfterRebalance);
      const expectedLiabilityPaidEth = ONE_ETHER;
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(
        vaultBalanceBefore - expectedLiabilityPaidEth,
      );
    });
    it("Should decrement reported yield by obligations paid", async () => {
      // Arrange - Set up Vault balance
      const vaultBalance = ONE_ETHER * 2n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, vaultBalance);
      // Arrange - Obligations paid
      const expectedObligationsPaid = ONE_ETHER;
      await mockVaultHub.setIsSettleVaultObligationsWithdrawingFromVault(true);
      await mockVaultHub.setSettleVaultObligationAmount(expectedObligationsPaid);
      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);

      // Act
      const availableYield = ONE_ETHER * 2n;
      const newReportedYield = await yieldManager
        .connect(securityCouncil)
        .handlePostiveYieldAccounting.staticCall(yieldProviderAddress, availableYield);
      await yieldManager.connect(securityCouncil).handlePostiveYieldAccounting(yieldProviderAddress, availableYield);

      // Assert
      expect(newReportedYield).eq(availableYield - expectedObligationsPaid);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(
        vaultBalanceBefore - expectedObligationsPaid,
      );
    });
    it("Should succeed when obligations paid > available yield", async () => {
      // Arrange - Set up Vault balance
      const vaultBalance = ONE_ETHER * 4n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, vaultBalance);
      // Arrange - Obligations paid
      const expectedObligationsPaid = ONE_ETHER * 3n;
      await mockVaultHub.setIsSettleVaultObligationsWithdrawingFromVault(true);
      await mockVaultHub.setSettleVaultObligationAmount(expectedObligationsPaid);
      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);

      // Act
      const availableYield = ONE_ETHER * 2n;
      const newReportedYield = await yieldManager
        .connect(securityCouncil)
        .handlePostiveYieldAccounting.staticCall(yieldProviderAddress, availableYield);
      await yieldManager.connect(securityCouncil).handlePostiveYieldAccounting(yieldProviderAddress, availableYield);

      // Assert
      expect(newReportedYield).eq(0);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(
        vaultBalanceBefore - expectedObligationsPaid,
      );
    });
    it("Will decrement reported yield by node operator fees paid", async () => {
      // Arrange - Set up Vault balance
      const vaultBalance = ONE_ETHER * 2n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, vaultBalance);
      // Arrange - Set up current fees
      const operatorFees = ONE_ETHER;
      await mockDashboard.setNodeOperatorDisbursableFeeReturn(operatorFees);
      await mockDashboard.setIsDisburseNodeOperatorFeeWithdrawingFromVault(true);
      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);

      // Act
      const availableYield = ONE_ETHER * 2n;
      const newReportedYield = await yieldManager
        .connect(securityCouncil)
        .handlePostiveYieldAccounting.staticCall(yieldProviderAddress, availableYield);
      await yieldManager.connect(securityCouncil).handlePostiveYieldAccounting(yieldProviderAddress, availableYield);

      // Assert
      expect(newReportedYield).eq(availableYield - operatorFees);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(vaultBalanceBefore - operatorFees);
    });
    it("Should succeed when node operator fees paid > available yield", async () => {
      // Arrange - Set up Vault balance
      const vaultBalance = ONE_ETHER * 5n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, vaultBalance);
      // Arrange - Set up current fees
      const operatorFees = ONE_ETHER * 4n;
      await mockDashboard.setNodeOperatorDisbursableFeeReturn(operatorFees);
      await mockDashboard.setIsDisburseNodeOperatorFeeWithdrawingFromVault(true);
      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);

      // Act
      const availableYield = ONE_ETHER * 2n;
      const newReportedYield = await yieldManager
        .connect(securityCouncil)
        .handlePostiveYieldAccounting.staticCall(yieldProviderAddress, availableYield);
      await yieldManager.connect(securityCouncil).handlePostiveYieldAccounting(yieldProviderAddress, availableYield);

      // Assert
      expect(newReportedYield).eq(0);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(vaultBalanceBefore - operatorFees);
    });
    it("Will succeed with LST liability payment, obligation payment and node operator fee payment", async () => {
      // Arrange - Setup node operator fees = 2 ETH
      const operatorFees = ONE_ETHER * 2n;
      await mockDashboard.setNodeOperatorDisbursableFeeReturn(operatorFees);
      await mockDashboard.setIsDisburseNodeOperatorFeeWithdrawingFromVault(true);
      // Arrange - Setup obligations paid  = 1 ETH
      const expectedObligationsPaid = ONE_ETHER;
      await mockVaultHub.setIsSettleVaultObligationsWithdrawingFromVault(true);
      await mockVaultHub.setSettleVaultObligationAmount(expectedObligationsPaid);
      // Arrange - Setup LST liability payment  = 1 ETH
      const expectedLiabilityPayment = ONE_ETHER;
      await mockDashboard.connect(securityCouncil).setLiabilitySharesReturn(expectedLiabilityPayment);
      await mockDashboard.setRebalanceVaultWithSharesWithdrawingFromVault(true);
      await mockSTETH.connect(securityCouncil).setSharesByPooledEthReturn(expectedLiabilityPayment);
      // Arrange - Setup initial Vault balance + userFunds + liabilityPrincipal = 10 ETH
      const initialVaultBalance = ONE_ETHER * 10n;
      await getWithdrawLSTCall(mockLineaRollup, yieldManager, yieldProvider, nativeYieldOperator, initialVaultBalance);
      // Arrange - Setup external liability settlement of 4 ETH
      const ethValueOfLidoLiabilitySharesAfterRebalance = ONE_ETHER * 6n;
      await mockSTETH
        .connect(securityCouncil)
        .setPooledEthBySharesRoundUpReturn(ethValueOfLidoLiabilitySharesAfterRebalance);

      // Arrange - Get before figures
      const vaultBalanceBefore = await ethers.provider.getBalance(mockStakingVaultAddress);
      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProvider);
      const lstLiabilityPrincipalBefore =
        await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress);

      // Arrange - Provide 3.5 ETH positive yield
      const availableYield = (ONE_ETHER * 7n) / 2n;

      // Act
      const newReportedYield = await yieldManager
        .connect(securityCouncil)
        .handlePostiveYieldAccounting.staticCall(yieldProviderAddress, availableYield);
      await yieldManager.connect(securityCouncil).handlePostiveYieldAccounting(yieldProviderAddress, availableYield);

      // Assert
      expect(newReportedYield).eq(0);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(
        vaultBalanceBefore - operatorFees - expectedObligationsPaid - expectedLiabilityPayment,
      );
      const expectedExternalLiabilitySettlement = initialVaultBalance - ethValueOfLidoLiabilitySharesAfterRebalance;
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(
        userFundsInYieldProvidersTotalBefore - expectedExternalLiabilitySettlement,
      );
      expect(await yieldManager.userFunds(yieldProvider)).eq(userFundsBefore - expectedExternalLiabilitySettlement);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProvider)).eq(
        lstLiabilityPrincipalBefore - expectedExternalLiabilitySettlement,
      );
    });
  });

  describe("reportYield", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).reportYield(yieldProviderAddress);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should revert if ossified", async () => {
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await yieldManager.connect(securityCouncil).processPendingOssification(yieldProviderAddress);
      const call = yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipientAddress);
      await expectRevertWithCustomError(yieldProvider, call, "OperationNotSupportedDuringOssification", [
        REPORT_YIELD_OPERATION_TYPE,
      ]);
    });
    it("If vault value > user funds, should report positive yield", async () => {
      // Arrange
      const userFunds = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, userFunds);
      const vaultValue = ONE_ETHER * 2n;
      await mockDashboard.setTotalValueReturn(vaultValue);

      // Act
      const newReportedYield = await yieldManager
        .connect(nativeYieldOperator)
        .reportYield.staticCall(yieldProviderAddress, l2YieldRecipientAddress);
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipientAddress);

      // Assert
      expect(newReportedYield).eq(vaultValue - userFunds);
    });
    it("If vault value <= user funds, should report 0 yield", async () => {
      // Arrange
      const userFunds = ONE_ETHER * 2n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, userFunds);
      const vaultValue = ONE_ETHER;
      await mockDashboard.setTotalValueReturn(vaultValue);

      // Act
      const newReportedYield = await yieldManager
        .connect(nativeYieldOperator)
        .reportYield.staticCall(yieldProviderAddress, l2YieldRecipientAddress);
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipientAddress);

      // Assert
      expect(newReportedYield).eq(0);
    });
  });
});
