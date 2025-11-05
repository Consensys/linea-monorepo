import { ethers } from "ethers";

export function validateEthereumAddress(argName: string, input: string) {
  if (!ethers.isAddress(input)) {
    throw new Error(`${argName} is not a valid Ethereum address.`);
  }
  return input;
}

export function isValidProtocolUrl(input: string, allowedProtocols: string[]): boolean {
  try {
    const url = new URL(input);
    return allowedProtocols.includes(url.protocol);
  } catch (e) {
    return false;
  }
}

export function validateUrl(argName: string, input: string, allowedProtocols: string[]) {
  if (!isValidProtocolUrl(input, allowedProtocols)) {
    throw new Error(`${argName}, with value: ${input} is not a valid URL`);
  }
  return input;
}

export function validateHexString(argName: string, input: string, expectedLength: number) {
  if (!ethers.isHexString(input, expectedLength)) {
    throw new Error(`${argName} must be hexadecimal string of length ${expectedLength}.`);
  }
  return input;
}
