import { Address } from "viem";

export interface Transaction {
  txHash: Address | undefined;
  chainId: number | undefined;
  name: string | undefined;
}
