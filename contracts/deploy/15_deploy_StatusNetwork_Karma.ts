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

  const contractName = "Karma";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);
  const provider = ethers.provider;

  const deployer = getRequiredEnvVar("STATUS_NETWORK_DEPLOYER");

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

  // Deploy Karma implementation
  const karmaImpl = await deployFromFactory("Karma", provider, await get1559Fees(provider));
  const karmaImplAddress = await karmaImpl.getAddress();
  
  console.log(`Karma implementation deployed at: ${karmaImplAddress}`);

  // Prepare initialization data
  const initializeData = ethers.concat([
    "0xc4d66de8", // initialize(address) function selector
    ethers.AbiCoder.defaultAbiCoder().encode(["address"], [deployer])
  ]);

  // Deploy ERC1967Proxy
  const proxyContract = await deployFromFactory(
    "ERC1967Proxy", 
    provider, 
    karmaImplAddress,
    initializeData,
    await get1559Fees(provider)
  );

  const contractAddress = await proxyContract.getAddress();
  await LogContractDeployment(contractName, proxyContract);

  await tryStoreAddress(hre.network.name, contractName, contractAddress, proxyContract.deploymentTransaction()!.hash);

  const args = [karmaImplAddress, initializeData];
  await tryVerifyContractWithConstructorArgs(contractAddress, "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol:ERC1967Proxy", args);
};

export default func;
func.tags = ["StatusNetworkKarma"];
func.dependencies = [];
