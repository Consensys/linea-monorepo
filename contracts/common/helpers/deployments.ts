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
  contractName: string,
  abi: ethers.InterfaceAbi,
  bytecode: ethers.BytesLike,
  wallet: AbstractSigner,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ...args: ethers.ContractMethodArgs<any[]>
) {
  const factory = new ethers.ContractFactory(abi, bytecode, wallet);
  const contract = await factory.deploy(...args);

  const chainId = (await wallet.provider!.getNetwork()).chainId;
  const txReceipt = await contract.deploymentTransaction()!.wait();

  if (!txReceipt) {
    throw `Contract deployment transaction receipt not found for contract=${contractName}`;
  }

  // This should match the regexes used in the coordinator - e.g.
  // contract=LineaRollup(?:.*)? deployed: address=(0x[0-9a-fA-F]{40}) blockNumber=(\\d+)
  console.log(
    `contract=${contractName} deployed: address=${contract.target} blockNumber=${txReceipt.blockNumber} chainId=${chainId}`,
  );

  return contract;
}
