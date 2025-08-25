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

  const contractName = "VaultFactory";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);
  const provider = ethers.provider;

  const deployer = getRequiredEnvVar("STATUS_NETWORK_DEPLOYER");
  const stakingToken = getRequiredEnvVar("STATUS_NETWORK_STAKING_TOKEN"); // SNT token address
  
  // Get StakeManager proxy address from previous deployment
  const stakeManagerAddress = await getDeployedContractAddress("StakeManager", deployments);
  if (!stakeManagerAddress) {
    throw new Error("StakeManager must be deployed first");
  }

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

  // Deploy StakeVault implementation
  const vaultImplementation = await deployFromFactory("StakeVault", provider, stakingToken, await get1559Fees(provider));
  const vaultImplAddress = await vaultImplementation.getAddress();
  
  console.log(`StakeVault implementation deployed at: ${vaultImplAddress}`);

  // Deploy VaultFactory
  const contract = await deployFromFactory(
    "VaultFactory",
    provider,
    deployer,
    stakeManagerAddress,
    vaultImplAddress,
    await get1559Fees(provider)
  );

  const contractAddress = await contract.getAddress();
  await LogContractDeployment(contractName, contract);

  await tryStoreAddress(hre.network.name, contractName, contractAddress, contract.deploymentTransaction()!.hash);

  const args = [deployer, stakeManagerAddress, vaultImplAddress];
  await tryVerifyContractWithConstructorArgs(contractAddress, "contracts/src/VaultFactory.sol:VaultFactory", args);

  // Whitelist the vault implementation in StakeManager
  console.log("Setting trusted codehash for StakeVault implementation...");
  
  // Create a proxy clone to get the codehash
  const proxyCloneFactory = await ethers.getContractFactory("Clones");
  const cloneAddress = await proxyCloneFactory.predictDeterministicAddress(vaultImplAddress, ethers.ZeroHash);
  
  // Get the StakeManager contract instance
  const stakeManager = await ethers.getContractAt("StakeManager", stakeManagerAddress);
  
  // Set trusted codehash (this would need to be called by the owner/deployer)
  console.log(`Setting trusted codehash for vault at ${cloneAddress}`);
  // Note: This would be done in a separate script or manually by the deployer
  // await stakeManager.setTrustedCodehash(cloneAddress.codehash, true);
};

export default func;
func.tags = ["StatusNetworkVaultFactory"];
func.dependencies = ["StatusNetworkStakeManager"];
