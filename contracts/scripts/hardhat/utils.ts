import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { DeployProxyOptions } from "@openzeppelin/hardhat-upgrades/dist/utils";
import { AbstractSigner, ContractFactory, JsonRpcProvider, Provider } from "ethers";
import { ethers, upgrades } from "hardhat";
import { FactoryOptions, HardhatEthersHelpers } from "hardhat/types";
import { resolveDeploymentRunner, setDeploymentUiNextTransactionContext } from "./deployment-ui";

type RunnerOrProvider = AbstractSigner | Provider | JsonRpcProvider | HardhatEthersHelpers["provider"] | null;

function jsonSafeForUi(value: unknown): unknown {
  if (value === undefined) {
    return undefined;
  }
  if (value === null) {
    return null;
  }
  if (typeof value === "bigint") {
    return value.toString();
  }
  if (value instanceof Uint8Array) {
    return `0x${Buffer.from(value).toString("hex")}`;
  }
  if (Array.isArray(value)) {
    return value.map((item) => jsonSafeForUi(item));
  }
  if (typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(([key, item]) => [key, jsonSafeForUi(item)]),
    );
  }
  return value;
}

function openZeppelinProxyKindFromOpts(opts?: DeployProxyOptions): "transparent" | "uups" | "beacon" {
  const k = opts?.kind;
  if (k === "transparent" || k === "uups" || k === "beacon") {
    return k;
  }
  return "transparent";
}

function pushUiDeployContext(
  contractName: string,
  details: {
    constructorArgs?: unknown;
    initializerArgs?: unknown;
    proxyOptions?: string;
    notes?: string;
    openZeppelinProxyKind?: "transparent" | "uups" | "beacon";
  },
): void {
  if (process.env.DEPLOY_WITH_UI !== "true") {
    return;
  }

  setDeploymentUiNextTransactionContext({
    contractName,
    constructorArgs: details.constructorArgs,
    initializerArgs: details.initializerArgs,
    proxyOptions: details.proxyOptions,
    notes: details.notes,
    openZeppelinProxyKind: details.openZeppelinProxyKind,
  });
}

function tryStringifyProxyOpts(opts?: DeployProxyOptions): string | undefined {
  if (opts === undefined) {
    return undefined;
  }
  try {
    const serialized = JSON.stringify(jsonSafeForUi(opts), null, 2);
    return serialized.length > 1200 ? `${serialized.slice(0, 1200)}…` : serialized;
  } catch {
    return "(proxy options not serializable)";
  }
}

async function deployFromFactory(
  contractName: string,
  runnerOrProvider: RunnerOrProvider = null,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: any[]
) {
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  const runner = await resolveDeploymentRunner(runnerOrProvider);
  if (!skipLog) {
    const signerAddress = "getAddress" in runner ? await runner.getAddress() : undefined;
    console.log(`Going to deploy ${contractName} with account ${signerAddress}...`);
  }

  const factory = await ethers.getContractFactory(contractName, runner);
  pushUiDeployContext(contractName, { constructorArgs: jsonSafeForUi(args) });
  const contract = await factory.connect(runner).deploy(...args);
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
  runnerOrProvider: RunnerOrProvider = null,
  factoryOpts: FactoryOptions,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: any[]
) {
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  const runner = await resolveDeploymentRunner(runnerOrProvider);
  if (!skipLog) {
    const signerAddress = "getAddress" in runner ? await runner.getAddress() : undefined;
    console.log(`Going to deploy ${contractName} with account ${signerAddress}...`);
  }

  const factory = await ethers.getContractFactory(contractName, factoryOpts);
  pushUiDeployContext(contractName, { constructorArgs: jsonSafeForUi(args) });
  const contract = await factory.connect(runner).deploy(...args);
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

async function deployUpgradableFromFactory(
  contractName: string,
  args?: unknown[],
  opts?: DeployProxyOptions,
  factoryOpts?: FactoryOptions,
) {
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  const runner = await resolveDeploymentRunner();
  if (!skipLog) {
    console.log(`Going to deploy upgradable ${contractName}`);
  }
  const factory = factoryOpts
    ? await ethers.getContractFactory(contractName, factoryOpts)
    : await ethers.getContractFactory(contractName, runner);
  pushUiDeployContext(contractName, {
    initializerArgs: jsonSafeForUi(args ?? []),
    constructorArgs: jsonSafeForUi(opts?.constructorArgs),
    proxyOptions: tryStringifyProxyOpts(opts),
    openZeppelinProxyKind: openZeppelinProxyKindFromOpts(opts),
  });
  const contract = await upgrades.deployProxy(factory.connect(runner), args, opts);
  if (!skipLog) {
    const deployTx = contract.deploymentTransaction();
    console.log(`Upgradable ${contractName} deployment transaction has been sent, waiting...`, {
      hash: deployTx?.hash,
      gasPrice: deployTx?.gasPrice?.toString(),
      gasLimit: deployTx?.gasLimit.toString(),
    });
  }
  const afterDeploy = await contract.waitForDeployment();
  const timeDiff = performance.now() - startTime;
  if (!skipLog) {
    console.log(
      `${contractName} artifact has been deployed in ${timeDiff / 1000}s` +
        ` tx-hash=${afterDeploy.deploymentTransaction()?.hash}`,
    );
  }
  return contract;
}

async function deployUpgradableWithAbiAndByteCode(
  deployer: SignerWithAddress | AbstractSigner,
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
  const factory: ContractFactory = new ContractFactory(abi, byteCode, deployer);

  pushUiDeployContext(contractName, {
    initializerArgs: jsonSafeForUi(args ?? []),
    constructorArgs: jsonSafeForUi(opts?.constructorArgs),
    proxyOptions: tryStringifyProxyOpts(opts),
    openZeppelinProxyKind: openZeppelinProxyKindFromOpts(opts),
  });
  const contract = await upgrades.deployProxy(factory, args, opts);

  if (!skipLog) {
    const deployTx = contract.deploymentTransaction();
    console.log(`Upgradable ${contractName} deployment transaction has been sent, waiting...`, {
      hash: deployTx?.hash,
      gasPrice: deployTx?.gasPrice?.toString(),
      gasLimit: deployTx?.gasLimit.toString(),
    });
  }
  const afterDeploy = await contract.waitForDeployment();
  if (!skipLog) {
    console.log(`${contractName} artifact has been deployed in tx-hash=${afterDeploy.deploymentTransaction()?.hash}`);
  }
  return contract;
}

async function deployUpgradableFromFactoryWithConstructorArgs(
  contractName: string,
  constructorArgs: unknown[] = [],
  initializerArgs: unknown[] = [],
  opts: DeployProxyOptions = {},
  factoryOpts?: FactoryOptions,
) {
  const startTime = performance.now();
  const skipLog = process.env.SKIP_DEPLOY_LOG === "true" || false;
  const runner = await resolveDeploymentRunner();
  if (!skipLog) {
    console.log(`Going to deploy upgradable ${contractName}`);
  }
  const factory = factoryOpts
    ? await ethers.getContractFactory(contractName, factoryOpts)
    : await ethers.getContractFactory(contractName, runner);
  pushUiDeployContext(contractName, {
    constructorArgs: jsonSafeForUi(constructorArgs),
    initializerArgs: jsonSafeForUi(initializerArgs),
    proxyOptions: tryStringifyProxyOpts(opts),
    openZeppelinProxyKind: openZeppelinProxyKindFromOpts(opts),
  });
  const contract = await upgrades.deployProxy(factory.connect(runner), initializerArgs, {
    ...opts,
    constructorArgs,
  });
  if (!skipLog) {
    const deployTx = contract.deploymentTransaction();
    console.log(`Upgradable ${contractName} deployment transaction has been sent, waiting...`, {
      hash: deployTx?.hash,
      gasPrice: deployTx?.gasPrice?.toString(),
      gasLimit: deployTx?.gasLimit.toString(),
    });
  }
  const afterDeploy = await contract.waitForDeployment();
  const timeDiff = performance.now() - startTime;
  if (!skipLog) {
    console.log(
      `${contractName} artifact has been deployed in ${timeDiff / 1000}s` +
        ` tx-hash=${afterDeploy.deploymentTransaction()?.hash}`,
    );
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

export {
  deployFromFactory,
  deployFromFactoryWithOpts,
  deployUpgradableFromFactory,
  deployUpgradableWithAbiAndByteCode,
  deployUpgradableFromFactoryWithConstructorArgs,
  requireEnv,
};
