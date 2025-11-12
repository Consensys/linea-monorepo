import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { deployUpgradableFromFactory } from "../scripts/hardhat/utils";
import {
  generateRoleAssignments,
  getEnvVarOrDefault,
  getRequiredEnvVar,
  tryVerifyContract,
  getDeployedContractAddress,
  tryStoreAddress,
  LogContractDeployment,
} from "../common/helpers";
import {
  DEAD_ADDRESS,
  LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  LINEA_ROLLUP_ROLES,
  LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
} from "../common/constants";
import { ZeroHash } from "ethers";

// Deploy script for TestLineaRollup on Hoodi for Native Yield testing.
// Don't need any finalization functionality.
const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments } = hre;

  const contractName = "TestLineaRollup";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);

  // TestLineaRollup DEPLOYED AS UPGRADEABLE PROXY
  const lineaRollupSecurityCouncil = getRequiredEnvVar("LINEA_ROLLUP_SECURITY_COUNCIL");
  const lineaRollupRateLimitPeriodInSeconds = getRequiredEnvVar("LINEA_ROLLUP_RATE_LIMIT_PERIOD");
  const lineaRollupRateLimitAmountInWei = getRequiredEnvVar("LINEA_ROLLUP_RATE_LIMIT_AMOUNT");

  const pauseTypeRoles = getEnvVarOrDefault("LINEA_ROLLUP_PAUSE_TYPE_ROLES", LINEA_ROLLUP_PAUSE_TYPES_ROLES);
  const unpauseTypeRoles = getEnvVarOrDefault("LINEA_ROLLUP_UNPAUSE_TYPE_ROLES", LINEA_ROLLUP_UNPAUSE_TYPES_ROLES);
  const defaultRoleAddresses = generateRoleAssignments(LINEA_ROLLUP_ROLES, lineaRollupSecurityCouncil, []);
  const roleAddresses = getEnvVarOrDefault("LINEA_ROLLUP_ROLE_ADDRESSES", defaultRoleAddresses);

  if (!existingContractAddress) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

  const contract = await deployUpgradableFromFactory(
    "TestLineaRollup",
    [
      {
        initialStateRootHash: ZeroHash,
        initialL2BlockNumber: 0,
        genesisTimestamp: 0,
        defaultVerifier: DEAD_ADDRESS,
        rateLimitPeriodInSeconds: lineaRollupRateLimitPeriodInSeconds,
        rateLimitAmountInWei: lineaRollupRateLimitAmountInWei,
        roleAddresses,
        pauseTypeRoles,
        unpauseTypeRoles,
        fallbackOperator: DEAD_ADDRESS,
        defaultAdmin: lineaRollupSecurityCouncil,
      },
    ],
    {
      initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor"],
    },
  );

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  await tryStoreAddress(hre.network.name, contractName, contractAddress, contract.deploymentTransaction()!.hash);

  await tryVerifyContract(contractAddress);
};

export default func;
func.tags = ["TestLineaRollup"];
