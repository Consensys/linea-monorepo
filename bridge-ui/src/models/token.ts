import { Address } from "viem";

export interface Extension {
  rootChainId: number;
  rootChainURI: string;
  rootAddress: Address;
}

export interface Token {
  chainId: number;
  chainURI: string;
  tokenId: string;
  tokenType: string[];
  address: Address;
  name: string;
  symbol: string;
  decimals: number;
  createdAt: string;
  updatedAt: string;
  logoURI: string;
  extension: Extension;
}
