import { MessageSent, OnChainMessageStatus } from "@consensys/linea-sdk";
import { MessageProps } from "../../../entities/Message";
import { IMessageServiceContract } from "../../../services/contracts/IMessageServiceContract";
import { LineaGasFees } from "../IGasProvider";

export interface IL2MessageServiceClient<
  Overrides,
  TransactionReceipt,
  TransactionResponse,
  ContractTransactionResponse,
  Signer,
  ErrorDescription,
> extends IMessageServiceContract<TransactionReceipt, TransactionResponse, ErrorDescription> {
  claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides?: Overrides,
  ): Promise<ContractTransactionResponse>;
  getMessageStatus(messageHash: string, overrides?: Overrides): Promise<OnChainMessageStatus>;
  encodeClaimMessageTransactionData(message: MessageProps & { feeRecipient?: string }): string;
  estimateClaimGasFees(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides?: Overrides,
  ): Promise<LineaGasFees>;
  getSigner(): Signer | undefined;
  getContractAddress(): string;
}
