import { DeployProxyOptions } from "@openzeppelin/hardhat-upgrades/src/utils";
import { ethers, upgrades } from "hardhat";
import { FactoryOptions } from "hardhat/types";

async function deployFromFactory(contractName: string, ...args: unknown[]) {
  const factory = await ethers.getContractFactory(contractName);
  const contract = await factory.deploy(...args);
  await contract.deployed();
  return contract;
}

async function deployUpgradableFromFactory(
  contractName: string,
  args?: unknown[],
  opts?: DeployProxyOptions,
  factoryOpts?: FactoryOptions,
) {
  const factory = await ethers.getContractFactory(contractName, factoryOpts);
  const contract = await upgrades.deployProxy(factory, args, opts);
  await contract.deployed();
  return contract;
}

export { deployFromFactory, deployUpgradableFromFactory };
