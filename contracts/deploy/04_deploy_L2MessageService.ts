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
  L1_L2_MESSAGE_SETTER_ROLE,
  L1_L2_PAUSE_TYPE,
  L2_L1_PAUSE_TYPE,
  MINIMUM_FEE_SETTER_ROLE,
  PAUSE_ALL_ROLE,
  PAUSE_FINALIZE_WITHPROOF_ROLE,
  PAUSE_L1_L2_ROLE,
  PAUSE_L2_BLOB_SUBMISSION_ROLE,
  PAUSE_L2_L1_ROLE,
  RATE_LIMIT_SETTER_ROLE,
  UNPAUSE_ALL_ROLE,
  UNPAUSE_FINALIZE_WITHPROOF_ROLE,
  UNPAUSE_L1_L2_ROLE,
  UNPAUSE_L2_BLOB_SUBMISSION_ROLE,
  UNPAUSE_L2_L1_ROLE,
  USED_RATE_LIMIT_RESETTER_ROLE,
} from "contracts/test/utils/constants";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments } = hre;
  validateDeployBranchAndTags(hre.network.name);

  const contractName = "L2MessageService";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);

  const L2MessageService_securityCouncil = requireEnv("L2MSGSERVICE_SECURITY_COUNCIL");
  const L2MessageService_l1l2MessageSetter = requireEnv("L2MSGSERVICE_L1L2_MESSAGE_SETTER");
  const L2MessageService_rateLimitPeriod = requireEnv("L2MSGSERVICE_RATE_LIMIT_PERIOD");
  const L2MessageService_rateLimitAmount = requireEnv("L2MSGSERVICE_RATE_LIMIT_AMOUNT");
  const L2MessageService_roleAddress = process.env["L2MSGSERVICE_ROLE_ADDRESSES"];
  const L2MessageService_pauseTypeRoles = process.env["L2MSGSERVICE_PAUSE_TYPE_ROLES"];
  const L2MessageService_unpauseTypeRoles = process.env["L2MSGSERVICE_UNPAUSE_TYPE_ROLES"];

  let pauseTypeRoles = {};
  const pauseTypeRolesDefault = [
    { pauseType: GENERAL_PAUSE_TYPE, role: PAUSE_ALL_ROLE },
    { pauseType: L1_L2_PAUSE_TYPE, role: PAUSE_L1_L2_ROLE },
    { pauseType: L2_L1_PAUSE_TYPE, role: PAUSE_L2_L1_ROLE },
    { pauseType: BLOB_SUBMISSION_PAUSE_TYPE, role: PAUSE_L2_BLOB_SUBMISSION_ROLE },
    { pauseType: CALLDATA_SUBMISSION_PAUSE_TYPE, role: PAUSE_L2_BLOB_SUBMISSION_ROLE },
    { pauseType: FINALIZATION_PAUSE_TYPE, role: PAUSE_FINALIZE_WITHPROOF_ROLE },
  ];

  if (L2MessageService_pauseTypeRoles !== undefined) {
    console.log("Using provided L2MSGSERVICE_PAUSE_TYPE_ROLES environment variable");
    pauseTypeRoles = L2MessageService_pauseTypeRoles;
  } else {
    console.log("Using default pauseTypeRoles");
    pauseTypeRoles = pauseTypeRolesDefault;
  }

  let unpauseTypeRoles = {};
  const unpauseTypeRolesDefault = [
    { pauseType: GENERAL_PAUSE_TYPE, role: UNPAUSE_ALL_ROLE },
    { pauseType: L1_L2_PAUSE_TYPE, role: UNPAUSE_L1_L2_ROLE },
    { pauseType: L2_L1_PAUSE_TYPE, role: UNPAUSE_L2_L1_ROLE },
    { pauseType: BLOB_SUBMISSION_PAUSE_TYPE, role: UNPAUSE_L2_BLOB_SUBMISSION_ROLE },
    { pauseType: CALLDATA_SUBMISSION_PAUSE_TYPE, role: UNPAUSE_L2_BLOB_SUBMISSION_ROLE },
    { pauseType: FINALIZATION_PAUSE_TYPE, role: UNPAUSE_FINALIZE_WITHPROOF_ROLE },
  ];

  if (L2MessageService_unpauseTypeRoles !== undefined) {
    console.log("Using provided L2MSGSERVICE_UNPAUSE_TYPE_ROLES environment variable");
    unpauseTypeRoles = L2MessageService_unpauseTypeRoles;
  } else {
    console.log("Using default unpauseTypeRoles");
    unpauseTypeRoles = unpauseTypeRolesDefault;
  }

  let roleAddresses = {};
  const roleAddressDefault = [
    { addressWithRole: L2MessageService_securityCouncil, role: DEFAULT_ADMIN_ROLE },
    { addressWithRole: L2MessageService_securityCouncil, role: MINIMUM_FEE_SETTER_ROLE },
    { addressWithRole: L2MessageService_securityCouncil, role: RATE_LIMIT_SETTER_ROLE },
    { addressWithRole: L2MessageService_securityCouncil, role: USED_RATE_LIMIT_RESETTER_ROLE },
    { addressWithRole: L2MessageService_securityCouncil, role: PAUSE_ALL_ROLE },
    { addressWithRole: L2MessageService_securityCouncil, role: UNPAUSE_ALL_ROLE },
    { addressWithRole: L2MessageService_securityCouncil, role: PAUSE_L1_L2_ROLE },
    { addressWithRole: L2MessageService_securityCouncil, role: UNPAUSE_L1_L2_ROLE },
    { addressWithRole: L2MessageService_securityCouncil, role: PAUSE_L2_L1_ROLE },
    { addressWithRole: L2MessageService_securityCouncil, role: UNPAUSE_L2_L1_ROLE },
    { addressWithRole: L2MessageService_l1l2MessageSetter, role: L1_L2_MESSAGE_SETTER_ROLE },
  ];
  if (L2MessageService_roleAddress !== undefined) {
    console.log("Using provided L2MSGSERVICE_ROLE_ADDRESSES environment variable");
    roleAddresses = L2MessageService_roleAddress;
  } else {
    console.log("Using default roleAddresses");
    roleAddresses = roleAddressDefault;
  }

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

  const contract = await deployUpgradableFromFactory(
    "L2MessageService",
    [
      L2MessageService_rateLimitPeriod,
      L2MessageService_rateLimitAmount,
      roleAddresses,
      pauseTypeRoles,
      unpauseTypeRoles,
    ],
    {
      initializer: "initialize(uint256,uint256,(address,bytes32)[],(uint8,bytes32)[],(uint8,bytes32)[])",
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
func.tags = ["L2MessageService"];
