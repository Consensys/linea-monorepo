import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers, upgrades } from "hardhat";

import { YIELD_MANAGER_PAUSE_TYPES_ROLES, YIELD_MANAGER_UNPAUSE_TYPES_ROLES } from "contracts/common/constants";
import {
  MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
  TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
  MINIMUM_WITHDRAWAL_RESERVE_AMOUNT,
  TARGET_WITHDRAWAL_RESERVE_AMOUNT,
  GI_FIRST_VALIDATOR,
  GI_FIRST_VALIDATOR_AFTER_CHANGE,
  CHANGE_SLOT,
} from "../../common/constants";
import { generateRoleAssignments } from "contracts/common/helpers";
import {
  YIELD_MANAGER_OPERATOR_ROLES,
  YIELD_MANAGER_OPERATIONAL_SAFE_ROLES,
  YIELD_MANAGER_SECURITY_COUNCIL_ROLES,
} from "contracts/common/constants";
import { YIELD_MANAGER_INITIALIZE_SIGNATURE } from "contracts/common/constants";
import { deployUpgradableWithConstructorArgs } from "../../common/deployment";
import {
  TestYieldManager,
  MockLineaRollup,
  MockYieldProvider,
  MockWithdrawTarget,
  MockVaultHub,
  MockSTETH,
} from "contracts/typechain-types";
import { YieldManagerInitializationData } from "./types";

import { getAccountsFixture } from "../../common/helpers";

async function getYieldManagerRoleAddressesFixture(): Promise<
  {
    role: string;
    addressWithRole: string;
  }[]
> {
  const { nativeYieldOperator, securityCouncil, operationalSafe } = await loadFixture(getAccountsFixture);
  const yieldManagerOpereratorRoleAssignments = generateRoleAssignments(
    YIELD_MANAGER_OPERATOR_ROLES,
    await nativeYieldOperator.getAddress(),
    [],
  );
  const securityCouncilRoleAssignments = generateRoleAssignments(
    YIELD_MANAGER_SECURITY_COUNCIL_ROLES,
    await securityCouncil.getAddress(),
    [],
  );
  const operationalSafeRoleAssignments = generateRoleAssignments(
    YIELD_MANAGER_OPERATIONAL_SAFE_ROLES,
    await operationalSafe.getAddress(),
    [],
  );
  return [
    ...yieldManagerOpereratorRoleAssignments,
    ...securityCouncilRoleAssignments,
    ...operationalSafeRoleAssignments,
  ];
}

export async function deployMockLineaRollup(): Promise<MockLineaRollup> {
  const mockYieldManagerFactory = await ethers.getContractFactory("MockLineaRollup");
  const mockYieldManager = await mockYieldManagerFactory.deploy();
  await mockYieldManager.waitForDeployment();
  return await mockYieldManager;
}

export async function deployMockWithdrawTarget(): Promise<MockWithdrawTarget> {
  const factory = await ethers.getContractFactory("MockWithdrawTarget");
  const mockWithdrawTarget = await factory.deploy();
  await mockWithdrawTarget.waitForDeployment();
  return await mockWithdrawTarget;
}

export async function deployMockYieldProvider(): Promise<MockYieldProvider> {
  const mockYieldManagerFactory = await ethers.getContractFactory("MockYieldProvider");
  const mockYieldManager = await mockYieldManagerFactory.deploy();
  await mockYieldManager.waitForDeployment();
  return await mockYieldManager;
}

// Deploys with MockLineaRollup and MockYieldProvider
export async function deployYieldManagerForUnitTest() {
  upgrades.silenceWarnings();
  const { securityCouncil, l2YieldRecipient } = await loadFixture(getAccountsFixture);
  const roleAddresses = await loadFixture(getYieldManagerRoleAddressesFixture);

  const mockLineaRollup = await deployMockLineaRollup();

  const initializationData: YieldManagerInitializationData = {
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
      initializer: YIELD_MANAGER_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor", "incorrect-initializer-order", "state-variable-immutable", "delegatecall"],
    },
  )) as unknown as TestYieldManager;

  return { mockLineaRollup, yieldManager, initializationData };
}

export async function deployYieldManagerForUnitTestWithMutatedInitData(
  mutatedInitData: YieldManagerInitializationData,
) {
  upgrades.silenceWarnings();
  const mockLineaRollup = await deployMockLineaRollup();
  await deployUpgradableWithConstructorArgs(
    "TestYieldManager",
    [await mockLineaRollup.getAddress()],
    [mutatedInitData],
    {
      // initializer: "initialize",
      initializer: YIELD_MANAGER_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor", "incorrect-initializer-order", "state-variable-immutable", "delegatecall"],
    },
  );
}

export async function deployMockVaultHub(): Promise<MockVaultHub> {
  const factory = await ethers.getContractFactory("MockVaultHub");
  const contract = await factory.deploy();
  await contract.waitForDeployment();
  return contract;
}

export async function deployMockSTETH(): Promise<MockSTETH> {
  const factory = await ethers.getContractFactory("MockSTETH");
  const contract = await factory.deploy();
  await contract.waitForDeployment();
  return contract;
}

export async function deployLidoStVaultYieldProviderFactory() {
  const { mockLineaRollup, yieldManager } = await loadFixture(deployYieldManagerForUnitTest);
  const yieldProviderFactory = await ethers.getContractFactory("LidoStVaultYieldProvider");
  const mockVaultHub = await deployMockVaultHub();
  const mockSTETH = await deployMockSTETH();

  const l1MessageServiceAddress = await mockLineaRollup.getAddress();
  const yieldManagerAddress = await yieldManager.getAddress();
  const mockVaultHubAddress = await mockVaultHub.getAddress();
  const mockSTETHAddress = await mockSTETH.getAddress();

  const beacon = await upgrades.deployBeacon(yieldProviderFactory, {
    unsafeAllow: ["constructor", "incorrect-initializer-order", "state-variable-immutable", "delegatecall"],
    constructorArgs: [
      l1MessageServiceAddress,
      yieldManagerAddress,
      mockVaultHubAddress,
      mockSTETHAddress,
      GI_FIRST_VALIDATOR,
      GI_FIRST_VALIDATOR_AFTER_CHANGE,
      CHANGE_SLOT,
    ],
  });

  await beacon.waitForDeployment();
  const beaconAddress = await beacon.getAddress();

  const yieldProviderFactoryFactory = await ethers.getContractFactory("LidoStVaultYieldProviderFactory");
  const lidoStVaultYieldProviderFactory = await yieldProviderFactoryFactory.deploy(beaconAddress);
  await lidoStVaultYieldProviderFactory.waitForDeployment();

  return { mockLineaRollup, yieldManager, mockVaultHub, mockSTETH, beacon, lidoStVaultYieldProviderFactory };
}
