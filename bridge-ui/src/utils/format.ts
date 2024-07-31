import { Address, getAddress } from "viem";

/**
 * Format Ethereum address
 * @param address
 * @param step
 * @returns
 */
export const formatAddress = (address: string | undefined, step = 5) => {
  if (!address) return "N/A";
  return address.substring(0, step) + "..." + address.substring(address.length - step, address.length);
};

/**
 * Format balance
 * @param balance
 * @param precision
 * @returns
 */
export const formatBalance = (balance: string | undefined, precision = 4) => {
  if (!balance) return "";
  const [whole, fraction = ""] = balance.split(".");
  if (fraction.length > precision) {
    return `${whole}.${fraction.slice(0, precision)}`;
  }
  return balance.toString();
};

/**
 * Safe get address
 * @param address
 * @returns
 */
export const safeGetAddress = (address: Address | null): string | null => {
  return address ? getAddress(address) : null;
};
