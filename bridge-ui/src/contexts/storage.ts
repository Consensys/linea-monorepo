import { Address } from "viem";

export enum StorageKeys {
  TRANSACTIONS = "transactions",
  HASHES = "hashes",
  USER_TOKENS = "user_tokens",
  STORAGE_VERSION = "storage_version",
}

export const generateKey = (baseKey: string, address: Address, networkType: string) => {
  return `${address}_${networkType}_${baseKey}`;
};
