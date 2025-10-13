// Test scenarios with LineaRollup + YieldManager + LidoStVaultYieldProvider
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { getAccountsFixture } from "../../common/helpers";
import { deployYieldManagerIntegrationTestFixture } from "../helpers";
import { TestYieldManager, TestLineaRollup } from "contracts/typechain-types";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { ONE_ETHER } from "../../common/constants";
// import { generateLidoUnstakePermissionlessWitness } from "../helpers/proof";

describe("Integration tests with LineaRollup, YieldManager and LidoStVaultYieldProvider", () => {
  let nativeYieldOperator: SignerWithAddress;
  let lineaRollup: TestLineaRollup;
  let yieldManager: TestYieldManager;

  let l1MessageServiceAddress: string;
  let yieldProviderAddress: string;
  before(async () => {
    ({ nativeYieldOperator } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ lineaRollup, yieldProviderAddress, yieldManager } = await loadFixture(deployYieldManagerIntegrationTestFixture));
    l1MessageServiceAddress = await lineaRollup.getAddress();
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
});
