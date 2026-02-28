import type { NewTaskActionFunction } from "hardhat/types/tasks";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper.js";

interface TaskArgs {
  adminAddress?: string;
  proxyAddress?: string;
  contractType?: string;
  contractRoles?: string;
}

const action: NewTaskActionFunction<TaskArgs> = async (taskArgs, hre) => {
  const connection = await hre.network.connect();
  const { ethers } = connection;

  const adminAddress = getTaskCliOrEnvValue(taskArgs, "adminAddress", "ADMIN_ADDRESS");
  const proxyAddress = getTaskCliOrEnvValue(taskArgs, "proxyAddress", "PROXY_ADDRESS");
  const contractType = getTaskCliOrEnvValue(taskArgs, "contractType", "CONTRACT_TYPE");
  const contractRoles = getTaskCliOrEnvValue(taskArgs, "contractRoles", "CONTRACT_ROLES");

  if (contractType === undefined) {
    throw new Error("Please specify a contract type e.g: --contract-type LineaRollup or CONTRACT_TYPE=LineaRollup");
  }

  if (proxyAddress === undefined) {
    throw new Error("Please specify a proxy address e.g: --proxy-address 0x... or PROXY_ADDRESS=0x...");
  }

  if (contractRoles === undefined) {
    throw new Error("Please specify roles e.g. --contract-roles 0x9a80e24e... or CONTRACT_ROLES=0x9a80e24e...");
  }

  if (adminAddress === undefined) {
    throw new Error("Please specify an admin address e.g. --admin-address 0x... or ADMIN_ADDRESS=0x...");
  }

  const contract = await ethers.getContractAt(contractType, proxyAddress);

  const rolesArray = contractRoles.split(",");
  for (let i = 0; i < rolesArray.length; i++) {
    console.log(`Granting ${rolesArray[i]} to ${adminAddress}`);
    const tx = await contract.grantRole(rolesArray[i], adminAddress);
    console.log("Waiting for transaction to process");
    await tx.wait();
  }

  console.log("Done");
};

export default action;
