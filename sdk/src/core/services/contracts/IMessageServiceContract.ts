import { OnChainMessageStatus } from "../../../core/enums/MessageEnums";
import { MessageProps } from "../../entities/Message";
import { MessageSent } from "../../types/events";

export interface IMessageServiceContract<
  Overrides,
  TransactionReceipt,
  TransactionResponse,
  ContractTransactionResponse,
> {
  getMessageStatus(messageHash: string, overrides?: Overrides): Promise<OnChainMessageStatus>;
  getMessageByMessageHash(messageHash: string): Promise<MessageSent | null>;
  getMessagesByTransactionHash(transactionHash: string): Promise<MessageSent[] | null>;
  getTransactionReceiptByMessageHash(messageHash: string): Promise<TransactionReceipt | null>;
  claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    overrides?: Overrides,
  ): Promise<ContractTransactionResponse>;
  retryTransactionWithHigherFee(transactionHash: string, priceBumpPercent?: number): Promise<TransactionResponse>;
  isRateLimitExceeded(messageFee: bigint, messageValue: bigint): Promise<boolean>;
  isRateLimitExceededError(transactionHash: string): Promise<boolean>;
}
