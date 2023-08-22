import { Contract, ContractFactory, Wallet, ethers, utils } from "ethers";
import { ProxyAdmin__factory, TransparentUpgradeableProxy__factory, ProxyAdmin } from "../typechain";

function getInitializerData(contractInterface: ethers.utils.Interface, args: unknown[]) {
  const initializer = "initialize";
  const fragment = contractInterface.getFunction(initializer);
  return contractInterface.encodeFunctionData(fragment, args);
}

export const encodeLibraryName = (libraryName: string) => {
  const encodedLibraryName = utils.solidityKeccak256(["string"], [libraryName]).slice(2, 36);
  return `__$${encodedLibraryName}$__`;
};

export const deployContract = async <T extends ContractFactory>(
  contractFactory: T,
  deployer: Wallet,
): Promise<Contract> => {
  const instance = await contractFactory.connect(deployer).deploy();
  await instance.deployed();
  return instance;
};

export const deployUpgradableContract = async <T extends ContractFactory>(
  contractFactory: T,
  deployer: Wallet,
  admin: ProxyAdmin,
  initializerData = "0x",
): Promise<Contract> => {
  const instance = await contractFactory.connect(deployer).deploy();
  await instance.deployed();

  const proxy = await new TransparentUpgradeableProxy__factory()
    .connect(deployer)
    .deploy(instance.address, admin.address, initializerData);
  await proxy.deployed();

  return instance.attach(proxy.address);
};

export async function deployUpgradableContractWithProxyAdmin<T extends ContractFactory>(
  contractFactory: T,
  deployer: Wallet,
  args: unknown[],
) {
  const proxyFactory = new ProxyAdmin__factory(deployer);
  const proxyAdmin = await proxyFactory.connect(deployer).deploy();
  await proxyAdmin.deployed();
  console.log(`ProxyAdmin contract deployed at address: ${proxyAdmin.address}`);

  const contract = await deployUpgradableContract(
    contractFactory,
    deployer,
    proxyAdmin,
    getInitializerData(contractFactory.interface, args),
  );
  console.log(`Contract deployed at address: ${contract.address}`);
  return contract;
}
