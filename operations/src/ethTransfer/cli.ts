import { ethers } from "ethers";

function sanitizeAddress(argName: string) {
  return (input: string) => {
    if (!ethers.isAddress(input)) {
      throw new Error(`${argName} is not a valid Ethereum address.`);
    }
    return input;
  };
}

function isValidUrl(input: string, allowedProtocols: string[]): boolean {
  try {
    const url = new URL(input);
    return allowedProtocols.includes(url.protocol);
  } catch (e) {
    return false;
  }
}

function sanitizeUrl(argName: string, allowedProtocols: string[]) {
  return (input: string) => {
    if (!isValidUrl(input, allowedProtocols)) {
      throw new Error(`${argName}, with value: ${input} is not a valid URL`);
    }
    return input;
  };
}

function sanitizeHexString(argName: string, expectedLength: number) {
  return (input: string) => {
    if (!ethers.isHexString(input, expectedLength)) {
      throw new Error(`${argName} must be hexadecimal string of length ${expectedLength}.`);
    }
    return input;
  };
}

function sanitizeETHThreshold() {
  return (input: string) => {
    if (parseInt(input) <= 1) {
      throw new Error("Threshold must be higher than 1 ETH");
    }
    return input;
  };
}

export { sanitizeAddress, sanitizeHexString, sanitizeETHThreshold, sanitizeUrl, isValidUrl };
