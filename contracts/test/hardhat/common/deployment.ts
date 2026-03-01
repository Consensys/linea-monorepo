import { Interface, Signer, BaseContract, ContractFactory } from "ethers";
import type { FactoryOptions } from "hardhat/types";
import { ethers } from "./hardhat-connection.js";

import ProxyAdminArtifact from "../../../deployments/bytecode/mainnet-proxy/ProxyAdmin.json" with { type: "json" };
import TransparentUpgradeableProxyArtifact from "../../../deployments/bytecode/mainnet-proxy/TransparentUpgradeableProxy.json" with { type: "json" };

export interface DeployProxyOptions {
  initializer?: string;
  constructorArgs?: unknown[];
  unsafeAllow?: string[];
}

async function deployFromFactory(contractName: string, ...args: unknown[]) {
  const factory = await ethers.getContractFactory(contractName);
  const contract = await factory.deploy(...args);
  await contract.waitForDeployment();
  return contract;
}

async function deployProxyAdmin(signer: Signer): Promise<BaseContract> {
  const factory = new ContractFactory(ProxyAdminArtifact.abi, ProxyAdminArtifact.bytecode, signer);
  const proxyAdmin = await factory.deploy();
  await proxyAdmin.waitForDeployment();
  return proxyAdmin;
}

async function deployTransparentProxy(
  implementationAddress: string,
  proxyAdminAddress: string,
  initData: string,
  signer: Signer,
): Promise<BaseContract> {
  const factory = new ContractFactory(
    TransparentUpgradeableProxyArtifact.abi,
    TransparentUpgradeableProxyArtifact.bytecode,
    signer,
  );
  const proxy = await factory.deploy(implementationAddress, proxyAdminAddress, initData);
  await proxy.waitForDeployment();
  return proxy;
}

function encodeInitializerData(
  contractInterface: Interface,
  initializerName: string | undefined,
  args: unknown[],
): string {
  if (!initializerName || initializerName === "") {
    return "0x";
  }

  const functionName = initializerName.includes("(") ? initializerName.split("(")[0] : initializerName;

  const fragment = contractInterface.getFunction(functionName);
  if (!fragment) {
    return "0x";
  }

  return contractInterface.encodeFunctionData(fragment, args);
}

async function deployUpgradableFromFactory(
  contractName: string,
  args?: unknown[],
  opts?: DeployProxyOptions,
  factoryOpts?: FactoryOptions,
) {
  const signers = await ethers.getSigners();
  const deployer = signers[0];

  const factory = await ethers.getContractFactory(contractName, factoryOpts);

  const implementation = await factory.deploy(...(opts?.constructorArgs || []));
  await implementation.waitForDeployment();
  const implementationAddress = await implementation.getAddress();

  const proxyAdmin = await deployProxyAdmin(deployer);
  const proxyAdminAddress = await proxyAdmin.getAddress();

  const initData = encodeInitializerData(factory.interface, opts?.initializer || "initialize", args || []);

  const proxy = await deployTransparentProxy(implementationAddress, proxyAdminAddress, initData, deployer);
  const proxyAddress = await proxy.getAddress();

  const contract = factory.attach(proxyAddress) as BaseContract;
  return contract;
}

async function deployUpgradableWithConstructorArgs(
  contractName: string,
  constructorArgs: unknown[] = [],
  initializerArgs: unknown[] = [],
  opts: DeployProxyOptions = {},
  factoryOpts?: FactoryOptions,
) {
  return deployUpgradableFromFactory(
    contractName,
    initializerArgs,
    {
      ...opts,
      constructorArgs,
    },
    factoryOpts,
  );
}

export { deployFromFactory, deployUpgradableFromFactory, deployUpgradableWithConstructorArgs };
