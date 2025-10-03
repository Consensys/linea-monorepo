// TODO rename to LineaRollupYieldExtension
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

import { MockLineaRollup, TestYieldManager } from "contracts/typechain-types";
import { deployYieldManagerForUnitTest, deployYieldManagerForUnitTestWithMutatedInitData } from "./helpers/deploy";
import { MINIMUM_FEE, EMPTY_CALLDATA } from "../common/constants";
import {
  // expectEvent,
  // buildAccessErrorMessage,
  expectRevertWithCustomError,
  // expectRevertWithReason,
  getAccountsFixture,
} from "../common/helpers";
import { YieldManagerInitializationData } from "./helpers/types";
import { ZeroAddress } from "ethers";

describe("Linea Rollup contract", () => {
  let yieldManager: TestYieldManager;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let securityCouncil: SignerWithAddress;
  // let nonAuthorizedAccount: SignerWithAddress;
  let nativeYieldOperator: SignerWithAddress;
  let operationalSafe: SignerWithAddress;
  let mockLineaRollup: MockLineaRollup;
  let initializationData: YieldManagerInitializationData;

  before(async () => {
    ({ securityCouncil, nativeYieldOperator, operationalSafe } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ yieldManager, initializationData, mockLineaRollup } = await loadFixture(deployYieldManagerForUnitTest));
  });

  describe("Fallback/Receive tests", () => {
    const sendEthToContract = async (data: string) => {
      return nativeYieldOperator.sendTransaction({
        to: await yieldManager.getAddress(),
        value: MINIMUM_FEE,
        data,
      });
    };

    it("Should fail to send eth to the yieldManager contract through the fallback function", async () => {
      await expect(sendEthToContract("0x1234")).to.be.reverted;
    });

    it("Should fail to send eth to the yieldManager contract through the receive caller when unexpected", async () => {
      await expectRevertWithCustomError(yieldManager, sendEthToContract(EMPTY_CALLDATA), "UnexpectedReceiveCaller");
    });

    it("Should keep rejecting receive calls when the transient caller is not set within the same transaction", async () => {
      await yieldManager.setTransientReceiveCaller(nativeYieldOperator.address);
      await expectRevertWithCustomError(yieldManager, sendEthToContract(EMPTY_CALLDATA), "UnexpectedReceiveCaller");
    });
  });

  describe("Initialisation", () => {
    const cloneInitializationData = (
      overrides: Partial<YieldManagerInitializationData> = {},
    ): YieldManagerInitializationData => ({
      pauseTypeRoles: initializationData.pauseTypeRoles.map((role) => ({ ...role })),
      unpauseTypeRoles: initializationData.unpauseTypeRoles.map((role) => ({ ...role })),
      roleAddresses: initializationData.roleAddresses.map((role) => ({ ...role })),
      initialL2YieldRecipients: [...initializationData.initialL2YieldRecipients],
      defaultAdmin: initializationData.defaultAdmin,
      initialMinimumWithdrawalReservePercentageBps: initializationData.initialMinimumWithdrawalReservePercentageBps,
      initialTargetWithdrawalReservePercentageBps: initializationData.initialTargetWithdrawalReservePercentageBps,
      initialMinimumWithdrawalReserveAmount: initializationData.initialMinimumWithdrawalReserveAmount,
      initialTargetWithdrawalReserveAmount: initializationData.initialTargetWithdrawalReserveAmount,
      ...overrides,
    });

    it("Should revert if the default admin address is zero address", async () => {
      const mutatedInitData = cloneInitializationData({ defaultAdmin: ZeroAddress });
      await expectRevertWithCustomError(
        yieldManager,
        deployYieldManagerForUnitTestWithMutatedInitData(mutatedInitData),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should revert if a 0 address is provided for an l2YieldRecipient", async () => {
      const mutatedInitData = cloneInitializationData({ initialL2YieldRecipients: [ZeroAddress] });
      await expectRevertWithCustomError(
        yieldManager,
        deployYieldManagerForUnitTestWithMutatedInitData(mutatedInitData),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should revert if >10000 bps is provided for initialTargetWithdrawalReservePercentageBps", async () => {
      const mutatedInitData = cloneInitializationData({ initialTargetWithdrawalReservePercentageBps: 10001 });
      await expectRevertWithCustomError(
        yieldManager,
        deployYieldManagerForUnitTestWithMutatedInitData(mutatedInitData),
        "BpsMoreThan10000",
      );
    });

    it("Should revert if >10000 bps is provided for initialMinimumWithdrawalReservePercentageBps", async () => {
      const mutatedInitData = cloneInitializationData({ initialMinimumWithdrawalReservePercentageBps: 10001 });
      await expectRevertWithCustomError(
        yieldManager,
        deployYieldManagerForUnitTestWithMutatedInitData(mutatedInitData),
        "BpsMoreThan10000",
      );
    });

    it("Should revert if initialMinimumWithdrawalReservePercentageBps > initialTargetWithdrawalReservePercentageBps", async () => {
      const mutatedInitData = cloneInitializationData({
        initialTargetWithdrawalReservePercentageBps: initializationData.initialMinimumWithdrawalReservePercentageBps,
        initialMinimumWithdrawalReservePercentageBps: initializationData.initialTargetWithdrawalReservePercentageBps,
      });
      await expectRevertWithCustomError(
        yieldManager,
        deployYieldManagerForUnitTestWithMutatedInitData(mutatedInitData),
        "TargetReservePercentageMustBeAboveMinimum",
      );
    });

    it("Should revert if initialMinimumWithdrawalReserveAmount > initialTargetWithdrawalReserveAmount", async () => {
      const mutatedInitData = cloneInitializationData({
        initialTargetWithdrawalReserveAmount: initializationData.initialMinimumWithdrawalReserveAmount,
        initialMinimumWithdrawalReserveAmount: initializationData.initialTargetWithdrawalReserveAmount,
      });
      await expectRevertWithCustomError(
        yieldManager,
        deployYieldManagerForUnitTestWithMutatedInitData(mutatedInitData),
        "TargetReserveAmountMustBeAboveMinimum",
      );
    });

    it("Should revert if the initialize function is called a second time", async () => {
      await expect(yieldManager.initialize(cloneInitializationData())).to.be.revertedWith(
        "Initializable: contract is already initialized",
      );
    });

    it("Should have correct L1_MESSAGE_SERVICE", async () => {
      const l1MessageServiceAddress = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [
        l1MessageServiceAddress,
        ethers.toBeHex(ethers.parseEther("1")),
      ]);
      await ethers.provider.send("hardhat_impersonateAccount", [l1MessageServiceAddress]);
      const l1MessageServiceSigner = await ethers.getSigner(l1MessageServiceAddress);

      await expect(yieldManager.connect(l1MessageServiceSigner).receiveFundsFromReserve({ value: 0 })).to.not.be
        .reverted;

      await ethers.provider.send("hardhat_stopImpersonatingAccount", [l1MessageServiceAddress]);
    });

    it("Should assign the OSSIFIER_ROLE to securityCouncil address", async () => {
      const ossifierRole = await yieldManager.OSSIFIER_ROLE();
      expect(await yieldManager.hasRole(ossifierRole, securityCouncil.address)).to.be.true;
    });

    it("Should assign the PAUSE_ALL_ROLE to securityCouncil address", async () => {
      const pauseAllRole = await yieldManager.PAUSE_ALL_ROLE();
      expect(await yieldManager.hasRole(pauseAllRole, securityCouncil.address)).to.be.true;
    });

    it("Should assign the UNPAUSE_ALL_ROLE to securityCouncil address", async () => {
      const unpauseAllRole = await yieldManager.UNPAUSE_ALL_ROLE();
      expect(await yieldManager.hasRole(unpauseAllRole, securityCouncil.address)).to.be.true;
    });

    it("Should not assign the OSSIFIER_ROLE to nativeYieldOperator address", async () => {
      const ossifierRole = await yieldManager.OSSIFIER_ROLE();
      expect(await yieldManager.hasRole(ossifierRole, nativeYieldOperator.address)).to.be.false;
    });

    it("Should assign the YIELD_PROVIDER_FUNDER_ROLE to nativeYieldOperator address", async () => {
      const role = await yieldManager.YIELD_PROVIDER_FUNDER_ROLE();
      expect(await yieldManager.hasRole(role, nativeYieldOperator.address)).to.be.true;
    });

    it("Should assign the YIELD_PROVIDER_UNSTAKER_ROLE to nativeYieldOperator address", async () => {
      const role = await yieldManager.YIELD_PROVIDER_UNSTAKER_ROLE();
      expect(await yieldManager.hasRole(role, nativeYieldOperator.address)).to.be.true;
    });

    it("Should assign the YIELD_REPORTER_ROLE to nativeYieldOperator address", async () => {
      const role = await yieldManager.YIELD_REPORTER_ROLE();
      expect(await yieldManager.hasRole(role, nativeYieldOperator.address)).to.be.true;
    });

    it("Should assign the STAKING_PAUSER_ROLE to nativeYieldOperator address", async () => {
      const role = await yieldManager.STAKING_PAUSER_ROLE();
      expect(await yieldManager.hasRole(role, nativeYieldOperator.address)).to.be.true;
    });

    it("Should assign the STAKING_UNPAUSER_ROLE to nativeYieldOperator address", async () => {
      const role = await yieldManager.STAKING_UNPAUSER_ROLE();
      expect(await yieldManager.hasRole(role, nativeYieldOperator.address)).to.be.true;
    });

    it("Should assign the WITHDRAWAL_RESERVE_SETTER_ROLE to nativeYieldOperator address", async () => {
      const role = await yieldManager.WITHDRAWAL_RESERVE_SETTER_ROLE();
      expect(await yieldManager.hasRole(role, nativeYieldOperator.address)).to.be.true;
    });

    it("Should assign the SET_YIELD_PROVIDER_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.SET_YIELD_PROVIDER_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the SET_L2_YIELD_RECIPIENT_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.SET_L2_YIELD_RECIPIENT_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the PAUSE_NATIVE_YIELD_UNSTAKING_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.PAUSE_NATIVE_YIELD_UNSTAKING_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the UNPAUSE_NATIVE_YIELD_UNSTAKING_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.UNPAUSE_NATIVE_YIELD_UNSTAKING_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the PAUSE_NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.PAUSE_NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the UNPAUSE_NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.UNPAUSE_NATIVE_YIELD_PERMISSIONLESS_UNSTAKING_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the PAUSE_NATIVE_YIELD_PERMISSIONLESS_REBALANCE_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.PAUSE_NATIVE_YIELD_PERMISSIONLESS_REBALANCE_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the UNPAUSE_NATIVE_YIELD_PERMISSIONLESS_REBALANCE_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.UNPAUSE_NATIVE_YIELD_PERMISSIONLESS_REBALANCE_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the PAUSE_NATIVE_YIELD_RESERVE_FUNDING_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.PAUSE_NATIVE_YIELD_RESERVE_FUNDING_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the UNPAUSE_NATIVE_YIELD_RESERVE_FUNDING_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.UNPAUSE_NATIVE_YIELD_RESERVE_FUNDING_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the PAUSE_NATIVE_YIELD_REPORTING_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.PAUSE_NATIVE_YIELD_REPORTING_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });

    it("Should assign the UNPAUSE_NATIVE_YIELD_REPORTING_ROLE to operationalSafe address", async () => {
      const role = await yieldManager.UNPAUSE_NATIVE_YIELD_REPORTING_ROLE();
      expect(await yieldManager.hasRole(role, operationalSafe.address)).to.be.true;
    });
  });
});
