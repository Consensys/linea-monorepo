import type { HardhatEthersSigner as SignerWithAddress } from "@nomicfoundation/hardhat-ethers/types";
import { ethers } from "ethers";

export const range = (start: number, end: number) => Array.from(Array(end - start + 1).keys()).map((x) => x + start);

export const generateRandomBytes = (length: number): string => ethers.hexlify(ethers.randomBytes(length));

export function buildAccessErrorMessage(account: SignerWithAddress, role: string): string {
  return `AccessControl: account ${account.address.toLowerCase()} is missing role ${role}`;
}
