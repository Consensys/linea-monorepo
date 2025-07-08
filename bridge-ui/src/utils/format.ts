import { formatDate, fromUnixTime } from "date-fns";
import { Address, formatUnits, getAddress } from "viem";
import { isUndefinedOrEmptyString } from "@/utils";

/**
 * Format Ethereum address
 * @param address
 * @param step
 * @returns
 */
export const formatAddress = (address: string | undefined, step = 5) => {
  if (isUndefinedOrEmptyString(address)) return "N/A";
  return address.substring(0, step) + "..." + address.substring(address.length - step, address.length);
};

/**
 * Formats a hexadecimal string by truncating it and adding ellipsis in the middle.
 * @param hexString - The hexadecimal string to format.
 * @param step - The number of characters to keep at the beginning and end of the string.
 * @returns The formatted hexadecimal string.
 */
export const formatHex = (hexString: string | undefined, step = 5) => {
  if (isUndefinedOrEmptyString(hexString)) return "N/A";
  return hexString.substring(0, step) + "..." + hexString.substring(hexString.length - step, hexString.length);
};

/**
 * Format balance
 * @param balance
 * @param precision
 * @returns
 */
export const formatBalance = (balance: string | 0n | undefined, precision = 4) => {
  if (balance === 0n || isUndefinedOrEmptyString(balance)) return "";
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

/**
 * Format a bigint to a string with a specified number of decimal places.
 * @param value - The value to format.
 * @param decimals - The number of token decimals to use for formatting (default is 18).
 * @returns A formatted string representation of the value.
 */
export function formatDigit(value: bigint, decimals: number = 18): string {
  if (value <= 0n) return "0.00";

  const valueStr = formatUnits(value, decimals);
  const num = Number(valueStr);
  if (isNaN(num)) return "0.00";
  if (num >= 0.00000001) return num.toFixed(8);

  const match = valueStr.match(/^0\.0+(?=\d)/);
  if (!match) return num.toFixed(8);

  const zeroCount = match[0].length - 2;
  const remainder = valueStr.slice(match[0].length);
  const rounded = Math.round(Number("0." + remainder) * 100)
    .toString()
    .padStart(2, "0");

  return `0.0<sub>${zeroCount}</sub>${rounded}`;
}
