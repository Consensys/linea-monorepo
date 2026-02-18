// Usage:
// npx hardhat deploy --network <network> --tags LineaRollupV8WithReinitialization
//
// Required environment variables:
//   L1_SECURITY_COUNCIL
//   LINEA_ROLLUP_ADDRESS
//   NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS
//
// This script deploys a LineaRollup implementation from audited artifacts and generates
// encoded calldata for upgrading an existing proxy via reinitializeV8.
// The encoded calldata should be used through the Safe for the actual upgrade.

import { DeployFunction } from "hardhat-deploy/types";
import { getRequiredEnvVar, deployContractFromArtifacts, generateRoleAssignments } from "../common/helpers";
import {
  UNPAUSE_NATIVE_YIELD_STAKING_ROLE,
  PAUSE_NATIVE_YIELD_STAKING_ROLE,
  YIELD_PROVIDER_STAKING_ROLE,
  SET_YIELD_MANAGER_ROLE,
  NATIVE_YIELD_STAKING_PAUSE_TYPE,
} from "contracts/common/constants";
import {
  contractName as LineaRollupContractName,
  abi as LineaRollupAbi,
  bytecode as LineaRollupBytecode,
} from "../deployments/bytecode/2026-01-14/LineaRollup.json";
import { LineaRollup__factory } from "contracts/typechain-types";
import { HardhatRuntimeEnvironment } from "hardhat/types";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { ethers, getNamedAccounts } = hre;
  const { deployer } = await getNamedAccounts();
  const signer = await ethers.getSigner(deployer);

  const securityCouncilAddress = getRequiredEnvVar("L1_SECURITY_COUNCIL");
  const automationServiceAddress = getRequiredEnvVar("NATIVE_YIELD_AUTOMATION_SERVICE_ADDRESS");
  const proxyAddress = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");

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

  const lineaRollupImpl = await deployContractFromArtifacts(
    LineaRollupContractName,
    LineaRollupAbi,
    LineaRollupBytecode,
    signer,
  );

  const reinitializeCalldata = LineaRollup__factory.createInterface().encodeFunctionData("reinitializeV8", [
    newRoleAddresses,
    newPauseRoles,
    newUnPauseRoles,
  ]);

  // The encoding should be used through the safe.
  // THIS IS JUST A SAMPLE AND WILL BE ADJUSTED WHEN NEEDED FOR GENERATING THE CALLDATA FOR THE UPGRADE CALL
  // https://www.4byte.directory/signatures/?bytes4_signature=0x9623609d
  const upgradeCallWithReinitializationUsingSecurityCouncil = ethers.concat([
    "0x9623609d",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address", "bytes"],
      [proxyAddress, await lineaRollupImpl.getAddress(), reinitializeCalldata],
    ),
  ]);

  console.log(
    "Encoded Tx Upgrade with Reinitialization from Security Council:",
    "\n",
    upgradeCallWithReinitializationUsingSecurityCouncil,
  );
  console.log("\n");
};

export default func;
func.tags = ["LineaRollupV8WithReinitialization"];
