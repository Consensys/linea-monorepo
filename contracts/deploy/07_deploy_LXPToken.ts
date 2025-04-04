import { ethers } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { deployFromFactory } from "../scripts/hardhat/utils";
import { get1559Fees } from "../scripts/utils";
import {
  tryVerifyContractWithConstructorArgs,
  getDeployedContractAddress,
  tryStoreAddress,
  getRequiredEnvVar,
  LogContractDeployment,
} from "../common/helpers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments } = hre;

  const contractName = "LineaVoyageXP";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);
  const provider = ethers.provider;

  const adminAddress = getRequiredEnvVar("LINEA_VOYAGE_XP_ADMIN_ADDRESS");

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }
  const contract = await deployFromFactory(contractName, provider, adminAddress, await get1559Fees(provider));

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  await tryStoreAddress(hre.network.name, contractName, contractAddress, contract.deploymentTransaction()!.hash);

  const args = [adminAddress];

  await tryVerifyContractWithConstructorArgs(contractAddress, "contracts/token/LineaVoyageXP.sol:LineaVoyageXP", args);
};
export default func;
func.tags = ["LineaVoyageXPToken"];
