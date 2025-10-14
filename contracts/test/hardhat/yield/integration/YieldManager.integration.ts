// Test scenarios with LineaRollup + YieldManager + LidoStVaultYieldProvider
import { loadFixture, setBalance } from "@nomicfoundation/hardhat-network-helpers";
import { encodeSendMessage, expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
import {
  deployAndAddAdditionalLidoStVaultYieldProvider,
  deployYieldManagerIntegrationTestFixture,
  fundLidoStVaultYieldProvider,
  incrementBalance,
  setupLineaRollupMessageMerkleTree,
  setWithdrawalReserveToMinimum,
  setWithdrawalReserveToTarget,
} from "../helpers";
import {
  TestYieldManager,
  TestLineaRollup,
  TestLidoStVaultYieldProvider,
  MockDashboard,
  MockStakingVault,
  TestLidoStVaultYieldProviderFactory,
  SSZMerkleTree,
  TestCLProofVerifier,
} from "contracts/typechain-types";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { ethers } from "hardhat";
import {
  EMPTY_CALLDATA,
  MAX_0X2_VALIDATOR_EFFECTIVE_BALANCE_GWEI,
  ONE_ETHER,
  ONE_GWEI,
  VALIDATOR_WITNESS_TYPE,
  ZERO_VALUE,
} from "../../common/constants";
import { generateLidoUnstakePermissionlessWitness } from "../helpers/proof";

describe("Integration tests with LineaRollup, YieldManager and LidoStVaultYieldProvider", () => {
  let nativeYieldOperator: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  let l2YieldRecipient: SignerWithAddress;
  let securityCouncil: SignerWithAddress;

  let lineaRollup: TestLineaRollup;
  let yieldManager: TestYieldManager;
  let yieldProvider: TestLidoStVaultYieldProvider;
  let mockDashboard: MockDashboard;
  let mockStakingVault: MockStakingVault;
  let lidoStVaultYieldProviderFactory: TestLidoStVaultYieldProviderFactory;
  let sszMerkleTree: SSZMerkleTree;
  let verifier: TestCLProofVerifier;

  let l1MessageServiceAddress: string;
  let yieldManagerAddress: string;
  let yieldProviderAddress: string;
  let mockStakingVaultAddress: string;

  before(async () => {
    ({ nativeYieldOperator, nonAuthorizedAccount, l2YieldRecipient, securityCouncil } =
      await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({
      lineaRollup,
      yieldProvider,
      yieldProviderAddress,
      yieldManager,
      mockDashboard,
      mockStakingVault,
      lidoStVaultYieldProviderFactory,
      sszMerkleTree,
      verifier,
    } = await loadFixture(deployYieldManagerIntegrationTestFixture));
    l1MessageServiceAddress = await lineaRollup.getAddress();
    yieldManagerAddress = await yieldManager.getAddress();
    mockStakingVaultAddress = await mockStakingVault.getAddress();
  });

  describe("Donations", () => {
    it("Donations should arrive on the LineaRollup", async () => {
      const rollupBalanceBefore = await ethers.provider.getBalance(l1MessageServiceAddress);
      const donationAmount = ONE_ETHER;
      await yieldManager.connect(nativeYieldOperator).donate(yieldProviderAddress, { value: donationAmount });
      const rollupBalanceAfter = await ethers.provider.getBalance(l1MessageServiceAddress);
      expect(rollupBalanceAfter).eq(rollupBalanceBefore + donationAmount);
    });
  });

  describe("Transfering to the YieldManager", () => {
    it("Should successfully transfer from LineaRollup to the YieldManager", async () => {
      const fundAmount = ONE_ETHER;
      await setWithdrawalReserveToMinimum(yieldManager);
      await incrementBalance(l1MessageServiceAddress, fundAmount);
      const rollupBalanceBefore = await ethers.provider.getBalance(l1MessageServiceAddress);
      const yieldManagerBalanceBefore = await ethers.provider.getBalance(yieldManagerAddress);
      // Act
      await lineaRollup.connect(nativeYieldOperator).transferFundsForNativeYield(fundAmount);
      // Assert
      const rollupBalanceAfter = await ethers.provider.getBalance(l1MessageServiceAddress);
      const yieldManagerBalanceAfter = await ethers.provider.getBalance(yieldManagerAddress);
      expect(rollupBalanceAfter).eq(rollupBalanceBefore - fundAmount);
      expect(yieldManagerBalanceAfter).eq(yieldManagerBalanceBefore + fundAmount);
    });
    it("Should revert when withdrawal reserve at minimum", async () => {
      await setWithdrawalReserveToMinimum(yieldManager);
      // Act
      const call = lineaRollup.connect(nativeYieldOperator).transferFundsForNativeYield(1);
      // Assert
      await expectRevertWithCustomError(yieldManager, call, "InsufficientWithdrawalReserve");
    });
  });

  describe("Withdraw LST", () => {
    it("Should allow LST withdrawal when in deficit", async () => {
      // Arrange - setup user funds
      const initialFundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);
      // Arrange - setup withdrawal reserve deficit
      await setBalance(l1MessageServiceAddress, ZERO_VALUE);
      // Arrange - setup L1MessageService message
      const withdrawAmount = initialFundAmount / 2n;
      const recipientAddress = await nonAuthorizedAccount.getAddress();
      const claimParams = await setupLineaRollupMessageMerkleTree(
        lineaRollup,
        recipientAddress,
        recipientAddress,
        withdrawAmount,
        EMPTY_CALLDATA,
      );
      // Arrange - Get before figures
      const lstPrincipalBefore = await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress);

      // Act
      const claimCall = lineaRollup
        .connect(nonAuthorizedAccount)
        .claimMessageWithProofAndWithdrawLST(claimParams, yieldProviderAddress);

      // Assert
      await expect(claimCall).to.not.be.reverted;
      const lstPrincipalAfter = await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress);
      expect(lstPrincipalAfter).eq(lstPrincipalBefore + withdrawAmount);
    });
    it("Should not be allowed from claimMessage", async () => {
      // Arrange - setup user funds
      const initialFundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);
      // Arrange - setup withdrawal reserve deficit
      await setBalance(l1MessageServiceAddress, ZERO_VALUE);
      // Arrange - setup L1MessageService message
      const withdrawAmount = 0n;
      const recipientAddress = await nonAuthorizedAccount.getAddress();
      const calldata = await yieldProvider.interface.encodeFunctionData("withdrawLST", [
        yieldProviderAddress,
        withdrawAmount,
        recipientAddress,
      ]);

      const claimParams = await setupLineaRollupMessageMerkleTree(
        lineaRollup,
        recipientAddress,
        yieldManagerAddress,
        withdrawAmount,
        calldata,
      );

      // Act
      const claimCall = lineaRollup.connect(nonAuthorizedAccount).claimMessageWithProof(claimParams);
      await expectRevertWithCustomError(yieldManager, claimCall, "LSTWithdrawalNotAllowed");
    });
    it("Should not allow LST withdrawal amount > fund amount", async () => {
      // Arrange - setup user funds
      const initialFundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);
      // Arrange - setup withdrawal reserve deficit
      await setBalance(l1MessageServiceAddress, ZERO_VALUE);
      // Arrange - setup L1MessageService message
      const withdrawAmount = initialFundAmount * 2n;
      const recipientAddress = await nonAuthorizedAccount.getAddress();
      const claimParams = await setupLineaRollupMessageMerkleTree(
        lineaRollup,
        recipientAddress,
        recipientAddress,
        withdrawAmount,
        EMPTY_CALLDATA,
      );

      // Act
      const claimCall = lineaRollup
        .connect(nonAuthorizedAccount)
        .claimMessageWithProofAndWithdrawLST(claimParams, yieldProviderAddress);
      await expectRevertWithCustomError(yieldManager, claimCall, "LSTWithdrawalExceedsYieldProviderFunds");
    });
    it("Should revert if LST withdrawal is for another recipient", async () => {
      // Arrange - setup user funds
      const initialFundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);
      // Arrange - setup withdrawal reserve deficit
      await setBalance(l1MessageServiceAddress, ZERO_VALUE);
      // Arrange - setup L1MessageService message
      const withdrawAmount = initialFundAmount / 2n;
      const recipientAddress = await nonAuthorizedAccount.getAddress();
      const claimParams = await setupLineaRollupMessageMerkleTree(
        lineaRollup,
        recipientAddress,
        recipientAddress,
        withdrawAmount,
        EMPTY_CALLDATA,
      );

      // Act - Execute with different sender than _to
      const claimCall = lineaRollup
        .connect(l2YieldRecipient)
        .claimMessageWithProofAndWithdrawLST(claimParams, yieldProviderAddress);

      // Assert
      expectRevertWithCustomError(lineaRollup, claimCall, "CallerNotLSTWithdrawalRecipient");
    });
  });

  describe("Yield reporting", () => {
    it("Should report positive yield successfully", async () => {
      // Arrange - setup user funds
      const initialFundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);
      // Arrange - Setup positive yield
      const yieldEarned = ONE_ETHER / 10n;
      await mockDashboard.setTotalValueReturn(initialFundAmount + yieldEarned);
      // Arrange - Get message params
      const nextMessageNumberBefore = await lineaRollup.nextMessageNumber();

      const expectedBytes = await encodeSendMessage(
        l1MessageServiceAddress,
        await l2YieldRecipient.getAddress(),
        0n,
        yieldEarned,
        nextMessageNumberBefore,
        EMPTY_CALLDATA,
      );
      const messageHash = ethers.keccak256(expectedBytes);

      // Act
      const call = await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);

      // Assert
      await expect(call)
        .to.emit(lineaRollup, "MessageSent")
        .withArgs(
          yieldManagerAddress,
          l2YieldRecipient,
          0,
          yieldEarned,
          nextMessageNumberBefore,
          EMPTY_CALLDATA,
          messageHash,
        );
    });
    it("Should report negative yield successfully", async () => {
      // Arrange - setup user funds
      const initialFundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);
      // Arrange - Setup negative yield
      const yieldEarned = 0n;
      await mockDashboard.setTotalValueReturn(yieldEarned);
      // Arrange - Get message params
      const nextMessageNumberBefore = await lineaRollup.nextMessageNumber();

      const expectedBytes = await encodeSendMessage(
        l1MessageServiceAddress,
        await l2YieldRecipient.getAddress(),
        0n,
        yieldEarned,
        nextMessageNumberBefore,
        EMPTY_CALLDATA,
      );
      const messageHash = ethers.keccak256(expectedBytes);
      const negativeYieldBefore = await yieldManager.getYieldProviderCurrentNegativeYield(yieldProviderAddress);

      // Act
      const call = await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);

      // Assert
      await expect(call)
        .to.emit(lineaRollup, "MessageSent")
        .withArgs(
          yieldManagerAddress,
          l2YieldRecipient,
          0,
          yieldEarned,
          nextMessageNumberBefore,
          EMPTY_CALLDATA,
          messageHash,
        );

      const negativeYieldAfter = await yieldManager.getYieldProviderCurrentNegativeYield(yieldProviderAddress);
      expect(negativeYieldAfter).eq(negativeYieldBefore + initialFundAmount - yieldEarned);
    });
  });

  describe("Migrating yield providers", () => {
    it("Cannot remove a yield provider with funds", async () => {
      // Arrange - fund a single vault
      const initialFundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);
      // Arrange - single positive vault report
      const yieldEarned = ONE_ETHER;
      await mockDashboard.setTotalValueReturn(initialFundAmount + yieldEarned);
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);

      // Act
      const call = yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress);

      // Assert
      await expectRevertWithCustomError(yieldManager, call, "YieldProviderHasRemainingFunds");
    });
    it("Can migrate funds successfully to a new YieldProvider", async () => {
      // Arrange - fund a single vault
      const initialFundAmount = ONE_ETHER;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);
      // Arrange - single positive vault report
      const yieldEarned = ONE_ETHER;
      await mockDashboard.setTotalValueReturn(initialFundAmount + yieldEarned);
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
      // Arrange - withdraw vault funds
      await incrementBalance(await mockStakingVault.getAddress(), ONE_ETHER);
      await setWithdrawalReserveToTarget(yieldManager);
      await yieldManager.connect(nativeYieldOperator).withdrawFromYieldProvider(yieldProviderAddress, ONE_ETHER * 2n);
      // Arrange - Remove yield provider
      await yieldManager.connect(securityCouncil).removeYieldProvider(yieldProviderAddress);
      // Arrange - Add new yield provider
      const { yieldProviderAddress: yieldProvider2Address } = await deployAndAddAdditionalLidoStVaultYieldProvider(
        lidoStVaultYieldProviderFactory,
        yieldManager,
        securityCouncil,
      );

      // Act - Move funds to new YieldProvider
      await yieldManager.connect(nativeYieldOperator).fundYieldProvider(yieldProvider2Address, ONE_ETHER);
    });
  });

  describe("Multiple yield providers", () => {
    it("Unstake permissionless cap should be shared globally across all yield providers", async () => {
      // Arrange - add additional yield provider
      const { yieldProviderAddress: yieldProvider2Address, mockStakingVaultAddress: mockStakingVault2Address } =
        await deployAndAddAdditionalLidoStVaultYieldProvider(
          lidoStVaultYieldProviderFactory,
          yieldManager,
          securityCouncil,
        );
      // Arrange - Prepare first unstakePermissionless
      const targetDeficit = await yieldManager.getTargetReserveDeficit();
      const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        verifier,
        mockStakingVaultAddress,
        MAX_0X2_VALIDATOR_EFFECTIVE_BALANCE_GWEI,
      );
      const refundAddress = nativeYieldOperator.address;
      const unstakeAmount = [targetDeficit / ONE_GWEI];
      const withdrawalParams = ethers.AbiCoder.defaultAbiCoder().encode(
        ["bytes", "uint64[]", "address"],
        [validatorWitness.pubkey, unstakeAmount, refundAddress],
      );
      const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness],
      );
      // Arrange - first unstake
      await yieldManager.unstakePermissionless(yieldProviderAddress, withdrawalParams, withdrawalParamsProof);
      expect(await yieldManager.pendingPermissionlessUnstake()).eq(targetDeficit);

      // Arrange - Prepare second unstakePermissionless
      const { validatorWitness: validatorWitness2 } = await generateLidoUnstakePermissionlessWitness(
        sszMerkleTree,
        verifier,
        mockStakingVault2Address,
        MAX_0X2_VALIDATOR_EFFECTIVE_BALANCE_GWEI,
      );

      const secondWithdrawalParams = ethers.AbiCoder.defaultAbiCoder().encode(
        ["bytes", "uint64[]", "address"],
        [validatorWitness2.pubkey, [1n], refundAddress],
      );
      const secondWithdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode(
        [VALIDATOR_WITNESS_TYPE],
        [validatorWitness2],
      );

      // Act
      const call = yieldManager.unstakePermissionless(
        yieldProvider2Address,
        secondWithdrawalParams,
        secondWithdrawalParamsProof,
      );
      await expectRevertWithCustomError(
        yieldManager,
        call,
        "PermissionlessUnstakeRequestPlusAvailableFundsExceedsTargetDeficit",
      );
    });
  });
});
