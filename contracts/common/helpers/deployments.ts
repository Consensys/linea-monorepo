import { ethers, AbstractSigner, Interface, InterfaceAbi } from "ethers";

export function getInitializerData(contractAbi: InterfaceAbi, initializerFunctionName: string, args: unknown[]) {
  const contractInterface = new Interface(contractAbi);
  const fragment = contractInterface.getFunction(initializerFunctionName);

  if (!fragment) {
    return "0x";
  }

  return contractInterface.encodeFunctionData(fragment, args);
}

export async function deployContractFromArtifacts(
  abi: ethers.InterfaceAbi,
  bytecode: ethers.BytesLike,
  wallet: AbstractSigner,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: ethers.ContractMethodArgs<any[]>
) {
  const factory = new ethers.ContractFactory(abi, bytecode, wallet);
  const contract = await factory.deploy(...args);
  return contract.waitForDeployment();
}
