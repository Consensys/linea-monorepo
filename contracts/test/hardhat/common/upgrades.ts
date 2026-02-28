import hre from "hardhat";
import { BaseContract, ContractFactory, Signer, Interface, AddressLike } from "ethers";

import ProxyAdminArtifact from "../../../deployments/bytecode/mainnet-proxy/ProxyAdmin.json" with { type: "json" };
import TransparentUpgradeableProxyArtifact from "../../../deployments/bytecode/mainnet-proxy/TransparentUpgradeableProxy.json" with { type: "json" };
import UpgradeableBeaconArtifact from "@openzeppelin/contracts/build/contracts/UpgradeableBeacon.json" with { type: "json" };
import BeaconProxyArtifact from "@openzeppelin/contracts/build/contracts/BeaconProxy.json" with { type: "json" };

export interface DeployProxyOptions {
  initializer?: string;
  unsafeAllow?: string[];
  constructorArgs?: unknown[];
}

export interface UpgradeProxyOptions {
  call?: {
    fn: string;
    args?: unknown[];
  };
  unsafeAllow?: string[];
  constructorArgs?: unknown[];
}

interface ProxyInfo {
  implementation: string;
  proxyAdmin: string;
  proxy: string;
}

const proxyRegistry = new Map<string, ProxyInfo>();
const beaconRegistry = new Map<string, { implementation: string; beacon: string }>();

async function getEthers() {
  const connection = await hre.network.connect();
  return connection.ethers;
}

async function getDeployer(): Promise<Signer> {
  const ethers = await getEthers();
  const signers = await ethers.getSigners();
  return signers[0];
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

async function deployProxyAdminContract(signer: Signer): Promise<BaseContract> {
  const factory = new ContractFactory(ProxyAdminArtifact.abi, ProxyAdminArtifact.bytecode, signer);
  const proxyAdmin = await factory.deploy();
  await proxyAdmin.waitForDeployment();
  return proxyAdmin;
}

async function deployTransparentProxyContract(
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

export async function deployProxy(
  factory: ContractFactory,
  args: unknown[] = [],
  opts: DeployProxyOptions = {},
): Promise<BaseContract> {
  const deployer = await getDeployer();

  const implementation = await factory.deploy(...(opts.constructorArgs || []));
  await implementation.waitForDeployment();
  const implementationAddress = await implementation.getAddress();

  const proxyAdmin = await deployProxyAdminContract(deployer);
  const proxyAdminAddress = await proxyAdmin.getAddress();

  const initData = encodeInitializerData(factory.interface, opts.initializer || "initialize", args);

  const proxy = await deployTransparentProxyContract(implementationAddress, proxyAdminAddress, initData, deployer);
  const proxyAddress = await proxy.getAddress();

  proxyRegistry.set(proxyAddress.toLowerCase(), {
    implementation: implementationAddress,
    proxyAdmin: proxyAdminAddress,
    proxy: proxyAddress,
  });

  return factory.attach(proxyAddress) as BaseContract;
}

export async function upgradeProxy(
  proxy: AddressLike | BaseContract,
  newFactory: ContractFactory,
  opts: UpgradeProxyOptions = {},
): Promise<BaseContract> {
  const deployer = await getDeployer();

  let proxyAddress: string;
  if (typeof proxy === "string") {
    proxyAddress = proxy;
  } else if ("getAddress" in proxy && typeof proxy.getAddress === "function") {
    proxyAddress = await proxy.getAddress();
  } else {
    proxyAddress = proxy as string;
  }

  const info = proxyRegistry.get(proxyAddress.toLowerCase());
  if (!info) {
    throw new Error(`Proxy at ${proxyAddress} was not deployed through this upgrades module`);
  }

  const newImplementation = await newFactory.deploy(...(opts.constructorArgs || []));
  await newImplementation.waitForDeployment();
  const newImplementationAddress = await newImplementation.getAddress();

  const proxyAdminFactory = new ContractFactory(ProxyAdminArtifact.abi, ProxyAdminArtifact.bytecode, deployer);
  const proxyAdmin = proxyAdminFactory.attach(info.proxyAdmin);

  if (opts.call) {
    const callData = newFactory.interface.encodeFunctionData(opts.call.fn, opts.call.args || []);
    const tx = await (
      proxyAdmin as BaseContract & { upgradeAndCall: (proxy: string, impl: string, data: string) => Promise<unknown> }
    ).upgradeAndCall(proxyAddress, newImplementationAddress, callData);
    await (tx as { wait: () => Promise<unknown> }).wait();
  } else {
    const tx = await (
      proxyAdmin as BaseContract & { upgrade: (proxy: string, impl: string) => Promise<unknown> }
    ).upgrade(proxyAddress, newImplementationAddress);
    await (tx as { wait: () => Promise<unknown> }).wait();
  }

  proxyRegistry.set(proxyAddress.toLowerCase(), {
    ...info,
    implementation: newImplementationAddress,
  });

  return newFactory.attach(proxyAddress) as BaseContract;
}

export async function deployBeacon(factory: ContractFactory): Promise<BaseContract> {
  const deployer = await getDeployer();

  const implementation = await factory.deploy();
  await implementation.waitForDeployment();
  const implementationAddress = await implementation.getAddress();

  const beaconFactory = new ContractFactory(
    UpgradeableBeaconArtifact.abi,
    UpgradeableBeaconArtifact.bytecode,
    deployer,
  );
  // OZ 4.9.6 UpgradeableBeacon takes only implementation address
  const beacon = await beaconFactory.deploy(implementationAddress);
  await beacon.waitForDeployment();
  const beaconAddress = await beacon.getAddress();

  beaconRegistry.set(beaconAddress.toLowerCase(), {
    implementation: implementationAddress,
    beacon: beaconAddress,
  });

  return beacon;
}

export async function deployBeaconProxy(
  beaconAddress: string,
  factory: ContractFactory,
  args: unknown[] = [],
): Promise<BaseContract> {
  const deployer = await getDeployer();

  const initData = encodeInitializerData(factory.interface, "initialize", args);

  const beaconProxyFactory = new ContractFactory(BeaconProxyArtifact.abi, BeaconProxyArtifact.bytecode, deployer);
  const beaconProxy = await beaconProxyFactory.deploy(beaconAddress, initData);
  await beaconProxy.waitForDeployment();
  const proxyAddress = await beaconProxy.getAddress();

  return factory.attach(proxyAddress) as BaseContract;
}

export function silenceWarnings(): void {}

/* eslint-disable @typescript-eslint/no-unused-vars */
export async function forceImport(
  proxyAddress: string,
  contract: BaseContract,
  opts?: { kind?: string },
): Promise<void> {
  console.warn("forceImport is not supported in this compatibility layer");
}

export async function validateUpgrade(proxyAddress: string, factory: ContractFactory): Promise<void> {
  console.warn("validateUpgrade is not supported in this compatibility layer - skipping storage layout validation");
}
/* eslint-enable @typescript-eslint/no-unused-vars */

export const upgrades = {
  deployProxy,
  upgradeProxy,
  deployBeacon,
  deployBeaconProxy,
  silenceWarnings,
  forceImport,
  validateUpgrade,
};

export default upgrades;
