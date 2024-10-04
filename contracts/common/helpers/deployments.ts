import { ContractFactory, Overrides, Wallet, ethers } from "ethers";
import { HardhatEthersSigner } from "@nomicfoundation/hardhat-ethers/signers";
import { GetContractTypeFromFactory } from "../../typechain-types/common";
import { ProxyAdmin, ProxyAdmin__factory, TransparentUpgradeableProxy__factory } from "../../typechain-types";

export function getInitializerData(
  contractInterface: ethers.Interface,
  initializerFunctionName: string,
  args: unknown[],
) {
  const fragment = contractInterface.getFunction(initializerFunctionName);

  if (!fragment) {
    return "0x";
  }

  return contractInterface.encodeFunctionData(fragment, args);
}

export const deployContract = async <TFactory extends ContractFactory>(
  contractFactory: TFactory,
  deployer: Wallet | HardhatEthersSigner,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  args?: any[],
  overrides: Overrides = {},
): Promise<GetContractTypeFromFactory<TFactory>> => {
  const deploymentArgs = args || [];
  const instance = await contractFactory.connect(deployer).deploy(...deploymentArgs, overrides);
  return instance.waitForDeployment() as GetContractTypeFromFactory<TFactory>;
};

export const deployUpgradableContract = async <TFactory extends ContractFactory>(
  contractFactory: TFactory,
  deployer: Wallet | HardhatEthersSigner,
  admin: ProxyAdmin,
  initializerData = "0x",
  overrides: Overrides = {},
): Promise<GetContractTypeFromFactory<TFactory>> => {
  const instance = await deployContract(contractFactory, deployer, [], overrides);

  const proxy = await deployContract(new TransparentUpgradeableProxy__factory(), deployer, [
    await instance.getAddress(),
    await admin.getAddress(),
    initializerData,
  ]);

  return proxy as GetContractTypeFromFactory<TFactory>;
};

export async function deployUpgradableContractWithProxyAdmin<TFactory extends ContractFactory>(
  contractFactory: TFactory,
  deployer: Wallet | HardhatEthersSigner,
  initializer?: {
    functionName: string;
    args: unknown[];
  },
  overrides: Overrides = {},
): Promise<GetContractTypeFromFactory<TFactory>> {
  const proxyFactory = new ProxyAdmin__factory(deployer);

  const proxyAdmin = await deployContract(proxyFactory, deployer, [], overrides);

  let contract: GetContractTypeFromFactory<TFactory>;
  if (initializer) {
    const initializerData = getInitializerData(contractFactory.interface, initializer.functionName, initializer.args);
    contract = await deployUpgradableContract(contractFactory, deployer, proxyAdmin, initializerData, overrides);
  } else {
    contract = await deployUpgradableContract(contractFactory, deployer, proxyAdmin, "0x", overrides);
  }
  return contract;
}
