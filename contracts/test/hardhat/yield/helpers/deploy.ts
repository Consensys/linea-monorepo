import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers, upgrades } from "hardhat";

import {
  LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  YIELD_MANAGER_PAUSE_TYPES_ROLES,
  YIELD_MANAGER_UNPAUSE_TYPES_ROLES,
} from "contracts/common/constants";
import {
  MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
  TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
  MINIMUM_WITHDRAWAL_RESERVE_AMOUNT,
  TARGET_WITHDRAWAL_RESERVE_AMOUNT,
  GI_FIRST_VALIDATOR,
  GI_FIRST_VALIDATOR_AFTER_CHANGE,
  CHANGE_SLOT,
  LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR,
  FUNDER_ROLE,
} from "../../common/constants";
import { generateRoleAssignments } from "contracts/common/helpers";
import { YIELD_MANAGER_OPERATOR_ROLES, YIELD_MANAGER_SECURITY_COUNCIL_ROLES } from "contracts/common/constants";
import { YIELD_MANAGER_INITIALIZE_SIGNATURE } from "contracts/common/constants";
import { deployUpgradableWithConstructorArgs } from "../../common/deployment";
import {
  TestYieldManager,
  MockLineaRollup,
  MockYieldProvider,
  MockWithdrawTarget,
  MockVaultHub,
  MockSTETH,
  MockDashboard,
  MockStakingVault,
  TestLidoStVaultYieldProvider,
  TestCLProofVerifier,
  SSZMerkleTree,
} from "contracts/typechain-types";
import { YieldManagerInitializationData, YieldProviderRegistration } from "./types";

import { getAccountsFixture } from "../../common/helpers";
import { deployLineaRollupFixture, reinitializeLineaRollupFixtureV7 } from "../../rollup/helpers";
import { LineaRollupV7ReinitializationData } from "../../rollup/helpers/types";

async function getYieldManagerRoleAddressesFixture(): Promise<
  {
    role: string;
    addressWithRole: string;
  }[]
> {
  const { nativeYieldOperator, securityCouncil } = await loadFixture(getAccountsFixture);
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
  return [...yieldManagerOpereratorRoleAssignments, ...securityCouncilRoleAssignments];
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

export async function deployTestCLProofVerifier(): Promise<TestCLProofVerifier> {
  const factory = await ethers.getContractFactory("TestCLProofVerifier");
  const contract = await factory.deploy(GI_FIRST_VALIDATOR, GI_FIRST_VALIDATOR_AFTER_CHANGE, CHANGE_SLOT);
  await contract.waitForDeployment();
  return contract;
}

export async function deploySSZMerkleTree(): Promise<SSZMerkleTree> {
  const factory = await ethers.getContractFactory("SSZMerkleTree");
  const contract = await factory.deploy(GI_FIRST_VALIDATOR);
  await contract.waitForDeployment();
  return contract;
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

export async function deployMockDashboard(): Promise<MockDashboard> {
  const factory = await ethers.getContractFactory("MockDashboard");
  const contract = await factory.deploy();
  await contract.waitForDeployment();
  return contract;
}

export async function deployMockStakingVault(): Promise<MockStakingVault> {
  const factory = await ethers.getContractFactory("MockStakingVault");
  const contract = await factory.deploy();
  await contract.waitForDeployment();
  return contract;
}

export async function deployLidoStVaultYieldProviderFactory() {
  const { mockLineaRollup, yieldManager } = await loadFixture(deployYieldManagerForUnitTest);
  const mockVaultHub = await deployMockVaultHub();
  const mockSTETH = await deployMockSTETH();

  const l1MessageServiceAddress = await mockLineaRollup.getAddress();
  const yieldManagerAddress = await yieldManager.getAddress();
  const mockVaultHubAddress = await mockVaultHub.getAddress();
  const mockSTETHAddress = await mockSTETH.getAddress();

  const yieldProviderFactoryFactory = await ethers.getContractFactory("LidoStVaultYieldProviderFactory");
  const lidoStVaultYieldProviderFactory = await yieldProviderFactoryFactory.deploy(
    l1MessageServiceAddress,
    yieldManagerAddress,
    mockVaultHubAddress,
    mockSTETHAddress,
    GI_FIRST_VALIDATOR,
    GI_FIRST_VALIDATOR_AFTER_CHANGE,
    CHANGE_SLOT,
  );
  await lidoStVaultYieldProviderFactory.waitForDeployment();

  return { mockLineaRollup, yieldManager, mockVaultHub, mockSTETH, lidoStVaultYieldProviderFactory };
}

async function deployLidoStVaultYieldProviderDependenciesFixture() {
  const { securityCouncil } = await getAccountsFixture();
  const { mockLineaRollup, yieldManager } = await deployYieldManagerForUnitTest();
  const mockVaultHub = await deployMockVaultHub();
  const mockSTETH = await deployMockSTETH();
  const mockDashboard = await deployMockDashboard();
  const mockStakingVault = await deployMockStakingVault();
  const sszMerkleTree = await deploySSZMerkleTree();
  const verifier = await deployTestCLProofVerifier();

  return {
    securityCouncil,
    mockLineaRollup,
    yieldManager,
    mockVaultHub,
    mockSTETH,
    mockDashboard,
    mockStakingVault,
    sszMerkleTree,
    verifier,
  };
}

export async function deployAndAddSingleLidoStVaultYieldProvider() {
  const {
    securityCouncil,
    mockLineaRollup,
    yieldManager,
    mockVaultHub,
    mockSTETH,
    mockDashboard,
    mockStakingVault,
    sszMerkleTree,
    verifier,
  } = await loadFixture(deployLidoStVaultYieldProviderDependenciesFixture);

  const l1MessageServiceAddress = await mockLineaRollup.getAddress();
  const yieldManagerAddress = await yieldManager.getAddress();
  const mockVaultHubAddress = await mockVaultHub.getAddress();
  const mockSTETHAddress = await mockSTETH.getAddress();

  // Deploy Factory
  const yieldProviderFactoryFactory = await ethers.getContractFactory("TestLidoStVaultYieldProviderFactory");
  const lidoStVaultYieldProviderFactory = await yieldProviderFactoryFactory.deploy(
    l1MessageServiceAddress,
    yieldManagerAddress,
    mockVaultHubAddress,
    mockSTETHAddress,
    GI_FIRST_VALIDATOR,
    GI_FIRST_VALIDATOR_AFTER_CHANGE,
    CHANGE_SLOT,
  );
  await lidoStVaultYieldProviderFactory.waitForDeployment();

  // Create YieldProvider
  const yieldProviderAddress = await lidoStVaultYieldProviderFactory.createTestLidoStVaultYieldProvider.staticCall();
  await lidoStVaultYieldProviderFactory.connect(securityCouncil).createTestLidoStVaultYieldProvider();
  const yieldProvider: TestLidoStVaultYieldProvider = (
    await ethers.getContractFactory("TestLidoStVaultYieldProvider")
  ).attach(yieldProviderAddress);

  // Add YieldProvider
  await mockDashboard.setStakingVaultReturn(await mockStakingVault.getAddress());
  const mockDashboardAddress = await mockDashboard.getAddress();
  const mockStakingVaultAddress = await mockStakingVault.getAddress();

  const registration: YieldProviderRegistration = {
    yieldProviderVendor: LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR,
    primaryEntrypoint: mockDashboardAddress,
    ossifiedEntrypoint: mockStakingVaultAddress,
    receiveCaller: mockStakingVaultAddress,
  };

  await yieldManager.connect(securityCouncil).addYieldProvider(yieldProviderAddress, registration);
  return {
    mockDashboard,
    mockStakingVault,
    yieldManager,
    yieldProvider,
    yieldProviderAddress,
    mockVaultHub,
    mockSTETH,
    mockLineaRollup,
    sszMerkleTree,
    verifier,
  };
}

export async function deployYieldManagerIntegrationTestFixture() {
  const { securityCouncil, l2YieldRecipient } = await loadFixture(getAccountsFixture);
  const yieldManagerRoleAddresses = await loadFixture(getYieldManagerRoleAddressesFixture);
  // Deploy LineaRollup
  const { lineaRollup, roleAddresses: lineaRollupRoleAddresses } = await loadFixture(deployLineaRollupFixture);
  const l1MessageServiceAddress = await lineaRollup.getAddress();

  // Deploy YieldManager
  const initializationData: YieldManagerInitializationData = {
    pauseTypeRoles: YIELD_MANAGER_PAUSE_TYPES_ROLES,
    unpauseTypeRoles: YIELD_MANAGER_UNPAUSE_TYPES_ROLES,
    roleAddresses: yieldManagerRoleAddresses,
    initialL2YieldRecipients: [l2YieldRecipient.address],
    defaultAdmin: securityCouncil.address,
    initialMinimumWithdrawalReservePercentageBps: MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
    initialTargetWithdrawalReservePercentageBps: TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS,
    initialMinimumWithdrawalReserveAmount: MINIMUM_WITHDRAWAL_RESERVE_AMOUNT,
    initialTargetWithdrawalReserveAmount: TARGET_WITHDRAWAL_RESERVE_AMOUNT,
  };

  upgrades.silenceWarnings();
  const yieldManager = (await deployUpgradableWithConstructorArgs(
    "TestYieldManager",
    [l1MessageServiceAddress],
    [initializationData],
    {
      initializer: YIELD_MANAGER_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor", "incorrect-initializer-order", "state-variable-immutable", "delegatecall"],
    },
  )) as unknown as TestYieldManager;

  // Deploy LidoStVaultYieldProviderFactory
  const mockVaultHub = await deployMockVaultHub();
  const mockSTETH = await deployMockSTETH();
  const mockDashboard = await deployMockDashboard();
  const mockStakingVault = await deployMockStakingVault();

  const yieldManagerAddress = await yieldManager.getAddress();
  const mockVaultHubAddress = await mockVaultHub.getAddress();
  const mockSTETHAddress = await mockSTETH.getAddress();

  const yieldProviderFactoryFactory = await ethers.getContractFactory("TestLidoStVaultYieldProviderFactory");
  const lidoStVaultYieldProviderFactory = await yieldProviderFactoryFactory.deploy(
    l1MessageServiceAddress,
    yieldManagerAddress,
    mockVaultHubAddress,
    mockSTETHAddress,
    GI_FIRST_VALIDATOR,
    GI_FIRST_VALIDATOR_AFTER_CHANGE,
    CHANGE_SLOT,
  );
  await lidoStVaultYieldProviderFactory.waitForDeployment();

  // Create YieldProvider
  const yieldProviderAddress = await lidoStVaultYieldProviderFactory.createTestLidoStVaultYieldProvider.staticCall();
  await lidoStVaultYieldProviderFactory.connect(securityCouncil).createTestLidoStVaultYieldProvider();
  const yieldProvider = await ethers.getContractAt("TestLidoStVaultYieldProvider", yieldProviderAddress);

  // Add YieldProvider
  await mockDashboard.setStakingVaultReturn(await mockStakingVault.getAddress());
  const mockDashboardAddress = await mockDashboard.getAddress();
  const mockStakingVaultAddress = await mockStakingVault.getAddress();

  const registration: YieldProviderRegistration = {
    yieldProviderVendor: LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR,
    primaryEntrypoint: mockDashboardAddress,
    ossifiedEntrypoint: mockStakingVaultAddress,
    receiveCaller: mockStakingVaultAddress,
  };

  await yieldManager.connect(securityCouncil).addYieldProvider(yieldProviderAddress, registration);

  // Reinit LineaRollup with actual YieldManager
  const reinitData: LineaRollupV7ReinitializationData = {
    pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
    unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
    roleAddresses: [
      ...lineaRollupRoleAddresses,
      {
        role: FUNDER_ROLE,
        addressWithRole: yieldManagerAddress,
      },
    ],
    yieldManager: yieldManagerAddress,
  };
  await reinitializeLineaRollupFixtureV7(lineaRollup, reinitData);

  return {
    lineaRollup,
    yieldManager,
    yieldManagerAddress,
    yieldProvider,
    yieldProviderAddress,
    mockDashboard,
    mockSTETH,
    mockVaultHub,
    mockStakingVault,
  };
}
