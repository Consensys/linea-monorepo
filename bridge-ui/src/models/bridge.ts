import { BridgeType, TokenInfo } from "@/config/config";
import { Address } from "viem";

export type BridgeForm = {
  amount: string;
  balance: string;
  submit: string;
  claim: string;
  minFees: bigint;
  token: TokenInfo;
  bridgingAllowed: boolean;
  gasFees: bigint;
  mode: BridgeType;
  destinationAddress: Address;
};

export type BridgeError = {
  name: string;
  message: string;
  link: string;
  displayInToast: boolean;
};

export enum BridgeErrors {
  ReservedToken = "ReservedToken",
  RateLimitExceeded = "RateLimitExceeded",
}
