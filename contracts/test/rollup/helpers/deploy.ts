import { ethers, upgrades } from "hardhat";
import firstCompressedDataContent from "../../testData/compressedData/blocks-1-46.json";
import {
  DEFAULT_LAST_FINALIZED_TIMESTAMP,
  INITIAL_WITHDRAW_LIMIT,
  LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  ONE_DAY_IN_SECONDS,
} from "../../common/constants";
import { TestLineaRollup } from "contracts/typechain-types";
import { deployUpgradableFromFactory } from "contracts/scripts/hardhat/utils";
import { LINEA_ROLLUP_PAUSE_TYPES_ROLES, LINEA_ROLLUP_UNPAUSE_TYPES_ROLES } from "contracts/common/constants";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { getAccountsFixture } from "contracts/test/common/helpers";
import { getRoleAddressesFixture } from "./before";
import { LineaRollupV7ReinitializationData } from "./types";
import { Contract } from "ethers";

export async function deployMockYieldManager(): Promise<string> {
  const mockYieldManagerFactory = await ethers.getContractFactory("MockYieldManager");
  const mockYieldManager = await mockYieldManagerFactory.deploy();
  await mockYieldManager.waitForDeployment();
  return await mockYieldManager.getAddress();
}

export async function deployLineaRollupFixture() {
  const { securityCouncil } = await loadFixture(getAccountsFixture);
  const plonkVerifierFactory = await ethers.getContractFactory("TestPlonkVerifierForDataAggregation");
  const plonkVerifier = await plonkVerifierFactory.deploy();
  await plonkVerifier.waitForDeployment();
  const roleAddresses = await loadFixture(getRoleAddressesFixture);

  const verifier = await plonkVerifier.getAddress();
  const { parentStateRootHash } = firstCompressedDataContent;

  const mockYieldManager = await deployMockYieldManager();

  const FALLBACK_OPERATOR_ADDRESS = "0xcA11bde05977b3631167028862bE2a173976CA11";
  const initializationData = {
    initialStateRootHash: parentStateRootHash,
    initialL2BlockNumber: 0,
    genesisTimestamp: DEFAULT_LAST_FINALIZED_TIMESTAMP,
    defaultVerifier: verifier,
    rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
    rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
    roleAddresses,
    pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
    unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
    initialYieldManager: mockYieldManager,
    fallbackOperator: FALLBACK_OPERATOR_ADDRESS,
    defaultAdmin: securityCouncil.address,
  };

  const lineaRollup = (await deployUpgradableFromFactory("TestLineaRollup", [initializationData], {
    initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
    unsafeAllow: ["constructor"],
  })) as unknown as TestLineaRollup;

  return { lineaRollup, mockYieldManager, roleAddresses };
}

export async function reinitializeLineaRollupFixtureV7(
  lineaRollup: TestLineaRollup,
  initData: LineaRollupV7ReinitializationData,
): Promise<Contract> {
  const rollupFactory = await ethers.getContractFactory("TestLineaRollup");

  const initArgs = [initData.roleAddresses, initData.pauseTypeRoles, initData.unpauseTypeRoles, initData.yieldManager];

  return upgrades.upgradeProxy(lineaRollup, rollupFactory, {
    kind: "transparent",
    call: { fn: "reinitializeLineaRollupV7", args: initArgs },
    unsafeAllow: ["incorrect-initializer-order"],
  });
}
