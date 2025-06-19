import { Message } from "../../types";
import { IMessageServiceContract } from "../IMessageServiceContract";
import { LineaGasFees } from "../IGasProvider";
import { OnChainMessageStatus } from "../../enums";

export interface IL2MessageServiceClient<
  Overrides,
  TransactionReceipt,
  TransactionResponse,
  ContractTransactionResponse,
  Signer,
  ErrorDescription,
> extends IMessageServiceContract<TransactionReceipt, TransactionResponse, ErrorDescription> {
  claim(message: Message & { feeRecipient?: string }, overrides?: Overrides): Promise<ContractTransactionResponse>;
  getMessageStatus(messageHash: string, overrides?: Overrides): Promise<OnChainMessageStatus>;
  encodeClaimMessageTransactionData(message: Message & { feeRecipient?: string }): string;
  estimateClaimGasFees(message: Message & { feeRecipient?: string }, overrides?: Overrides): Promise<LineaGasFees>;
  getSigner(): Signer | undefined;
  getContractAddress(): string;
}
