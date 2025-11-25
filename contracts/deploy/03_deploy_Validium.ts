import { DeployFunction } from "hardhat-deploy/types";
import { deployUpgradableFromFactory } from "../scripts/hardhat/utils";
import {
  generateRoleAssignments,
  getEnvVarOrDefault,
  getRequiredEnvVar,
  tryVerifyContract,
  LogContractDeployment,
} from "../common/helpers";
import {
  VALIDIUM_INITIALIZE_SIGNATURE,
  VALIDIUM_PAUSE_TYPES_ROLES,
  VALIDIUM_ROLES,
  VALIDIUM_UNPAUSE_TYPES_ROLES,
  OPERATOR_ROLE,
} from "../common/constants";

const func: DeployFunction = async function () {
  const contractName = "LineaRollup";

  // LineaRollup DEPLOYED AS UPGRADEABLE PROXY
  const verifierAddress = getRequiredEnvVar("PLONKVERIFIER_ADDRESS");
  const lineaRollupInitialStateRootHash = getRequiredEnvVar("VALIDIUM_INITIAL_STATE_ROOT_HASH");
  const lineaRollupInitialL2BlockNumber = getRequiredEnvVar("VALIDIUM_INITIAL_L2_BLOCK_NUMBER");
  const lineaRollupSecurityCouncil = getRequiredEnvVar("VALIDIUM_SECURITY_COUNCIL");
  const lineaRollupOperators = getRequiredEnvVar("VALIDIUM_OPERATORS").split(",");
  const lineaRollupRateLimitPeriodInSeconds = getRequiredEnvVar("VALIDIUM_RATE_LIMIT_PERIOD");
  const lineaRollupRateLimitAmountInWei = getRequiredEnvVar("VALIDIUM_RATE_LIMIT_AMOUNT");
  const lineaRollupGenesisTimestamp = getRequiredEnvVar("VALIDIUM_GENESIS_TIMESTAMP");

  const pauseTypeRoles = getEnvVarOrDefault("VALIDIUM_PAUSE_TYPE_ROLES", VALIDIUM_PAUSE_TYPES_ROLES);
  const unpauseTypeRoles = getEnvVarOrDefault("VALIDIUM_UNPAUSE_TYPE_ROLES", VALIDIUM_UNPAUSE_TYPES_ROLES);
  const defaultRoleAddresses = generateRoleAssignments(VALIDIUM_ROLES, lineaRollupSecurityCouncil, [
    { role: OPERATOR_ROLE, addresses: lineaRollupOperators },
  ]);
  const roleAddresses = getEnvVarOrDefault("VALIDIUM_ROLE_ADDRESSES", defaultRoleAddresses);

  const contract = await deployUpgradableFromFactory(
    "Validium",
    [
      {
        initialStateRootHash: lineaRollupInitialStateRootHash,
        initialL2BlockNumber: lineaRollupInitialL2BlockNumber,
        genesisTimestamp: lineaRollupGenesisTimestamp,
        defaultVerifier: verifierAddress,
        rateLimitPeriodInSeconds: lineaRollupRateLimitPeriodInSeconds,
        rateLimitAmountInWei: lineaRollupRateLimitAmountInWei,
        roleAddresses,
        pauseTypeRoles,
        unpauseTypeRoles,
        defaultAdmin: lineaRollupSecurityCouncil,
      },
    ],
    {
      initializer: VALIDIUM_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor"],
    },
  );

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  await tryVerifyContract(contractAddress);
};

export default func;
func.tags = ["LineaRollup"];
