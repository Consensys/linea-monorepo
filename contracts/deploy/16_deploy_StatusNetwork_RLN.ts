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
  getEnvVarOrDefault,
  LogContractDeployment,
} from "../common/helpers";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments } = hre;

  const contractName = "RLN";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);
  const provider = ethers.provider;

  const deployer = getRequiredEnvVar("STATUS_NETWORK_DEPLOYER");
  const rlnDepth = getEnvVarOrDefault("STATUS_NETWORK_RLN_DEPTH", "20"); // Default depth of 20 for 1M users
  
  // Get Karma contract address from previous deployment
  const karmaAddress = await getDeployedContractAddress("Karma", deployments);
  if (!karmaAddress) {
    throw new Error("Karma contract must be deployed first");
  }

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

  // Deploy RLN implementation
  const rlnImpl = await deployFromFactory("RLN", provider, await get1559Fees(provider));
  const rlnImplAddress = await rlnImpl.getAddress();
  
  console.log(`RLN implementation deployed at: ${rlnImplAddress}`);

  // Prepare initialization data
  // initialize(address owner, address admin, address registrar, uint256 depth, address karmaContract)
  const initializeData = ethers.concat([
    "0x", // initialize function selector (would need actual selector)
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address", "address", "uint256", "address"], 
      [deployer, deployer, deployer, parseInt(rlnDepth), karmaAddress]
    )
  ]);

  // Deploy ERC1967Proxy
  const proxyContract = await deployFromFactory(
    "ERC1967Proxy", 
    provider, 
    rlnImplAddress,
    initializeData,
    await get1559Fees(provider)
  );

  const contractAddress = await proxyContract.getAddress();
  await LogContractDeployment(contractName, proxyContract);

  await tryStoreAddress(hre.network.name, contractName, contractAddress, proxyContract.deploymentTransaction()!.hash);

  const args = [rlnImplAddress, initializeData];
  await tryVerifyContractWithConstructorArgs(contractAddress, "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol:ERC1967Proxy", args);

  console.log(`RLN deployed with depth ${rlnDepth} and karma contract at ${karmaAddress}`);
};

export default func;
func.tags = ["StatusNetworkRLN"];
func.dependencies = ["StatusNetworkKarma"];
