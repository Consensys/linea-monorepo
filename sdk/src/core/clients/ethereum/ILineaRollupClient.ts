import { Message } from "../../types";
import { OnChainMessageStatus } from "../../enums";
import { IMessageServiceContract } from "../IMessageServiceContract";
import { FinalizationMessagingInfo, Proof } from "./IMerkleTreeService";

export interface ILineaRollupClient<
  Overrides,
  TransactionReceipt,
  TransactionResponse,
  ContractTransactionResponse,
  ErrorDescription,
> extends IMessageServiceContract<TransactionReceipt, TransactionResponse, ErrorDescription> {
  getFinalizationMessagingInfo(transactionHash: string): Promise<FinalizationMessagingInfo>;
  getL2MessageHashesInBlockRange(fromBlock: number, toBlock: number): Promise<string[]>;
  getMessageSiblings(messageHash: string, messageHashes: string[], treeDepth: number): string[];
  getMessageProof(
    messageHash: string,
    opts?: {
      l1LogsFromBlock?: string | number;
      l2LogsFromBlock?: string | number;
    },
  ): Promise<Proof>;
  getMessageStatus(
    messageHash: string,
    opts?: { overrides?: Overrides; l1LogsFromBlock?: string | number; l2LogsFromBlock?: string | number },
  ): Promise<OnChainMessageStatus>;
  getMessageStatusUsingMessageHash(messageHash: string, overrides: Overrides): Promise<OnChainMessageStatus>;
  getMessageStatusUsingMerkleTree(
    messageHash: string,
    opts?: { overrides?: Overrides; l1LogsFromBlock?: string | number; l2LogsFromBlock?: string | number },
  ): Promise<OnChainMessageStatus>;
  estimateClaimGas(
    message: Message & { feeRecipient?: string },
    opts?: {
      overrides?: Overrides;
      l1LogsFromBlock?: string | number;
      l2LogsFromBlock?: string | number;
    },
  ): Promise<bigint>;
  claim(
    message: Message & { feeRecipient?: string },
    opts?: {
      overrides?: Overrides;
      l1LogsFromBlock?: string | number;
      l2LogsFromBlock?: string | number;
    },
  ): Promise<ContractTransactionResponse>;
  estimateClaimWithoutProofGas(message: Message & { feeRecipient?: string }, overrides: Overrides): Promise<bigint>;
  claimWithoutProof(
    message: Message & { feeRecipient?: string },
    overrides: Overrides,
  ): Promise<ContractTransactionResponse>;
}
