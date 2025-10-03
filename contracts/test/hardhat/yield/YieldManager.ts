// TODO rename to LineaRollupYieldExtension
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

import { MockLineaRollup, TestYieldManager } from "contracts/typechain-types";
import { deployYieldManagerForUnitTest, deployYieldManagerForUnitTestWithMutatedInitData } from "./helpers/deploy";
import { MINIMUM_FEE, EMPTY_CALLDATA, ONE_THOUSAND_ETHER } from "../common/constants";
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

  let securityCouncil: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  let nativeYieldOperator: SignerWithAddress;
  let operationalSafe: SignerWithAddress;
  let mockLineaRollup: MockLineaRollup;
  let initializationData: YieldManagerInitializationData;

  before(async () => {
    ({
      securityCouncil,
      operator: nonAuthorizedAccount,
      nativeYieldOperator,
      operationalSafe,
    } = await loadFixture(getAccountsFixture));
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

    it("Should have the initial l2YieldRecipients in state", async () => {
      const existingRecipient = initializationData.initialL2YieldRecipients[0];
      expect(await yieldManager.isL2YieldRecipientKnown(existingRecipient)).to.be.true;
    });

    it("Should have the initial minimum reserve percentage in state", async () => {
      expect(await yieldManager.minimumWithdrawalReservePercentageBps()).to.equal(
        BigInt(initializationData.initialMinimumWithdrawalReservePercentageBps),
      );
    });

    it("Should have the initial minimum reserve amount in state", async () => {
      expect(await yieldManager.minimumWithdrawalReserveAmount()).to.equal(
        initializationData.initialMinimumWithdrawalReserveAmount,
      );
    });

    it("Should have the initial target reserve percentage in state", async () => {
      expect(await yieldManager.targetWithdrawalReservePercentageBps()).to.equal(
        BigInt(initializationData.initialTargetWithdrawalReservePercentageBps),
      );
    });

    it("Should have the initial target reserve amount in state", async () => {
      expect(await yieldManager.targetWithdrawalReserveAmount()).to.equal(
        initializationData.initialTargetWithdrawalReserveAmount,
      );
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

  describe("Adding and removing L2YieldRecipients", () => {
    it("Should revert when adding if the caller does not have the SET_L2_YIELD_RECIPIENT_ROLE", async () => {
      const requiredRole = await yieldManager.SET_L2_YIELD_RECIPIENT_ROLE();
      await expect(
        yieldManager.connect(nonAuthorizedAccount).addL2YieldRecipient(nonAuthorizedAccount.address),
      ).to.be.revertedWith(
        `AccessControl: account ${nonAuthorizedAccount.address.toLowerCase()} is missing role ${requiredRole}`,
      );
    });

    it("Should add the new l2YieldRecipient address and emit the correct event", async () => {
      const newRecipient = ethers.Wallet.createRandom().address;
      const addTx = yieldManager.connect(operationalSafe).addL2YieldRecipient(newRecipient);

      await expect(addTx).to.emit(yieldManager, "L2YieldRecipientAdded").withArgs(newRecipient);

      expect(await yieldManager.isL2YieldRecipientKnown(newRecipient)).to.be.true;
    });

    it("Should revert if the address being added has already been added", async () => {
      const existingRecipient = initializationData.initialL2YieldRecipients[0];
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(operationalSafe).addL2YieldRecipient(existingRecipient),
        "L2YieldRecipientAlreadyAdded",
      );
    });

    it("Should revert if the address being removed is unknown", async () => {
      const unknownRecipient = ethers.Wallet.createRandom().address;
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager.connect(operationalSafe).removeL2YieldRecipient(unknownRecipient),
        "UnknownL2YieldRecipient",
      );
    });

    it("Should revert when removing if the caller does not have the SET_L2_YIELD_RECIPIENT_ROLE", async () => {
      const requiredRole = await yieldManager.SET_L2_YIELD_RECIPIENT_ROLE();
      const existingRecipient = initializationData.initialL2YieldRecipients[0];
      await expect(
        yieldManager.connect(nonAuthorizedAccount).removeL2YieldRecipient(existingRecipient),
      ).to.be.revertedWith(
        `AccessControl: account ${nonAuthorizedAccount.address.toLowerCase()} is missing role ${requiredRole}`,
      );
    });

    it("Should remove the new l2YieldRecipient address and emit the correct event", async () => {
      const recipientToRemove = initializationData.initialL2YieldRecipients[0];
      const removeTx = yieldManager.connect(operationalSafe).removeL2YieldRecipient(recipientToRemove);

      await expect(removeTx).to.emit(yieldManager, "L2YieldRecipientRemoved").withArgs(recipientToRemove);

      expect(await yieldManager.isL2YieldRecipientKnown(recipientToRemove)).to.be.false;
    });
  });

  describe("Setting withdrawal reserve parameters", () => {
    const buildSetWithdrawalReserveParams = (
      overrides: Partial<{ minPct: number; targetPct: number; minAmount: bigint; targetAmount: bigint }> = {},
    ) => ({
      minimumWithdrawalReservePercentageBps:
        overrides.minPct ?? initializationData.initialMinimumWithdrawalReservePercentageBps,
      targetWithdrawalReservePercentageBps:
        overrides.targetPct ?? initializationData.initialTargetWithdrawalReservePercentageBps,
      minimumWithdrawalReserveAmount: overrides.minAmount ?? initializationData.initialMinimumWithdrawalReserveAmount,
      targetWithdrawalReserveAmount: overrides.targetAmount ?? initializationData.initialTargetWithdrawalReserveAmount,
    });

    it("Should revert set withdrawal reserve parameters when the caller does not have the WITHDRAWAL_RESERVE_SETTER_ROLE role", async () => {
      const role = await yieldManager.WITHDRAWAL_RESERVE_SETTER_ROLE();
      await expect(
        yieldManager.connect(nonAuthorizedAccount).setWithdrawalReserveParameters(buildSetWithdrawalReserveParams()),
      ).to.be.revertedWith(
        `AccessControl: account ${nonAuthorizedAccount.address.toLowerCase()} is missing role ${role}`,
      );
    });

    it("Should revert if minimum withdrawal percentage higher than 10000 bps", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .setWithdrawalReserveParameters(buildSetWithdrawalReserveParams({ minPct: 10001 })),
        "BpsMoreThan10000",
      );
    });

    it("Should revert if target withdrawal percentage higher than 10000 bps", async () => {
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .setWithdrawalReserveParameters(buildSetWithdrawalReserveParams({ targetPct: 10001 })),
        "BpsMoreThan10000",
      );
    });

    it("Should revert if minimum withdrawal reserve percentage > target", async () => {
      const target = initializationData.initialTargetWithdrawalReservePercentageBps;
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .setWithdrawalReserveParameters(buildSetWithdrawalReserveParams({ minPct: target + 1, targetPct: target })),
        "TargetReservePercentageMustBeAboveMinimum",
      );
    });

    it("Should revert if minimum withdrawal amount > target", async () => {
      const target = initializationData.initialTargetWithdrawalReserveAmount;
      await expectRevertWithCustomError(
        yieldManager,
        yieldManager
          .connect(nativeYieldOperator)
          .setWithdrawalReserveParameters(
            buildSetWithdrawalReserveParams({ minAmount: target + 1n, targetAmount: target }),
          ),
        "TargetReserveAmountMustBeAboveMinimum",
      );
    });

    it("Should successfully set withdrawal reserve parameters and emit logs", async () => {
      const params = buildSetWithdrawalReserveParams({
        minPct: initializationData.initialMinimumWithdrawalReservePercentageBps + 1,
        targetPct: initializationData.initialTargetWithdrawalReservePercentageBps + 2,
        minAmount: initializationData.initialMinimumWithdrawalReserveAmount + 5n,
        targetAmount: initializationData.initialTargetWithdrawalReserveAmount + 10n,
      });

      const prevMinPct = await yieldManager.minimumWithdrawalReservePercentageBps();
      const prevMinAmount = await yieldManager.minimumWithdrawalReserveAmount();
      const prevTargetPct = await yieldManager.targetWithdrawalReservePercentageBps();
      const prevTargetAmount = await yieldManager.targetWithdrawalReserveAmount();

      const tx = yieldManager.connect(nativeYieldOperator).setWithdrawalReserveParameters(params);

      await expect(tx)
        .to.emit(yieldManager, "WithdrawalReserveParametersSet")
        .withArgs(
          prevMinPct,
          params.minimumWithdrawalReservePercentageBps,
          prevMinAmount,
          params.minimumWithdrawalReserveAmount,
          prevTargetPct,
          params.targetWithdrawalReservePercentageBps,
          prevTargetAmount,
          params.targetWithdrawalReserveAmount,
        );

      await tx;

      expect(await yieldManager.minimumWithdrawalReservePercentageBps()).to.equal(
        BigInt(params.minimumWithdrawalReservePercentageBps),
      );
      expect(await yieldManager.minimumWithdrawalReserveAmount()).to.equal(params.minimumWithdrawalReserveAmount);
      expect(await yieldManager.targetWithdrawalReservePercentageBps()).to.equal(
        BigInt(params.targetWithdrawalReservePercentageBps),
      );
      expect(await yieldManager.targetWithdrawalReserveAmount()).to.equal(params.targetWithdrawalReserveAmount);
    });
  });

  describe("getTotalSystemBalance", () => {
    const ONE_HUNDRED_FIFTY_ETHER = ethers.parseEther("150");

    it("Return correct value with 1000 ETH on L1MessageService only", async () => {
      const l1MessageServiceAddress = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [l1MessageServiceAddress, ethers.toBeHex(ONE_THOUSAND_ETHER)]);

      const total = await yieldManager.getTotalSystemBalance();
      expect(total).to.equal(ONE_THOUSAND_ETHER);
    });

    it("Return correct value with 1000 ETH on L1MessageService and 150 ETH on YieldManager", async () => {
      const l1MessageServiceAddress = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [l1MessageServiceAddress, ethers.toBeHex(ONE_THOUSAND_ETHER)]);
      const yieldManagerAddress = await yieldManager.getAddress();
      await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(ONE_HUNDRED_FIFTY_ETHER)]);

      const total = await yieldManager.getTotalSystemBalance();
      expect(total).to.equal(ONE_THOUSAND_ETHER + ONE_HUNDRED_FIFTY_ETHER);
    });
  });

  describe("getEffectiveMinimumWithdrawalReserve", () => {
    it("With 0 balance in total system, should return minimum amount", async () => {
      const l1MessageServiceAddress = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [l1MessageServiceAddress, "0x0"]);
      await ethers.provider.send("hardhat_setBalance", [await yieldManager.getAddress(), "0x0"]);
      expect(await yieldManager.getEffectiveMinimumWithdrawalReserve()).to.equal(
        await yieldManager.minimumWithdrawalReserveAmount(),
      );
    });

    it("When total system balance > MINIMUM_AMOUNT / MINIMUM_WITHDRAWAL_PERCENTAGE, should return value dictated by MINIMUM_WITHDRAWAL_PERCENTAGE", async () => {
      // Arrange
      const minPct = initializationData.initialMinimumWithdrawalReservePercentageBps;
      const minAmount = initializationData.initialMinimumWithdrawalReserveAmount;
      const SYSTEM_BALANCE_WITH_EFFECTIVE_MINIMUM_ABOVE_AMOUNT = (minAmount * 10000n) / BigInt(minPct);
      const l1MessageServiceAddress = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [
        l1MessageServiceAddress,
        ethers.toBeHex(SYSTEM_BALANCE_WITH_EFFECTIVE_MINIMUM_ABOVE_AMOUNT),
      ]);
      await ethers.provider.send("hardhat_setBalance", [
        await yieldManager.getAddress(),
        ethers.toBeHex(SYSTEM_BALANCE_WITH_EFFECTIVE_MINIMUM_ABOVE_AMOUNT),
      ]);

      // Assert
      expect(await yieldManager.getEffectiveMinimumWithdrawalReserve()).to.be.above(
        await yieldManager.minimumWithdrawalReserveAmount(),
      );
      expect(await yieldManager.getEffectiveMinimumWithdrawalReserve()).to.equal(
        ((await yieldManager.getTotalSystemBalance()) * (await yieldManager.minimumWithdrawalReservePercentageBps())) /
          10000n,
      );
    });
  });

  describe("getEffectiveTargetWithdrawalReserve", () => {
    it("With 0 balance in total system, should return target amount", async () => {
      const l1MessageServiceAddress = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [l1MessageServiceAddress, "0x0"]);
      await ethers.provider.send("hardhat_setBalance", [await yieldManager.getAddress(), "0x0"]);
      expect(await yieldManager.getEffectiveTargetWithdrawalReserve()).to.equal(
        await yieldManager.targetWithdrawalReserveAmount(),
      );
    });

    it("When total system balance > TARGET_AMOUNT / TARGET_WITHDRAWAL_PERCENTAGE, should return value dictated by TARGET_WITHDRAWAL_PERCENTAGE", async () => {
      // Arrange
      const targetPct = initializationData.initialTargetWithdrawalReservePercentageBps;
      const targetAmount = initializationData.initialTargetWithdrawalReserveAmount;
      const SYSTEM_BALANCE_WITH_EFFECTIVE_TARGET_ABOVE_AMOUNT = (targetAmount * 10000n) / BigInt(targetPct);
      const l1MessageServiceAddress = await mockLineaRollup.getAddress();
      await ethers.provider.send("hardhat_setBalance", [
        l1MessageServiceAddress,
        ethers.toBeHex(SYSTEM_BALANCE_WITH_EFFECTIVE_TARGET_ABOVE_AMOUNT),
      ]);
      await ethers.provider.send("hardhat_setBalance", [
        await yieldManager.getAddress(),
        ethers.toBeHex(SYSTEM_BALANCE_WITH_EFFECTIVE_TARGET_ABOVE_AMOUNT),
      ]);

      // Assert
      expect(await yieldManager.getEffectiveTargetWithdrawalReserve()).to.be.above(
        await yieldManager.targetWithdrawalReserveAmount(),
      );
      expect(await yieldManager.getEffectiveTargetWithdrawalReserve()).to.equal(
        ((await yieldManager.getTotalSystemBalance()) * (await yieldManager.targetWithdrawalReservePercentageBps())) /
          10000n,
      );
    });
  });
});
