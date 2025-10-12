import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
import {
  deployAndAddSingleLidoStVaultYieldProvider,
  deployMockStakingVault,
  fundLidoStVaultYieldProvider,
  getWithdrawLSTCall,
  incrementBalance,
  ossifyYieldProvider,
  setWithdrawalReserveToMinimum,
  YieldProviderRegistration,
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
import { ethers } from "hardhat";
import { parseUnits, ZeroAddress } from "ethers";
import {
  GI_FIRST_VALIDATOR,
  GI_FIRST_VALIDATOR_AFTER_CHANGE,
  CHANGE_SLOT,
  ONE_ETHER,
  ZERO_VALUE,
  LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR,
  UNUSED_YIELD_PROVIDER_VENDOR,
  LIDO_DASHBOARD_NOT_LINKED_TO_VAULT,
  LIDO_VAULT_IS_EXPECTED_RECEIVE_CALLER_AND_OSSIFIED_ENTRYPOINT,
  EMPTY_CALLDATA,
  VALIDATOR_WITNESS_TYPE,
  THIRTY_TWO_ETH_IN_GWEI,
} from "../../common/constants";
import { generateLidoUnstakePermissionlessWitness } from "../helpers/proof";

describe("LidoStVaultYieldProvider contract - basic operations", () => {
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
  });

  describe("Constructor", () => {
    it("Should revert if 0 address provided for _l1MessageService", async () => {
      const contractFactory = await ethers.getContractFactory("TestLidoStVaultYieldProvider");
      const call = contractFactory.deploy(
        ZeroAddress,
        yieldManagerAddress,
        vaultHubAddress,
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
    it("Should successfully withdraw when ossifed", async () => {
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

  describe("undoInitiateOssification", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).undoInitiateOssification(yieldProviderAddress);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should succeed when vault is connected", async () => {
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await mockVaultHub.connect(securityCouncil).setIsVaultConnectedReturn(true);
      await yieldManager.connect(securityCouncil).undoInitiateOssification(yieldProviderAddress);
    });
    it("Should succeed when vault is not connected", async () => {
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await mockVaultHub.connect(securityCouncil).setIsVaultConnectedReturn(false);
      await yieldManager.connect(securityCouncil).undoInitiateOssification(yieldProviderAddress);
    });
  });

  describe("process pending ossification", () => {
    it("Should revert if not invoked via delegatecall", async () => {
      const call = yieldProvider.connect(securityCouncil).processPendingOssification(yieldProviderAddress);
      await expectRevertWithCustomError(yieldProvider, call, "ContextIsNotYieldManager");
    });
    it("Should succeed", async () => {
      await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
      await yieldManager.connect(securityCouncil).processPendingOssification(yieldProviderAddress);
    });
  });

  describe("validateAdditionToYieldManager", () => {
    it("Should revert if registration is for unknown YieldProvider Vendor", async () => {
      await yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress);

      const registration: YieldProviderRegistration = {
        yieldProviderVendor: UNUSED_YIELD_PROVIDER_VENDOR,
        primaryEntrypoint: mockDashboardAddress,
        ossifiedEntrypoint: mockStakingVaultAddress,
        receiveCaller: mockStakingVaultAddress,
      };

      const call = yieldManager.connect(securityCouncil).addYieldProvider(yieldProviderAddress, registration);
      await expectRevertWithCustomError(yieldProvider, call, "UnknownYieldProviderVendor");
    });
    it("Should revert if dashboard is not linked to staking vault", async () => {
      await yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress);

      const newVault = await deployMockStakingVault();

      const registration: YieldProviderRegistration = {
        yieldProviderVendor: LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR,
        primaryEntrypoint: mockDashboardAddress,
        ossifiedEntrypoint: await newVault.getAddress(),
        receiveCaller: mockStakingVaultAddress,
      };

      const call = yieldManager.connect(securityCouncil).addYieldProvider(yieldProviderAddress, registration);
      await expectRevertWithCustomError(yieldProvider, call, "InvalidYieldProviderRegistration", [
        LIDO_DASHBOARD_NOT_LINKED_TO_VAULT,
      ]);
    });
    it("Should revert if receiveCaller is not set to the staking vault", async () => {
      await yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress);

      const registration: YieldProviderRegistration = {
        yieldProviderVendor: LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR,
        primaryEntrypoint: mockDashboardAddress,
        ossifiedEntrypoint: mockStakingVaultAddress,
        receiveCaller: mockDashboardAddress,
      };

      const call = yieldManager.connect(securityCouncil).addYieldProvider(yieldProviderAddress, registration);
      await expectRevertWithCustomError(yieldProvider, call, "InvalidYieldProviderRegistration", [
        LIDO_VAULT_IS_EXPECTED_RECEIVE_CALLER_AND_OSSIFIED_ENTRYPOINT,
      ]);
    });
    it("Should succeed for a correct registration", async () => {
      await yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress);

      const registration: YieldProviderRegistration = {
        yieldProviderVendor: LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR,
        primaryEntrypoint: mockDashboardAddress,
        ossifiedEntrypoint: mockStakingVaultAddress,
        receiveCaller: mockStakingVaultAddress,
      };

      await yieldManager.connect(securityCouncil).addYieldProvider(yieldProviderAddress, registration);
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
      const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(mockStakingVaultAddress);
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
        .withArgs(mockStakingVaultAddress, refundAddress, maxUnstakeAmountGwei, validatorWitness.pubkey, unstakeAmount);
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
      const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(mockStakingVaultAddress);
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness],
      );

      // Act
      const maxUnstakeAmountGwei = await yieldManager
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

      expect(expectedMaxUnstakeAmountGwei).eq(maxUnstakeAmountGwei);
    });
    it("If withdrawal amount leaves validator > activation balance, return amount", async () => {
      const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(mockStakingVaultAddress);
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness],
      );
      const WITHDRAWAL_AMOUNT = validatorWitness.effectiveBalance - THIRTY_TWO_ETH_IN_GWEI - parseUnits("1", "gwei");

      // Act
      const maxUnstakeAmountGwei = await yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessHarness.staticCall(
          yieldProviderAddress,
          validatorWitness.pubkey,
          [WITHDRAWAL_AMOUNT],
          withdrawalParamsProof,
        );

      // Assert
      expect(WITHDRAWAL_AMOUNT).eq(maxUnstakeAmountGwei);
    });
    it("If effective balance < 32 ETH, return 0", async () => {
      const WITHDRAWAL_AMOUNT = parseUnits("2049", "gwei");
      const INSUFFICIENT_EFFECTIVE_BALANCE = parseUnits("31", "gwei");
      const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(
        mockStakingVaultAddress,
        INSUFFICIENT_EFFECTIVE_BALANCE,
      );
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness],
      );

      // Act
      const maxUnstakeAmountGwei = await yieldManager
        .connect(securityCouncil)
        .validateUnstakePermissionlessHarness.staticCall(
          yieldProviderAddress,
          validatorWitness.pubkey,
          [WITHDRAWAL_AMOUNT],
          withdrawalParamsProof,
        );

      // Assert
      expect(0).eq(maxUnstakeAmountGwei);
    });
  });
});
