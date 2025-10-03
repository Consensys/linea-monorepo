import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";

import { YIELD_MANAGER_PAUSE_TYPES_ROLES, YIELD_MANAGER_UNPAUSE_TYPES_ROLES } from "contracts/common/constants";
import {
  MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
  TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
  MINIMUM_WITHDRAWAL_RESERVE_AMOUNT,
  TARGET_WITHDRAWAL_RESERVE_AMOUNT,
} from "../../common/constants";
import { generateRoleAssignments } from "contracts/common/helpers";
import { YIELD_MANAGER_ROLES } from "contracts/common/constants";
import { YIELD_MANAGER_INITIALIZE_SIGNATURE } from "contracts/common/constants";
import { deployUpgradableWithConstructorArgs } from "../../common/deployment";
import { TestYieldManager, MockLineaRollup, MockYieldProvider } from "contracts/typechain-types";

import { getAccountsFixture } from "../../common/helpers";

async function getYieldManagerRoleAddressesFixture() {
  const { nativeYieldOperator } = await loadFixture(getAccountsFixture);
  const roleAddresses = generateRoleAssignments(YIELD_MANAGER_ROLES, await nativeYieldOperator.getAddress(), []);
  return roleAddresses;
}

export async function deployMockLineaRollup(): Promise<MockLineaRollup> {
  const mockYieldManagerFactory = await ethers.getContractFactory("MockLineaRollup");
  const mockYieldManager = await mockYieldManagerFactory.deploy();
  await mockYieldManager.waitForDeployment();
  return await mockYieldManager;
}

export async function deployMockYieldProvider(): Promise<MockYieldProvider> {
  const mockYieldManagerFactory = await ethers.getContractFactory("MockYieldProvider");
  const mockYieldManager = await mockYieldManagerFactory.deploy();
  await mockYieldManager.waitForDeployment();
  return await mockYieldManager;
}

// Deploys with MockLineaRollup and MockYieldProvider
export async function deployYieldManagerForUnitTest() {
  const { securityCouncil, l2YieldRecipient } = await loadFixture(getAccountsFixture);
  const roleAddresses = await loadFixture(getYieldManagerRoleAddressesFixture);

  const mockLineaRollup = await deployMockLineaRollup();

  const initializationData = {
    pauseTypeRoles: YIELD_MANAGER_PAUSE_TYPES_ROLES,
    unpauseTypeRoles: YIELD_MANAGER_UNPAUSE_TYPES_ROLES,
    roleAddresses,
    initialL2YieldRecipients: [l2YieldRecipient.address],
    defaultAdmin: securityCouncil.address,
    initialMinimumWithdrawalReservePercentageBps: MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
    initialTargetWithdrawalReservePercentageBps: TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
    initialMinimumWithdrawalReserveAmount: MINIMUM_WITHDRAWAL_RESERVE_AMOUNT,
    initialTargetWithdrawalReserveAmount: TARGET_WITHDRAWAL_RESERVE_AMOUNT,
  };

  const yieldManager = (await deployUpgradableWithConstructorArgs(
    "TestYieldManager",
    [await mockLineaRollup.getAddress()],
    [initializationData],
    {
      // initializer: "initialize",
      initializer: YIELD_MANAGER_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor", "incorrect-initializer-order", "state-variable-immutable"],
    },
  )) as unknown as TestYieldManager;

  return { mockLineaRollup, yieldManager };
}
