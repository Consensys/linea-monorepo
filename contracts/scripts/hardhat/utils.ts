import { ContractFactory, JsonRpcProvider, Interface, BaseContract, Signer } from "ethers";
import { ethers } from "hardhat";
import { FactoryOptions, HardhatEthersHelpers } from "hardhat/types";

import ProxyAdminArtifact from "../../deployments/bytecode/mainnet-proxy/ProxyAdmin.json" with { type: "json" };
import TransparentUpgradeableProxyArtifact from "../../deployments/bytecode/mainnet-proxy/TransparentUpgradeableProxy.json" with { type: "json" };

export interface DeployProxyOptions {
  initializer?: string;
  constructorArgs?: unknown[];
  unsafeAllow?: string[];
}

async function deployFromFactory(
  contractName: string,
  provider: JsonRpcProvider | HardhatEthersHelpers["provider"] | null = null,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: any[]
) {
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  if (!skipLog) {
    const signer = await provider?.getSigner();
    console.log(`Going to deploy ${contractName} with account ${await signer?.getAddress()}...`);
  }

  const factory = await ethers.getContractFactory(contractName);
  if (provider) {
    factory.connect(await provider.getSigner());
  }
  const contract = await factory.deploy(...args);
  if (!skipLog) {
    const deployTx = contract.deploymentTransaction();

    console.log(`${contractName} deployment transaction has been sent, waiting...`, {
      nonce: deployTx?.nonce,
      hash: deployTx?.hash,
      gasPrice: deployTx?.gasPrice?.toString(),
      maxFeePerGas: deployTx?.maxFeePerGas?.toString(),
      maxPriorityFeePerGas: deployTx?.maxPriorityFeePerGas?.toString(),
      gasLimit: deployTx?.gasLimit.toString(),
    });
  }
  const afterDeploy = await contract.waitForDeployment();
  const timeDiff = performance.now() - startTime;
  if (!skipLog) {
    console.log(
      `${contractName} deployed: time=${timeDiff / 1000}s blockNumber=${afterDeploy.deploymentTransaction()?.blockNumber}` +
        ` tx-hash=${afterDeploy.deploymentTransaction()?.hash}`,
    );
  }
  return contract;
}

async function deployFromFactoryWithOpts(
  contractName: string,
  provider: JsonRpcProvider | HardhatEthersHelpers["provider"] | null = null,
  factoryOpts: FactoryOptions,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: any[]
) {
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  if (!skipLog) {
    const signer = await provider?.getSigner();
    console.log(`Going to deploy ${contractName} with account ${await signer?.getAddress()}...`);
  }

  const factory = await ethers.getContractFactory(contractName, factoryOpts);
  if (provider) {
    factory.connect(await provider.getSigner());
  }
  const contract = await factory.deploy(...args);
  if (!skipLog) {
    const deployTx = contract.deploymentTransaction();

    console.log(`${contractName} deployment transaction has been sent, waiting...`, {
      nonce: deployTx?.nonce,
      hash: deployTx?.hash,
      gasPrice: deployTx?.gasPrice?.toString(),
      maxFeePerGas: deployTx?.maxFeePerGas?.toString(),
      maxPriorityFeePerGas: deployTx?.maxPriorityFeePerGas?.toString(),
      gasLimit: deployTx?.gasLimit.toString(),
    });
  }
  const afterDeploy = await contract.waitForDeployment();
  const timeDiff = performance.now() - startTime;
  if (!skipLog) {
    console.log(
      `${contractName} deployed: time=${timeDiff / 1000}s blockNumber=${afterDeploy.deploymentTransaction()?.blockNumber}` +
        ` tx-hash=${afterDeploy.deploymentTransaction()?.hash}`,
    );
  }
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
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  if (!skipLog) {
    console.log(`Going to deploy upgradable ${contractName}`);
  }

  const signers = await ethers.getSigners();
  const deployer = signers[0];

  const factory = await ethers.getContractFactory(contractName, factoryOpts);

  const implementation = await factory.deploy(...(opts?.constructorArgs || []));
  await implementation.waitForDeployment();
  const implementationAddress = await implementation.getAddress();

  const proxyAdmin = await deployProxyAdmin(deployer);
  const proxyAdminAddress = await proxyAdmin.getAddress();

  const initData = encodeInitializerData(factory.interface, opts?.initializer, args || []);

  const proxy = await deployTransparentProxy(implementationAddress, proxyAdminAddress, initData, deployer);
  const proxyAddress = await proxy.getAddress();

  const timeDiff = performance.now() - startTime;
  if (!skipLog) {
    console.log(`${contractName} artifact has been deployed in ${timeDiff / 1000}s at ${proxyAddress}`);
  }

  const contract = factory.attach(proxyAddress) as BaseContract;
  return contract;
}

async function deployUpgradableWithAbiAndByteCode(
  deployer: Signer,
  contractName: string,
  abi: string,
  byteCode: string,
  args?: unknown[],
  opts?: DeployProxyOptions,
) {
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  if (!skipLog) {
    console.log(`Going to deploy upgradable ${contractName}`);
  }

  const factory = new ContractFactory(abi, byteCode, deployer);

  const implementation = await factory.deploy(...(opts?.constructorArgs || []));
  await implementation.waitForDeployment();
  const implementationAddress = await implementation.getAddress();

  const proxyAdmin = await deployProxyAdmin(deployer);
  const proxyAdminAddress = await proxyAdmin.getAddress();

  const initData = encodeInitializerData(factory.interface, opts?.initializer, args || []);

  const proxy = await deployTransparentProxy(implementationAddress, proxyAdminAddress, initData, deployer);
  const proxyAddress = await proxy.getAddress();

  if (!skipLog) {
    console.log(`${contractName} artifact has been deployed at ${proxyAddress}`);
  }

  const contract = factory.attach(proxyAddress) as BaseContract;
  return contract;
}

async function deployUpgradableFromFactoryWithConstructorArgs(
  contractName: string,
  constructorArgs: unknown[] = [],
  initializerArgs: unknown[] = [],
  opts: DeployProxyOptions = {},
  factoryOpts?: FactoryOptions,
) {
  return deployUpgradableFromFactory(contractName, initializerArgs, {
    ...opts,
    constructorArgs,
  }, factoryOpts);
}

function requireEnv(name: string): string {
  const envVariable = process.env[name];
  if (!envVariable) {
    throw new Error(`Missing ${name} environment variable`);
  }

  return envVariable;
}

export {
  deployFromFactory,
  deployFromFactoryWithOpts,
  deployUpgradableFromFactory,
  deployUpgradableWithAbiAndByteCode,
  deployUpgradableFromFactoryWithConstructorArgs,
  requireEnv,
};
