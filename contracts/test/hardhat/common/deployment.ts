import { upgrades as createUpgrades } from "@openzeppelin/hardhat-upgrades";
import { getInitializerData } from "contracts/common/helpers";
import hre, { network as hardhatNetwork } from "hardhat";

import type { FactoryOptions } from "@nomicfoundation/hardhat-ethers/types";
import type { UpgradeOptions } from "@openzeppelin/hardhat-upgrades";
import type { ProxyAdmin } from "contracts/typechain-types";
import type { BaseContract, ContractFactory, ContractTransactionResponse, InterfaceAbi, Overrides } from "ethers";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const upgrades = await createUpgrades(hre, hardhatConnection);

type DeployProxyOptions = UpgradeOptions & {
  initializer?: string | false;
  initialOwner?: string;
  unsafeSkipProxyAdminCheck?: boolean;
  txOverrides?: Overrides;
  proxyFactory?: ContractFactory;
};

async function deployFromFactory<TContract extends BaseContract = BaseContract>(
  contractName: string,
  ...args: unknown[]
): Promise<TContract> {
  const factory = await ethers.getContractFactory(contractName);
  const contract = await factory.deploy(...args);
  await contract.waitForDeployment();
  return contract as unknown as TContract;
}

async function deployUpgradableFromFactory(
  contractName: string,
  args?: unknown[],
  opts?: DeployProxyOptions,
  factoryOpts?: FactoryOptions,
) {
  const factory = await ethers.getContractFactory(contractName, factoryOpts);
  const contract = await upgrades.deployProxy(factory, args, opts);
  await contract.waitForDeployment();
  return contract;
}

// Use constructor in upgradeable contract to set immutable for static global constants.
async function deployUpgradableWithConstructorArgs(
  contractName: string,
  constructorArgs: unknown[] = [],
  initializerArgs: unknown[] = [],
  opts: DeployProxyOptions = {},
  factoryOpts?: FactoryOptions,
) {
  const factory = await ethers.getContractFactory(contractName, factoryOpts);

  const contract = await upgrades.deployProxy(factory, initializerArgs, {
    ...opts,
    constructorArgs,
  });

  await contract.waitForDeployment();
  return contract;
}

/**
 * Reinitializes an upgradeable proxy contract by calling upgradeAndCall on the ProxyAdmin.
 *
 * @param proxyContract - The proxy contract instance to reinitialize
 * @param contractAbi - The ABI of the contract (used to encode the initializer call)
 * @param initializerFunctionName - The name of the reinitializer function to call
 * @param initializerArgs - The arguments to pass to the reinitializer function
 */
async function reinitializeUpgradeableProxy(
  proxyContract: BaseContract,
  contractAbi: InterfaceAbi,
  initializerFunctionName: string,
  initializerArgs: unknown[],
): Promise<ContractTransactionResponse> {
  const proxyAddress = await proxyContract.getAddress();
  const proxyAdminAddress = await upgrades.erc1967.getAdminAddress(proxyAddress);
  const proxyAdmin = (await ethers.getContractAt("ProxyAdmin", proxyAdminAddress)) as unknown as ProxyAdmin;
  const implementationAddress = await upgrades.erc1967.getImplementationAddress(proxyAddress);

  const encodedCall = getInitializerData(contractAbi, initializerFunctionName, initializerArgs);

  return await proxyAdmin.upgradeAndCall(proxyAddress, implementationAddress, encodedCall);
}

export {
  deployFromFactory,
  deployUpgradableFromFactory,
  deployUpgradableWithConstructorArgs,
  reinitializeUpgradeableProxy,
};
