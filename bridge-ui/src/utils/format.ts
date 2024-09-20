import { formatDate, fromUnixTime } from "date-fns";
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
 * Formats a hexadecimal string by truncating it and adding ellipsis in the middle.
 * @param hexString - The hexadecimal string to format.
 * @param step - The number of characters to keep at the beginning and end of the string.
 * @returns The formatted hexadecimal string.
 */
export const formatHex = (hexString: string | undefined, step = 5) => {
  if (!hexString) return "N/A";
  return hexString.substring(0, step) + "..." + hexString.substring(hexString.length - step, hexString.length);
};

/**
 * Format balance
 * @param balance
 * @param precision
 * @returns
 */
export const formatBalance = (balance: string | 0n | undefined, precision = 4) => {
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

/**
 * Format timestamp
 * @param timestamp
 * @param formatStr
 * @returns
 */
export const formatTimestamp = (timestamp: number, formatStr: string) => {
  return formatDate(fromUnixTime(timestamp), formatStr);
};
