import { ethers, upgrades } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { tryVerifyContract, getDeployedContractAddress, getRequiredEnvVar } from "../common/helpers";
import { TokenBridge__factory } from "contracts/typechain-types";
import {
  L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES,
  L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES,
  PAUSE_ALL_ROLE,
  PAUSE_L1_L2_ROLE,
  PAUSE_L2_L1_ROLE,
  UNPAUSE_ALL_ROLE,
  UNPAUSE_L1_L2_ROLE,
  UNPAUSE_L2_L1_ROLE,
  USED_RATE_LIMIT_RESETTER_ROLE,
} from "contracts/common/constants";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const securityCouncilAddress = "0xcA11bde05977b3631167028862bE2a173976CA11";

  const newRoleAddresses = [
    { addressWithRole: securityCouncilAddress, role: USED_RATE_LIMIT_RESETTER_ROLE },
    { addressWithRole: securityCouncilAddress, role: PAUSE_ALL_ROLE },
    { addressWithRole: securityCouncilAddress, role: PAUSE_L1_L2_ROLE },
    { addressWithRole: securityCouncilAddress, role: PAUSE_L2_L1_ROLE },
    { addressWithRole: securityCouncilAddress, role: UNPAUSE_ALL_ROLE },
    { addressWithRole: securityCouncilAddress, role: UNPAUSE_L1_L2_ROLE },
    { addressWithRole: securityCouncilAddress, role: UNPAUSE_L2_L1_ROLE },
  ];

  const { deployments } = hre;
  const contractName = "L2MessageService";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);

  const proxyAddress = getRequiredEnvVar("L2_MESSAGE_SERVICE_ADDRESS");

  const factory = await ethers.getContractFactory(contractName);

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

  console.log("Deploying Contract...");
  const newContract = await upgrades.deployImplementation(factory, {
    kind: "transparent",
  });

  const contract = newContract.toString();

  console.log(`Contract deployed at ${contract}`);

  // The encoding should be used through the safe.
  // THIS IS JUST A SAMPLE AND WILL BE ADJUSTED WHEN NEEDED FOR GENERATING THE CALLDATA FOR THE UPGRADE CALL
  // https://www.4byte.directory/signatures/?bytes4_signature=0x9623609d
  const upgradeCallWithReinitializationUsingSecurityCouncil = ethers.concat([
    "0x9623609d",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address", "bytes"],
      [
        proxyAddress,
        newContract,
        TokenBridge__factory.createInterface().encodeFunctionData("reinitializePauseTypesAndPermissions", [
          securityCouncilAddress,
          newRoleAddresses,
          L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES,
          L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES,
        ]),
      ],
    ),
  ]);

  console.log(
    "Encoded Tx Upgrade with Reinitialization from Security Council:",
    "\n",
    upgradeCallWithReinitializationUsingSecurityCouncil,
  );
  console.log("\n");

  await tryVerifyContract(contract);
};

export default func;
func.tags = ["TokenBridgeWithReinitialization"];
