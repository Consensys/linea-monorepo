import { parseEther } from "viem";

export function etherToWei(amount: string): bigint {
  return parseEther(amount.toString());
}
