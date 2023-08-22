import { ethers, upgrades } from "hardhat";
import { providers } from "ethers";
import { DeployProxyOptions } from "@openzeppelin/hardhat-upgrades/dist/utils";
import { FactoryOptions } from "hardhat/types";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function deployFromFactory(
  contractName: string,
  provider: providers.JsonRpcProvider | null = null,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: any[]
) {
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  if (!skipLog) {
    console.log(`Going to deploy ${contractName}`);
  }

  const factory = await ethers.getContractFactory(contractName);
  if (provider) {
    factory.connect(provider.getSigner());
  }
  const contract = await factory.deploy(...args);
  if (!skipLog) {
    console.log(`${contractName} deployment transaction has been sent, waiting...`, {
      nonce: contract.deployTransaction.nonce,
      hash: contract.deployTransaction.hash,
      gasPrice: contract.deployTransaction.gasPrice?.toString(),
      maxFeePerGas: contract.deployTransaction.maxFeePerGas?.toString(),
      maxPriorityFeePerGas: contract.deployTransaction.maxPriorityFeePerGas?.toString(),
      gasLimit: contract.deployTransaction.gasLimit.toString(),
    });
  }
  const afterDeploy = await contract.deployed();
  if (!skipLog) {
    console.log(`${contractName} artifact has been deployed in tx-hash=${afterDeploy.deployTransaction.hash}`);
  }
  return contract;
}

async function deployUpgradableFromFactory(
  contractName: string,
  args?: unknown[],
  opts?: DeployProxyOptions,
  factoryOpts?: FactoryOptions,
) {
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  if (!skipLog) {
    console.log(`Going to deploy upgradable ${contractName}`);
  }
  const factory = await ethers.getContractFactory(contractName, factoryOpts);
  const contract = await upgrades.deployProxy(factory, args, opts);
  if (!skipLog) {
    console.log(`Upgradable ${contractName} deployment transaction has been sent, waiting...`, {
      hash: contract.deployTransaction.hash,
      gasPrice: contract.deployTransaction.gasPrice?.toString(),
      gasLimit: contract.deployTransaction.gasLimit.toString(),
    });
  }
  const afterDeploy = await contract.deployed();
  if (!skipLog) {
    console.log(`${contractName} artifact has been deployed in tx-hash=${afterDeploy.deployTransaction.hash}`);
  }
  return contract;
}

function requireEnv(name: string): string {
  const envVariable = process.env[name];
  if (!envVariable) {
    throw new Error(`Missing ${name} environment variable`);
  }

  return envVariable;
}

export { deployFromFactory, deployUpgradableFromFactory, requireEnv };
