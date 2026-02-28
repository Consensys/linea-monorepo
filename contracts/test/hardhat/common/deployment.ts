import { Contract, ContractFactory } from "ethers";
import { ethers } from "./connection.js";

interface DeployProxyOptions {
  initializer?: string;
  unsafeAllow?: string[];
  constructorArgs?: unknown[];
  kind?: string;
}

interface FactoryOptions {
  signer?: unknown;
  libraries?: Record<string, string>;
}

async function deployFromFactory(contractName: string, ...args: unknown[]) {
  const factory = await ethers.getContractFactory(contractName);
  const contract = await factory.deploy(...args);
  await contract.waitForDeployment();
  return contract;
}

async function deployUpgradableFromFactory(
  contractName: string,
  args?: unknown[],
  opts?: DeployProxyOptions,
  factoryOpts?: FactoryOptions,
) {
  const factory = await ethers.getContractFactory(
    contractName,
    factoryOpts as Parameters<typeof ethers.getContractFactory>[1],
  );
  return deployTransparentProxy(factory, args, opts);
}

async function deployUpgradableWithConstructorArgs(
  contractName: string,
  constructorArgs: unknown[] = [],
  initializerArgs: unknown[] = [],
  opts: DeployProxyOptions = {},
  factoryOpts?: FactoryOptions,
) {
  const factory = await ethers.getContractFactory(
    contractName,
    factoryOpts as Parameters<typeof ethers.getContractFactory>[1],
  );
  return deployTransparentProxy(factory, initializerArgs, { ...opts, constructorArgs });
}

async function deployTransparentProxy(
  factory: ContractFactory,
  args?: unknown[],
  opts?: DeployProxyOptions,
): Promise<Contract> {
  const constructorArgs = opts?.constructorArgs ?? [];
  const implementation = await factory.deploy(...constructorArgs);
  await implementation.waitForDeployment();
  const implementationAddress = await implementation.getAddress();

  const proxyAdminFactory = await ethers.getContractFactory("src/_testing/integration/ProxyAdmin.sol:ProxyAdmin");
  const proxyAdmin = await proxyAdminFactory.deploy();
  await proxyAdmin.waitForDeployment();
  const proxyAdminAddress = await proxyAdmin.getAddress();

  let initData = "0x";
  if (args && args.length > 0) {
    const initializerName = opts?.initializer ?? "initialize";
    const iface = factory.interface;
    const fragment = iface.getFunction(initializerName);
    if (fragment) {
      initData = iface.encodeFunctionData(fragment, args);
    }
  }

  const proxyFactory = await ethers.getContractFactory(
    "src/_testing/integration/TransparentUpgradeableProxy.sol:TransparentUpgradeableProxy",
  );
  const proxy = await proxyFactory.deploy(implementationAddress, proxyAdminAddress, initData);
  await proxy.waitForDeployment();
  const proxyAddress = await proxy.getAddress();

  const contract = factory.attach(proxyAddress) as Contract;
  const proxyDeployTx = proxy.deploymentTransaction();
  if (proxyDeployTx) {
    (contract as unknown as { deploymentTransaction: () => typeof proxyDeployTx }).deploymentTransaction = () =>
      proxyDeployTx;
  }
  return contract;
}

async function upgradeProxy(
  proxyAddress: string,
  newFactory: ContractFactory,
  opts?: {
    call?: { fn: string; args: unknown[] };
    unsafeAllow?: string[];
    unsafeAllowRenames?: boolean;
    constructorArgs?: unknown[];
  },
): Promise<Contract> {
  const constructorArgs = opts?.constructorArgs ?? [];
  const newImplementation = await newFactory.deploy(...constructorArgs);
  await newImplementation.waitForDeployment();
  const newImplementationAddress = await newImplementation.getAddress();

  const adminSlot = "0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103";
  const adminStorageValue = await ethers.provider.getStorage(proxyAddress, adminSlot);
  const proxyAdminAddress = "0x" + adminStorageValue.slice(26);

  const proxyAdmin = await ethers.getContractAt(
    "src/_testing/integration/ProxyAdmin.sol:ProxyAdmin",
    proxyAdminAddress,
  );

  if (opts?.call) {
    const callData = newFactory.interface.encodeFunctionData(opts.call.fn, opts.call.args);
    const [signer] = await ethers.getSigners();
    await signer.sendTransaction({
      to: proxyAdminAddress,
      data: proxyAdmin.interface.encodeFunctionData("upgradeAndCall", [
        proxyAddress,
        newImplementationAddress,
        callData,
      ]),
    });
  } else {
    const [signer] = await ethers.getSigners();
    await signer.sendTransaction({
      to: proxyAdminAddress,
      data: proxyAdmin.interface.encodeFunctionData("upgrade", [proxyAddress, newImplementationAddress]),
    });
  }

  const contract = newFactory.attach(proxyAddress);
  return contract as Contract;
}

export {
  deployFromFactory,
  deployUpgradableFromFactory,
  deployUpgradableWithConstructorArgs,
  deployTransparentProxy,
  upgradeProxy,
};
