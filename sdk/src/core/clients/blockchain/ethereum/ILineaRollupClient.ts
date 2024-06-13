import { FinalizationMessagingInfo, Proof } from "../../../../services/merkleTree/types";
import { MessageProps } from "../../../entities/Message";
import { OnChainMessageStatus } from "../../../enums/MessageEnums";
import { IMessageServiceContract } from "../../../services/contracts/IMessageServiceContract";
import { MessageSent } from "../../../types/Events";

export interface ILineaRollupClient<Overrides, ContractTransactionResponse>
  extends IMessageServiceContract<unknown, unknown, unknown, unknown> {
  getFinalizationMessagingInfo(transactionHash: string): Promise<FinalizationMessagingInfo>;
  getL2MessageHashesInBlockRange(fromBlock: number, toBlock: number): Promise<string[]>;
  getMessageSiblings(messageHash: string, messageHashes: string[], treeDepth: number): string[];
  getMessageProof(messageHash: string): Promise<Proof>;
  getMessageStatusUsingMessageHash(messageHash: string, overrides: Overrides): Promise<OnChainMessageStatus>;
  getMessageStatusUsingMerkleTree(messageHash: string, overrides: Overrides): Promise<OnChainMessageStatus>;
  estimateClaimWithoutProofGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides: Overrides,
  ): Promise<bigint>;
  claimWithoutProof(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides: Overrides,
  ): Promise<ContractTransactionResponse>;
}
