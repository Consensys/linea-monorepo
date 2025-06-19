import { MessageSent } from "@consensys/linea-sdk";

export interface IMessageServiceContract<TransactionReceipt, TransactionResponse, ErrorDescription> {
  getMessageByMessageHash(messageHash: string): Promise<MessageSent | null>;
  getMessagesByTransactionHash(transactionHash: string): Promise<MessageSent[] | null>;
  getTransactionReceiptByMessageHash(messageHash: string): Promise<TransactionReceipt | null>;
  retryTransactionWithHigherFee(transactionHash: string, priceBumpPercent?: number): Promise<TransactionResponse>;
  isRateLimitExceeded(messageFee: bigint, messageValue: bigint): Promise<boolean>;
  isRateLimitExceededError(transactionHash: string): Promise<boolean>;
  parseTransactionError(transactionHash: string): Promise<ErrorDescription | string>;
}
