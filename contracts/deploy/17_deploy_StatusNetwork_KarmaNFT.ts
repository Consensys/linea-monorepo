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

  const contractName = "KarmaNFT";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);
  const provider = ethers.provider;

  // Get Karma contract address from previous deployment
  const karmaAddress = await getDeployedContractAddress("Karma", deployments);
  if (!karmaAddress) {
    throw new Error("Karma contract must be deployed first");
  }

  // Deploy metadata generator first
  const metadataGenerator = await deployFromFactory("NFTMetadataGeneratorSVG", provider, await get1559Fees(provider));
  const metadataGeneratorAddress = await metadataGenerator.getAddress();
  
  console.log(`NFT Metadata Generator deployed at: ${metadataGeneratorAddress}`);

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }

  // Deploy KarmaNFT
  const contract = await deployFromFactory(
    "KarmaNFT",
    provider,
    karmaAddress,
    metadataGeneratorAddress,
    await get1559Fees(provider)
  );

  const contractAddress = await contract.getAddress();
  await LogContractDeployment(contractName, contract);

  await tryStoreAddress(hre.network.name, contractName, contractAddress, contract.deploymentTransaction()!.hash);

  const args = [karmaAddress, metadataGeneratorAddress];
  await tryVerifyContractWithConstructorArgs(contractAddress, "contracts/src/KarmaNFT.sol:KarmaNFT", args);

  console.log(`KarmaNFT deployed with Karma at ${karmaAddress} and metadata generator at ${metadataGeneratorAddress}`);
};

export default func;
func.tags = ["StatusNetworkKarmaNFT"];
func.dependencies = ["StatusNetworkKarma"];
