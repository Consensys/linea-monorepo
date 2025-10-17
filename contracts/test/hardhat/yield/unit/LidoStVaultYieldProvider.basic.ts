import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
import {
  buildVendorExitData,
  buildVendorInitializationData,
  deployAndAddSingleLidoStVaultYieldProvider,
  fundLidoStVaultYieldProvider,
  getWithdrawLSTCall,
  incrementBalance,
  ossifyYieldProvider,
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
  TestCLProofVerifier,
  SSZMerkleTree,
} from "contracts/typechain-types";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { parseUnits, ZeroAddress } from "ethers";
import {
  GI_FIRST_VALIDATOR,
  GI_FIRST_VALIDATOR_AFTER_CHANGE,
  CHANGE_SLOT,
  ONE_ETHER,
  ZERO_VALUE,
  EMPTY_CALLDATA,
  VALIDATOR_WITNESS_TYPE,
  THIRTY_TWO_ETH_IN_GWEI,
  ONE_GWEI,
  MAX_0X2_VALIDATOR_EFFECTIVE_BALANCE_GWEI,
  ProgressOssificationResult,
  YieldProviderVendor,
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
  let verifier: TestCLProofVerifier;

  let l1MessageServiceAddress: string;
  let yieldManagerAddress: string;
  let vaultHubAddress: string;
  let vaultFactoryAddress: string;
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
      mockVaultFactory,
      mockSTETH,
      mockLineaRollup,
      sszMerkleTree,
      verifier,
    } = await loadFixture(deployAndAddSingleLidoStVaultYieldProvider));

    l1MessageServiceAddress = await mockLineaRollup.getAddress();
    yieldManagerAddress = await yieldManager.getAddress();
    vaultHubAddress = await mockVaultHub.getAddress();
    vaultFactoryAddress = await mockVaultFactory.getAddress();
    stethAddress = await mockSTETH.getAddress();
    mockDashboardAddress = await mockDashboard.getAddress();
    mockStakingVaultAddress = await mockStakingVault.getAddress();
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
        GI_FIRST_VALIDATOR,
        GI_FIRST_VALIDATOR_AFTER_CHANGE,
        CHANGE_SLOT,
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
          GI_FIRST_VALIDATOR,
          GI_FIRST_VALIDATOR_AFTER_CHANGE,
          CHANGE_SLOT,
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
        GI_FIRST_VALIDATOR,
        GI_FIRST_VALIDATOR_AFTER_CHANGE,
        CHANGE_SLOT,
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
        GI_FIRST_VALIDATOR,
        GI_FIRST_VALIDATOR_AFTER_CHANGE,
        CHANGE_SLOT,
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
        GI_FIRST_VALIDATOR,
        GI_FIRST_VALIDATOR_AFTER_CHANGE,
        CHANGE_SLOT,
      );
      await expectRevertWithCustomError(contractFactory, call, "ZeroAddressNotAllowed");
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
    it("If not ossified, should fund the Dashboard", async () => {
      const beforeVaultBalance = await ethers.provider.getBalance(mockStakingVaultAddress);
      const fundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, fundAmount);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(beforeVaultBalance + fundAmount);
    });
    it("If ossified, should fund the StakingVault", async () => {
      await ossifyYieldProvider(yieldManager, yieldProviderAddress, securityCouncil);
      const beforeVaultBalance = await ethers.provider.getBalance(mockStakingVaultAddress);
      const fundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, fundAmount);
      expect(await ethers.provider.getBalance(mockStakingVaultAddress)).eq(beforeVaultBalance + fundAmount);
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
      await yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(yieldProviderAddress, withdrawAmount);
    });
    it("Should successfully withdraw when ossified", async () => {
      await ossifyYieldProvider(yieldManager, yieldProviderAddress, securityCouncil);
      const withdrawAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, withdrawAmount);
      await yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(yieldProviderAddress, withdrawAmount);
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
      const withdrawAmount = ONE_ETHER;
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      const call = getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        withdrawAmount,
      );
      await expectRevertWithCustomError(yieldProvider, call, "MintLSTDisabledDuringOssification");
    });
    it("Should revert withdraw LST when ossifed", async () => {
      const withdrawAmount = ONE_ETHER;
      await yieldManager.connect(securityCouncil).setYieldProviderIsOssified(yieldProviderAddress, true);
      const call = getWithdrawLSTCall(
        mockLineaRollup,
        yieldManager,
        yieldProvider,
        nativeYieldOperator,
        withdrawAmount,
      );
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
    it("Should succeed with expected return values", async () => {
      await yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress, buildVendorExitData());
      const expectedVaultAddress = ethers.Wallet.createRandom().address;
      const expectedDashboardAddress = ethers.Wallet.createRandom().address;
      await mockVaultFactory.setReturnValues(expectedVaultAddress, expectedDashboardAddress);

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
    });
  });

  describe("vendor exit", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).exitVendorContracts(yieldProviderAddress, EMPTY_CALLDATA);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should succeed with empty vendorExitData", async () => {
      const call = yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress, EMPTY_CALLDATA);
      await expect(call).to.not.be.reverted;
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
      const call = yieldProvider
        .connect(securityCouncil)
        .unstakePermissionless(yieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should revert if incorrect withdrawal params type", async () => {
      const mockWithdrawalParams = ethers.hexlify(ethers.randomBytes(8));
      const mockWithdrawalParamsProof = ethers.hexlify(ethers.randomBytes(8));
      const call = yieldManager
        .connect(securityCouncil)
        .unstakePermissionless(yieldProviderAddress, mockWithdrawalParams, mockWithdrawalParamsProof);
      await expect(call).to.be.reverted;
    });
    it("Should succeed and emit the expected event", async () => {
      const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        verifier,
        mockStakingVaultAddress,
      );
      const refundAddress = nativeYieldOperator.address;
      const unstakeAmount = [32000000000n];
      const withdrawalParams = ethers.AbiCoder.defaultAbiCoder().encode(
        ["bytes", "uint64[]", "address"],
        [validatorWitness.pubkey, unstakeAmount, refundAddress],
      );
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness],
      );

      // Act
      const call = yieldManager
        .connect(securityCouncil)
        .unstakePermissionless(yieldProviderAddress, withdrawalParams, withdrawalParamsProof);

      // Assert
      let maxUnstakeAmountGwei: bigint;
      if (unstakeAmount[0] < validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI) {
        maxUnstakeAmountGwei = unstakeAmount[0];
      } else if (validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI > 0n) {
        maxUnstakeAmountGwei = validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI;
      } else {
        maxUnstakeAmountGwei = 0n;
      }
      await expect(call)
        .to.emit(yieldManager, "LidoVaultUnstakePermissionlessRequest")
        .withArgs(
          mockStakingVaultAddress,
          refundAddress,
          maxUnstakeAmountGwei * ONE_GWEI,
          validatorWitness.pubkey,
          unstakeAmount,
        );
    });
  });

  describe("validateUnstakePermissionless", () => {
    it("Should revert if pubkeys argument is not 48 bytes exactly", async () => {
      const invalidPubkeys = "0x" + "22".repeat(32);
      const call = yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessHarness(yieldProviderAddress, invalidPubkeys, [1n], EMPTY_CALLDATA);

      await expectRevertWithCustomError(yieldProvider, call, "SingleValidatorOnlyForUnstakePermissionless");
    });

    it("Should revert if more than a single amounts element is provided", async () => {
      const pubkeys = "0x" + "11".repeat(48);
      const amounts = [1n, 1n];
      const call = yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessHarness(yieldProviderAddress, pubkeys, amounts, EMPTY_CALLDATA);

      await expectRevertWithCustomError(yieldProvider, call, "SingleValidatorOnlyForUnstakePermissionless");
    });

    it("Should revert if 0 amount is provided", async () => {
      const pubkeys = "0x" + "11".repeat(48);
      const call = yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessHarness(yieldProviderAddress, pubkeys, [0n], EMPTY_CALLDATA);

      await expectRevertWithCustomError(yieldProvider, call, "NoValidatorExitForUnstakePermissionless");
    });

    it("Should revert if incorrect type is provided for proof", async () => {
      const pubkeys = "0x" + "11".repeat(48);
      const call = yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessHarness(yieldProviderAddress, pubkeys, [1n], EMPTY_CALLDATA);

      await expect(call).to.be.reverted;
    });
    it("If withdrawal amount leaves validator < activation balance, return maximum available unstake", async () => {
      // Choose 2049 ETH which is > maximum effective balance of 2048 ETH.
      const EXCESSIVE_WITHDRAWAL_AMOUNT = parseUnits("2049", "gwei");
      const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        verifier,
        mockStakingVaultAddress,
      );
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness],
      );

      // Act
      const maxUnstakeAmount = await yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessHarness.staticCall(
          yieldProviderAddress,
          validatorWitness.pubkey,
          [EXCESSIVE_WITHDRAWAL_AMOUNT],
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
      const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        verifier,
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
        .validateUnstakePermissionlessHarness.staticCall(
          yieldProviderAddress,
          validatorWitness.pubkey,
          [WITHDRAWAL_AMOUNT_GWEI],
          withdrawalParamsProof,
        );

      // Assert
      expect(WITHDRAWAL_AMOUNT_GWEI * ONE_GWEI).eq(maxUnstakeAmount);
    });
    it("If effective balance < 32 ETH, return 0", async () => {
      const WITHDRAWAL_AMOUNT = parseUnits("2049", "gwei");
      const INSUFFICIENT_EFFECTIVE_BALANCE = parseUnits("31", "gwei");
      const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        verifier,
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
        .validateUnstakePermissionlessHarness.staticCall(
          yieldProviderAddress,
          validatorWitness.pubkey,
          [WITHDRAWAL_AMOUNT],
          withdrawalParamsProof,
        );

      // Assert
      expect(0).eq(maxUnstakeAmount);
    });
  });
});
