import hre from "hardhat";
import { ContractFactory, Contract } from "ethers";

const { ethers } = await hre.network.connect();

async function deployFromFactory(
  contractName: string,
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  _provider: unknown = null,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: any[]
) {
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  if (!skipLog) {
    const signer = await ethers.provider?.getSigner();
    console.log(`Going to deploy ${contractName} with account ${await signer?.getAddress()}...`);
  }

  const factory = await ethers.getContractFactory(contractName);
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
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  _provider: unknown = null,
  factoryOpts: Parameters<typeof ethers.getContractFactory>[1],
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: any[]
) {
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  if (!skipLog) {
    const signer = await ethers.provider?.getSigner();
    console.log(`Going to deploy ${contractName} with account ${await signer?.getAddress()}...`);
  }

  const factory = await ethers.getContractFactory(contractName, factoryOpts);
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

interface DeployProxyOptions {
  initializer?: string;
  unsafeAllow?: string[];
  constructorArgs?: unknown[];
  kind?: string;
  txOverrides?: Record<string, unknown>;
}

async function deployUpgradableFromFactory(
  contractName: string,
  args?: unknown[],
  opts?: DeployProxyOptions,
  factoryOpts?: Parameters<typeof ethers.getContractFactory>[1],
) {
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  if (!skipLog) {
    console.log(`Going to deploy upgradable ${contractName}`);
  }
  const factory = await ethers.getContractFactory(contractName, factoryOpts);

  const constructorArgs = opts?.constructorArgs ?? [];
  const implementation = await factory.deploy(...constructorArgs);
  await implementation.waitForDeployment();
  const implementationAddress = await implementation.getAddress();

  const proxyAdminFactory = await ethers.getContractFactory("ProxyAdmin");
  const proxyAdmin = await proxyAdminFactory.deploy();
  await proxyAdmin.waitForDeployment();
  const proxyAdminAddress = await proxyAdmin.getAddress();

  let initData = "0x";
  if (args && args.length > 0) {
    const initializerName = opts?.initializer ?? "initialize";
    const fragment = factory.interface.getFunction(initializerName);
    if (fragment) {
      initData = factory.interface.encodeFunctionData(fragment, args);
    }
  }

  const proxyFactory = await ethers.getContractFactory("TransparentUpgradeableProxy");
  const proxy = await proxyFactory.deploy(implementationAddress, proxyAdminAddress, initData);
  if (!skipLog) {
    const deployTx = proxy.deploymentTransaction();
    console.log(`Upgradable ${contractName} deployment transaction has been sent, waiting...`, {
      hash: deployTx?.hash,
      gasPrice: deployTx?.gasPrice?.toString(),
      gasLimit: deployTx?.gasLimit.toString(),
    });
  }
  await proxy.waitForDeployment();
  const proxyAddress = await proxy.getAddress();
  const timeDiff = performance.now() - startTime;
  if (!skipLog) {
    console.log(`${contractName} artifact has been deployed in ${timeDiff / 1000}s` + ` proxy=${proxyAddress}`);
  }
  return factory.attach(proxyAddress) as Contract;
}

async function deployUpgradableWithAbiAndByteCode(
  deployer: unknown,
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
  const [signer] = await ethers.getSigners();
  const factory: ContractFactory = new ContractFactory(abi, byteCode, signer);

  const constructorArgs = opts?.constructorArgs ?? [];
  const implementation = await factory.deploy(...constructorArgs);
  await implementation.waitForDeployment();
  const implementationAddress = await implementation.getAddress();

  const proxyAdminFactory = await ethers.getContractFactory("ProxyAdmin");
  const proxyAdmin = await proxyAdminFactory.deploy();
  await proxyAdmin.waitForDeployment();
  const proxyAdminAddress = await proxyAdmin.getAddress();

  let initData = "0x";
  if (args && args.length > 0) {
    const initializerName = opts?.initializer ?? "initialize";
    const fragment = factory.interface.getFunction(initializerName);
    if (fragment) {
      initData = factory.interface.encodeFunctionData(fragment, args);
    }
  }

  const proxyFactory = await ethers.getContractFactory("TransparentUpgradeableProxy");
  const proxy = await proxyFactory.deploy(implementationAddress, proxyAdminAddress, initData);
  await proxy.waitForDeployment();
  const proxyAddress = await proxy.getAddress();

  if (!skipLog) {
    console.log(`${contractName} artifact has been deployed proxy=${proxyAddress}`);
  }
  return factory.attach(proxyAddress) as Contract;
}

async function deployUpgradableFromFactoryWithConstructorArgs(
  contractName: string,
  constructorArgs: unknown[] = [],
  initializerArgs: unknown[] = [],
  opts: DeployProxyOptions = {},
  factoryOpts?: Parameters<typeof ethers.getContractFactory>[1],
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
