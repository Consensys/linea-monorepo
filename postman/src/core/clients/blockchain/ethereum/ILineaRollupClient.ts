import { MessageProps } from "../../../entities/Message";
import { MessageSent, OnChainMessageStatus } from "@consensys/linea-sdk";
import { IMessageServiceContract } from "../../../services/contracts/IMessageServiceContract";
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
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    opts?: {
      overrides?: Overrides;
      l1LogsFromBlock?: string | number;
      l2LogsFromBlock?: string | number;
    },
  ): Promise<bigint>;
  estimateClaimWithoutProofGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides: Overrides,
  ): Promise<bigint>;
  claimWithoutProof(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides: Overrides,
  ): Promise<ContractTransactionResponse>;
  claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    opts?: {
      overrides?: Overrides;
      l1LogsFromBlock?: string | number;
      l2LogsFromBlock?: string | number;
    },
  ): Promise<ContractTransactionResponse>;
}
