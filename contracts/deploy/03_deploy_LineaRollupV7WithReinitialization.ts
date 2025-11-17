import { ethers, upgrades } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import {
  tryVerifyContract,
  getDeployedContractAddress,
  getRequiredEnvVar,
  generateRoleAssignments,
} from "../common/helpers";
import { LineaRollup__factory } from "contracts/typechain-types";
import {
  SET_YIELD_MANAGER_ROLE,
  YIELD_PROVIDER_STAKING_ROLE,
  PAUSE_NATIVE_YIELD_STAKING_ROLE,
  UNPAUSE_NATIVE_YIELD_STAKING_ROLE,
  NATIVE_YIELD_STAKING_PAUSE_TYPE,
} from "contracts/common/constants";

// Prerequisite - run 14_deploy_YieldManager.ts
const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const securityCouncilAddress = getRequiredEnvVar("LINEA_ROLLUP_SECURITY_COUNCIL");
  const yieldManager = getRequiredEnvVar("YIELD_MANAGER");
  const proxyAddress = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");
  const automationServiceAddress = getRequiredEnvVar("AUTOMATION_SERVICE_ADDRESS");

  const newRoles = [
    SET_YIELD_MANAGER_ROLE,
    YIELD_PROVIDER_STAKING_ROLE,
    PAUSE_NATIVE_YIELD_STAKING_ROLE,
    UNPAUSE_NATIVE_YIELD_STAKING_ROLE,
  ];

  const newRoleAddresses = [
    ...generateRoleAssignments(newRoles, securityCouncilAddress, []),
    {
      role: YIELD_PROVIDER_STAKING_ROLE,
      addressWithRole: automationServiceAddress,
    },
  ];
  console.log("New role addresses", newRoleAddresses);

  const newPauseRoles = [{ pauseType: NATIVE_YIELD_STAKING_PAUSE_TYPE, role: PAUSE_NATIVE_YIELD_STAKING_ROLE }];
  const newUnPauseRoles = [{ pauseType: NATIVE_YIELD_STAKING_PAUSE_TYPE, role: UNPAUSE_NATIVE_YIELD_STAKING_ROLE }];

  const { deployments } = hre;
  const contractName = "LineaRollup";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);

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
        LineaRollup__factory.createInterface().encodeFunctionData("reinitializeLineaRollupV7", [
          newRoleAddresses,
          newPauseRoles,
          newUnPauseRoles,
          yieldManager,
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
func.tags = ["LineaRollupV7WithReinitialization"];
