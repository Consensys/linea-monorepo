import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { getAccountsFixture } from "../../common/helpers";
import { deployAndAddSingleLidoStVaultYieldProvider, getWithdrawLSTCall } from "../helpers";
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
import { ONE_ETHER, ZERO_VALUE } from "../../common/constants";

describe("LidoStVaultYieldProvider contract - yield operations", () => {
  let yieldProvider: TestLidoStVaultYieldProvider;
  let nativeYieldOperator: SignerWithAddress;
  let securityCouncil: SignerWithAddress;
  let mockVaultHub: MockVaultHub;
  let mockSTETH: MockSTETH;
  let mockLineaRollup: MockLineaRollup;
  let yieldManager: TestYieldManager;
  let mockDashboard: MockDashboard;
  let mockStakingVault: MockStakingVault;

  let l1MessageServiceAddress: string;
  let yieldManagerAddress: string;
  let vaultHubAddress: string;
  let stethAddress: string;
  let mockDashboardAddress: string;
  let mockStakingVaultAddress: string;
  let yieldProviderAddress: string;
  before(async () => {
    ({ nativeYieldOperator, securityCouncil } = await loadFixture(getAccountsFixture));
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

    l1MessageServiceAddress = await mockLineaRollup.getAddress();
    yieldManagerAddress = await yieldManager.getAddress();
    vaultHubAddress = await mockVaultHub.getAddress();
    stethAddress = await mockSTETH.getAddress();
    mockDashboardAddress = await mockDashboard.getAddress();
    mockStakingVaultAddress = await mockStakingVault.getAddress();

    console.log(l1MessageServiceAddress);
    console.log(yieldManagerAddress);
    console.log(vaultHubAddress);
    console.log(stethAddress);
    console.log(mockDashboardAddress);
    console.log(mockStakingVaultAddress);
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
    });
  });
});
