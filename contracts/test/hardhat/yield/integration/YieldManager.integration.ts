// Test scenarios with LineaRollup + YieldManager + LidoStVaultYieldProvider
import { loadFixture, setBalance } from "@nomicfoundation/hardhat-network-helpers";
import { expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
import {
  deployYieldManagerIntegrationTestFixture,
  fundLidoStVaultYieldProvider,
  incrementBalance,
  setWithdrawalReserveToMinimum,
} from "../helpers";
import { TestYieldManager, TestLineaRollup, TestLidoStVaultYieldProvider } from "contracts/typechain-types";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { ethers } from "hardhat";
import {
  EMPTY_CALLDATA,
  MESSAGE_FEE,
  MESSAGE_VALUE_1ETH,
  ONE_ETHER,
  VALID_MERKLE_PROOF,
  ZERO_VALUE,
} from "../../common/constants";
import { ZeroAddress } from "ethers";
// import { generateLidoUnstakePermissionlessWitness } from "../helpers/proof";

describe("Integration tests with LineaRollup, YieldManager and LidoStVaultYieldProvider", () => {
  let nativeYieldOperator: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  let lineaRollup: TestLineaRollup;
  let yieldManager: TestYieldManager;
  let yieldProvider: TestLidoStVaultYieldProvider;

  let l1MessageServiceAddress: string;
  let yieldManagerAddress: string;
  let yieldProviderAddress: string;

  before(async () => {
    ({ nativeYieldOperator, nonAuthorizedAccount } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ lineaRollup, yieldProvider, yieldProviderAddress, yieldManager } = await loadFixture(
      deployYieldManagerIntegrationTestFixture,
    ));
    l1MessageServiceAddress = await lineaRollup.getAddress();
    yieldManagerAddress = await yieldManager.getAddress();
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

  describe("transferFundsForNativeYield", () => {
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
      await lineaRollup.addL2MerkleRoots([VALID_MERKLE_PROOF.merkleRoot], VALID_MERKLE_PROOF.proof.length);

      // const withdrawAmount = ONE_ETHER / 2n;
      const claimParams = {
        proof: VALID_MERKLE_PROOF.proof,
        messageNumber: 1n,
        leafIndex: VALID_MERKLE_PROOF.index,
        from: nonAuthorizedAccount.address,
        to: nonAuthorizedAccount.address,
        fee: MESSAGE_FEE,
        value: MESSAGE_FEE + MESSAGE_VALUE_1ETH,
        feeRecipient: ZeroAddress,
        merkleRoot: VALID_MERKLE_PROOF.merkleRoot,
        data: EMPTY_CALLDATA,
      };

      console.log(claimParams);

      // Act
      // const claimCall = lineaRollup
      //   .connect(nonAuthorizedAccount)
      //   .claimMessageWithProofAndWithdrawLST(claimParams, yieldProviderAddress);
      // await claimCall;
    });
  });
});
