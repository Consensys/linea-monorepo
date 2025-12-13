import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
import {
  buildVendorExitData,
  buildVendorInitializationData,
  deployAndAddSingleLidoStVaultYieldProvider,
  fundLidoStVaultYieldProvider,
  getBalance,
  getWithdrawLSTCall,
  incrementBalance,
  ossifyYieldProvider,
  setupLSTPrincipalDecrementForPaxMaximumPossibleLSTLiability,
  setBalance,
  setWithdrawalReserveToMinimum,
} from "../helpers";
import {
  MockVaultHub,
  MockVaultFactory,
  MockSTETH,
  MockLineaRollup,
  TestYieldManager,
  MockDashboard,
  MockStakingVault,
  TestLidoStVaultYieldProvider,
  ValidatorContainerProofVerifier,
  SSZMerkleTree,
  TestValidatorContainerProofVerifier,
} from "contracts/typechain-types";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { parseUnits, ZeroAddress } from "ethers";
import {
  ONE_ETHER,
  ZERO_VALUE,
  EMPTY_CALLDATA,
  VALIDATOR_WITNESS_TYPE,
  THIRTY_TWO_ETH_IN_GWEI,
  ONE_GWEI,
  MAX_0X2_VALIDATOR_EFFECTIVE_BALANCE_GWEI,
  ProgressOssificationResult,
  YieldProviderVendor,
  OperationType,
  BEACON_PROOF_WITNESS_TYPE,
} from "../../common/constants";
import { generateLidoUnstakePermissionlessWitness } from "../helpers/proof";

describe("LidoStVaultYieldProvider contract - basic operations", () => {
  let yieldProvider: TestLidoStVaultYieldProvider;
  let nativeYieldOperator: SignerWithAddress;
  let securityCouncil: SignerWithAddress;
  let mockVaultHub: MockVaultHub;
  let mockVaultFactory: MockVaultFactory;
  let mockSTETH: MockSTETH;
  let mockLineaRollup: MockLineaRollup;
  let yieldManager: TestYieldManager;
  let mockDashboard: MockDashboard;
  let mockStakingVault: MockStakingVault;
  let sszMerkleTree: SSZMerkleTree;
  let verifier: ValidatorContainerProofVerifier;
  let testVerifier: TestValidatorContainerProofVerifier;

  let l1MessageServiceAddress: string;
  let yieldManagerAddress: string;
  let vaultHubAddress: string;
  let vaultFactoryAddress: string;
  let stethAddress: string;
  let mockDashboardAddress: string;
  let mockStakingVaultAddress: string;
  let yieldProviderAddress: string;
  let verifierAddress: string;

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
      mockVaultFactory,
      mockSTETH,
      mockLineaRollup,
      sszMerkleTree,
      verifier,
      testVerifier,
    } = await loadFixture(deployAndAddSingleLidoStVaultYieldProvider));

    l1MessageServiceAddress = await mockLineaRollup.getAddress();
    yieldManagerAddress = await yieldManager.getAddress();
    vaultHubAddress = await mockVaultHub.getAddress();
    vaultFactoryAddress = await mockVaultFactory.getAddress();
    stethAddress = await mockSTETH.getAddress();
    mockDashboardAddress = await mockDashboard.getAddress();
    mockStakingVaultAddress = await mockStakingVault.getAddress();
    verifierAddress = await verifier.getAddress();
  });

  describe("Constructor", () => {
    it("Should revert if 0 address provided for _l1MessageService", async () => {
      const contractFactory = await ethers.getContractFactory("TestLidoStVaultYieldProvider");
      const call = contractFactory.deploy(
        ZeroAddress,
        yieldManagerAddress,
        vaultHubAddress,
        vaultFactoryAddress,
        stethAddress,
        verifierAddress,
      );
      await expectRevertWithCustomError(contractFactory, call, "ZeroAddressNotAllowed");
    });
    it("Should revert if 0 address provided for _yieldManager", async () => {
      const contractFactory = await ethers.getContractFactory("TestLidoStVaultYieldProvider");
      const call = contractFactory
        .connect(nativeYieldOperator)
        .deploy(
          l1MessageServiceAddress,
          ZeroAddress,
          vaultHubAddress,
          vaultFactoryAddress,
          stethAddress,
          verifierAddress,
        );
      await expectRevertWithCustomError(contractFactory, call, "ZeroAddressNotAllowed");
    });
    it("Should revert if 0 address provided for _vaultHub", async () => {
      const contractFactory = await ethers.getContractFactory("TestLidoStVaultYieldProvider");
      const call = contractFactory.deploy(
        l1MessageServiceAddress,
        yieldManagerAddress,
        ZeroAddress,
        vaultFactoryAddress,
        stethAddress,
        verifierAddress,
      );
      await expectRevertWithCustomError(contractFactory, call, "ZeroAddressNotAllowed");
    });
    it("Should revert if 0 address provided for _vaultFactory", async () => {
      const contractFactory = await ethers.getContractFactory("TestLidoStVaultYieldProvider");
      const call = contractFactory.deploy(
        l1MessageServiceAddress,
        yieldManagerAddress,
        vaultHubAddress,
        ZeroAddress,
        stethAddress,
        verifierAddress,
      );
      await expectRevertWithCustomError(contractFactory, call, "ZeroAddressNotAllowed");
    });

    it("Should revert if 0 address provided for _steth", async () => {
      const contractFactory = await ethers.getContractFactory("TestLidoStVaultYieldProvider");
      const call = contractFactory.deploy(
        l1MessageServiceAddress,
        yieldManagerAddress,
        vaultHubAddress,
        vaultFactoryAddress,
        ZeroAddress,
        verifierAddress,
      );
      await expectRevertWithCustomError(contractFactory, call, "ZeroAddressNotAllowed");
    });

    it("Should revert if 0 address provided for _validatorContainerProofVerifier", async () => {
      const contractFactory = await ethers.getContractFactory("TestLidoStVaultYieldProvider");
      const call = contractFactory.deploy(
        l1MessageServiceAddress,
        yieldManagerAddress,
        vaultHubAddress,
        vaultFactoryAddress,
        stethAddress,
        ZeroAddress,
      );
      await expectRevertWithCustomError(contractFactory, call, "ZeroAddressNotAllowed");
    });

    it("Should succeed and emit the correct deployment event", async () => {
      const contractFactory = await ethers.getContractFactory("TestLidoStVaultYieldProvider");
      const contract = await contractFactory.deploy(
        l1MessageServiceAddress,
        yieldManagerAddress,
        vaultHubAddress,
        vaultFactoryAddress,
        stethAddress,
        verifierAddress,
      );
      expect(contract.deploymentTransaction)
        .to.emit(contract, "LidoStVaultYieldProviderDeployed")
        .withArgs(
          l1MessageServiceAddress,
          yieldManagerAddress,
          vaultHubAddress,
          vaultFactoryAddress,
          stethAddress,
          verifierAddress,
        );
    });
  });

  describe("Immutables", () => {
    it("Should deploy with correct VaultHub address", async () => {
      expect(await yieldProvider.VAULT_HUB()).eq(await mockVaultHub.getAddress());
    });
    it("Should deploy with correct STETH address", async () => {
      expect(await yieldProvider.STETH()).eq(await mockSTETH.getAddress());
    });
    it("Should deploy with correct L1MessageService address", async () => {
      expect(await yieldProvider.L1_MESSAGE_SERVICE()).eq(await mockLineaRollup.getAddress());
    });
    it("Should deploy with correct YieldManager address", async () => {
      expect(await yieldProvider.YIELD_MANAGER()).eq(await yieldManager.getAddress());
    });
    it("Should deploy with correct ValidatorContainerProofVerifier address", async () => {
      expect(await yieldProvider.VALIDATOR_CONTAINER_PROOF_VERIFIER()).eq(await verifier.getAddress());
    });
  });

  describe("withdrawableValue", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.withdrawableValue(yieldProviderAddress);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("If not ossified, should return withdrawable value from Dashboard", async () => {
      // Arrange
      const expectedWithdrawableValue = 99n;
      await mockDashboard.setWithdrawableValueReturn(expectedWithdrawableValue);
      await yieldManager.setYieldProviderUserFunds(yieldProviderAddress, expectedWithdrawableValue + 10n);
      // Act
      const withdrawableValue = await yieldManager.withdrawableValue.staticCall(yieldProviderAddress);
      // Assert
      expect(withdrawableValue).eq(expectedWithdrawableValue);
    });
    it("If ossified, should return staking vault balance", async () => {
      // Arrange - Ossify
      await ossifyYieldProvider(yieldManager, yieldProviderAddress, securityCouncil);
      // Arrange - Set Staking Vault Balance
      const expectedWithdrawableValue = 99n;
      await incrementBalance(mockStakingVaultAddress, expectedWithdrawableValue);
      await yieldManager.setYieldProviderUserFunds(yieldProviderAddress, expectedWithdrawableValue + 10n);
      // Act
      const withdrawableValue = await yieldManager.withdrawableValue.staticCall(yieldProviderAddress);
      // Assert
      expect(withdrawableValue).eq(expectedWithdrawableValue);
    });
  });

  describe("getEntrypointContract", () => {
    it("If not ossified, should return the Dashboard address", async () => {
      const entryPoint = await yieldManager.getEntrypointContract.staticCall(yieldProviderAddress);
      expect(entryPoint).eq(mockDashboardAddress);
    });
    it("If ossified, should return the Vault address", async () => {
      await ossifyYieldProvider(yieldManager, yieldProviderAddress, securityCouncil);
      const entryPoint = await yieldManager.getEntrypointContract.staticCall(yieldProviderAddress);
      expect(entryPoint).eq(mockStakingVaultAddress);
    });
  });

  describe("getDashboard", () => {
    it("should return the Dashboard address", async () => {
      const contract = await yieldManager.getEntrypointContract.staticCall(yieldProviderAddress);
      expect(contract).eq(mockDashboardAddress);
    });
  });

  describe("getVault", () => {
    it("should return the Dashboard address", async () => {
      const contract = await yieldManager.getVault.staticCall(yieldProviderAddress);
      expect(contract).eq(mockStakingVaultAddress);
    });
  });

  describe("fund YieldProvider", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).fundYieldProvider(yieldProviderAddress, ZERO_VALUE);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should revert if ossification initiatied", async () => {
      await yieldManager.setYieldProviderIsOssificationInitiated(yieldProviderAddress, true);
      const fundAmount = ONE_ETHER;
      const call = fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, fundAmount);
      await expectRevertWithCustomError(yieldProvider, call, "OperationNotSupportedDuringOssification", [
        OperationType.FUND_YIELD_PROVIDER,
      ]);
    });

    it("Should revert if staking paused", async () => {
      await yieldManager.connect(securityCouncil).pauseStaking(yieldProviderAddress);
      const fundAmount = ONE_ETHER;
      const call = fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, fundAmount);
      await expectRevertWithCustomError(yieldProvider, call, "OperationNotSupportedDuringStakingPause", [
        OperationType.FUND_YIELD_PROVIDER,
      ]);
    });

    it("Should revert if initiated ossification", async () => {
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await yieldManager.connect(securityCouncil).setYieldProviderIsStakingPaused(yieldProviderAddress, false);
      const fundAmount = ONE_ETHER;
      const call = fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, fundAmount);
      await expectRevertWithCustomError(yieldProvider, call, "OperationNotSupportedDuringOssification", [
        OperationType.FUND_YIELD_PROVIDER,
      ]);
    });

    it("Should revert if ossified", async () => {
      await ossifyYieldProvider(yieldManager, yieldProviderAddress, securityCouncil);
      await yieldManager.connect(securityCouncil).setYieldProviderIsStakingPaused(yieldProviderAddress, false);
      const fundAmount = ONE_ETHER;
      const call = fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, fundAmount);
      await expectRevertWithCustomError(yieldProvider, call, "OperationNotSupportedDuringOssification", [
        OperationType.FUND_YIELD_PROVIDER,
      ]);
    });
    it("If not ossified, should fund the Dashboard and pay max LST liability", async () => {
      const beforeVaultBalance = await ethers.provider.getBalance(mockStakingVaultAddress);
      const fundAmount = ONE_ETHER;

      // Setup LST liability principal < fundAmount
      const lstLiabilityPrincipal = ONE_ETHER / 2n;
      await yieldManager.setYieldProviderLstLiabilityPrincipal(yieldProviderAddress, lstLiabilityPrincipal);
      await setupLSTPrincipalDecrementForPaxMaximumPossibleLSTLiability(
        lstLiabilityPrincipal,
        yieldManager,
        yieldProviderAddress,
        mockSTETH,
        mockDashboard,
      );

      // Do funding
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, fundAmount);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(beforeVaultBalance + fundAmount);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress)).eq(0);
    });
  });

  describe("withdraw from YieldProvider", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).withdrawFromYieldProvider(yieldProviderAddress, ZERO_VALUE);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should successfully withdraw when not ossifed", async () => {
      const withdrawAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, withdrawAmount);
      await yieldManager
        .connect(nativeYieldOperator)
        .safeWithdrawFromYieldProvider(yieldProviderAddress, withdrawAmount);
    });
    it("Should successfully withdraw when ossified", async () => {
      const withdrawAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, withdrawAmount);
      await ossifyYieldProvider(yieldManager, yieldProviderAddress, securityCouncil);
      await yieldManager
        .connect(nativeYieldOperator)
        .safeWithdrawFromYieldProvider(yieldProviderAddress, withdrawAmount);
    });
    it("If VaultHub is connected, should perform max possible LST liability", async () => {
      // Setup LST liability principal < fundAmount
      const lstLiabilityPrincipal = ONE_ETHER / 2n;
      await yieldManager.setYieldProviderLstLiabilityPrincipal(yieldProviderAddress, lstLiabilityPrincipal);
      await setupLSTPrincipalDecrementForPaxMaximumPossibleLSTLiability(
        lstLiabilityPrincipal,
        yieldManager,
        yieldProviderAddress,
        mockSTETH,
        mockDashboard,
      );
      // Setup VaultHub connected
      await mockVaultHub.setIsVaultConnectedReturn(true);

      // Act
      const withdrawAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, withdrawAmount);
      await yieldManager
        .connect(nativeYieldOperator)
        .safeWithdrawFromYieldProvider(yieldProviderAddress, withdrawAmount);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress)).eq(0);
    });
  });

  describe("pause staking", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).pauseStaking(yieldProviderAddress);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should successfully pause when not ossifed", async () => {
      await yieldManager.connect(securityCouncil).pauseStaking(yieldProviderAddress);
    });
    it("Should successfully pause when ossifed", async () => {
      await yieldManager.connect(securityCouncil).setYieldProviderIsOssified(yieldProviderAddress, true);
      await yieldManager.connect(securityCouncil).pauseStaking(yieldProviderAddress);
    });
  });

  describe("unpause staking", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).unpauseStaking(yieldProviderAddress);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should successfully unpause when not ossifed", async () => {
      await yieldManager.connect(securityCouncil).pauseStaking(yieldProviderAddress);
      await setWithdrawalReserveToMinimum(yieldManager);
      await yieldManager.connect(securityCouncil).unpauseStaking(yieldProviderAddress);
    });
    it("Should revert unpause when ossifed", async () => {
      await yieldManager.connect(securityCouncil).pauseStaking(yieldProviderAddress);
      await yieldManager.connect(securityCouncil).setYieldProviderIsOssified(yieldProviderAddress, true);
      await setWithdrawalReserveToMinimum(yieldManager);
      const call = yieldManager.connect(securityCouncil).unpauseStaking(yieldProviderAddress);
      await expectRevertWithCustomError(yieldProvider, call, "UnpauseStakingForbiddenWhenOssified");
    });
    it("Should successfully sync LST liability", async () => {
      await yieldManager.connect(nativeYieldOperator).pauseStaking(yieldProviderAddress);
      await yieldManager.setYieldProviderLstLiabilityPrincipal(yieldProviderAddress, 1n);
      await expect(yieldManager.connect(nativeYieldOperator).unpauseStaking(yieldProviderAddress)).to.be.reverted;
      // Will sync to this if below current lstLiabilityPrincipal.
      await mockSTETH.setPooledEthBySharesRoundUpReturn(0n);
      // If it didn't sync lstLiabilityPrincipal to 0, the below will revert.
      yieldManager.connect(nativeYieldOperator).unpauseStaking(yieldProviderAddress);
    });
  });

  describe("withdraw LST", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).withdrawLST(yieldProviderAddress, ZERO_VALUE, ZeroAddress);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should successfully withdraw LST when not ossifed, and update state", async () => {
      const withdrawAmount = ONE_ETHER;
      const lstPrincipalLiabilityBefore =
        await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress);
      const call = getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        withdrawAmount,
      );
      await call;
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress)).eq(
        lstPrincipalLiabilityBefore + withdrawAmount,
      );
    });
    it("Should revert withdraw LST when ossification pending", async () => {
      // Setup fund
      const withdrawAmount = ONE_ETHER;
      const recipient = ethers.Wallet.createRandom().address;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, withdrawAmount);

      // Add L1MessageService balance
      const l1MessageService = await yieldManager.L1_MESSAGE_SERVICE();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(ONE_ETHER)]);
      const l1Signer = await ethers.getImpersonatedSigner(l1MessageService);
      await mockLineaRollup.setWithdrawLSTAllowed(true);

      // Initiate ossification
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);

      const call = yieldManager
        .connect(l1Signer)
        .withdrawLST(await yieldProvider.getAddress(), withdrawAmount, recipient);
      await expectRevertWithCustomError(yieldProvider, call, "MintLSTDisabledDuringOssification");
    });
    it("Should revert withdraw LST when ossifed", async () => {
      const withdrawAmount = ONE_ETHER;
      const recipient = ethers.Wallet.createRandom().address;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, withdrawAmount);

      // Add gas fees
      const l1MessageService = await yieldManager.L1_MESSAGE_SERVICE();
      await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(ONE_ETHER)]);
      const l1Signer = await ethers.getImpersonatedSigner(l1MessageService);
      await mockLineaRollup.setWithdrawLSTAllowed(true);

      await yieldManager.connect(securityCouncil).setYieldProviderIsOssified(yieldProviderAddress, true);

      // Act
      const call = yieldManager
        .connect(l1Signer)
        .withdrawLST(await yieldProvider.getAddress(), withdrawAmount, recipient);

      // Assert
      await expectRevertWithCustomError(yieldProvider, call, "MintLSTDisabledDuringOssification");
    });
  });

  describe("initiate ossification", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should succeed", async () => {
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
    });
    it("Should succeed even if dashboard.voluntaryDisconnect call fails", async () => {
      await mockDashboard.setIsVoluntaryDisconnectRevert(true);
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
    });
  });

  describe("progress pending ossification", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).progressPendingOssification(yieldProviderAddress);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("If vault is disconnected, should succeed and return complete", async () => {
      // Arrange
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await mockVaultHub.setIsVaultConnectedReturn(false);
      // Act
      const isOssificationComplete = await yieldManager
        .connect(securityCouncil)
        .progressPendingOssification.staticCall(yieldProviderAddress);
      const call = yieldManager.connect(securityCouncil).progressPendingOssification(yieldProviderAddress);

      // Assert
      await expect(call).to.not.be.reverted;
      expect(isOssificationComplete).eq(ProgressOssificationResult.COMPLETE);
    });
    it("If vault is connected and pending disconnect, should succeed with no-op and return false", async () => {
      // Arrange
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await mockVaultHub.setIsVaultConnectedReturn(true);
      await mockVaultHub.setIsPendingDisconnectReturn(true);
      // Act
      const isOssificationComplete = await yieldManager
        .connect(securityCouncil)
        .progressPendingOssification.staticCall(yieldProviderAddress);
      const call = yieldManager.connect(securityCouncil).progressPendingOssification(yieldProviderAddress);

      // Assert
      await expect(call).to.not.be.reverted;
      expect(isOssificationComplete).eq(ProgressOssificationResult.NOOP);
    });
    it("If vault is connected and not pending disconnect, should successfully redo initiate ossification and return false", async () => {
      // Arrange
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await mockVaultHub.setIsVaultConnectedReturn(true);
      await mockVaultHub.setIsPendingDisconnectReturn(false);
      // Act
      const isOssificationComplete = await yieldManager
        .connect(securityCouncil)
        .progressPendingOssification.staticCall(yieldProviderAddress);
      const call = yieldManager.connect(securityCouncil).progressPendingOssification(yieldProviderAddress);

      // Assert
      await expect(call).to.not.be.reverted;
      expect(isOssificationComplete).eq(ProgressOssificationResult.REINITIATED);
    });
  });

  describe("vendor initialization", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).initializeVendorContracts(EMPTY_CALLDATA);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should revert if incorrect type for _vendorInitializationData", async () => {
      const call = yieldManager
        .connect(securityCouncil)
        .initializeVendorContracts(yieldProviderAddress, EMPTY_CALLDATA);
      await expect(call).to.be.reverted;
    });
    it("Should succeed with expected return values, and transfer 1 ether to Lido contracts", async () => {
      await yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress, buildVendorExitData());
      const expectedVaultAddress = ethers.Wallet.createRandom().address;
      const expectedDashboardAddress = ethers.Wallet.createRandom().address;
      await mockVaultFactory.setReturnValues(expectedVaultAddress, expectedDashboardAddress);
      await incrementBalance(yieldManagerAddress, ONE_ETHER);
      const beforeYieldManagerBalance = await getBalance(yieldManager);
      const beforeVaultFactoryBalance = await getBalance(mockVaultFactory);

      const registrationData = await yieldManager
        .connect(securityCouncil)
        .initializeVendorContracts.staticCall(yieldProviderAddress, buildVendorInitializationData());

      const call = yieldManager
        .connect(securityCouncil)
        .initializeVendorContracts(yieldProviderAddress, buildVendorInitializationData());
      await expect(call).to.not.be.reverted;
      expect(registrationData[0]).eq(YieldProviderVendor.LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR);
      expect(registrationData[1]).eq(expectedDashboardAddress);
      expect(registrationData[2]).eq(expectedVaultAddress);
      expect(await getBalance(yieldManager)).eq(beforeYieldManagerBalance - ONE_ETHER);
      expect(await getBalance(mockVaultFactory)).eq(beforeVaultFactoryBalance + ONE_ETHER);
    });
    it("Should revert if YieldManager lacks the CONNECT_DEPOSIT of 1 ether", async () => {
      await yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress, buildVendorExitData());
      const expectedVaultAddress = ethers.Wallet.createRandom().address;
      const expectedDashboardAddress = ethers.Wallet.createRandom().address;
      await mockVaultFactory.setReturnValues(expectedVaultAddress, expectedDashboardAddress);

      const call = yieldManager
        .connect(securityCouncil)
        .initializeVendorContracts(yieldProviderAddress, buildVendorInitializationData());
      await expect(call).to.be.reverted;
    });
  });

  describe("vendor exit", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).exitVendorContracts(yieldProviderAddress, EMPTY_CALLDATA);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should revert with empty vendorExitData", async () => {
      const call = yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress, EMPTY_CALLDATA);
      await expectRevertWithCustomError(yieldProvider, call, "NoVendorExitDataProvided");
    });
    it("Should revert if newVaultAddress = 0", async () => {
      const call = yieldManager
        .connect(securityCouncil)
        .exitVendorContracts(yieldProviderAddress, buildVendorExitData({ newVaultOwner: ZeroAddress }));
      expectRevertWithCustomError(yieldManager, call, "ZeroAddressNotAllowed");
    });
    it("When non-ossified, should succeed with call to Dashboard", async () => {
      await yieldManager.setYieldProviderIsOssified(yieldProviderAddress, false);
      const call = yieldManager
        .connect(securityCouncil)
        .removeYieldProvider(yieldProviderAddress, buildVendorExitData());
      await expect(call).to.not.be.reverted;
      expect(await mockDashboard.transferVaultOwnershipCallCount()).eq(1);
      expect(await mockStakingVault.transferOwnershipCallCount()).eq(0);
    });
    it("When ossified, should succeed with call to StakingVault", async () => {
      await yieldManager.setYieldProviderIsOssified(yieldProviderAddress, true);
      const call = yieldManager
        .connect(securityCouncil)
        .removeYieldProvider(yieldProviderAddress, buildVendorExitData());
      await expect(call).to.not.be.reverted;
      expect(await mockDashboard.transferVaultOwnershipCallCount()).eq(0);
      expect(await mockStakingVault.transferOwnershipCallCount()).eq(1);
    });
  });

  describe("unstake", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const mockWithdrawalParams = ethers.hexlify(ethers.randomBytes(8));
      const call = yieldProvider.connect(securityCouncil).unstake(yieldProviderAddress, mockWithdrawalParams);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should revert if incorrect withdrawal params type", async () => {
      const mockWithdrawalParams = ethers.hexlify(ethers.randomBytes(8));
      const call = yieldManager.connect(securityCouncil).unstake(yieldProviderAddress, mockWithdrawalParams);
      await expect(call).to.be.reverted;
    });
    it("Should succeed with valid withdrawal params type", async () => {
      const pubkey = "0x" + "a".repeat(96); // 48 bytes
      const amounts = [32000000000n]; // 32 ETH in Gwei
      const refundRecipient = nativeYieldOperator.address;
      const withdrawalParams = ethers.AbiCoder.defaultAbiCoder().encode(
        ["bytes", "uint64[]", "address"],
        [pubkey, amounts, refundRecipient],
      );

      const call = yieldManager.connect(securityCouncil).unstake(yieldProviderAddress, withdrawalParams);
      await call;
    });
  });

  describe("unstakePermissionless", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const mockWithdrawalParams = ethers.hexlify(ethers.randomBytes(8));
      const mockWithdrawalParamsProof = ethers.hexlify(ethers.randomBytes(8));
      const mockRequiredUnstakeAmountWei = ONE_ETHER;
      const mockValidatorIndex = 0n;
      const mockSlot = 100000n;
      const call = yieldProvider
        .connect(securityCouncil)
        .unstakePermissionless(
          yieldProviderAddress,
          mockRequiredUnstakeAmountWei,
          mockValidatorIndex,
          mockSlot,
          mockWithdrawalParams,
          mockWithdrawalParamsProof,
        );
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should revert if incorrect withdrawal params type", async () => {
      const mockWithdrawalParams = ethers.hexlify(ethers.randomBytes(8));
      const mockWithdrawalParamsProof = ethers.hexlify(ethers.randomBytes(8));
      const mockValidatorIndex = 0n;
      const mockSlot = 100000n;
      const call = yieldManager
        .connect(securityCouncil)
        .unstakePermissionless(
          yieldProviderAddress,
          mockValidatorIndex,
          mockSlot,
          mockWithdrawalParams,
          mockWithdrawalParamsProof,
        );
      await expect(call).to.be.reverted;
    });
    it("Should succeed and update pendingPermissionlessUnstake", async () => {
      // Arrange - Set up withdrawal reserve in deficit
      await setBalance(l1MessageServiceAddress, 0n);

      const { eip4788Witness, pubkey, validatorIndex, slot } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        testVerifier,
        mockStakingVaultAddress,
        THIRTY_TWO_ETH_IN_GWEI,
      );
      const refundAddress = nativeYieldOperator.address;

      const withdrawalParams = ethers.AbiCoder.defaultAbiCoder().encode(["bytes", "address"], [pubkey, refundAddress]);
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [BEACON_PROOF_WITNESS_TYPE],
        [eip4788Witness.beaconProofWitness],
      );

      // Calculate expected unstaked amount (clamped by validator effective balance)
      let expectedUnstakeAmountGwei: bigint;
      if (unstakeAmountGwei < validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI) {
        expectedUnstakeAmountGwei = unstakeAmountGwei;
      } else if (validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI > 0n) {
        expectedUnstakeAmountGwei = validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI;
      } else {
        expectedUnstakeAmountGwei = 0n;
      }
      const expectedUnstakeAmountWei = expectedUnstakeAmountGwei * ONE_GWEI;

      // Act
      await expect(
        yieldManager
          .connect(securityCouncil)
          .unstakePermissionless(yieldProviderAddress, validatorIndex, slot, withdrawalParams, withdrawalParamsProof),
      ).to.not.be.reverted;

      // Assert - Verify pendingPermissionlessUnstake was updated with expected amount
      expect(await yieldManager.pendingPermissionlessUnstake()).to.equal(expectedUnstakeAmountWei);
    });
  });

  describe.skip("validateUnstakePermissionless", () => {
    it("Should revert if pubkeys argument is not 48 bytes exactly", async () => {
      const invalidPubkeys = "0x" + "22".repeat(32);
      const call = yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessRequestHarness(yieldProviderAddress, 1n, invalidPubkeys, 0n, 0n, EMPTY_CALLDATA);

      await expectRevertWithCustomError(yieldProvider, call, "SingleValidatorOnlyForUnstakePermissionless");
    });

    it("Should revert if more than a single amounts element is provided", async () => {
      const pubkeys = "0x" + "11".repeat(48);
      const call = yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessRequestHarness(yieldProviderAddress, 1n, pubkeys, 0n, 0n, EMPTY_CALLDATA);

      await expectRevertWithCustomError(yieldProvider, call, "SingleValidatorOnlyForUnstakePermissionless");
    });

    it("Should revert if incorrect type is provided for proof", async () => {
      const pubkeys = "0x" + "11".repeat(48);
      const call = yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessRequestHarness(yieldProviderAddress, 1n, pubkeys, 0n, 0n, EMPTY_CALLDATA);

      await expect(call).to.be.reverted;
    });
    it("If withdrawal amount leaves validator < activation balance, return maximum available unstake", async () => {
      // Choose 2049 ETH which is > maximum effective balance of 2048 ETH.
      const EXCESSIVE_WITHDRAWAL_AMOUNT = parseUnits("2049", "gwei");
      const { validatorWitness, pubkey } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        testVerifier,
        mockStakingVaultAddress,
      );
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness],
      );

      // Act
      const maxUnstakeAmount = await yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessRequestHarness.staticCall(
          yieldProviderAddress,
          EXCESSIVE_WITHDRAWAL_AMOUNT,
          pubkey,
          validatorWitness.validatorIndex,
          validatorWitness.slot,
          withdrawalParamsProof,
        );

      // Assert
      const expectedMaxUnstakeAmountGwei =
        validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI > 0n
          ? validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI
          : 0n;

      expect(expectedMaxUnstakeAmountGwei * ONE_GWEI).eq(maxUnstakeAmount);
    });
    it("If withdrawal amount leaves validator > activation balance, return amount", async () => {
      const { validatorWitness, pubkey } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        testVerifier,
        mockStakingVaultAddress,
        MAX_0X2_VALIDATOR_EFFECTIVE_BALANCE_GWEI,
      );
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness],
      );
      const WITHDRAWAL_AMOUNT_GWEI =
        validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI - parseUnits("1", "gwei");

      // Act
      const maxUnstakeAmount = await yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessRequestHarness.staticCall(
          yieldProviderAddress,
          WITHDRAWAL_AMOUNT_GWEI,
          pubkey,
          validatorWitness.validatorIndex,
          validatorWitness.slot,
          withdrawalParamsProof,
        );

      // Assert
      expect(WITHDRAWAL_AMOUNT_GWEI * ONE_GWEI).eq(maxUnstakeAmount);
    });
    it("If effective balance < 32 ETH, return 0", async () => {
      const WITHDRAWAL_AMOUNT = parseUnits("2049", "gwei");
      const INSUFFICIENT_EFFECTIVE_BALANCE = parseUnits("31", "gwei");
      const { validatorWitness, pubkey } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        testVerifier,
        mockStakingVaultAddress,
        INSUFFICIENT_EFFECTIVE_BALANCE,
      );
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness],
      );

      // Act
      const maxUnstakeAmount = await yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessRequestHarness.staticCall(
          yieldProviderAddress,
          WITHDRAWAL_AMOUNT,
          pubkey,
          validatorWitness.validatorIndex,
          validatorWitness.slot,
          withdrawalParamsProof,
        );

      // Assert
      expect(0).eq(maxUnstakeAmount);
    });
  });
});
