import { DeployFunction } from "hardhat-deploy/types";
import { HardhatRuntimeEnvironment } from "hardhat/types";
import { deployUpgradableWithAbiAndByteCode, requireEnv } from "../scripts/hardhat/utils";
import { validateDeployBranchAndTags } from "../utils/auditedDeployVerifier";
import { getDeployedContractAddress, tryStoreAddress } from "../utils/storeAddress";
import { tryVerifyContract } from "../utils/verifyContract";

import fs from "fs";
import { ethers } from "hardhat";
import path from "path";
import { abi, bytecode } from "./V1/ZkEvmV2Deployed.json";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments } = hre;
  validateDeployBranchAndTags(hre.network.name);

  const mainnetDeployedZkEvmCacheFolder = path.resolve("./deploy/V1/ZkEvmV2Cache/");

  const validationFilePath = path.join(hre.config.paths.cache, "validations.json");
  const validationFileBackupPath = path.join(hre.config.paths.cache, "validations_backup.json");

  if (fs.existsSync(validationFilePath)) {
    fs.copyFileSync(validationFilePath, validationFileBackupPath);
  }

  fs.copyFileSync(path.join(mainnetDeployedZkEvmCacheFolder, "validations.json"), validationFilePath);

  const contractName = "ZkEvmV2Mainnet";
  const verifierName = "PlonkVerifier";
  const existingContractAddress = await getDeployedContractAddress(contractName, deployments);
  let verifierAddress = await getDeployedContractAddress(verifierName, deployments);
  if (verifierAddress === undefined) {
    if (process.env["PLONKVERIFIER_ADDRESS"] !== undefined) {
      console.log(`Using environment variable for PlonkVerifier , ${process.env["PLONKVERIFIER_ADDRESS"]}`);
      verifierAddress = process.env["PLONKVERIFIER_ADDRESS"];
    } else {
      throw "Missing PLONKVERIFIER_ADDRESS environment variable";
    }
  } else {
    console.log(`Using deployed variable for PlonkVerifier , ${verifierAddress}`);
  }

  // LineaRollup DEPLOYED AS UPGRADEABLE PROXY
  const LineaRollup_initialStateRootHash = requireEnv("LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH");
  const LineaRollup_initialL2BlockNumber = requireEnv("LINEA_ROLLUP_INITIAL_L2_BLOCK_NUMBER");
  const LineaRollup_securityCouncil = requireEnv("LINEA_ROLLUP_SECURITY_COUNCIL");
  const LineaRollup_operators = requireEnv("LINEA_ROLLUP_OPERATORS");
  const LineaRollup_rateLimitPeriodInSeconds = requireEnv("LINEA_ROLLUP_RATE_LIMIT_PERIOD");
  const LineaRollup_rateLimitAmountInWei = requireEnv("LINEA_ROLLUP_RATE_LIMIT_AMOUNT");

  console.log(`Setting operators ${LineaRollup_operators}`);

  const [deployer] = await ethers.getSigners();

  if (existingContractAddress === undefined) {
    console.log(`Deploying initial version, NB: the address will be saved if env SAVE_ADDRESS=true.`);
  } else {
    console.log(`Deploying new version, NB: ${existingContractAddress} will be overwritten if env SAVE_ADDRESS=true.`);
  }
  const contract = await deployUpgradableWithAbiAndByteCode(
    deployer,
    "ZkEvmV2Mainnet",
    JSON.stringify(abi),
    bytecode,
    [
      LineaRollup_initialStateRootHash,
      LineaRollup_initialL2BlockNumber,
      verifierAddress,
      LineaRollup_securityCouncil,
      LineaRollup_operators?.split(","),
      LineaRollup_rateLimitPeriodInSeconds,
      LineaRollup_rateLimitAmountInWei,
    ],
    {
      initializer: "initialize(bytes32,uint256,address,address,address[],uint256,uint256)",
      unsafeAllow: ["constructor"],
    },
  );

  const contractAddress = await contract.getAddress();
  const txReceipt = await contract.deploymentTransaction()?.wait();
  if (!txReceipt) {
    throw "Contract deployment transaction receipt not found.";
  }

  console.log(`${contractName} deployed at ${contractAddress}`);

  await tryStoreAddress(hre.network.name, contractName, contractAddress, txReceipt.hash);

  await tryVerifyContract(contractAddress);

  fs.unlinkSync(path.join(hre.config.paths.cache, "validations.json"));
  if (fs.existsSync(validationFileBackupPath)) {
    fs.copyFileSync(validationFileBackupPath, validationFilePath);
    fs.unlinkSync(validationFileBackupPath);
  }
};

export default func;
func.tags = ["ZkEvmV2Mainnet"];
