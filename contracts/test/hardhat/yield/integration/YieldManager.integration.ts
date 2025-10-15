// Test scenarios with LineaRollup + YieldManager + LidoStVaultYieldProvider
import { loadFixture, setBalance } from "@nomicfoundation/hardhat-network-helpers";
import { encodeSendMessage, expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
import {
  decrementBalance,
  deployAndAddAdditionalLidoStVaultYieldProvider,
  deployYieldManagerIntegrationTestFixture,
  fundLidoStVaultYieldProvider,
  getBalance,
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
  MockSTETH,
  MockVaultHub,
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
  let mockSTETH: MockSTETH;
  let mockVaultHub: MockVaultHub;
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
      mockSTETH,
      mockVaultHub,
      lidoStVaultYieldProviderFactory,
      sszMerkleTree,
      verifier,
    } = await loadFixture(deployYieldManagerIntegrationTestFixture));
    l1MessageServiceAddress = await lineaRollup.getAddress();
    yieldManagerAddress = await yieldManager.getAddress();
    mockStakingVaultAddress = await mockStakingVault.getAddress();
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
      const [newReportedYield, outstandingNegativeYield] = await yieldManager
        .connect(nativeYieldOperator)
        .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
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
      expect(newReportedYield).eq(yieldEarned);
      expect(outstandingNegativeYield).eq(0);
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

      // Act
      const [newReportedYield, outstandingNegativeYield] = await yieldManager
        .connect(nativeYieldOperator)
        .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
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
      expect(newReportedYield).eq(0n);
      expect(outstandingNegativeYield).eq(initialFundAmount - yieldEarned);
    });
    it("Immediate subsequent positive yield report with no fund movement should have 0 incremental earned yield", async () => {
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
      const [newReportedYield, outstandingNegativeYield] = await yieldManager
        .connect(nativeYieldOperator)
        .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
      const call = await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
      const [newReportedYield2, outstandingNegativeYield2] = await yieldManager
        .connect(nativeYieldOperator)
        .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);

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
      expect(newReportedYield).eq(yieldEarned);
      expect(outstandingNegativeYield).eq(0);
      expect(newReportedYield2).eq(0);
      expect(outstandingNegativeYield2).eq(0);
    });
    it("Immediate subsequent negative yield report with no fund movement should have same outstanding negative yield", async () => {
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

      // Act
      const [newReportedYield, outstandingNegativeYield] = await yieldManager
        .connect(nativeYieldOperator)
        .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
      const call = await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
      const [newReportedYield2, outstandingNegativeYield2] = await yieldManager
        .connect(nativeYieldOperator)
        .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);

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
      expect(newReportedYield).eq(0n);
      expect(outstandingNegativeYield).eq(initialFundAmount - yieldEarned);
      expect(newReportedYield2).eq(0n);
      expect(outstandingNegativeYield2).eq(outstandingNegativeYield);
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

  describe("Native yield resilience scenarios", () => {
    it("Should not reduce user funds after repeated negative yield reports", async () => {
      // Arrange - setup user funds
      const initialFundAmount = ONE_ETHER * 10n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);
      const userFundsInYieldProvidersTotalBefore = await yieldManager.userFundsInYieldProvidersTotal();
      const userFundsBefore = await yieldManager.userFunds(yieldProviderAddress);
      // Arrange - Setup first negative yield
      const firstNegativeYield = ONE_ETHER * 5n;
      await mockDashboard.setTotalValueReturn(initialFundAmount - firstNegativeYield);
      // Arrange - Setup second negative yield
      const secondNegativeYield = ONE_ETHER * 5n;
      await mockDashboard.setTotalValueReturn(initialFundAmount - firstNegativeYield - secondNegativeYield);

      // Act
      const [newReportedYield, outstandingNegativeYield] = await yieldManager
        .connect(nativeYieldOperator)
        .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);

      // Assert
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(userFundsInYieldProvidersTotalBefore);
      expect(await yieldManager.userFunds(yieldProviderAddress)).eq(userFundsBefore);
      expect(newReportedYield).eq(0);
      expect(outstandingNegativeYield).eq(firstNegativeYield + secondNegativeYield);
    });

    it("Should be able to recover after temporary negative yield and max LST withdrawal", async () => {
      // Initial funding with 10 ETH
      const initialFundAmount = ONE_ETHER * 10n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);

      // Negative yield event for -5 ETH
      const firstNegativeYield = ONE_ETHER * 5n;
      await mockDashboard.setTotalValueReturn(initialFundAmount - firstNegativeYield);
      await decrementBalance(mockStakingVaultAddress, firstNegativeYield);
      {
        const [newReportedYield, outstandingNegativeYield] = await yieldManager
          .connect(nativeYieldOperator)
          .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
        expect(newReportedYield).eq(0);
        expect(outstandingNegativeYield).eq(firstNegativeYield);
      }
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);

      // Drain all L1MessageService funds
      await setBalance(l1MessageServiceAddress, ZERO_VALUE);

      // Max LST withdrawal for 10 ETH
      await lineaRollup.connect(securityCouncil).resetRateLimitAmount(ONE_ETHER * 100n);
      const withdrawLSTAmount = initialFundAmount;
      const recipientAddress = await nonAuthorizedAccount.getAddress();
      const claimParams = await setupLineaRollupMessageMerkleTree(
        lineaRollup,
        recipientAddress,
        recipientAddress,
        withdrawLSTAmount,
        EMPTY_CALLDATA,
      );
      await lineaRollup
        .connect(nonAuthorizedAccount)
        .claimMessageWithProofAndWithdrawLST(claimParams, yieldProviderAddress);
      await mockSTETH.setPooledEthBySharesRoundUpReturn(withdrawLSTAmount);

      // Assert - Cannot do another LST withdrawal
      const claimParams2 = await setupLineaRollupMessageMerkleTree(
        lineaRollup,
        recipientAddress,
        recipientAddress,
        1n,
        EMPTY_CALLDATA,
      );
      const secondWithdrawLSTCall = lineaRollup
        .connect(nonAuthorizedAccount)
        .claimMessageWithProofAndWithdrawLST(claimParams2, yieldProviderAddress);
      expectRevertWithCustomError(yieldManager, secondWithdrawLSTCall, "LSTWithdrawalExceedsYieldProviderFunds");
      expectRevertWithCustomError(
        yieldManager,
        yieldManager.replenishWithdrawalReserve(yieldProviderAddress),
        "NoAvailableFundsToReplenishWithdrawalReserve",
      );

      // replenishWithdrawalReserve
      let stakingVaultBalance = await ethers.provider.getBalance(mockStakingVaultAddress);
      await mockDashboard.setWithdrawableValueReturn(stakingVaultBalance);
      let l1MessageBalance = await ethers.provider.getBalance(l1MessageServiceAddress);
      let userFunds = await yieldManager.userFunds(yieldProviderAddress);
      await yieldManager.replenishWithdrawalReserve(yieldProviderAddress);
      expect(await ethers.provider.getBalance(l1MessageServiceAddress)).eq(l1MessageBalance + stakingVaultBalance);
      expect(await yieldManager.userFunds(yieldProviderAddress)).eq(userFunds - stakingVaultBalance);

      // Earn 3 ETH positive yield
      const firstPositiveYield = ONE_ETHER * 3n;
      await incrementBalance(mockStakingVaultAddress, firstPositiveYield);
      stakingVaultBalance = await ethers.provider.getBalance(mockStakingVaultAddress);
      await mockDashboard.setTotalValueReturn(stakingVaultBalance);
      userFunds = await yieldManager.userFunds(yieldProviderAddress);
      {
        const [newReportedYield, outstandingNegativeYield] = await yieldManager
          .connect(nativeYieldOperator)
          .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
        expect(newReportedYield).eq(0);
        expect(outstandingNegativeYield).eq(firstNegativeYield - firstPositiveYield);
      }
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
      expect(await yieldManager.userFunds(yieldProviderAddress)).eq(userFunds);
      expect(await yieldManager.getYieldProviderYieldReportedCumulative(yieldProviderAddress)).eq(0);
      expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(
        await yieldManager.userFunds(yieldProviderAddress),
      );

      // Add to withdrawal reserve (expect route all funds to L1MessageService)
      stakingVaultBalance = await ethers.provider.getBalance(mockStakingVaultAddress);
      userFunds = await yieldManager.userFunds(yieldProviderAddress);
      l1MessageBalance = await ethers.provider.getBalance(l1MessageServiceAddress);
      await yieldManager.connect(nativeYieldOperator).addToWithdrawalReserve(yieldProviderAddress, stakingVaultBalance);
      expect(await yieldManager.userFunds(yieldProviderAddress)).eq(userFunds - stakingVaultBalance);
      expect(await ethers.provider.getBalance(l1MessageServiceAddress)).eq(l1MessageBalance + stakingVaultBalance);

      // Donation to replenish reserve to target
      await lineaRollup.connect(securityCouncil).fund({ value: await yieldManager.getTargetReserveDeficit() });

      // Fresh deposit of 10 ETH -> should clear all existing LST principal
      const secondFundAmount = ONE_ETHER * 10n;
      await lineaRollup.connect(nativeYieldOperator).transferFundsForNativeYield(secondFundAmount);
      await yieldManager.connect(nativeYieldOperator).fundYieldProvider(yieldProviderAddress, secondFundAmount);
      expect(await yieldManager.getYieldProviderLstLiabilityPrincipal(yieldProviderAddress)).eq(0);
    });

    it("Should be able to safely handle different yield report situations", async () => {
      // Initial funding with 10 ETH
      const initialFundAmount = ONE_ETHER * 10n;
      await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, nativeYieldOperator, initialFundAmount);

      // First negative yield event for -5 ETH
      const firstNegativeYield = ONE_ETHER * 5n;
      await mockDashboard.setTotalValueReturn(initialFundAmount - firstNegativeYield);
      await decrementBalance(mockStakingVaultAddress, firstNegativeYield);
      {
        const [newReportedYield, outstandingNegativeYield] = await yieldManager
          .connect(nativeYieldOperator)
          .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
        expect(newReportedYield).eq(0);
        expect(outstandingNegativeYield).eq(firstNegativeYield);
      }
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);

      // Second negative yield event for -1 ETH, with -2 ETH for obligation settlement
      // Expect obligations not paid
      const secondNegativeYield = ONE_ETHER;
      await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) - secondNegativeYield);
      await decrementBalance(mockStakingVaultAddress, secondNegativeYield);
      const firstObligationsPaid = ONE_ETHER * 2n;
      await mockVaultHub.setIsSettleVaultObligationsWithdrawingFromVault(true);
      await mockVaultHub.setSettleVaultObligationAmount(firstObligationsPaid);
      {
        const [newReportedYield, outstandingNegativeYield] = await yieldManager
          .connect(nativeYieldOperator)
          .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
        expect(newReportedYield).eq(0);
        expect(outstandingNegativeYield).eq(firstNegativeYield + secondNegativeYield);
      }
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
      expect(await yieldManager.userFunds(yieldProviderAddress)).eq(initialFundAmount);
      expect(await getBalance(mockStakingVault)).eq(initialFundAmount - firstNegativeYield - secondNegativeYield);

      // First positive yield of 4 ETH, not enough to recover from all negative yield incurred so far
      const firstPositiveYield = ONE_ETHER * 4n;
      await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) + firstPositiveYield);
      await incrementBalance(mockStakingVaultAddress, firstPositiveYield);
      {
        const [newReportedYield, outstandingNegativeYield] = await yieldManager
          .connect(nativeYieldOperator)
          .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
        expect(newReportedYield).eq(0);
        expect(outstandingNegativeYield).eq(firstNegativeYield + secondNegativeYield - firstPositiveYield);
      }
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
      expect(await yieldManager.getYieldProviderYieldReportedCumulative(yieldProviderAddress)).eq(0n);

      // Second positive yield of 3 ETH
      // +1 ETH in Vault relative to start, with -2 ETH of obligations
      const secondPositiveYield = ONE_ETHER * 3n;
      await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) + secondPositiveYield);
      await incrementBalance(mockStakingVaultAddress, secondPositiveYield);

      {
        const [newReportedYield, outstandingNegativeYield] = await yieldManager
          .connect(nativeYieldOperator)
          .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
        expect(newReportedYield).eq(0);
        expect(outstandingNegativeYield).eq(ONE_ETHER);
      }
      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
      expect(await yieldManager.getYieldProviderYieldReportedCumulative(yieldProviderAddress)).eq(0n);
      await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) - firstObligationsPaid);

      // Start with -1 ETH
      // Third positive yield of 3 ETH
      // -2 ETH of obligations, -2 ETH of node operators fees
      // Expect end with -2 ETH
      const thirdPositiveYield = ONE_ETHER * 3n;
      await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) + thirdPositiveYield);
      await incrementBalance(mockStakingVaultAddress, thirdPositiveYield);
      await mockVaultHub.setIsSettleVaultObligationsWithdrawingFromVault(true);
      await mockVaultHub.setSettleVaultObligationAmount(ONE_ETHER * 2n);
      await mockDashboard.setIsDisburseNodeOperatorFeeWithdrawingFromVault(true);
      await mockDashboard.setNodeOperatorDisbursableFeeReturn(ONE_ETHER * 2n);

      {
        const [newReportedYield, outstandingNegativeYield] = await yieldManager
          .connect(nativeYieldOperator)
          .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
        expect(newReportedYield).eq(0);
        // Obligations was paid, node operator fees was not
        expect(outstandingNegativeYield).eq(0n);
      }

      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
      await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) - ONE_ETHER * 4n);
      await mockVaultHub.setSettleVaultObligationAmount(0n);
      expect(await yieldManager.getYieldProviderYieldReportedCumulative(yieldProviderAddress)).eq(0n);

      // Start with -2 ETH, and -2 ETH of node operator fees owing
      // Fourth positive yield of 7 ETH
      // LST liability of -2 ETH
      // Expect end of +1 ETH
      await mockSTETH.setSharesByPooledEthReturn(ONE_ETHER * 2n);
      await mockDashboard.setLiabilitySharesReturn(ONE_ETHER * 2n);
      await mockDashboard.setRebalanceVaultWithSharesWithdrawingFromVault(true);
      const fourthPositiveYield = ONE_ETHER * 7n;
      await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) + fourthPositiveYield);
      await incrementBalance(mockStakingVaultAddress, fourthPositiveYield);

      {
        const [newReportedYield, outstandingNegativeYield] = await yieldManager
          .connect(nativeYieldOperator)
          .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
        expect(newReportedYield).eq(ONE_ETHER);
        // Obligations was paid, node operator fees was not
        expect(outstandingNegativeYield).eq(0n);
      }

      await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
      await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) - ONE_ETHER * 4n);
      await mockSTETH.setSharesByPooledEthReturn(0);
      await mockDashboard.setLiabilitySharesReturn(0);
      await mockDashboard.setNodeOperatorDisbursableFeeReturn(0n);
      expect(await yieldManager.getYieldProviderYieldReportedCumulative(yieldProviderAddress)).eq(ONE_ETHER);
    });
  });
});
