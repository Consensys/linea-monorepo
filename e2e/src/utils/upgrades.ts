import { ContractReceipt, Wallet, ethers } from "ethers";
import { ProxyAdmin__factory } from "../typechain";

export function getInitializerData(
  contractInterface: ethers.utils.Interface,
  initializerFunctionName: string,
  args: unknown[],
) {
  const fragment = contractInterface.getFunction(initializerFunctionName);
  return contractInterface.encodeFunctionData(fragment, args);
}

export async function upgradeContractAndCall(
  deployer: Wallet,
  proxyAdminContractAddress: string,
  proxyContractAddress: string,
  implementationContractAddress: string,
  initializerData = "0x",
): Promise<ContractReceipt> {
  const proxyAdminFactory = new ProxyAdmin__factory(deployer);
  const proxyAdmin = proxyAdminFactory.connect(deployer).attach(proxyAdminContractAddress);
  const tx = await proxyAdmin.upgradeAndCall(proxyContractAddress, implementationContractAddress, initializerData);

  return tx.wait();
}

export async function upgradeContract(
  deployer: Wallet,
  proxyAdminContractAddress: string,
  proxyContractAddress: string,
  implementationContractAddress: string,
): Promise<ContractReceipt> {
  const proxyAdminFactory = new ProxyAdmin__factory(deployer);
  const proxyAdmin = proxyAdminFactory.connect(deployer).attach(proxyAdminContractAddress);
  const tx = await proxyAdmin.upgrade(proxyContractAddress, implementationContractAddress);

  return tx.wait();
}
