import { ethers } from "ethers";

export function getWallet(privateKey: string, provider: ethers.providers.JsonRpcProvider) {
  return new ethers.Wallet(privateKey, provider);
}

export function encodeFunctionCall(contractInterface: ethers.utils.Interface, functionName: string, args: unknown[]) {
  return contractInterface.encodeFunctionData(functionName, args);
}
