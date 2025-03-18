import { Proof } from "@consensys/linea-sdk/dist/lib/sdk/merkleTree/types";
import { Chain, Token, TransactionStatus } from "@/types";
import { Address } from "viem";

export type NativeBridgeMessage = {
  from: Address;
  to: Address;
  fee: bigint;
  value: bigint;
  nonce: bigint;
  calldata: string;
  messageHash: string;
  proof?: Proof;
  amountSent: bigint;
};

// TODO - Move params to claim component, not transaction history component
// Params expected for `receiveMessage` as per https://developers.circle.com/stablecoins/transfer-usdc-on-testnet-from-ethereum-to-avalanche
export type CCTPV2BridgeMessage = {
  message: string;
  attestation: string;
  amountSent: bigint;
};

// BridgeTransaction object that is populated when user opens "TransactionHistory" component, and is passed to child components.
export interface BridgeTransaction {
  type: "ETH" | "ERC20";
  status: TransactionStatus;
  timestamp: bigint;
  fromChain: Chain;
  toChain: Chain;
  token: Token;
  message: NativeBridgeMessage | CCTPV2BridgeMessage;
  bridgingTx: string;
  claimingTx?: string;
}
