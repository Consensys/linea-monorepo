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

  const contractName = "StakeManager";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);
  const provider = ethers.provider;

  const deployer = getRequiredEnvVar("STATUS_NETWORK_DEPLOYER");
  const stakingToken = getRequiredEnvVar("STATUS_NETWORK_STAKING_TOKEN"); // SNT token address

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

  // Deploy StakeManager implementation
  const stakeManagerImpl = await deployFromFactory("StakeManager", provider, await get1559Fees(provider));
  const stakeManagerImplAddress = await stakeManagerImpl.getAddress();
  
  console.log(`StakeManager implementation deployed at: ${stakeManagerImplAddress}`);

  // Prepare initialization data
  const initializeData = ethers.concat([
    "0x485cc955", // initialize(address,address) function selector
    ethers.AbiCoder.defaultAbiCoder().encode(["address", "address"], [deployer, stakingToken])
  ]);

  // Deploy TransparentProxy
  const proxyContract = await deployFromFactory(
    "TransparentProxy", 
    provider, 
    stakeManagerImplAddress,
    initializeData,
    await get1559Fees(provider)
  );

  const contractAddress = await proxyContract.getAddress();
  await LogContractDeployment(contractName, proxyContract);

  await tryStoreAddress(hre.network.name, contractName, contractAddress, proxyContract.deploymentTransaction()!.hash);

  const args = [stakeManagerImplAddress, initializeData];
  await tryVerifyContractWithConstructorArgs(contractAddress, "contracts/src/proxies/TransparentProxy.sol:TransparentProxy", args);
};

export default func;
func.tags = ["StatusNetworkStakeManager"];
func.dependencies = []; // Can add dependencies if needed
