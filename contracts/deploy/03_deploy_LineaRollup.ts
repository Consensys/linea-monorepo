import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { deployUpgradableFromFactory, requireEnv } from "../scripts/hardhat/utils";
import { validateDeployBranchAndTags } from "../utils/auditedDeployVerifier";
import { getDeployedContractAddress, tryStoreAddress } from "../utils/storeAddress";
import { tryVerifyContract } from "../utils/verifyContract";
import {
  BLOB_SUBMISSION_PAUSE_TYPE,
  CALLDATA_SUBMISSION_PAUSE_TYPE,
  DEFAULT_ADMIN_ROLE,
  FINALIZATION_PAUSE_TYPE,
  GENERAL_PAUSE_TYPE,
  L1_L2_PAUSE_TYPE,
  L2_L1_PAUSE_TYPE,
  PAUSE_ALL_ROLE,
  PAUSE_FINALIZE_WITHPROOF_ROLE,
  PAUSE_L1_L2_ROLE,
  PAUSE_L2_BLOB_SUBMISSION_ROLE,
  PAUSE_L2_L1_ROLE,
  UNPAUSE_ALL_ROLE,
  UNPAUSE_FINALIZE_WITHPROOF_ROLE,
  UNPAUSE_L1_L2_ROLE,
  UNPAUSE_L2_BLOB_SUBMISSION_ROLE,
  UNPAUSE_L2_L1_ROLE,
  VERIFIER_SETTER_ROLE,
  VERIFIER_UNSETTER_ROLE,
  OPERATOR_ROLE,
  LINEA_ROLLUP_INITIALIZE_SIGNATURE,
} from "contracts/test/utils/constants";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments } = hre;
  validateDeployBranchAndTags(hre.network.name);

  const contractName = "LineaRollup";
  const verifierName = "PlonkVerifier";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);
  let verifierAddress = await getDeployedContractAddress(verifierName, deployments);
  if (verifierAddress === undefined) {
    if (process.env["PLONKVERIFIER_ADDRESS"] !== undefined) {
      console.log(`Using environment variable for PlonkVerifier , ${process.env["PLONKVERIFIER_ADDRESS"]}`);
      verifierAddress = process.env["PLONKVERIFIER_ADDRESS"];
    } else {
      throw "Missing PLONKVERIFIER_ADDRESS environment variable";
    }
  } else {
    console.log(`Using deployed variable for PlonkVerifier , ${verifierAddress}`);
  }

  // LineaRollup DEPLOYED AS UPGRADEABLE PROXY
  const LineaRollup_initialStateRootHash = requireEnv("LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH");
  const LineaRollup_initialL2BlockNumber = requireEnv("LINEA_ROLLUP_INITIAL_L2_BLOCK_NUMBER");
  const LineaRollup_securityCouncil = requireEnv("LINEA_ROLLUP_SECURITY_COUNCIL");
  const LineaRollup_operators = requireEnv("LINEA_ROLLUP_OPERATORS").split(",");
  const LineaRollup_rateLimitPeriodInSeconds = requireEnv("LINEA_ROLLUP_RATE_LIMIT_PERIOD");
  const LineaRollup_rateLimitAmountInWei = requireEnv("LINEA_ROLLUP_RATE_LIMIT_AMOUNT");
  const LineaRollup_genesisTimestamp = requireEnv("LINEA_ROLLUP_GENESIS_TIMESTAMP");
  const MultiCallAddress = "0xcA11bde05977b3631167028862bE2a173976CA11";
  const LineaRollup_roleAddresses = process.env["LINEA_ROLLUP_ROLE_ADDRESSES"];
  const LineaRollup_pauseTypeRoles = process.env["LINEA_ROLLUP_PAUSE_TYPE_ROLES"];
  const LineaRollup_unpauseTypeRoles = process.env["LINEA_ROLLUP_UNPAUSE_TYPE_ROLES"];

  let pauseTypeRoles = [];
  const pauseTypeRolesDefault = [
    { pauseType: GENERAL_PAUSE_TYPE, role: PAUSE_ALL_ROLE },
    { pauseType: L1_L2_PAUSE_TYPE, role: PAUSE_L1_L2_ROLE },
    { pauseType: L2_L1_PAUSE_TYPE, role: PAUSE_L2_L1_ROLE },
    { pauseType: BLOB_SUBMISSION_PAUSE_TYPE, role: PAUSE_L2_BLOB_SUBMISSION_ROLE },
    { pauseType: CALLDATA_SUBMISSION_PAUSE_TYPE, role: PAUSE_L2_BLOB_SUBMISSION_ROLE },
    { pauseType: FINALIZATION_PAUSE_TYPE, role: PAUSE_FINALIZE_WITHPROOF_ROLE },
  ];

  if (LineaRollup_pauseTypeRoles !== undefined) {
    console.log("Using provided LINEA_ROLLUP_PAUSE_TYPE_ROLES environment variable");
    pauseTypeRoles = JSON.parse(LineaRollup_pauseTypeRoles);
  } else {
    console.log("Using default pauseTypeRoles");
    pauseTypeRoles = pauseTypeRolesDefault;
  }

  let unpauseTypeRoles = [];
  const unpauseTypeRolesDefault = [
    { pauseType: GENERAL_PAUSE_TYPE, role: UNPAUSE_ALL_ROLE },
    { pauseType: L1_L2_PAUSE_TYPE, role: UNPAUSE_L1_L2_ROLE },
    { pauseType: L2_L1_PAUSE_TYPE, role: UNPAUSE_L2_L1_ROLE },
    { pauseType: BLOB_SUBMISSION_PAUSE_TYPE, role: UNPAUSE_L2_BLOB_SUBMISSION_ROLE },
    { pauseType: CALLDATA_SUBMISSION_PAUSE_TYPE, role: UNPAUSE_L2_BLOB_SUBMISSION_ROLE },
    { pauseType: FINALIZATION_PAUSE_TYPE, role: UNPAUSE_FINALIZE_WITHPROOF_ROLE },
  ];

  if (LineaRollup_unpauseTypeRoles !== undefined) {
    console.log("Using provided LINEA_ROLLUP_UNPAUSE_TYPE_ROLES environment variable");
    unpauseTypeRoles = JSON.parse(LineaRollup_unpauseTypeRoles);
  } else {
    console.log("Using default unpauseTypeRoles");
    unpauseTypeRoles = unpauseTypeRolesDefault;
  }

  let roleAddresses = [];
  const roleAddressesDefault = [
    { addressWithRole: LineaRollup_securityCouncil, role: DEFAULT_ADMIN_ROLE },
    { addressWithRole: LineaRollup_securityCouncil, role: VERIFIER_SETTER_ROLE },
    { addressWithRole: LineaRollup_securityCouncil, role: VERIFIER_UNSETTER_ROLE },
    { addressWithRole: LineaRollup_securityCouncil, role: PAUSE_ALL_ROLE },
    { addressWithRole: LineaRollup_securityCouncil, role: UNPAUSE_ALL_ROLE },
    { addressWithRole: LineaRollup_securityCouncil, role: PAUSE_L2_BLOB_SUBMISSION_ROLE },
    { addressWithRole: LineaRollup_securityCouncil, role: UNPAUSE_L2_BLOB_SUBMISSION_ROLE },
    { addressWithRole: LineaRollup_securityCouncil, role: PAUSE_FINALIZE_WITHPROOF_ROLE },
    { addressWithRole: LineaRollup_securityCouncil, role: UNPAUSE_FINALIZE_WITHPROOF_ROLE },
  ];

  for (let i = 0; i < LineaRollup_operators.length; i++) {
    roleAddressesDefault.push({ addressWithRole: LineaRollup_operators[i], role: OPERATOR_ROLE });
  }

  if (LineaRollup_roleAddresses !== undefined) {
    console.log("Using provided LINEA_ROLLUP_ROLE_ADDRESSES environment variable");
    roleAddresses = JSON.parse(LineaRollup_roleAddresses);
  } else {
    console.log("Using default roleAddresses");
    roleAddresses = roleAddressesDefault;
  }

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }
  const contract = await deployUpgradableFromFactory(
    "LineaRollup",
    [
      {
        initialStateRootHash: LineaRollup_initialStateRootHash,
        initialL2BlockNumber: LineaRollup_initialL2BlockNumber,
        genesisTimestamp: LineaRollup_genesisTimestamp,
        defaultVerifier: verifierAddress,
        rateLimitPeriodInSeconds: LineaRollup_rateLimitPeriodInSeconds,
        rateLimitAmountInWei: LineaRollup_rateLimitAmountInWei,
        roleAddresses: roleAddresses,
        pauseTypeRoles: pauseTypeRoles,
        unpauseTypeRoles: unpauseTypeRoles,
        fallbackOperator: MultiCallAddress,
      },
    ],
    {
      initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor"],
    },
  );
  const contractAddress = await contract.getAddress();
  const txReceipt = await contract.deploymentTransaction()?.wait();
  if (!txReceipt) {
    throw "Contract deployment transaction receipt not found.";
  }

  console.log(`${contractName} deployed: address=${contractAddress} blockNumber=${txReceipt.blockNumber}`);

  await tryStoreAddress(hre.network.name, contractName, contractAddress, txReceipt.hash);

  await tryVerifyContract(contractAddress);
};

export default func;
func.tags = ["LineaRollup"];
